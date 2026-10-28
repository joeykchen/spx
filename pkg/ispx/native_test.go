//go:build !js

package ispx

import (
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/goplus/ixgo"
	"github.com/goplus/ixgo/xgobuild"
	"github.com/goplus/spx/v4/internal/engine"
	"github.com/goplus/spx/v4/internal/enginewrap"
	gdx "github.com/goplus/spx/v4/pkg/spx/pkg/engine"
)

type interpreterExitBackend struct {
	gdx.IExtMgr
	exit func(int64)
}

type interpreterPlatform struct{ gdx.IPlatformMgr }

func TestInterpreterShutdownWaitsForRun(t *testing.T) {
	isolateInterpreter(t)
	entered, release := make(chan struct{}), make(chan struct{})
	releaseRun := sync.OnceFunc(func() { close(release) })
	defer releaseRun()
	requested := make(chan int64, 1)
	installInterpreterExitBackend(t, func(code int64) { requested <- code })
	ctx := ixgo.NewContext(ixgo.SupportMultipleInterp | xgobuild.StaticLoad)
	ctx.RegisterExternal("main.wait", func() { close(entered); <-release })
	if err := Init(ctx); err != nil {
		t.Fatal(err)
	}
	if err := Build(map[string][]byte{"main.xgo": []byte("func wait() {}\nfunc main() { wait() }\n")}); err != nil {
		t.Fatal(err)
	}
	finished := make(chan interpreterRunResult, 1)
	go runInterpreter(finished)
	t.Cleanup(func() {
		releaseRun()
		select {
		case <-finished:
		case <-time.After(3 * time.Second):
			t.Error("interpreter did not finish during cleanup")
		}
	})
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not enter the program")
	}
	requireInterpreterPanic(t, "ispx: already running", func() { _, _ = Run() })
	stopped := make(chan error, 1)
	go func() {
		defer close(stopped)
		stopped <- Shutdown()
	}()
	t.Cleanup(func() {
		releaseRun()
		select {
		case <-stopped:
		case <-time.After(3 * time.Second):
			t.Error("Shutdown did not finish during cleanup")
		}
	})
	select {
	case code := <-requested:
		if code != 0 {
			t.Errorf("exit request = %d, want 0", code)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Shutdown did not request engine exit")
	}
	select {
	case err := <-stopped:
		t.Fatalf("Shutdown returned before Run completed: %v", err)
	default:
	}
	releaseRun()
	select {
	case result := <-finished:
		if result != (interpreterRunResult{}) {
			t.Fatalf("Run() = %+v", result)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not finish after release")
	}
	select {
	case err := <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Shutdown did not observe completion")
	}
	if interpreter.runningLocked() {
		t.Fatal("completed interpreter still reports a running instance")
	}
}

func TestInterpreterRunWaitsForRebuild(t *testing.T) {
	for _, result := range []string{"successful rebuild", "failed rebuild"} {
		t.Run(result, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				isolateInterpreter(t)
				entered, release := make(chan struct{}), make(chan struct{})
				releaseRun := sync.OnceFunc(func() { close(release) })
				defer releaseRun()
				requested := make(chan struct{}, 1)
				installInterpreterExitBackend(t, func(int64) { requested <- struct{}{} })
				ctx := ixgo.NewContext(ixgo.SupportMultipleInterp | xgobuild.StaticLoad)
				ctx.RegisterExternal("main.wait", func() { close(entered); <-release })
				if err := Init(ctx); err != nil {
					t.Fatal(err)
				}
				if err := Build(map[string][]byte{"main.xgo": []byte("func wait() {}\nfunc main() { wait() }\n")}); err != nil {
					t.Fatal(err)
				}
				oldRun := make(chan interpreterRunResult, 1)
				go runInterpreter(oldRun)
				<-entered
				built := make(chan error, 1)
				go func() {
					source := "var result int\nfunc main() { result = 42 }\nfunc Result() int { return result }\n"
					if result == "failed rebuild" {
						source = "func main( {"
					}
					built <- Build(map[string][]byte{"main.xgo": []byte(source)})
				}()
				<-requested
				synctest.Wait()
				// The rebuild must retain admission while it waits for the old run.
				if interpreter.mu.TryLock() {
					interpreter.mu.Unlock()
					t.Error("Build released the lifecycle lock while waiting for Run")
				}
				newRun := make(chan interpreterRunResult, 1)
				started := make(chan struct{})
				go func() {
					close(started)
					runInterpreter(newRun)
				}()
				<-started
				releaseRun()
				if got := <-oldRun; got != (interpreterRunResult{}) {
					t.Fatalf("old Run() = %+v", got)
				}
				err := <-built
				if result == "failed rebuild" {
					if err == nil || !strings.Contains(err.Error(), "failed to build XGo source") {
						t.Fatalf("rebuild error = %v", err)
					}
					if got := <-newRun; got != (interpreterRunResult{panic: "ispx: not built"}) || interpreter.interp != nil {
						t.Fatalf("failed rebuild admitted stale interpreter: %+v", got)
					}
					if err := Shutdown(); err != nil {
						t.Fatalf("shutdown after failed rebuild = %v", err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				if got := <-newRun; got != (interpreterRunResult{}) {
					t.Fatalf("Run during rebuild = %+v", got)
				}
				if value, err := interpreter.interp.RunFunc("Result"); value != 42 || err != nil {
					t.Fatalf("rebuilt program result = %v, %v", value, err)
				}
			})
		})
	}
}

func (b interpreterExitBackend) RequestExit(code int64) { b.exit(code) }

func (interpreterPlatform) IsMainThread() bool { return true }

func installInterpreterExitBackend(t *testing.T, exit func(int64)) {
	t.Helper()
	previousExt, previousPlatform := gdx.ExtMgr, gdx.PlatformMgr
	enginewrap.Init(func(call func()) { call() })
	gdx.PlatformMgr = interpreterPlatform{}
	gdx.ExtMgr = interpreterExitBackend{exit: exit}
	t.Cleanup(func() {
		gdx.ExtMgr, gdx.PlatformMgr = previousExt, previousPlatform
		enginewrap.Init(engine.WaitMainThread)
	})
}

type interpreterRunResult struct {
	code  int
	err   error
	panic any
}

func runInterpreter(finished chan<- interpreterRunResult) {
	var result interpreterRunResult
	defer close(finished)
	defer func() {
		result.panic = recover()
		finished <- result
	}()
	result.code, result.err = Run()
}
