# Servicemaker Code Simplification Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix logic errors and simplify the core servicemaker package

**Architecture:** Simplify shutdown flow, remove redundant guards, fix context propagation, clean up run.go

**Tech Stack:** Go 1.27, errgroup, signal handling

**Spec:** None — this is refactoring based on code review

## Global Constraints

- Do NOT `git commit` or `git push` unless user explicitly asks
- All documentation in English
- Follow existing code conventions

---

## Issues Found

### Logic Errors

1. **`closeAll()` uses cancelled context** (`close.go:15`): `closer(s.gCtx)` is called when `s.gCtx` is already cancelled by the signal handler. Components receiving a cancelled context can't log or do useful work during cleanup.

2. **Double-call to `closeAll()`** (`run.go` select + `graceful_shutdown.go` signal handler): Both the select and the signal handler call `closeAll()`. The `closed` flag prevents double execution, but the `closerWg.Wait()` in the signal handler can race with the select's `closerWg.Wait()`.

3. **Signal handler blocks unnecessarily** (`graceful_shutdown.go:16`): `s.closerWg.Wait()` in the signal handler goroutine blocks until all closers finish, but the select in `run.go` also calls `closerWg.Wait()`. This is redundant and can race.

4. **Unnecessary goroutine for `eg.Wait()`** (`run.go:36-38`): `errCh` with a goroutine wrapping `eg.Wait()` is unnecessary — `errCh` is buffered with capacity 1, so `eg.Wait()` can be called directly in a goroutine or the select can be simplified.

### Simplification Opportunities

5. **`close.go` redundant double-close guard**: The `closed` flag + `closerMu` protects against double-close, but `closeAll()` should only be called from one place (the select in `run.go`). Remove the guard.

6. **`run.go` overly complex**: Nested selects, goroutine-wrapped `eg.Wait()`, duplicate code in both select cases. Can be simplified to a linear flow.

7. **`Register[T]()` returns `T` but stores `*T`** (`register.go:45,52`): Inconsistent with `Get[T]()` which returns `*T`. Should return `*T` for consistency.

8. **`servicemaker.go` unnecessary field init**: `sync.Map{}` and `sync.WaitGroup{}` zero values are fine, no need to explicitly initialize.

---

## File Changes

### Files to Modify
- `close.go` — Remove double-close guard, use `context.Background()` for closers
- `run.go` — Simplify to linear flow, remove nested selects
- `graceful_shutdown.go` — Remove `closeAll()` and `closerWg.Wait()`, just close channel
- `register.go` — Return `*T` instead of `T`
- `servicemaker.go` — Remove unnecessary field init

### Files Unchanged
- `interface.go` — No changes needed
- `config.go` — No changes needed
- `get.go` — No changes needed
- `error.go` — File doesn't exist (ErrGracefulShutdown referenced but not defined — need to create or inline)

---

## Task 1: Fix `close.go`

**Files:**
- Modify: `close.go`

- [ ] **Step 1: Simplify closeAll()**

Remove the `closed` flag and mutex guard. Use `context.Background()` for closer calls so they get a valid context.

```go
package servicemaker

import "github.com/ksckaan1/logger"

func (s *ServiceMaker) closeAll() {
	ctx := context.Background()
	for _, closer := range s.closers {
		err := closer(ctx)
		if err != nil {
			logger.Default.Error(ctx, "error when closing component", "error", err)
		}
		s.closerWg.Done()
	}
}
```

- [ ] **Step 2: Run build check**

Run: `go vet ./...`
Expected: PASS

---

## Task 2: Simplify `graceful_shutdown.go`

**Files:**
- Modify: `graceful_shutdown.go`

- [ ] **Step 1: Remove closeAll/wait from signal handler**

The signal handler should only close `shutdownCh`. All cleanup happens in the select in `run.go`.

```go
package servicemaker

import (
	"context"
	"os"
	"os/signal"
)

func (s *ServiceMaker) gracefulShutdown(ctx context.Context) context.Context {
	gCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	go func() {
		<-gCtx.Done()
		defer cancel()
		close(s.shutdownCh)
	}()
	return gCtx
}
```

- [ ] **Step 2: Run build check**

Run: `go vet ./...`
Expected: PASS

---

## Task 3: Simplify `run.go`

**Files:**
- Modify: `run.go`

- [ ] **Step 1: Rewrite Run() with linear flow**

