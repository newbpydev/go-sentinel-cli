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

### Changed
- Updated `ROADMAP.md` to reflect completed milestones

### Fixed
- N/A (initial implementation)

---
