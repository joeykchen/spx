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
	"github.com/goplus/spx/v4/internal/engine/wrap"
	gdx "github.com/goplus/spx/v4/pkg/spx/pkg/engine"
)

type interpreterExitBackend struct {
	gdx.IExtMgr
	exit func(int64)
}

type interpreterPlatform struct{ gdx.IPlatformMgr }

func TestInterpreterShutdownWaitsForRun(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		isolateInterpreter(t)
		requested := make(chan int64, 1)
		installInterpreterExitBackend(t, func(code int64) { requested <- code })
		releaseRun, finished := startWaitingInterpreter(t)
		requireInterpreterPanic(t, "ispx: already running", func() { _, _ = Run() })

		stopped := make(chan error, 1)
		go func() { stopped <- Shutdown() }()
		if code := <-requested; code != 0 {
			t.Errorf("exit request = %d, want 0", code)
		}
		synctest.Wait()
		select {
		case err := <-stopped:
			t.Fatalf("Shutdown returned before Run completed: %v", err)
		default:
		}

		releaseRun()
		if result := <-finished; result != (interpreterRunResult{}) {
			t.Fatalf("Run() = %+v", result)
		}
		if err := <-stopped; err != nil {
			t.Fatal(err)
		}
	})
}

func TestInterpreterRunWaitsForRebuild(t *testing.T) {
	for _, result := range []string{"successful rebuild", "failed rebuild"} {
		t.Run(result, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				isolateInterpreter(t)
				requested := make(chan struct{}, 1)
				installInterpreterExitBackend(t, func(int64) { requested <- struct{}{} })
				releaseRun, oldRun := startWaitingInterpreter(t)
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

func TestInterpreterExitCallbackCanCheckInitialization(t *testing.T) {
	for _, stop := range []struct {
		name string
		call func() error
	}{
		{"shutdown", Shutdown},
		{"rebuild", func() error {
			return Build(map[string][]byte{"main.xgo": []byte("func main() {}")})
		}},
	} {
		t.Run(stop.name, func(t *testing.T) {
			isolateInterpreter(t)
			releaseRun, finished := startWaitingInterpreter(t)
			checked := make(chan struct{})
			installInterpreterExitBackend(t, func(int64) {
				// Bound regressions without leaving the callback or interpreter stuck.
				go func() {
					defer close(checked)
					if err := ConfigureFilesystemRoots("", ""); err == nil || !strings.Contains(err.Error(), "before Init") {
						t.Errorf("configuring roots during exit = %v", err)
					}
					requireInterpreterPanic(t, "ispx: already initialized", func() { _ = Init(nil) })
				}()
				select {
				case <-checked:
				case <-time.After(3 * time.Second):
					t.Error("exit callback blocked checking an already initialized interpreter")
				}
				releaseRun()
			})
			if err := stop.call(); err != nil {
				t.Error(err)
			}
			<-checked
			if got := <-finished; got != (interpreterRunResult{}) {
				t.Errorf("stopped Run() = %+v", got)
			}
		})
	}
}

// startWaitingInterpreter runs a program that can only finish after release.
func startWaitingInterpreter(t *testing.T) (release func(), finished <-chan interpreterRunResult) {
	t.Helper()
	entered, resume := make(chan struct{}), make(chan struct{})
	ctx := ixgo.NewContext(ixgo.SupportMultipleInterp | xgobuild.StaticLoad)
	ctx.RegisterExternal("main.wait", func() { close(entered); <-resume })
	if err := Init(ctx); err != nil {
		t.Fatal(err)
	}
	if err := Build(map[string][]byte{"main.xgo": []byte("func wait() {}\nfunc main() { wait() }\n")}); err != nil {
		t.Fatal(err)
	}
	result := make(chan interpreterRunResult, 1)
	go runInterpreter(result)
	<-entered
	release = sync.OnceFunc(func() { close(resume) })
	t.Cleanup(func() {
		release()
		<-result
	})
	return release, result
}

func (b interpreterExitBackend) RequestExit(code int64) { b.exit(code) }

func (interpreterPlatform) IsMainThread() bool { return true }

func installInterpreterExitBackend(t *testing.T, exit func(int64)) {
	t.Helper()
	previousExt, previousPlatform := gdx.ExtMgr, gdx.PlatformMgr
	wrap.Init(func(call func()) { call() })
	gdx.PlatformMgr = interpreterPlatform{}
	gdx.ExtMgr = interpreterExitBackend{exit: exit}
	t.Cleanup(func() {
		gdx.ExtMgr, gdx.PlatformMgr = previousExt, previousPlatform
		wrap.Init(engine.WaitMainThread)
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