Remove the goroutine-wrapped `eg.Wait()`. Simplify to: run components → wait → closeAll → check if graceful.

```go
package servicemaker

import (
	"context"

	"github.com/ksckaan1/logger"
	"golang.org/x/sync/errgroup"
)

func (s *ServiceMaker) Run() error {
	eg, ctx := errgroup.WithContext(s.gCtx)

	s.comps.Range(func(_, comp any) bool {
		compRunner, ok := comp.(Runner)
		if !ok {
			return true
		}

		eg.Go(func() error {
			done := make(chan error, 1)
			go func() {
				done <- compRunner.Run(ctx)
			}()
			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		})

		return true
	})

	err := eg.Wait()

	s.closeAll()
	s.closerWg.Wait()

	select {
	case <-s.shutdownCh:
		logger.Default.Info(context.Background(), "graceful shutdown completed")
		return nil
	default:
		return err
	}
}
```

- [ ] **Step 2: Run build check**

Run: `go vet ./...`
Expected: PASS

---

## Task 4: Fix `servicemaker.go`

**Files:**
- Modify: `servicemaker.go`

- [ ] **Step 1: Remove unnecessary field init, remove closed/closerMu**

```go
package servicemaker

import (
	"context"
	"sync"
)

type ServiceMaker struct {
	comps      sync.Map
	gCtx       context.Context
	closers    []func(context.Context) error
	closerWg   sync.WaitGroup
	shutdownCh chan struct{}
}

func New(ctx context.Context) *ServiceMaker {
	sm := &ServiceMaker{
		shutdownCh: make(chan struct{}),
	}

	sm.gCtx = sm.gracefulShutdown(ctx)

	return sm
}
```

- [ ] **Step 2: Run build check**

Run: `go vet ./...`
Expected: PASS

---

## Task 5: Fix `register.go` return type

**Files:**
- Modify: `register.go`

- [ ] **Step 1: Change Register[T] to return *T**

Change `return c` to `return &c` at the end, and update the return type to `*T`.

```go
func (s *ServiceMaker) Register[T any]() *T {
	var c T

	err := ParseConfig(&c)
	if err != nil {
		logger.Default.Fatal(
			s.gCtx, "error when parsing config",
			"component", fmt.Sprintf("%T", c),
			"error", err,
		)
		return &c
	}

	compInitializer, ok := any(&c).(Initializer)
	if ok {
		err := compInitializer.Init(s.gCtx)
		if err != nil {
			logger.Default.Fatal(
				s.gCtx, "error when initializing component",
				"error", fmt.Errorf("error when initializing component (%T): %w", compInitializer, err),
			)
			return &c
		}
	}

	name := fmt.Sprintf("%T", &c)

	_, ok = s.comps.Load(name)
	if ok {
		logger.Default.Fatal(
			s.gCtx, "error when registering component",
			"component", name,
			"error", fmt.Errorf("component %s already registered", name),
		)
	}

	s.comps.Store(name, &c)

	if closer, ok := any(&c).(Closer); ok {
		s.closers = append(s.closers, closer.Close)
		s.closerWg.Add(1)
	}

	return &c
}
```

- [ ] **Step 2: Run build check**

Run: `go vet ./...`
Expected: PASS

---

## Task 6: Create `error.go` (missing file)

**Files:**
- Create: `error.go`

The `ErrGracefulShutdown` variable is used in `run.go` but `error.go` doesn't exist. Create it.

- [ ] **Step 1: Create error.go**

```go
package servicemaker

import "errors"

var ErrGracefulShutdown = errors.New("graceful shutdown")
```

- [ ] **Step 2: Run build check**

Run: `go vet ./...`
Expected: PASS

---

## Task 7: Update example if needed

**Files:**
- Check: `example/` directory

- [ ] **Step 1: Check if example uses Register[T]() return value**

If example uses `var comp = sm.Register[T]()` and assigns to a variable, it may need updating due to `*T` return type change.

- [ ] **Step 2: Update example if needed**

- [ ] **Step 3: Run build check**

Run: `go vet ./...` from example directory
Expected: PASS

---

## Verification

- [ ] `go vet ./...` passes from root
- [ ] `go vet ./...` passes from each component directory
- [ ] `make run` works in example directory
- [ ] CTRL+C shows "graceful shutdown completed" log
- [ ] Component error (e.g., grpccomp debug error) shows the actual error, not "context canceled"
