# Gogen Release Notes

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