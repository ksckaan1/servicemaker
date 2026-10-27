package servicemaker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeComponent struct {
	mu      sync.Mutex
	inits   int
	runs    int
	closes  int
	runErr  error
	closeFn func(context.Context) error
}

func (f *fakeComponent) Init(context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.inits++
	return nil
}

func (f *fakeComponent) Run(ctx context.Context) error {
	f.mu.Lock()
	f.runs++
	f.mu.Unlock()

	if f.runErr != nil {
		return f.runErr
	}

	<-ctx.Done()
	return nil
}

func (f *fakeComponent) Close(ctx context.Context) error {
	f.mu.Lock()
	f.closes++
	closeFn := f.closeFn
	f.mu.Unlock()

	if closeFn != nil {
		return closeFn(ctx)
	}
	return nil
}

func (f *fakeComponent) counts() (inits, runs, closes int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.inits, f.runs, f.closes
}

func newTestServiceMaker(t *testing.T) (*ServiceMaker, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	return New(ctx), cancel
}

// Verifies that every registered closer runs exactly once on shutdown.
func TestRunClosesEveryComponentOnce(t *testing.T) {
	sm, cancel := newTestServiceMaker(t)

	a := &fakeComponent{}
	b := &fakeComponent{}
	sm.RegisterGracefulShutdown(a.Close)
	sm.RegisterGracefulShutdown(b.Close)

	done := make(chan error, 1)
	go func() { done <- sm.Run() }()

	// Give Run a moment to start before asking for shutdown.
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v, want nil on graceful shutdown", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after shutdown")
	}

	for name, c := range map[string]*fakeComponent{"a": a, "b": b} {
		if _, _, closes := c.counts(); closes != 1 {
			t.Fatalf("%s closed %d times, want 1", name, closes)
		}
	}
}

// Verifies that a close sequence that hangs still ends within the timeout
// instead of keeping the process alive forever.
func TestCloseAllIsBounded(t *testing.T) {
	sm, _ := newTestServiceMaker(t)

	orig := closeTimeout
	closeTimeout = 50 * time.Millisecond
	t.Cleanup(func() { closeTimeout = orig })

	sm.RegisterGracefulShutdown(func(ctx context.Context) error {
		<-ctx.Done() // unblocks only because closeAll bounds the context
		return ctx.Err()
	})

	start := time.Now()
	sm.closeAll()

	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("closeAll took %v, want about %v", elapsed, closeTimeout)
	}
}

// Verifies that a Runner returning nil does not surface a fake error.
func TestRunIgnoresNilRunnerError(t *testing.T) {
	sm, cancel := newTestServiceMaker(t)

	comp := &fakeComponent{runErr: nil}
	sm.comps.Store("nil-runner", comp)

	done := make(chan error, 1)
	go func() { done <- sm.Run() }()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after shutdown")
	}
}

// Verifies that the real error of a Runner is reported.
func TestRunReportsRunnerError(t *testing.T) {
	sm, _ := newTestServiceMaker(t)

	want := errors.New("boom")
	sm.comps.Store("broken-runner", &fakeComponent{runErr: want})

	err := sm.Run()
	if !errors.Is(err, want) {
		t.Fatalf("Run returned %v, want it to wrap %v", err, want)
	}
}
