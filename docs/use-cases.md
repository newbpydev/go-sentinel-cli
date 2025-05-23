# Common Use Cases

Real-world examples of how to use Go Sentinel CLI in different development scenarios.

## Development Workflows

### TDD (Test-Driven Development)

**Scenario**: Writing new features with test-first approach

**Setup:**
```bash
# Create development config
cp docs/examples/development.json sentinel.config.json

# Start watch mode
go-sentinel run --watch
```

**Workflow:**
1. Write a failing test
2. Save the file → Go Sentinel automatically runs tests
3. See the failure in beautiful output
4. Write minimal code to make test pass
5. Save → Tests run again, now passing
6. Refactor with confidence

**Benefits:**
- Immediate feedback on test failures
- Clear visual indication of test status
- No manual test running required

### Bug Fixing

**Scenario**: Investigating and fixing a reported bug

**Commands:**
```bash
# Run tests with maximum verbosity for debugging
go-sentinel run -vvv --test="TestProblemArea*"

# Use debugging configuration
cp docs/examples/debugging.json sentinel.config.json
go-sentinel run --watch
```

**Features Used:**
- High verbosity to see detailed test execution
- Test filtering to focus on problematic area
- Watch mode for rapid iteration
- Sequential execution (no parallel) for clear debugging

### Code Review

**Scenario**: Reviewing a pull request with test changes

**Commands:**
```bash
# Run all tests to ensure nothing is broken
go-sentinel run ./...

# Run specific package tests that were modified
go-sentinel run ./pkg/modified-package

# Check test coverage and performance
go-sentinel run -v --parallel=1 ./pkg/modified-package
```

**Benefits:**
- Beautiful output makes it easy to spot issues
- Clear summary shows overall test health
- Performance metrics help identify slow tests

## CI/CD Integration

### GitHub Actions

**.github/workflows/test.yml:**
```yaml
name: Test Suite

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.23
    
    - name: Install Go Sentinel CLI
      run: go install github.com/newbpydev/go-sentinel-cli/cmd/go-sentinel-cli@latest
    
    - name: Run tests
      run: go-sentinel run --no-color --fail-fast --parallel=4 ./...
    
    - name: Upload coverage
      if: always()
      run: go-sentinel run --coverage ./...
```

### GitLab CI

**.gitlab-ci.yml:**
```yaml
stages:
  - test

test:
  stage: test
  image: golang:1.23
  before_script:
    - go install github.com/newbpydev/go-sentinel-cli/cmd/go-sentinel-cli@latest
  script:
    - go-sentinel run --no-color --timeout=10m ./...
  artifacts:
    reports:
      junit: test-results.xml
```

### Jenkins Pipeline

**Jenkinsfile:**
```groovy
pipeline {
    agent any
    
    stages {
        stage('Setup') {
            steps {
                sh 'go install github.com/newbpydev/go-sentinel-cli/cmd/go-sentinel-cli@latest'
            }
        }
        
        stage('Test') {
            steps {
                sh 'go-sentinel run --no-color --parallel=${BUILD_NUMBER} ./...'
            }
        }
    }
}
```

### Docker Integration

**Dockerfile.test:**
```dockerfile
FROM golang:1.23-alpine

RUN go install github.com/newbpydev/go-sentinel-cli/cmd/go-sentinel-cli@latest

WORKDIR /app
COPY . .

CMD ["go-sentinel", "run", "--no-color", "--fail-fast", "./..."]
```

**Usage:**
```bash
# Build test image
docker build -f Dockerfile.test -t myapp-test .

# Run tests in container
docker run --rm myapp-test
```

## Performance Testing

### Large Codebases

**Scenario**: Testing a large monorepo with many packages

**Configuration (performance.json):**
```json
{
  "parallel": 8,
  "timeout": "10m",
  "verbosity": 1,
  "colors": true
}
```

**Commands:**
```bash
# High-performance testing
go-sentinel run --parallel=8 --timeout=10m ./...

# Package-specific performance testing
go-sentinel run --parallel=4 ./pkg/...
go-sentinel run --parallel=4 ./internal/...
go-sentinel run --parallel=4 ./cmd/...
```

### Microservices

**Scenario**: Testing multiple microservices in a monorepo

**Script (test-services.sh):**
```bash
#!/bin/bash

services=("user-service" "order-service" "payment-service" "notification-service")

for service in "${services[@]}"; do
    echo "Testing $service..."
    go-sentinel run --parallel=2 "./services/$service/..."
    if [ $? -ne 0 ]; then
        echo "Tests failed for $service"
        exit 1
    fi
done

echo "All services tested successfully!"
```

### Benchmark Testing

**Commands:**
```bash
# Run benchmark tests with detailed output
go-sentinel run -v --test="Benchmark*" --timeout=30m ./...

# Performance regression testing
go-sentinel run --parallel=1 --test="Benchmark*" ./performance/...
```

## Team Collaboration

### Onboarding New Developers

