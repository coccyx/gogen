# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Gogen is an open-source data generator for creating demo and test data with first-class support for time series data. It generates logs and metrics for testing time series systems.

**Multi-component architecture:**
- **Go CLI** (`main.go`, `internal/`, `run/`, `outputter/`, `generator/`, `rater/`, `timer/`) - Core event generation engine
- **Python API** (`gogen-api/`) - AWS Lambda REST API for configuration storage (DynamoDB + S3)
- **React UI** (`ui/`) - TypeScript/React frontend for configuration management
- **WebAssembly** - Go compiled to WASM for in-browser execution

## Common Commands

### Go CLI
```bash
make build          # Build for Linux, macOS, Windows, and WASM
make install        # Install gogen locally
make test           # Run Go tests (go test -v ./...)
go test -v ./outputter  # Run tests for a single package
```

### UI (from ui/ directory)
```bash
npm run dev         # Start Vite dev server (copies WASM first)
npm run build       # Production build
npm run test        # Jest unit tests
npm run test:watch  # Jest watch mode
```

### API (from gogen-api/ directory)
```bash
./start_dev.sh      # Start local dev environment (DynamoDB Local + MinIO + SAM)
./stop_dev.sh       # Stop local dev environment
./setup_local_db.sh # Set up local DynamoDB schema
sam build           # Build Lambda functions
sam local start-api --host 0.0.0.0 --port 4000 --warm-containers EAGER --docker-network lambda-local
```

### Python Virtual Environment
```bash
source .pyvenv/bin/activate  # Activate from project root
pip install -r gogen-api/requirements.txt
```

## Architecture

### Event Generation Flow
1. CLI parses YAML/JSON configuration
2. Configuration specifies generators (sample-based, Lua scripts, or custom)
3. Generators produce events with token substitution
4. Events flow through outputter to destination (stdout, file, HTTP, Kafka, network)
5. Rater controls generation speed (default, kbps, or Lua-based)
6. Timer manages scheduling across intervals

### Key Go Packages
- `internal/` - Configuration parsing, sharing, GitHub integration
- `generator/` - Event generation engines including Lua scripting support
- `outputter/` - Output destinations (stdout, file, HTTP, Kafka, network)
- `rater/` - Rate limiting implementations
- `run/` - Event generation runner and execution logic

### API Endpoints (port 4000 local, api.gogen.io production)
- `GET /v1/get/{gogen}` - Get configuration
- `POST /v1/upsert` - Create/update configuration
- `GET /v1/list` - List configurations
- `GET /v1/search` - Search configurations

### Local Services
- DynamoDB Local: http://localhost:8000
- MinIO (S3): http://localhost:9000 (API), http://localhost:9001 (console, user: minioadmin)
- SAM API: http://localhost:4000

## Development Notes

- Go version: 1.23.0 (toolchain go1.24.1)
- Python: 3.13
- UI: React 18, TypeScript 5.3, Vite 5.1, Tailwind CSS 3.4

### Build Flags
Version info is injected at build time via ldflags from `VERSION` file and git describe.

### Configuration Storage
Configurations are stored in S3 (`gogen-configs` bucket) with metadata in DynamoDB. The API migrated from GitHub Gists to S3 in v0.12.
