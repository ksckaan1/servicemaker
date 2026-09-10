# AGENTS.md

## What this is

Go library that orchestrates service lifecycle (init, run, close, health check) via env-based config. Multiple separate Go modules — not a workspace with replace directives.

## Module structure

| Module | Path | Go version |
|--------|------|------------|
| `github.com/ksckaan1/servicemaker` | root | 1.27 |
| `github.com/ksckaan1/servicemaker/components/fibercomp` | `components/fibercomp/` | 1.27.0 |
| `github.com/ksckaan1/servicemaker/components/rediscomp` | `components/rediscomp/` | 1.27 |
| `github.com/ksckaan1/servicemaker/components/grpccomp` | `components/grpccomp/` | 1.27.0 |
| `github.com/ksckaan1/servicemaker/components/rabbitmqcomp` | `components/rabbitmqcomp/` | 1.27.0 |
| `github.com/ksckaan1/servicemaker/components/graphqlcomp` | `components/graphqlcomp/` | 1.27.0 |
| `github.com/ksckaan1/servicemaker/components/meilisearchcomp` | `components/meilisearchcomp/` | 1.27.0 |
| `github.com/ksckaan1/servicemaker/components/pgcomp` | `components/pgcomp/` | 1.27.0 |

Each module with its own `go.mod` has its own `go.sum`. Run `go build`/`go vet` etc. from the relevant directory, not from root.

## Core contracts (interface.go)

Components can optionally implement: `Initializer`, `Runner`, `Closer`, `HealthChecker`. `ServiceMaker` checks for these at runtime via type assertions — no registration of interfaces needed.

## Config parsing

All component config is parsed from environment variables using `caarlos0/env/v11`. Struct tags are `env:"VAR_NAME" envDefault:"value"`. Fatal on parse failure — no error returned to caller.

## API

- `Register[T any]() *T` — parses env config, calls `Init(ctx)` if implemented, registers component. Returns `*T`.
- `Get[T any]() *T` — retrieves a registered component by type.
- `Run()` — starts all `Runner` implementations concurrently. Returns `nil` on graceful shutdown (SIGINT/SIGTERM), or the component error otherwise.

## Shutdown flow

1. Signal handler listens for `SIGINT`/`SIGTERM`
2. On signal: closes `shutdownCh`, cancels context
3. `Run()` detects shutdown, calls `closeAll()` with `context.Background()`
4. All `Closer` implementations are called in registration order
5. `Run()` returns `nil` (graceful) or component error

## Pre-built components

`components/fibercomp/`, `components/rediscomp/`, `components/grpccomp/`, `components/rabbitmqcomp/`, `components/graphqlcomp/`, `components/meilisearchcomp/`, `components/pgcomp/` are separate modules. Users can import and register them directly without writing boilerplate.

## Workflow rules

- Do NOT run `git commit` or `git push` unless explicitly asked by the user.

## Current state

- No tests, no CI, no linting config, no Makefile
- `example/` is gitignored