**Setup Script (setup-dev.sh):**
```bash
#!/bin/bash

echo "Setting up Go Sentinel CLI for development..."

# Install Go Sentinel CLI
go install github.com/newbpydev/go-sentinel-cli/cmd/go-sentinel-cli@latest

# Copy development configuration
cp docs/examples/development.json sentinel.config.json

echo "Running initial test suite..."
go-sentinel run ./...

echo "Setup complete! Use 'go-sentinel run --watch' to start development."
```

### Shared Configuration

**Team Configuration (team.json):**
```json
{
  "colors": true,
  "verbosity": 1,
  "parallel": 4,
  "timeout": "5m",
  "visual": {
    "colors": true,
    "icons": "unicode",
    "theme": "dark"
  },
  "paths": {
    "includePatterns": ["**/*.go"],
    "excludePatterns": [
      "vendor/**",
      ".git/**",
      "**/*_generated.go",
      "**/mocks/**"
    ]
  },
  "watch": {
    "enabled": false,
    "debounce": "100ms",
    "ignorePatterns": [
      "**/*.log",
      "**/*.tmp",
      "**/coverage.*"
    ],
    "clearOnRerun": true,
    "runOnStart": true
  }
}
```

### Code Quality Gates

**Pre-commit Hook (.git/hooks/pre-commit):**
```bash
#!/bin/sh

echo "Running tests before commit..."
go-sentinel run --no-color --fail-fast ./...

if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

echo "All tests passed. Proceeding with commit."
```

## IDE Integration

### VS Code

**tasks.json:**
```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Go Sentinel: Run Tests",
            "type": "shell",
            "command": "go-sentinel",
            "args": ["run", "./..."],
            "group": "test",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        },
        {
            "label": "Go Sentinel: Watch Mode",
            "type": "shell",
            "command": "go-sentinel",
            "args": ["run", "--watch"],
            "group": "test",
            "isBackground": true,
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "dedicated"
            }
        }
    ]
}
```

**keybindings.json:**
```json
[
    {
        "key": "ctrl+t ctrl+r",
        "command": "workbench.action.tasks.runTask",
        "args": "Go Sentinel: Run Tests"
    },
    {
        "key": "ctrl+t ctrl+w",
        "command": "workbench.action.tasks.runTask",
        "args": "Go Sentinel: Watch Mode"
    }
]
```

### IntelliJ IDEA / GoLand

**External Tool Configuration:**
- **Name**: Go Sentinel Run
- **Program**: `go-sentinel`
- **Arguments**: `run $FilePath$`
- **Working Directory**: `$ProjectFileDir$`

**File Watcher:**
- **File Type**: Go files
- **Scope**: Project files
- **Program**: `go-sentinel`
- **Arguments**: `run --test="Test*" $FileDir$`

## Troubleshooting Scenarios

### Flaky Tests

**Problem**: Tests that sometimes pass, sometimes fail

**Solution:**
```bash
# Run tests multiple times to identify flaky behavior
for i in {1..10}; do
    echo "Run $i:"
    go-sentinel run --test="TestFlakyFunction" ./...
    sleep 1
done

# Use debugging configuration for detailed analysis
cp docs/examples/debugging.json sentinel.config.json
go-sentinel run --watch --test="TestFlakyFunction"
```

### Slow Tests

**Problem**: Test suite takes too long to run

**Analysis:**
```bash
# Run with timing information
go-sentinel run -vv ./...

# Test specific packages to isolate slow areas
go-sentinel run -v ./pkg/slow-package/...

# Use parallel execution to improve performance
go-sentinel run --parallel=8 ./...
```

### Memory Issues

**Problem**: Tests consuming too much memory

**Monitoring:**
```bash
# Run with memory tracking
go-sentinel run -v ./...

# Use debugging config with sequential execution
cp docs/examples/debugging.json sentinel.config.json
go-sentinel run ./...
```

## Advanced Patterns

### Multi-Environment Testing

**Script (test-environments.sh):**
```bash
#!/bin/bash

environments=("development" "staging" "production")

for env in "${environments[@]}"; do
    echo "Testing $env environment..."
    
    # Load environment-specific config
    cp "configs/$env.json" sentinel.config.json
    
    # Run tests with environment tag
    go-sentinel run --test="TestIntegration*" -tags="$env" ./...
done
```

### Conditional Testing

**Based on Changed Files:**
```bash
#!/bin/bash

# Get changed files
changed_files=$(git diff --name-only HEAD~1 HEAD | grep '\.go$')

if [ -n "$changed_files" ]; then
    # Extract unique packages
    packages=$(echo "$changed_files" | xargs -I {} dirname {} | sort -u | tr '\n' ' ')
    
    echo "Running tests for changed packages: $packages"
    go-sentinel run $packages
else
    echo "No Go files changed, skipping tests"
fi
```

### Test Data Management

**With Temporary Directories:**
```bash
# Create isolated test environment
test_dir=$(mktemp -d)
cd "$test_dir"

# Copy test data
cp -r "$original_dir/testdata" .
cp "$original_dir/sentinel.config.json" .

# Run tests in isolation
go-sentinel run ./...

# Cleanup
cd "$original_dir"
rm -rf "$test_dir"
``` 