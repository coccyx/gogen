# AGENTS.md

This file gives coding agents the repo-specific context needed to work effectively in `gogen`.

## Project Overview

Gogen is a data generator for demo and test data, especially time-series logs and metrics. The repo contains:

- a Go CLI core
- a Python AWS Lambda API backend in `gogen-api/`
- a React/TypeScript UI in `ui/`

## Common Commands

### Go

```bash
make install          # Preferred install path; injects ldflags from Makefile
make build            # Cross-compiles linux, darwin, windows, wasm
make test             # go test -v ./...
go test -v ./internal
go test -v -run TestName ./internal
```

Notes:

- Use `make install` instead of bare `go install`; version/build metadata and OAuth settings are injected through `-ldflags`.
- Dependencies are vendored. After dependency changes, run `go mod vendor`.

### Python API

```bash
cd gogen-api
./start_dev.sh
./setup_local_db.sh
./deploy_lambdas.sh
```

Repo-standard Python environment:

```bash
source /home/clint/local/src/gogen/.pyvenv/bin/activate
```

Focused API unit tests:

```bash
make api-test
```

### UI

```bash
cd ui
npm run dev
npm run build
npm test
```

## Architecture

### Go Package Layout

- `main.go`: CLI entry point using `urfave/cli.v1`; maps flags to `GOGEN_*` env vars
- `internal/`: core config, sample, token, API/share logic
- `generator/`: generation workers
- `outputter/`: output workers and destinations
- `run/`: pipeline orchestration
- `timer/`: one timer goroutine per sample
- `rater/`: event-rate control
- `template/`: output formatting
- `logger/`: log wrapper

### Data Flow

```text
YAML/JSON config -> internal.Config singleton
  -> timer goroutines
  -> generator worker pool
  -> outputter worker pool
  -> output destination
```

### Config System

- Config is a singleton guarded by `sync.Once`
- Remote configs default to `https://api.gogen.io` and can be overridden by `GOGEN_APIURL`
- In Go tests, reset config state with `config.ResetConfig()` before `config.NewConfig()`
- Tests often use `config.SetupFromString(...)` for inline YAML

### Python API

- Lambda handlers live as separate files in `gogen-api/api/`
- Backed by DynamoDB + S3
- Local development uses Docker Compose plus SAM
- Use `.pyvenv` rather than system Python when running repo Python commands

### UI

- Vite + React 18 + TypeScript + Tailwind
- Components live in `ui/src/components/`
- Pages live in `ui/src/pages/`
- Tests are colocated as `.test.tsx`

## CI/CD

- `.github/workflows/ci.yml` runs Go tests on pushes to `master`/`dev` and on PRs
- CI also runs `make api-test`
- Branch builds/deploys happen on `master` and `dev`
- Release workflow is handled separately in `.github/workflows/release.yml`

## Practical Notes

- Prefer minimal, targeted edits; this repo spans Go, Python, and frontend code in one tree
- For Python work, prefer adding tests that avoid external AWS dependencies unless the task explicitly needs integration coverage
- For UI tests, keep them aligned with the current design system rather than hardcoding old color classes
