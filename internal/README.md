# Go Sentinel Internal Packages

This directory contains the core implementation of Go Sentinel CLI, organized into focused, well-encapsulated packages following clean architecture principles. These packages are internal to the project and implement the modular architecture established after comprehensive refactoring.

**🚧 Current Status**: Architecture refactoring is 85% complete (7/10 violations fixed). The app package still contains 3 violations (595 lines) that need to be extracted to dedicated packages plus a configuration validation fix before Phase 1 can proceed.

## 🏗️ Architecture Overview

The internal packages implement a layered, modular architecture with clear separation of concerns:

```
internal/
├── app/          # Application orchestration & lifecycle management
├── config/       # Configuration management & CLI argument parsing  
├── test/         # Test execution, processing & caching system
├── watch/        # File monitoring & intelligent watch mode
└── ui/           # User interface & display components
```

## 📦 Package Overview

### Core Application Layer

#### `app/` - Application Orchestration 🚧 **3 VIOLATIONS REMAIN**
**Purpose**: Coordinates all application components and manages the overall application lifecycle.

**Current Status**: Contains 3 architecture violations that need extraction:
- ❌ **Event Handling** (198 lines) - Should be in `internal/events/`
- ❌ **Lifecycle Management** (160 lines) - Should be in `internal/lifecycle/`  
- ❌ **Dependency Container** (237 lines) - Should be in `internal/container/`

**Completed Components**:
- ✅ **ApplicationController**: Main application flow orchestration (clean)
- ✅ **Adapters**: Factory + Adapter pattern for all dependencies (clean)

**Key Interfaces**:
- `ApplicationController` - Main application coordination
- `LifecycleManager` - Application lifecycle management
- `DependencyContainer` - Dependency injection and component management

#### `config/` - Configuration Management
**Purpose**: Handles all configuration loading, validation, and CLI argument parsing.

**Key Components**:
- **Configuration Loading**: JSON config file parsing with defaults
- **CLI Arguments**: Command-line argument parsing and validation
- **Configuration Merging**: Merging CLI args with config files (CLI takes precedence)
- **Validation**: Comprehensive configuration validation and error reporting

**Key Types**:
- `Config` - Main configuration structure
- `Args` - CLI argument structure
- `ConfigLoader` - Configuration loading interface

### Test Execution System

#### `test/` - Test Execution & Processing
**Purpose**: Comprehensive test execution, result processing, and caching system.

**Subpackages**:
- `runner/` - Test execution engines (basic, optimized, parallel)
- `processor/` - Test output parsing, aggregation, and result processing
- `cache/` - Test result caching and optimization
- `metrics/` - Performance metrics and benchmarking
- `recovery/` - Error recovery and resilience mechanisms

**Key Features**:
- `go test -json` output parsing
- Parallel test execution with resource management
- Intelligent test result caching
- Performance metrics collection
- Comprehensive error recovery

### File Monitoring System

#### `watch/` - File Monitoring & Watch Mode
**Purpose**: Intelligent file system monitoring with debounced test execution.

**Subpackages**:
- `core/` - Core watch interfaces and types
- `watcher/` - File system monitoring implementation
- `debouncer/` - Event debouncing and deduplication
- `coordinator/` - Watch mode orchestration and coordination

**Key Features**:
- Cross-platform file system monitoring
- Intelligent debouncing of rapid file changes
- Pattern-based file filtering
- Smart test selection based on changed files
- Watch mode lifecycle management

### User Interface System

#### `ui/` - User Interface & Display
**Purpose**: Beautiful terminal UI with Vitest-style output and rich display components.

**Subpackages**:
- `display/` - Test result rendering and formatting
- `colors/` - Color themes and terminal detection
- `icons/` - Icon providers and visual elements
- `renderer/` - Progressive rendering and live updates

**Key Features**:
- Three-part display structure (header, content, summary)
- Color themes with terminal capability detection
- Multiple icon sets (Unicode, ASCII, minimal, none)
- Real-time progress indicators
- Live updating during test execution

## 🛠️ Development Guidelines

### Package Principles

1. **Single Responsibility**: Each package has one clear, well-defined purpose
2. **Dependency Inversion**: Packages depend on interfaces, not concrete implementations
3. **Interface Segregation**: Small, focused interfaces rather than large ones
4. **Open/Closed**: Open for extension, closed for modification
5. **Clean Boundaries**: Clear package boundaries with minimal coupling

### Code Organization Patterns

