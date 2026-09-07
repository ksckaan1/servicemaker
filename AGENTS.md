# AGENTS.md

## What this is

Go library that orchestrates service lifecycle (init, run, close, health check) via env-based config. Three separate Go modules — not a workspace with replace directives.

## Module structure

| Module | Path | Go version |
|--------|------|------------|
| `github.com/ksckaan1/servicemaker` | root | 1.27 |
| `github.com/ksckaan1/servicemaker/components/fibercomp` | `components/fibercomp/` | 1.27.0 |
| `github.com/ksckaan1/servicemaker/components/rediscomp` | `components/rediscomp/` | 1.27 |

Each module has its own `go.mod` and `go.sum`. Run `go build`/`go vet` etc. from the relevant directory, not from root.

## Core contracts (interface.go)

Components can optionally implement: `Initializer`, `Runner`, `Closer`, `HealthChecker`. `ServiceMaker` checks for these at runtime via type assertions — no registration of interfaces needed.

## Config parsing

All component config is parsed from environment variables using `caarlos0/env/v11`. Struct tags are `env:"VAR_NAME" envDefault:"value"`. Fatal on parse failure — no error returned to caller.

## Pre-built components

`components/fibercomp/` and `components/rediscomp/` are ready-made components. Users can import and register them directly without writing boilerplate.

## Workflow rules

- Do NOT run `git commit` or `git push` unless explicitly asked by the user.

## Current state

- No tests, no CI, no linting config, no Makefile
- `example/` is gitignored
- `redis/` module is a stub (no dependencies, no-op Init/Run)
- No commits on master yet
