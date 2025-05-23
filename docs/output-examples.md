# Output Examples

Visual examples of Go Sentinel CLI output in different scenarios.

## Successful Test Run

```
🚀 Running tests with go-sentinel...

github.com/newbpydev/go-sentinel/internal/cli (70 tests) 7480ms 0 MB heap used
  ✓ Suite passed (70 tests)

────────────────────────────────────────────────────────────────────────────────
Test Summary:
Test Files: 2 passed (total: 2)
Tests: 127 passed (total: 127)
Start at: 20:32:50
Duration: 2ms
────────────────────────────────────────────────────────────────────────────────

⏱️  Tests completed in 6.1s
```

## Test Run with Failures

```
🚀 Running tests with go-sentinel...

github.com/myproject/pkg/utils (15 tests | 2 failed) 1240ms 2.1 MB heap used
  ✓ TestStringHelper 45ms
  ✗ TestValidation 230ms
  ✓ TestFormatter 12ms
  ✗ TestParser 89ms
  ✓ TestConfig 156ms
  ✓ TestHelpers 67ms
  ✓ TestUtilities 34ms

github.com/myproject/pkg/api (8 tests | 1 failed) 890ms 1.2 MB heap used
  ✓ TestHandlerSuccess 120ms
  ✗ TestHandlerError 340ms
  ✓ TestMiddleware 78ms
  ✓ TestRouter 45ms

────────────────────────────────────────────────────────────────────────────────
                                 Failed Tests 3
────────────────────────────────────────────────────────────────────────────────
FAIL github.com/myproject/pkg/utils > TestValidation

    validation_test.go:25
    Expected validation to pass but got error: invalid input

FAIL github.com/myproject/pkg/utils > TestParser

    parser_test.go:67
    Parse error: unexpected token at line 15

FAIL github.com/myproject/pkg/api > TestHandlerError

    handler_test.go:89
    Expected status code 400, got 500

────────────────────────────────────────────────────────────────────────────────
Test Summary:
Test Files: 1 passed, 1 failed (total: 2)
Tests: 20 passed, 3 failed (total: 23)
Start at: 14:32:15
Duration: 2.1s
────────────────────────────────────────────────────────────────────────────────

⏱️  Tests completed in 2.13s
```

## Watch Mode

```
👀 Starting watch mode...
📁 Watching for changes in: [**/*.go]
🚫 Ignoring: [vendor/** .git/**]
⌨️  Press Ctrl+C to exit

🚀 Running tests with go-sentinel...

github.com/myproject/pkg/utils (5 tests) 450ms 0.8 MB heap used
  ✓ Suite passed (5 tests)

────────────────────────────────────────────────────────────────────────────────
Test Summary:
Test Files: 1 passed (total: 1)
Tests: 5 passed (total: 5)
Start at: 16:45:22
Duration: 450ms
────────────────────────────────────────────────────────────────────────────────

⏱️  Tests completed in 0.45s

📝 File changed: pkg/utils/helper.go
🏃 Running tests: [pkg/utils]

🚀 Running tests with go-sentinel...

github.com/myproject/pkg/utils (5 tests) 380ms 0.8 MB heap used
  ✓ Suite passed (5 tests)

────────────────────────────────────────────────────────────────────────────────
Test Summary:
Test Files: 1 passed (total: 1)
Tests: 5 passed (total: 5)
Start at: 16:45:34
Duration: 380ms
────────────────────────────────────────────────────────────────────────────────

⏱️  Tests completed in 0.38s
```

## Verbose Output

```bash
go-sentinel run -vv ./internal/cli
```

