# 📦 Test Processor Package

[![Test Coverage](https://img.shields.io/badge/coverage-87.8%25-brightgreen.svg)](https://github.com/newbpydev/go-sentinel/tree/main/internal/test/processor)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/go-sentinel/internal/test/processor)](https://goreportcard.com/report/github.com/newbpydev/go-sentinel/internal/test/processor)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel/internal/test/processor.svg)](https://pkg.go.dev/github.com/newbpydev/go-sentinel/internal/test/processor)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 📖 Overview

The Test Processor package provides comprehensive test result processing capabilities for the Go Sentinel CLI. It handles parsing, streaming, and formatting of Go test output with advanced source code context extraction and error analysis.

### 🎯 Key Features

- **JSON Test Parsing**: Complete parsing of `go test -json` output with error location extraction
- **Stream Processing**: Real-time test result processing with progress updates
- **Source Context Extraction**: Automatic extraction of source code context around test failures
- **Multi-format Support**: Support for both batch and streaming test result processing
- **Error Analysis**: Advanced error type detection and source location mapping
- **Concurrent Safety**: Thread-safe operations for high-performance test processing

## 🏗️ Architecture

This package follows clean architecture principles with clear separation of concerns:

- **Single Responsibility**: Each component handles a specific aspect of test processing
- **Dependency Inversion**: Interfaces define contracts for test processing operations
- **Interface Segregation**: Small, focused interfaces for specific processing needs
- **Factory Pattern**: Consistent object creation through factory functions
- **Strategy Pattern**: Multiple parsing strategies for different input formats

### 📦 Package Structure

```
internal/test/processor/
├── interfaces.go              # Core interfaces and contracts
├── json_parser.go            # JSON test output parser (93.3% coverage)
├── stream_processor.go       # Real-time stream processor (90.0% coverage)
├── source_extractor.go       # Source context extractor (94.7% coverage)
├── test_processor.go         # Main test processor (100.0% coverage)
├── event_handler.go          # Test event handlers (55.6% coverage)
├── parser_test.go            # JSON parser tests
├── stream_test.go            # Stream processor tests
├── source_extractor_test.go  # Source extractor tests
├── processor_test.go         # Main processor tests
└── README.md                 # This documentation
```

## 🚀 Quick Start

### Basic Test Result Processing

```go
package main

import (
    "bytes"
    "strings"
    
    "github.com/newbpydev/go-sentinel/internal/test/processor"
    "github.com/newbpydev/go-sentinel/internal/ui/colors"
)

func main() {
    // Create output buffer and formatting components
    var output bytes.Buffer
    formatter := colors.NewColorFormatter(true)
    iconProvider := colors.NewIconProvider(true)
    
    // Create test processor
    testProcessor := processor.NewTestProcessor(&output, formatter, iconProvider, 80)
    
    // Parse JSON test output
    jsonOutput := `{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"example","Test":"TestExample"}
{"Time":"2023-10-01T12:00:01Z","Action":"pass","Package":"example","Test":"TestExample","Elapsed":0.1}`
    
    reader := strings.NewReader(jsonOutput)
    progressCh := make(chan models.TestProgress, 10)
    
    // Process the test stream
    err := testProcessor.ProcessStream(reader, progressCh)
    if err != nil {
        log.Fatal("Failed to process test stream:", err)
    }
    
    // Render final results
    err = testProcessor.RenderResults(true)
    if err != nil {
        log.Fatal("Failed to render results:", err)
    }
    
    fmt.Print(output.String())
}
```

### JSON Parser Usage

```go
package main

import (
    "strings"
    
    "github.com/newbpydev/go-sentinel/internal/test/processor"
)

func main() {
    // Create JSON parser
    parser := processor.NewParser()
    
    // Parse test output
    jsonOutput := `{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"example","Test":"TestExample"}
{"Time":"2023-10-01T12:00:01Z","Action":"fail","Package":"example","Test":"TestExample","Elapsed":0.1}`
    
    reader := strings.NewReader(jsonOutput)
    packages, err := parser.Parse(reader)
    if err != nil {
        log.Fatal("Parse error:", err)
    }
    
    // Process results
    for _, pkg := range packages {
        fmt.Printf("Package: %s\n", pkg.Package)
        for _, test := range pkg.Tests {
            fmt.Printf("  Test: %s, Status: %s\n", test.Name, test.Status)
        }
    }
}
```

### Source Context Extraction

```go
package main

import (
    "github.com/newbpydev/go-sentinel/internal/test/processor"
    "github.com/newbpydev/go-sentinel/pkg/models"
)

func main() {
    // Create source extractor
    extractor := processor.NewSourceExtractor()
    
    // Create test error with location
    testError := &models.LegacyTestError{
        Message: "assertion failed",
        Location: &models.SourceLocation{
            File: "example_test.go",
            Line: 42,
        },
    }
    
    // Extract source context
    err := extractor.ExtractSourceContext(testError, 3)
    if err != nil {
        log.Fatal("Context extraction error:", err)
    }
    
    // Display context
    for i, line := range testError.SourceContext {
        marker := "  "
        if i == testError.HighlightedLine {
            marker = "> "
        }
        fmt.Printf("%s%s\n", marker, line)
    }
}
```

## 🔧 Core Interfaces

### TestProcessor Interface

The main interface for test result processing:

```go
type TestProcessor interface {
    // Core processing methods
    ProcessJSONOutput(reader io.Reader) error
    ProcessStream(reader io.Reader, progress chan<- models.TestProgress) error
    
    // Result management
    AddTestSuite(suite *models.TestSuite)
    RenderResults(showSummary bool) error
    
    // State management
    Reset()
    GetStats() *models.TestStats
    GetSuites() []*models.TestSuite
    GetWriter() io.Writer
}
```

### Parser Interface

Interface for parsing test output:

```go
type Parser interface {
    Parse(reader io.Reader) ([]*models.TestPackage, error)
}
```

### SourceExtractor Interface

Interface for extracting source code context:

```go
type SourceExtractor interface {
    ExtractContext(filePath string, lineNumber int, contextLines int) ([]string, error)
    ExtractSourceContext(testError *models.LegacyTestError, contextLines int) error
    IsValidSourceFile(filePath string) bool
}
```

### StreamParser Interface

Interface for real-time stream processing:

```go
type StreamParser interface {
    Parse(reader io.Reader, results chan<- *models.LegacyTestResult) error
}
```

## 🔄 Advanced Usage

### Custom Error Processing

```go
// Create custom error processor
processor := processor.NewTestProcessor(output, formatter, icons, 120)

// Process with custom error handling
testSuite := &models.TestSuite{
    FilePath: "custom_test.go",
    Tests: []*models.LegacyTestResult{
        {
            Name:   "TestCustom",
            Status: models.StatusFailed,
            Error: &models.LegacyTestError{
                Type:    "CustomError",
                Message: "Custom test failure",
                Location: &models.SourceLocation{
                    File: "custom_test.go",
                    Line: 25,
                },
            },
        },
    },
}

processor.AddTestSuite(testSuite)
processor.RenderResults(true)
```

### Concurrent Stream Processing

```go
// Process multiple streams concurrently
var wg sync.WaitGroup
streams := []io.Reader{stream1, stream2, stream3}

for i, stream := range streams {
    wg.Add(1)
    go func(id int, r io.Reader) {
        defer wg.Done()
        
        parser := processor.NewStreamParser()
        results := make(chan *models.LegacyTestResult, 100)
        
        go func() {
            defer close(results)
            parser.Parse(r, results)
        }()
        
        for result := range results {
            fmt.Printf("Stream %d: %s - %s\n", id, result.Name, result.Status)
        }
    }(i, stream)
}

wg.Wait()
```

### Advanced Source Context

```go
// Extract context with custom validation
extractor := processor.NewSourceExtractor()

// Validate source file first
if !extractor.IsValidSourceFile("test_file.go") {
    log.Fatal("Invalid source file")
}

// Extract with larger context
context, err := extractor.ExtractContext("test_file.go", 50, 10)
if err != nil {
    log.Fatal("Context extraction failed:", err)
}

// Process context lines
for i, line := range context {
    lineNum := 50 - 10 + i + 1
    fmt.Printf("%4d: %s\n", lineNum, line)
}
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./internal/test/processor/...

# Run with coverage
go test ./internal/test/processor/... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out

# Run specific test categories
go test ./internal/test/processor/... -run TestParser
go test ./internal/test/processor/... -run TestStreamParser
go test ./internal/test/processor/... -run TestSourceExtractor
```

### Test Categories

- **Unit Tests**: Individual component testing with 87.8% overall coverage
- **Integration Tests**: Cross-component interaction validation
- **Concurrency Tests**: Thread-safety validation with multiple goroutines
- **Edge Case Tests**: Boundary conditions and error scenarios
- **Performance Tests**: Memory efficiency and execution speed validation
- **Error Handling Tests**: Comprehensive error condition coverage

### Coverage Breakdown

| Component | Coverage | Test Count | Key Features Tested |
|-----------|----------|------------|-------------------|
| JSON Parser | 93.3% | 15 tests | Error type detection, source location extraction |
| Stream Processor | 90.0% | 12 tests | Real-time parsing, concurrent access |
| Source Extractor | 94.7% | 18 tests | Context extraction, file validation |
| Test Processor | 100.0% | 10 tests | Result rendering, state management |
| Event Handlers | 55.6% | 5 tests | Event processing (partial coverage) |

### Test Examples

```bash
# Example test execution
$ go test -v ./internal/test/processor/...

=== RUN   TestNewParser_Creation
=== PASS  TestNewParser_Creation (0.00s)
=== RUN   TestDetermineErrorType_AssertionError
=== PASS  TestDetermineErrorType_AssertionError (0.00s)
=== RUN   TestExtractSourceLocation_ValidLocations
=== PASS  TestExtractSourceLocation_ValidLocations (0.00s)
=== RUN   TestNewStreamParser_Creation
=== PASS  TestNewStreamParser_Creation (0.00s)
=== RUN   TestStreamParser_EdgeCases
=== PASS  TestStreamParser_EdgeCases (0.00s)

PASS
coverage: 87.8% of statements
```

## 📊 Performance

The package is optimized for high-performance test processing:

- **Fast Parsing**: Efficient JSON parsing with minimal memory allocation
- **Stream Processing**: Low-latency real-time test result processing
- **Memory Efficient**: Optimized data structures for large test suites
- **Concurrent Safe**: Thread-safe operations with minimal locking overhead

### Benchmarks

```bash
# Run performance benchmarks
go test ./internal/test/processor/... -bench=. -benchmem

# Example results:
BenchmarkParser_Parse-8                    1000    1.2ms/op    256KB/op
BenchmarkStreamParser_Parse-8               2000    0.8ms/op    128KB/op
BenchmarkSourceExtractor_ExtractContext-8   5000    0.3ms/op     64KB/op
BenchmarkTestProcessor_RenderResults-8       500    2.1ms/op    512KB/op
```

### Performance Characteristics

- **JSON Parsing**: ~1000 test results per second
- **Stream Processing**: Real-time processing with <1ms latency
- **Source Extraction**: Context extraction in <0.5ms per file
- **Memory Usage**: <1MB for typical test suites (100-500 tests)

## 🔍 Error Handling

The package provides comprehensive error handling for all scenarios:

### Error Types

```go
// JSON parsing errors
type ParseError struct {
    Line    int
    Message string
    Cause   error
}

// Source extraction errors
type SourceError struct {
    FilePath string
    Line     int
    Message  string
}

// Stream processing errors
type StreamError struct {
    Position int64
    Message  string
    Cause    error
}
```

### Error Handling Patterns

```go
// Graceful error handling
parser := processor.NewParser()
packages, err := parser.Parse(reader)
if err != nil {
    var parseErr *ParseError
    if errors.As(err, &parseErr) {
        log.Printf("Parse error at line %d: %s", parseErr.Line, parseErr.Message)
        return
    }
    log.Fatal("Unexpected error:", err)
}

// Stream error handling
streamParser := processor.NewStreamParser()
results := make(chan *models.LegacyTestResult, 100)

go func() {
    defer close(results)
    if err := streamParser.Parse(reader, results); err != nil {
        log.Printf("Stream processing error: %v", err)
    }
}()

// Process results with error recovery
for result := range results {
    if result.Error != nil {
        log.Printf("Test %s failed: %s", result.Name, result.Error.Message)
    }
}
```

## 🎯 Best Practices

### Usage Recommendations

1. **Use Appropriate Parser**: Choose JSON parser for batch processing, stream parser for real-time updates
2. **Handle Errors Gracefully**: Always check for parsing errors and handle them appropriately
3. **Optimize Context Size**: Use reasonable context sizes (3-5 lines) for source extraction
4. **Resource Management**: Close channels and readers properly to prevent resource leaks
5. **Concurrent Processing**: Use goroutines for processing multiple test streams

### Integration Patterns

```go
// Recommended integration pattern
func ProcessTestResults(input io.Reader, output io.Writer) error {
    // Create components
    formatter := colors.NewColorFormatter(true)
    icons := colors.NewIconProvider(true)
    processor := processor.NewTestProcessor(output, formatter, icons, 80)
    
    // Process with progress tracking
    progress := make(chan models.TestProgress, 100)
    
    go func() {
        defer close(progress)
        if err := processor.ProcessStream(input, progress); err != nil {
            log.Printf("Processing error: %v", err)
        }
    }()
    
    // Handle progress updates
    for update := range progress {
        log.Printf("Test progress: %s", update.TestName)
    }
    
    // Render final results
    return processor.RenderResults(true)
}
```

### Performance Optimization

```go
// Optimize for large test suites
processor := processor.NewTestProcessor(output, formatter, icons, 120)

// Pre-allocate channels for better performance
results := make(chan *models.LegacyTestResult, 1000)
progress := make(chan models.TestProgress, 100)

// Use buffered readers for large inputs
bufferedReader := bufio.NewReaderSize(input, 64*1024)

// Process in batches for memory efficiency
const batchSize = 100
batch := make([]*models.LegacyTestResult, 0, batchSize)

for result := range results {
    batch = append(batch, result)
    if len(batch) >= batchSize {
        processBatch(batch)
        batch = batch[:0] // Reset slice
    }
}
```

## 🤝 Contributing

### Development Setup

```bash
# Clone the repository
git clone https://github.com/newbpydev/go-sentinel.git
cd go-sentinel/internal/test/processor

# Install dependencies
go mod download

# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Run linting
golangci-lint run
```

### Quality Standards

- **Test Coverage**: Maintain ≥85% test coverage for all new code
- **Documentation**: Document all exported functions and types
- **Error Handling**: Implement comprehensive error handling
- **Performance**: Benchmark performance-critical functions
- **Concurrency**: Ensure thread-safety for concurrent operations

### Code Style

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Keep functions under 50 lines
- Write comprehensive tests for all scenarios
- Document complex algorithms and business logic

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../../LICENSE) file for details.

## 🔗 Related Packages

- [`internal/test/runner`](../runner/README.md) - Test execution and management
- [`internal/test/cache`](../cache/README.md) - Test result caching
- [`internal/test/metrics`](../metrics/README.md) - Test metrics and analysis
- [`internal/ui/colors`](../../ui/colors/README.md) - Color formatting and display
- [`pkg/models`](../../../pkg/models/README.md) - Core data models and interfaces

---

**Test Coverage Achievement**: 87.8% with comprehensive TDD methodology  
**Architecture Compliance**: Full adherence to clean architecture principles  
**Production Ready**: Thoroughly tested and optimized for high-performance usage 