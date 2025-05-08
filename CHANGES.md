# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Initial project scaffolding following Go best practices (`cmd/`, `internal/`, etc.)
- Added `.gitignore`, `LICENSE`, and CI/CD setup
- Implemented file watcher with recursive `fsnotify` support
- Exclusion of `vendor/`, hidden, and symlinked directories
- TDD-driven watcher tests for file change detection, exclusion, and edge cases
- Implemented event debouncer supporting per-package buffering and quiet period
- TDD-driven debouncer tests for rapid, single, and overlapping events
- Pre-commit hooks and linting configuration
- **Phase 3: Go Test Runner**
  - Comprehensive TDD-driven runner tests for: correct package execution, stdout/stderr capture, error handling, build errors, real-time output, and goroutine pipeline integration
  - Runner implementation using `os/exec` for `go test -json`, robust output streaming with `bufio.Scanner`, and per-test rerun support
  - Utilities for running `go version`, `go env`, `go list`, `go mod tidy`, and `go fmt`
  - Cleaned up debug/test output and improved test reliability

### Changed
- Updated `ROADMAP.md` to reflect completed Phase 3.1 and 3.2 milestones and next steps
- Refactored runner and utility code for maintainability and extensibility

### Fixed
- Cleaned up test log output and removed obsolete debug code
- Resolved all runner test flakiness and output issues

---

---