```
🚀 Running tests with go-sentinel...

Parsing CLI arguments: [./internal/cli]
Loading configuration from: sentinel.config.json
Configuration loaded successfully

Executing: go test -json ./internal/cli
Processing test output...

github.com/newbpydev/go-sentinel/internal/cli (70 tests) 7480ms 0 MB heap used
  ✓ TestArgumentParser_Parse 45ms
  ✓ TestArgumentParser_ParseVerbosity 23ms
  ✓ TestArgumentParser_ParseWatch 12ms
  ✓ TestArgumentParser_ParseColor 8ms
  ✓ TestArgumentParser_ParsePackages 34ms
  ✓ TestConfigLoader_LoadFromFile 67ms
  ✓ TestConfigLoader_ParseConfigData 89ms
  ✓ TestConfigLoader_ValidateConfig 45ms
  ... (62 more tests)

Test execution completed
Rendering results...

────────────────────────────────────────────────────────────────────────────────
Test Summary:
Test Files: 2 passed (total: 2)
Tests: 127 passed (total: 127)
Start at: 20:32:50
Duration: 7.48s (parsing: 150ms, execution: 7.3s, rendering: 30ms)
────────────────────────────────────────────────────────────────────────────────

⏱️  Tests completed in 7.48s
```

## Parallel Execution

```bash
go-sentinel run --parallel=4 ./...
```

```
🚀 Running tests with go-sentinel...

Running 4 packages in parallel...

github.com/myproject/pkg/utils (15 tests) 890ms 1.2 MB heap used
  ✓ Suite passed (15 tests)

github.com/myproject/pkg/api (12 tests) 1240ms 2.1 MB heap used
  ✓ Suite passed (12 tests)

github.com/myproject/internal/config (8 tests) 567ms 0.9 MB heap used
  ✓ Suite passed (8 tests)

github.com/myproject/cmd/server (6 tests) 743ms 1.4 MB heap used
  ✓ Suite passed (6 tests)

────────────────────────────────────────────────────────────────────────────────
Test Summary:
Test Files: 4 passed (total: 4)
Tests: 41 passed (total: 41)
Start at: 15:23:45
Duration: 1.24s (parallel execution saved ~2.8s)
────────────────────────────────────────────────────────────────────────────────

⏱️  Tests completed in 1.24s
```

## Test Pattern Filtering

```bash
go-sentinel run --test="TestConfig*" ./...
```

```
🚀 Running tests with go-sentinel...

Filtering tests with pattern: TestConfig*

github.com/myproject/internal/config (3 tests) 234ms 0.4 MB heap used
  ✓ TestConfigLoad 89ms
  ✓ TestConfigValidation 78ms  
  ✓ TestConfigDefaults 67ms

github.com/myproject/pkg/utils (1 tests) 45ms 0.1 MB heap used
  ✓ TestConfigHelper 45ms

────────────────────────────────────────────────────────────────────────────────
Test Summary:
Test Files: 2 passed (total: 2)
Tests: 4 passed (total: 4)
Start at: 17:12:33
Duration: 279ms
────────────────────────────────────────────────────────────────────────────────

⏱️  Tests completed in 0.28s
```

## No Color Output (CI/CD)

```bash
go-sentinel run --no-color --fail-fast ./...
```

```
Running tests with go-sentinel...

github.com/myproject/pkg/utils (15 tests | 1 failed) 1240ms 2.1 MB heap used
  + TestStringHelper 45ms
  - TestValidation 230ms
  + TestFormatter 12ms
  + TestParser 89ms

Failed Tests 1

FAIL github.com/myproject/pkg/utils > TestValidation

    validation_test.go:25
    Expected validation to pass but got error: invalid input

Test Summary:
Test Files: 0 passed, 1 failed (total: 1)
Tests: 14 passed, 1 failed (total: 15)
Start at: 09:15:22
Duration: 1.24s

Tests completed in 1.24s
```

## Help Output

```bash
go-sentinel run --help
```

