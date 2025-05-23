# Configuration Examples

This directory contains example configurations for different use cases with Go Sentinel CLI.

## Available Examples

### Development Configuration
- **File**: `development.json`
- **Use Case**: Local development with watch mode and full visual features
- **Features**: Unicode icons, colors, watch mode enabled, low debounce

### CI/CD Configuration  
- **File**: `ci-cd.json`
- **Use Case**: Continuous integration pipelines
- **Features**: No colors, fail-fast, parallel execution, minimal output

### Performance Configuration
- **File**: `performance.json`
- **Use Case**: Performance testing and benchmarking
- **Features**: High parallelism, extended timeouts, detailed timing

### Debugging Configuration
- **File**: `debugging.json`
- **Use Case**: Debugging test failures
- **Features**: Maximum verbosity, no parallel execution, extended timeouts

### Minimal Configuration
- **File**: `minimal.json`
- **Use Case**: Simple setups with basic features
- **Features**: ASCII icons, basic colors, standard settings

## Usage

Copy any example configuration to your project root as `sentinel.config.json`:

```bash
# Development setup
cp docs/examples/development.json sentinel.config.json

# CI/CD setup
cp docs/examples/ci-cd.json sentinel.config.json

# Performance testing
cp docs/examples/performance.json sentinel.config.json
```

## Customization

All configurations can be overridden by CLI arguments:

```bash
# Override config file colors setting
go-sentinel run --no-color

# Override config file watch setting  
go-sentinel run --watch

# Override config file parallel setting
go-sentinel run --parallel=8
``` 