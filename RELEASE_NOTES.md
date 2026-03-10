# Gogen Release Notes

## Version 0.13.0

### New Features
- Added GitHub OAuth authentication for configuration management
  - OAuth login flow with GitHub for user identity
  - Users can create, edit, and delete their own configurations
  - My Configurations page for managing owned configs
  - Protected routes for authenticated operations

### UI Overhaul
- Redesigned UI with dark terminal-style developer theme
  - Terminal color palette with JetBrains Mono font
  - Monaco editor and xterm terminal with dark themes
  - Compact, developer-friendly layout
- Unified create/edit screens with integrated execution panel
  - Test configurations while editing with terminal and structured output tabs
  - Interactive JSON browser with per-event collapse/expand

### Code Quality
- Improved test coverage from 52% to 75% with simplified test helpers
- Added 28 integration tests covering config, templates, tokens, and concurrency
- DRY refactoring: uniform rate maps, generic setDefault helper, string builders
- Replaced deprecated io/ioutil with io and os equivalents
- Split config.go (900+ lines) into focused files for maintainability
- Added HTTP helper with proper resource cleanup and structured error types
- Fixed error handling: return parse errors, log marshal errors
- Fixed logger package name to match directory convention
- Fixed concurrency safety with sync.Once for ROT initialization
- Replaced deprecated oauth2.NoContext with context.Background()

### Infrastructure
- Added CORS preflight OPTIONS handlers for all API endpoints
- Added staging environment support for API Gateway and Lambda
- Updated CI/CD with OAuth credential management per environment

### Bug Fixes
- Fixed race condition in ROT channel initialization
- Fixed flaky TestTCPRFC5424Output trailing newline issue
- Fixed Splunk HEC field transform duplication

## Version 0.12.1

### New Features
- Added automated release cycle with Docker image versioning
  - Docker images now tagged with semantic version numbers (e.g., `clintsharp/gogen:0.12.1`)
  - Added GitHub Actions workflow for automated releases on version tags
  - Both `gogen` and `gogen-api` images are versioned and published to Docker Hub

### Improvements
- Enhanced CI/CD pipeline to support version-based releases
  - CI workflow now skips on tag pushes to avoid conflicts with release workflow
  - Updated `docker-push.sh` script to support version tagging
  - Release workflow creates GitHub releases with build artifacts

### Bug Fixes
- Fixed race condition with nil pointer in outputter.go

## Version 0.12.0 

### Breaking Changes
- Configuration storage backend migrated from GitHub Gists to S3 (managed via `gogen-api`).
  - Direct interaction with GitHub Gists for configuration management is no longer supported.

### New Features
- Added a web-based UI for managing Gogen configurations.
  - Built with React, TypeScript, and Tailwind CSS.
  - Allows creating, viewing, editing, and deleting configurations.

### Improvements
- Refactored core Go codebase (`gogen`) to utilize the `gogen-api` for configuration management.
- Significantly updated the backend API (`gogen-api`):
  - Refactored core API endpoints (CRUD, search).
  - Migrated build and deployment to AWS SAM (`template.yaml`, `deploy_lambdas.sh`).
  - Improved local development environment with Docker Compose (`docker-compose.yml`) and setup scripts (`setup_local_db.sh`).
  - Enhanced utilities for CORS, database interactions, and S3 operations.
- Updated Go dependencies (`go.mod`, `go.sum`).
- Minor refactoring in core Go components (`internal/`, `outputter/`).

### Infrastructure
- Updated GitHub Actions CI workflow (`.github/workflows/ci.yml`).
- Added setup scripts for Python virtual environment (`setup_venv.sh`) and local DynamoDB (`gogen-api/setup_local_db.sh`).
- Added Docker Compose configuration (`gogen-api/docker-compose.yml`) for local API development.

### Internal Changes
- Added Cursor AI rules for API and UI development (`.cursor/rules/`).
- Updated `.gitignore`.

### Bug Fixes
- Fixed bug where statistics were not collected correctly during shutdown on short-duration runs.

## Version 0.11.0

### Breaking Changes
- Removed Splunk modinput and Splunk app support
  - All Splunk-specific functionality has been removed to focus on core event generation capabilities
  - Users relying on Splunk integration will need to use alternative methods for data ingestion

### Improvements
- Significantly improved test coverage across multiple components:
  - Added network testing suite
  - Added devnull and stdout output tests
  - Enhanced HTTP output testing
  - Added RunOnce operation tests
  - Improved timer and core functionality tests

### Infrastructure
- Migrated from Travis CI to GitHub Actions
- Added Coveralls integration for code coverage reporting
- Updated deployment process

### Bug Fixes
- Fixed issue with long intervals not shutting down in a reasonable time
- Fixed timezone extraction bug in tests
- Addressed flaky test behaviors

### Internal Changes
- Refactored core components for better maintainability
- Updated configuration handling
- Enhanced logging system reliability

For more information, please refer to the [documentation](README/Reference.md). 