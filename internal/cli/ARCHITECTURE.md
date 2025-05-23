# Go Sentinel CLI - Architecture & Design

## 🏗️ Architecture Overview

This document outlines the modular architecture of the Go Sentinel CLI, designed following Go best practices for testability, maintainability, and extensibility.

## 📁 Module Structure

```
internal/cli/
├── core/                    # Core business logic
│   ├── interfaces.go        # Central interface definitions
│   ├── types.go            # Core data types and enums
│   └── errors.go           # Custom error types
├── execution/              # Test execution engine
│   ├── runner.go           # Primary test runner implementation
│   ├── strategy.go         # Execution strategies (smart caching, etc.)
│   ├── cache.go           # Test result caching system
│   └── README.md          # Execution engine documentation
├── watch/                  # File watching and change detection
│   ├── watcher.go         # File system watcher
│   ├── debouncer.go       # Event debouncing
│   ├── analyzer.go        # File change analysis
│   └── README.md          # Watch system documentation
├── rendering/              # Output rendering and formatting
│   ├── renderer.go        # Test result renderer
│   ├── formatter.go       # Output formatting
│   ├── colors.go          # Color and styling
│   ├── icons.go           # Icon providers
│   └── README.md          # Rendering system documentation
├── config/                 # Configuration management
│   ├── config.go          # Configuration loading and validation
│   ├── args.go            # CLI argument parsing
│   └── README.md          # Configuration documentation
├── controller/             # Application coordination
│   ├── app.go             # Main application controller
│   └── README.md          # Controller documentation
└── testing/               # Testing utilities and test organization
    ├── complexity/        # Tests organized by complexity
    │   ├── unit/         # Simple unit tests
    │   ├── integration/  # Integration tests
    │   └── stress/       # Performance and stress tests
    ├── helpers/          # Test helper functions
    │   ├── fixtures.go   # Test fixtures and data
    │   ├── mocks.go      # Mock implementations
    │   └── builders.go   # Test data builders
    └── README.md         # Testing documentation
```

## 🔧 Core Design Principles

### 1. Separation of Concerns
- **Execution**: Handles test running and caching logic
- **Watch**: Manages file system monitoring and change detection
- **Rendering**: Handles all output formatting and display
- **Config**: Manages configuration and CLI parsing
- **Controller**: Coordinates between modules

### 2. Interface-Driven Design
All modules interact through well-defined interfaces, enabling:
- Easy testing with mocks
- Pluggable implementations
- Clear dependency boundaries

### 3. Dependency Injection
Components receive their dependencies explicitly:
- Improves testability
- Makes dependencies clear
- Enables easier refactoring

### 4. Immutable Data Structures
Where possible, data structures are immutable:
- Reduces bugs from shared state
- Improves concurrent safety
- Makes reasoning about code easier

## 🔄 Data Flow

```
CLI Args → Config → Controller → Execution Engine
                       ↓
File Changes → Watch System → Change Analyzer → Execution Engine
                                                      ↓
Test Results → Renderer → Formatted Output
```

## 📊 Complexity Levels

### Level 1: Unit Tests (Simple)
- Individual function testing
- No external dependencies
- Fast execution (< 10ms)

### Level 2: Integration Tests (Medium)
- Multiple component interaction
- May use file system or network
- Medium execution time (10ms - 1s)

### Level 3: Stress Tests (Complex)
- Performance and load testing
- Full system integration
- Longer execution time (> 1s)

## 🔌 Extension Points

The architecture provides several extension points:

1. **Execution Strategies**: New caching and optimization strategies
2. **Renderers**: Different output formats (JSON, XML, etc.)
3. **Watch Sources**: Different file system watchers
4. **Config Sources**: Environment variables, remote config, etc.

## 🎯 Benefits

- **Modularity**: Each package has a single responsibility
- **Testability**: Interfaces enable easy mocking
- **Maintainability**: Clear structure and documentation
- **Extensibility**: Well-defined extension points
- **Performance**: Optimized caching and execution strategies 