- **Interfaces First**: Define interfaces before implementations
- **Composition over Inheritance**: Use embedding and composition
- **Testable Design**: All components designed for easy testing
- **Context Awareness**: Use `context.Context` for cancellation and timeouts
- **Error Handling**: Rich error context with proper error wrapping

### Dependency Rules

```
app/ ←────────────── (orchestrates all)
├─── config/        (configuration)
├─── test/          (test execution)
├─── watch/         (file monitoring)  
└─── ui/            (display)
     └─── pkg/      (shared models)
```

**Rules**:
- `app/` can import from all other internal packages
- Other packages can only import from `pkg/` and standard library
- No circular dependencies between internal packages
- All external dependencies managed at package level

### Testing Standards

Each package must include:

1. **Unit Tests**: ≥90% coverage for all exported functionality
2. **Integration Tests**: Test component interactions
3. **Interface Tests**: Test against interface contracts
4. **Benchmark Tests**: For performance-critical code
5. **Example Tests**: Demonstrating package usage

### Quality Standards

- **Complexity**: ≤2.5 average cyclomatic complexity
- **Maintainability**: ≥90% maintainability index
- **Documentation**: All exported symbols documented
- **Linting**: Zero violations above "Warning" level
- **Performance**: Benchmarked critical paths

## 🧪 Testing

### Running Tests

```bash
# Run all internal package tests
go test ./internal/...

# Run with coverage
go test -cover ./internal/...

# Run with race detection
go test -race ./internal/...

# Run specific package
go test ./internal/app/...

# Run benchmarks
go test -bench=. ./internal/...

# Generate coverage report
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out
```

### Test Organization

- **Package Tests**: `*_test.go` files alongside source
- **Integration Tests**: `integration_test.go` files  
- **Benchmark Tests**: `*_bench_test.go` files
- **Example Tests**: `example_*_test.go` files

## 📚 Documentation

### Package Documentation

Each package contains:
- **README.md**: Package overview, usage, examples
- **doc.go**: Package-level documentation
- **examples/**: Usage examples and tutorials

### API Documentation

- All exported types and functions documented
- Usage examples in documentation
- Integration examples between packages
- Performance characteristics documented

## 🔄 Dependencies

### Allowed Dependencies

**Internal packages may depend on**:
- Go standard library
- `pkg/models` and `pkg/events` (shared packages)
- Well-vetted external dependencies (minimal)

**External Dependencies**:
- `github.com/spf13/cobra` - CLI framework
- `github.com/fsnotify/fsnotify` - File system notifications
- `github.com/fatih/color` - Terminal colors
- Testing and development tools only

### Forbidden Dependencies

- Direct dependencies between internal packages (except through interfaces)
- Large external frameworks
- Platform-specific libraries (use build tags instead)
- Packages with restrictive licenses

## 🚀 Performance Considerations

- **Memory Efficiency**: Use `sync.Pool` for frequently allocated objects
- **Goroutine Management**: Proper lifecycle management for goroutines
- **Context Propagation**: Use context for cancellation and timeouts
- **Resource Cleanup**: Proper cleanup in defer statements
- **Profiling**: Regular profiling of performance-critical paths

## 🔒 Security Guidelines

- **Input Validation**: Validate all inputs at package boundaries
- **Context Timeouts**: Use timeouts for all external operations
- **Error Sanitization**: Sanitize error messages (no sensitive data)
- **Dependency Updates**: Keep dependencies updated
- **Secure Defaults**: Use secure defaults for all configuration

## 📈 Monitoring & Observability

- **Structured Logging**: Use structured logging throughout
- **Metrics Collection**: Collect key performance metrics
- **Error Tracking**: Comprehensive error tracking and reporting
- **Tracing**: Request/operation tracing where applicable
- **Health Checks**: Health check endpoints for monitoring

## 🔧 Package-Specific Guidelines

### Configuration Package (`config/`)
- Immutable configuration objects
- Validation at load time
- Clear error messages for invalid config
- Support for configuration hot-reloading

### Test Package (`test/`)
- Resource cleanup after test execution
- Timeout handling for long-running tests
- Proper isolation between test runs
- Streaming output processing

### Watch Package (`watch/`)
- Efficient file system event handling
- Proper debouncing of rapid events
- Resource cleanup on watch cancellation
- Cross-platform compatibility

### UI Package (`ui/`)
- Terminal capability detection
- Graceful degradation for limited terminals
- Responsive layout handling
- Accessible color schemes

## 📝 License

This code is part of the Go Sentinel project and is licensed under the [MIT License](../LICENSE).

---

For detailed package-specific documentation, see the README.md file in each package directory.
