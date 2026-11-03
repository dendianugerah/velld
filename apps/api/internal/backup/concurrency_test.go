package backup

import (
	"sync"
	"testing"
	"time"
)

func newConcurrencyService(max int) *BackupService {
	return &BackupService{
		sem:     make(chan struct{}, max),
		running: make(map[string]struct{}),
	}
}

func TestAcquireRunRejectsDuplicateSchedule(t *testing.T) {
	s := newConcurrencyService(2)

	if !s.acquireRun("schedule-1") {
		t.Fatal("first acquire should succeed")
	}
	if s.acquireRun("schedule-1") {
		t.Fatal("second acquire of the same schedule should be rejected")
	}

	s.releaseRun("schedule-1")

	if !s.acquireRun("schedule-1") {
		t.Fatal("acquire should succeed after release")
	}
	s.releaseRun("schedule-1")
}

func TestAcquireRunBoundsConcurrency(t *testing.T) {
	s := newConcurrencyService(1)

	if !s.acquireRun("a") {
		t.Fatal("first acquire should succeed")
	}

	acquired := make(chan struct{})
	go func() {
		s.acquireRun("b")
		close(acquired)
	}()

	select {
	case <-acquired:
		t.Fatal("second acquire should block while the slot is taken")
	case <-time.After(100 * time.Millisecond):
	}

	s.releaseRun("a")

	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("second acquire should proceed after release")
	}
}

func TestDifferentSchedulesRunConcurrently(t *testing.T) {
	s := newConcurrencyService(2)

	if !s.acquireRun("a") || !s.acquireRun("b") {
		t.Fatal("two different schedules should acquire both slots")
	}
	s.releaseRun("a")
	s.releaseRun("b")
}

func TestAcquireRunIsRaceFree(t *testing.T) {
	s := newConcurrencyService(8)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()
			if s.acquireRun("schedule-x") {
				s.releaseRun("schedule-x")
			}
		}(i)
	}
	wg.Wait()
}
