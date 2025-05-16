# Go Sentinel Internal Packages

This directory contains the core implementation of Go Sentinel, organized into focused, well-encapsulated packages. These packages are internal to the project and not intended for direct use by external applications.

## 📦 Package Overview

### Core Components

#### `api/`
RESTful API server implementation with OpenAPI/Swagger documentation.
- **Handlers**: HTTP request handlers for API endpoints
- **Middleware**: Authentication, logging, and request processing
- **Models**: Data structures and validation
- **Server**: API server setup and configuration

#### `config/`
Configuration management with support for multiple sources (flags, env vars, config files).
- Environment-aware configuration loading
- Type-safe configuration access
- Default values and validation

#### `debouncer/`
Efficient event debouncing to handle rapid file system events.
- Configurable debounce intervals
- Per-package event coalescing
- Thread-safe implementation

#### `event/`
Event types and interfaces used for inter-package communication.
- Centralized event definitions
- Type-safe event publishing/subscribing
- Event filtering and transformation

#### `parser/`
Test output parsing and result extraction.
- `go test -json` output parsing
- Test result aggregation
- Failure analysis and categorization

#### `runner/`
Test execution and process management.
- Concurrent test execution
- Timeout and cancellation support
- Process lifecycle management

#### `ui/`
Terminal user interface components.
- Interactive console output
- Progress indicators
- Keyboard event handling

#### `watcher/`
File system monitoring with efficient event handling.
- Recursive directory watching
- Configurable file patterns
- Cross-platform file system notifications

### Web Components

#### `web/server/`
HTTP server implementation with support for both API and web UI.
- Static file serving
- Template rendering
- WebSocket support

#### `web/handlers/`
HTTP request handlers for web interface.
- Page rendering
- Form handling
- Error pages

#### `web/middleware/`
HTTP middleware components.
- Request logging
- Error handling
- Security headers
- Session management

## 🛠 Development Guidelines

### Package Principles

1. **Single Responsibility**: Each package should have a single, well-defined purpose
2. **Encapsulation**: Hide implementation details behind clean interfaces
3. **Dependency Direction**: Dependencies should point inward, toward more stable packages
4. **Testability**: Design for easy testing with minimal mocking

### Code Organization

- Keep package interfaces small and focused
- Use subpackages to break down large packages
- Document exported types and functions
- Follow Go's standard project layout

### Error Handling

- Use sentinel errors for expected error conditions
- Wrap errors with context using `fmt.Errorf("%w", err)`
- Log errors at the appropriate level
- Return errors that are useful to callers

### Concurrency

- Use `sync` primitives for synchronization
- Prefer channels for communication between goroutines
- Use `context.Context` for cancellation and timeouts
- Document goroutine ownership and lifecycle

## 🔍 Testing

Each package should include:

1. Unit tests for all exported functionality
2. Table-driven tests for functions with multiple code paths
3. Test helpers for common test scenarios
4. Benchmarks for performance-critical code

Run tests with:

```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📚 Documentation

- Document all exported types and functions
- Include usage examples in package documentation
- Keep README.md files up to date
- Document any non-obvious implementation details

## 🔄 Dependencies

Internal packages may depend on:

- Standard library packages
- Other internal packages (with care to avoid cycles)
- A minimal set of well-vetted external dependencies

Avoid depending on:

- Implementation details of other packages
- External packages when standard library alternatives exist
- Packages with restrictive licenses

## 🚀 Performance Considerations

- Profile before optimizing
- Use sync.Pool for frequently allocated objects
- Be mindful of memory allocations in hot paths
- Consider using `-race` during development

## 🔒 Security

- Validate all inputs
- Use context timeouts for all external operations
- Sanitize output to prevent XSS
- Keep dependencies updated

## 📈 Monitoring

- Use structured logging
- Include request IDs for tracing
- Expose metrics for key operations
- Monitor error rates and performance metrics

## 📝 License

This code is part of the Go Sentinel project and is licensed under the [MIT License](../LICENSE).

## WebSocket Integration
- WebSocket handler for real-time test updates
- Broadcaster pattern for pushing updates to all clients

## API Overview
- RESTful endpoints for tests, metrics, history, coverage, and settings
- All endpoints documented in `ROADMAP-API.md`

---
For more details, see `web/server/server.go` and `ROADMAP-API.md`.
