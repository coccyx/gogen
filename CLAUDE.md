# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Gogen is an open source data generator for generating demo and test data, especially time series log and metric data. It's a Go CLI tool with an embedded Lua scripting engine, a Python AWS Lambda API backend, and a React/TypeScript UI.

## Common Commands

### Go (core CLI)

```bash
make install          # Build and install to $GOPATH/bin (default target)
make build            # Cross-compile for linux, darwin, windows, wasm
make test             # Run all Go tests: go test -v ./...
go test -v ./internal # Run tests for a single package
go test -v -run TestName ./internal  # Run a single test
```

Version, git summary, build date, and GitHub OAuth credentials are injected via `-ldflags` in the Makefile. Always use `make install` rather than bare `go install`.

Dependencies are vendored in `vendor/`. After adding deps, run `go mod vendor`.

### Python API (`gogen-api/`)

```bash
cd gogen-api
./start_dev.sh        # Starts DynamoDB Local + MinIO via docker-compose, then SAM local API on port 4000
./setup_local_db.sh   # Seeds local DynamoDB schema
sam build && sam local start-api --port 4000 --docker-network lambda-local
./deploy_lambdas.sh   # Deploy to AWS (requires credentials)
```

### UI (`ui/`)

```bash
cd ui
npm run dev           # Vite dev server (copies wasm from build/wasm/ first)
npm run build         # Production build
npm test              # Jest tests
```

## Architecture

### Go Package Layout

All packages are at the top level (no `cmd/` or `pkg/` convention):

- **`main.go`** — CLI entry point using `urfave/cli.v1`. Maps CLI flags to `GOGEN_*` env vars.
- **`internal/`** — Core package. Config singleton, `Sample` struct, `Token` processing, API client, sharing. Imported as `config` throughout (`config "github.com/coccyx/gogen/internal"`).
- **`generator/`** — Reads `GenQueueItem` from channel, dispatches to sample-based or Lua generators.
- **`outputter/`** — Reads `OutQueueItem` from channel, dispatches to output destinations (stdout, file, HTTP, Kafka, network, devnull, buf).
- **`run/`** — Orchestrates the pipeline: timers -> generator worker pool -> outputter worker pool.
- **`timer/`** — One timer goroutine per Sample; handles backfill and realtime intervals.
- **`rater/`** — Controls event rate (config-based, time-of-day/weekday, kbps, Lua script).
- **`template/`** — Output formatting (raw, JSON, CSV, splunkhec, syslog, elasticsearch).
- **`logger/`** — Thin logrus wrapper with file/func/line context hook.

### Data Flow

```
YAML/JSON Config -> internal.Config singleton (sync.Once)
    -> [Timer goroutine per Sample]
    -> GenQueueItem channel -> [Generator worker pool]
    -> OutQueueItem channel -> [Outputter worker pool]
    -> output destination
```

Concurrency is channel + goroutine worker pools. Worker counts set by `GeneratorWorkers` and `OutputWorkers` config fields.

### Key Interfaces

- `internal.Generator` — `Gen(item *GenQueueItem) error`
- `internal.Outputter` — `Send(events []map[string]string, sample *Sample, outputTemplate string) error`
- `internal.Rater` — `EventsPerInterval(s *Sample) int`

### Config System

Config is a **singleton** via `sync.Once`. Controlled by environment variables:
- `GOGEN_HOME`, `GOGEN_FULLCONFIG`, `GOGEN_CONFIG_DIR`, `GOGEN_SAMPLES_DIR`
- Remote configs fetched from `https://api.gogen.io` (override with `GOGEN_APIURL`)

In tests, call `config.ResetConfig()` before `config.NewConfig()` to get a fresh instance. Tests commonly use `config.SetupFromString(yamlStr)` to inject inline YAML config.

### gogen-api (Python Lambda)

Each Lambda function is a separate `.py` file in `gogen-api/`. Backed by DynamoDB + S3. Originally Python 2.7, being updated to Python 3. AWS SAM template at `gogen-api/template.yaml`.

### UI (React/TypeScript)

Vite + React 18 + TypeScript + Tailwind CSS. Components in `src/components/`, pages in `src/pages/`, API clients in `src/api/`, types in `src/types/`. Tests use Jest + React Testing Library, placed adjacent to source as `.test.tsx`.

## CI/CD

GitHub Actions (`.github/workflows/ci.yml`):
- Push to `master`/`dev` or any PR: runs `make test`, then on `master`/`dev` cross-compiles, builds Docker, pushes artifacts to S3, deploys UI and Lambdas.
- Tag pushes (`v*.*.*`): full release workflow via `release.yml` — builds, creates GitHub release, pushes Docker images, deploys to production.

## Lua Scripting

Generators (`generator/lua.go`) and raters (`rater/script.go`) support embedded Lua via `gopher-lua` + `gopher-luar`. Lua state persists across calls within a run.
