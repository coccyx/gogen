# Gogen Release Notes

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