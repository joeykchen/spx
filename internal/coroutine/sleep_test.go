package coroutine

import (
	"slices"
	"testing"
	"time"

	itime "github.com/goplus/spx/v3/internal/time"
)

func TestParkedWaitKeepsOrderWithNewFrameWork(t *testing.T) {
	co := New(nil)
	itime.Start(nil)
	t.Cleanup(func() { co.StopAllAndWait(time.Second) })

	var order []string
	sleeper := co.Create("sleeper", func(Thread) {
		co.Wait(1)
		order = append(order, "time")
	})
	co.JoinYieldedOrDone(sleeper)
	co.Update()
	if got := co.parkedJobs.Count(); got != 1 {
		t.Fatalf("parked jobs = %d, want 1", got)
	}
	co.enqueuePriorityJob(&WaitJob{Type: waitTypeMainThread, Call: func() {
		order = append(order, "main")
	}})
	co.Update()
	if want := []string{"main"}; !slices.Equal(order, want) {
		t.Fatalf("priority order = %v, want %v", order, want)
	}

	frame := co.Create("frame", func(Thread) {
		co.WaitNextFrame()
		order = append(order, "frame")
	})
	co.JoinYieldedOrDone(frame)
	co.Update()
	if len(order) != 1 {
		t.Fatalf("jobs resumed in their issuing frame: %v", order)
	}

	itime.Update(1, 30)
	co.Update()
	if want := []string{"main", "time", "frame"}; !slices.Equal(order, want) {
		t.Fatalf("resume order = %v, want %v", order, want)
	}
}

func TestCancelUnparksSleepingWaits(t *testing.T) {
	co := New(nil)
	itime.Start(nil)
	t.Cleanup(func() { co.StopAllAndWait(time.Second) })

	threads := make([]Thread, 2)
	for i := range threads {
		threads[i] = co.Create("sleeper", func(Thread) { co.Wait(1000) })
		co.JoinYieldedOrDone(threads[i])
	}
	co.Update()
	if got := co.GetLastUpdateStats().NextCount; got != 2 {
		t.Fatalf("deferred jobs = %d, want 2", got)
	}
	co.Update()
	if got := co.GetLastUpdateStats().NextCount; got != 2 {
		t.Fatalf("idle deferred jobs = %d, want 2", got)
	}

	threads[0].Cancel()
	waitForThreadSignal(t, threads[0].done, "canceled sleeper did not stop")
	co.Update()
	if got := co.parkedJobs.Count(); got != 1 {
		t.Fatalf("parked jobs after cancel = %d, want 1", got)
	}
	if threads[1].Stopped() {
		t.Fatal("canceling one sleeper stopped its peer")
	}
}

func TestParkUsesCurrentDeferredDeadline(t *testing.T) {
	co := New(nil)
	itime.Start(nil)
	stats := &UpdateJobsStats{}
	co.processWaitJob(&updateState{frame: 0}, stats, &WaitJob{Type: waitTypeTime, Time: 1000})
	co.deferredJobs.PopFront()

	ran := false
	co.deferredJobs.PushBack(&WaitJob{Type: waitTypeTime, Time: 1, Call: func() { ran = true }})
	co.promoteDeferredJobs(stats)
	itime.Update(1, 30)
	co.Update()
	if !ran {
		t.Fatal("parked wait missed its current deadline")
	}
}

func TestFinishedThreadReleasesManager(t *testing.T) {
	check := func(t *testing.T, th Thread) {
		t.Helper()
		waitForThreadSignal(t, th.done, "thread did not finish")
		if th.owner.Load() != nil {
			t.Fatal("finished thread retained its manager")
		}
	}

	t.Run("normal", func(t *testing.T) {
		co := New(nil)
		check(t, co.Create("done", func(Thread) {}))
	})
	t.Run("canceled", func(t *testing.T) {
		co := New(nil)
		itime.Start(nil)
		th := co.Create("waiting", func(Thread) { co.Wait(1000) })
		co.JoinYieldedOrDone(th)
		co.Update()
		th.Cancel()
		check(t, th)
	})
	t.Run("rejected", func(t *testing.T) {
		co := New(nil)
		var th Thread
		if !co.RunAfterStopAll(time.Second, func() {
			th = co.Create("rejected", func(Thread) {})
		}) {
			t.Fatal("stop barrier did not complete")
		}
		check(t, th)
	})
}

func TestEarlyResumeDrainsParkedWait(t *testing.T) {
	co := New(nil)
	itime.Start(nil)
	th := co.Create("waiting", func(Thread) { co.Wait(1000) })
	co.JoinYieldedOrDone(th)
	co.Update()
	if co.parkedJobs.Count() != 1 {
		t.Fatal("time wait did not park")
	}

	co.Resume(th)
	waitForThreadSignal(t, th.done, "resumed thread did not finish")
	co.Update()
	if co.parkedJobs.Count() != 0 {
		t.Fatal("completed thread remained in parked waits")
	}
}