```
Run Go tests with beautiful, Vitest-style output.
If no packages are specified, runs tests in the current directory and subdirectories.

Usage:
  go-sentinel run [flags] [packages]

Flags:
  -c, --color              Use colored output (default true)
  -f, --fail-fast          Stop on first failure
  -h, --help               help for run
      --no-color           Disable colored output
  -j, --parallel int       Number of tests to run in parallel
  -t, --test string        Run only tests matching pattern
      --timeout duration   Timeout for test execution
  -v, --verbose            Enable verbose output
  -q, --verbosity count    Verbosity level (can be repeated: -v, -vv, -vvv)
  -w, --watch              Watch for file changes and re-run tests

Global Flags:
  -c, --color   Enable/disable colored output
  -w, --watch   Enable watch mode

Examples:
  # Run tests with beautiful output
  go-sentinel run

  # Run in watch mode with verbose output
  go-sentinel run -w -v

  # Run specific package with test filtering
  go-sentinel run --test="TestHandler*" ./api

  # High performance execution
  go-sentinel run --parallel=8 --timeout=5m ./...
```

## Demo Output

```bash
go-sentinel demo --phase=7
```

```
🚀 Phase 7-D: CLI Options & Configuration Demonstration
──────────────────────────────────────────────────────────

📋 CLI Arguments Parsing
Testing various CLI argument combinations and validation

  ⏱️ Basic Watch Mode
    Args: [-w ./internal]
    Desc: Simple watch mode with package
    ✓ Parsed successfully
      → Watch: true, Colors: true, Verbosity: 0

  ⏱️ Verbose Testing
    Args: [-vvv --test=TestUnit* ./...]
    Desc: High verbosity with test pattern
    ✓ Parsed successfully
      → Verbosity: 3, TestPattern: TestUnit*, Packages: [./...]

  ⏱️ Performance Mode
    Args: [--parallel=8 --timeout=5m --color]
    Desc: High performance with extended timeout
    ✓ Parsed successfully
      → Parallel: 8, Timeout: 5m0s, Colors: true

📄 Configuration File Management
Testing configuration loading and validation

  📂 Development Configuration
    ✓ Loaded from: development.json
      → Icons: unicode, Watch: enabled, Debounce: 50ms

  🏭 CI/CD Configuration  
    ✓ Loaded from: ci-cd.json
      → Colors: disabled, Parallel: 4, Icons: minimal

  ⚡ Performance Configuration
    ✓ Loaded from: performance.json
      → Parallel: 8, Timeout: 10m, Verbosity: 2

🎛️ CLI Options Variations
Demonstrating different CLI argument combinations

  🎨 Color Options
    ✓ --color: Colors enabled
    ✓ --no-color: Colors disabled

  📊 Verbosity Levels
    ✓ Default: Level 0
    ✓ -v: Level 1
    ✓ -vv: Level 2  
    ✓ -vvv: Level 3

🔄 Configuration Precedence & Merging
Testing CLI arguments override configuration files

  📝 Base Config: {"colors": false, "verbosity": 1}
  🎛️ CLI Args: [--color -vv]
  ✅ Final Result: {"colors": true, "verbosity": 2}
    → CLI arguments successfully override configuration

✅ Phase 7-D demonstration completed successfully!
   All CLI options and configuration features working correctly.
```

## Icon Styles Comparison

### Unicode Icons (Default)
```
github.com/myproject/pkg/utils (15 tests | 2 failed) 1240ms 2.1 MB heap used
  ✓ TestStringHelper 45ms
  ✗ TestValidation 230ms
  ✓ TestFormatter 12ms
```

### ASCII Icons
```
github.com/myproject/pkg/utils (15 tests | 2 failed) 1240ms 2.1 MB heap used
  + TestStringHelper 45ms
  - TestValidation 230ms
  + TestFormatter 12ms
```

### Minimal Icons
```
github.com/myproject/pkg/utils (15 tests | 2 failed) 1240ms 2.1 MB heap used
  > TestStringHelper 45ms
  x TestValidation 230ms
  > TestFormatter 12ms
```

### No Icons
```
github.com/myproject/pkg/utils (15 tests | 2 failed) 1240ms 2.1 MB heap used
  TestStringHelper 45ms
  TestValidation 230ms
  TestFormatter 12ms
``` 