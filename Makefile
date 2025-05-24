# Go Sentinel CLI - Makefile
# Provides common development tasks for building, testing, and maintaining the CLI

# Variables
BINARY_NAME := go-sentinel-cli
BINARY_NAME_V2 := go-sentinel-cli-v2
BUILD_DIR := build
COVERAGE_DIR := coverage
VERSION := $(shell git describe --tags --dirty --always 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S%z)

# Go related variables
GO_VERSION := 1.23
GOPATH := $(shell go env GOPATH)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Build flags
LDFLAGS := -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME} -s -w"

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
MAGENTA := \033[0;35m
CYAN := \033[0;36m
WHITE := \033[0;37m
RESET := \033[0m

# Quality thresholds
COVERAGE_THRESHOLD := 90
LINT_TIMEOUT := 5m

.PHONY: help
help: ## Show this help message
	@echo "$(CYAN)Go Sentinel CLI - Development Commands$(RESET)"
	@echo ""
	@echo "$(YELLOW)Usage:$(RESET)"
	@echo "  make [target]"
	@echo ""
	@echo "$(YELLOW)Available targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'

# ==============================================================================
# Build Commands
# ==============================================================================

.PHONY: build
build: clean ## Build the main CLI binary
	@echo "$(BLUE)Building $(BINARY_NAME)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/go-sentinel-cli
	@echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

.PHONY: build-v2
build-v2: clean ## Build the v2 CLI binary
	@echo "$(BLUE)Building $(BINARY_NAME_V2)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_V2) ./cmd/go-sentinel-cli-v2
	@echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME_V2)$(RESET)"

.PHONY: build-all
build-all: build build-v2 ## Build both CLI binaries

.PHONY: install
install: ## Install the CLI binary to GOPATH/bin
	@echo "$(BLUE)Installing $(BINARY_NAME)...$(RESET)"
	go install $(LDFLAGS) ./cmd/go-sentinel-cli
	@echo "$(GREEN)✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)$(RESET)"

.PHONY: cross-compile
cross-compile: clean ## Cross-compile for multiple platforms
	@echo "$(BLUE)Cross-compiling for multiple platforms...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/go-sentinel-cli
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/go-sentinel-cli
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/go-sentinel-cli
	# macOS ARM64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/go-sentinel-cli
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/go-sentinel-cli
	@echo "$(GREEN)✓ Cross-compilation complete$(RESET)"
	@ls -la $(BUILD_DIR)/

# ==============================================================================
# Test Commands
# ==============================================================================

.PHONY: test
test: ## Run all tests
	@echo "$(BLUE)Running tests...$(RESET)"
	go test -race -v ./...
	@echo "$(GREEN)✓ All tests passed$(RESET)"

.PHONY: test-short
test-short: ## Run tests with short flag
	@echo "$(BLUE)Running short tests...$(RESET)"
	go test -short -race -v ./...
	@echo "$(GREEN)✓ Short tests passed$(RESET)"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(RESET)"
	go test -tags=integration -race -v ./...
	@echo "$(GREEN)✓ Integration tests passed$(RESET)"

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "$(BLUE)Running tests with coverage...$(RESET)"
	@mkdir -p $(COVERAGE_DIR)
	go test -race -covermode=atomic -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	go tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1
	@echo "$(GREEN)✓ Coverage report generated: $(COVERAGE_DIR)/coverage.html$(RESET)"

.PHONY: test-coverage-ci
test-coverage-ci: ## Run tests with coverage for CI (no HTML)
	@echo "$(BLUE)Running tests with coverage for CI...$(RESET)"
	@mkdir -p $(COVERAGE_DIR)
	go test -race -covermode=atomic -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@echo "$(GREEN)✓ Coverage data generated: $(COVERAGE_DIR)/coverage.out$(RESET)"

# ==============================================================================
# Benchmark Commands
# ==============================================================================

.PHONY: benchmark
benchmark: ## Run all benchmarks
	@echo "$(BLUE)Running benchmarks...$(RESET)"
	go test -bench=. -benchmem -run=^$$ ./internal/cli
	@echo "$(GREEN)✓ Benchmarks complete$(RESET)"

.PHONY: benchmark-short
benchmark-short: ## Run short benchmarks (100ms each)
	@echo "$(BLUE)Running short benchmarks...$(RESET)"
	go test -bench=. -benchmem -benchtime=100ms -run=^$$ ./internal/cli
	@echo "$(GREEN)✓ Short benchmarks complete$(RESET)"

.PHONY: benchmark-filesystem
benchmark-filesystem: ## Run file system operation benchmarks
	@echo "$(BLUE)Running file system benchmarks...$(RESET)"
	go test -bench=BenchmarkFile -benchmem -run=^$$ ./internal/cli
	go test -bench=BenchmarkPattern -benchmem -run=^$$ ./internal/cli
	go test -bench=BenchmarkDirectory -benchmem -run=^$$ ./internal/cli
	@echo "$(GREEN)✓ File system benchmarks complete$(RESET)"

.PHONY: benchmark-execution
benchmark-execution: ## Run test execution pipeline benchmarks
	@echo "$(BLUE)Running execution benchmarks...$(RESET)"
	go test -bench=BenchmarkTest -benchmem -run=^$$ ./internal/cli
	go test -bench=BenchmarkOptimized -benchmem -run=^$$ ./internal/cli
	go test -bench=BenchmarkParallel -benchmem -run=^$$ ./internal/cli
	@echo "$(GREEN)✓ Execution benchmarks complete$(RESET)"

.PHONY: benchmark-rendering
benchmark-rendering: ## Run rendering and output benchmarks
	@echo "$(BLUE)Running rendering benchmarks...$(RESET)"
	go test -bench=BenchmarkColor -benchmem -run=^$$ ./internal/cli
	go test -bench=BenchmarkIcon -benchmem -run=^$$ ./internal/cli
	go test -bench=BenchmarkSuite -benchmem -run=^$$ ./internal/cli
	go test -bench=BenchmarkIncremental -benchmem -run=^$$ ./internal/cli
	@echo "$(GREEN)✓ Rendering benchmarks complete$(RESET)"

.PHONY: benchmark-integration
benchmark-integration: ## Run integration and end-to-end benchmarks
	@echo "$(BLUE)Running integration benchmarks...$(RESET)"
	go test -bench=BenchmarkEndToEnd -benchmem -run=^$$ ./internal/cli
	go test -bench=BenchmarkWatch -benchmem -run=^$$ ./internal/cli
	go test -bench=BenchmarkConcurrent -benchmem -run=^$$ ./internal/cli
	@echo "$(GREEN)✓ Integration benchmarks complete$(RESET)"

.PHONY: benchmark-memory
benchmark-memory: ## Run memory-intensive benchmarks
	@echo "$(BLUE)Running memory benchmarks...$(RESET)"
	go test -bench=BenchmarkMemory -benchmem -run=^$$ ./internal/cli
	go test -bench=BenchmarkCache -benchmem -run=^$$ ./internal/cli
	@echo "$(GREEN)✓ Memory benchmarks complete$(RESET)"

.PHONY: benchmark-compare
benchmark-compare: ## Run benchmarks and save results for comparison
	@echo "$(BLUE)Running benchmarks for comparison...$(RESET)"
	@mkdir -p $(BUILD_DIR)/benchmarks
	go test -bench=. -benchmem -run=^$$ ./internal/cli > $(BUILD_DIR)/benchmarks/current.txt
	@echo "$(GREEN)✓ Benchmark results saved to $(BUILD_DIR)/benchmarks/current.txt$(RESET)"

.PHONY: benchmark-profile
benchmark-profile: ## Run benchmarks with CPU profiling
	@echo "$(BLUE)Running benchmarks with CPU profiling...$(RESET)"
	@mkdir -p $(BUILD_DIR)/profiles
	go test -bench=BenchmarkTestProcessor -benchmem -cpuprofile=$(BUILD_DIR)/profiles/cpu.prof -run=^$$ ./internal/cli
	@echo "$(GREEN)✓ CPU profile saved to $(BUILD_DIR)/profiles/cpu.prof$(RESET)"
	@echo "$(CYAN)View profile with: go tool pprof $(BUILD_DIR)/profiles/cpu.prof$(RESET)"

.PHONY: benchmark-memprofile
benchmark-memprofile: ## Run benchmarks with memory profiling
	@echo "$(BLUE)Running benchmarks with memory profiling...$(RESET)"
	@mkdir -p $(BUILD_DIR)/profiles
	go test -bench=BenchmarkMemoryAllocation -benchmem -memprofile=$(BUILD_DIR)/profiles/mem.prof -run=^$$ ./internal/cli
	@echo "$(GREEN)✓ Memory profile saved to $(BUILD_DIR)/profiles/mem.prof$(RESET)"
	@echo "$(CYAN)View profile with: go tool pprof $(BUILD_DIR)/profiles/mem.prof$(RESET)"

.PHONY: benchmark-regression
benchmark-regression: ## Check for performance regressions
	@echo "$(BLUE)Checking for performance regressions...$(RESET)"
	@if [ -f "$(BUILD_DIR)/benchmarks/baseline.txt" ]; then \
		go test -bench=. -benchmem -run=^$$ ./internal/cli > $(BUILD_DIR)/benchmarks/current.txt; \
		echo "$(CYAN)Comparing with baseline...$(RESET)"; \
		if command -v benchcmp >/dev/null 2>&1; then \
			benchcmp $(BUILD_DIR)/benchmarks/baseline.txt $(BUILD_DIR)/benchmarks/current.txt; \
		else \
			echo "$(YELLOW)benchcmp not installed. Install with: go install golang.org/x/tools/cmd/benchcmp@latest$(RESET)"; \
			diff $(BUILD_DIR)/benchmarks/baseline.txt $(BUILD_DIR)/benchmarks/current.txt || true; \
		fi; \
	else \
		echo "$(YELLOW)No baseline found. Creating baseline...$(RESET)"; \
		go test -bench=. -benchmem -run=^$$ ./internal/cli > $(BUILD_DIR)/benchmarks/baseline.txt; \
		echo "$(GREEN)✓ Baseline created$(RESET)"; \
	fi

# ==============================================================================
# Code Quality Commands
# ==============================================================================

.PHONY: lint
lint: ## Run golangci-lint
	@echo "$(BLUE)Running linter...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --config .golangci.yml ./...; \
		echo "$(GREEN)✓ Linting complete$(RESET)"; \
	else \
		echo "$(RED)✗ golangci-lint not installed. Run: make install-tools$(RESET)"; \
		exit 1; \
	fi

.PHONY: fmt
fmt: ## Format Go code
	@echo "$(BLUE)Formatting code...$(RESET)"
	go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(RESET)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(RESET)"
	go vet ./...
	@echo "$(GREEN)✓ Vet complete$(RESET)"

.PHONY: mod-tidy
mod-tidy: ## Tidy Go modules
	@echo "$(BLUE)Tidying modules...$(RESET)"
	go mod tidy
	go mod verify
	@echo "$(GREEN)✓ Modules tidied$(RESET)"

.PHONY: check
check: fmt vet lint test ## Run all code quality checks

# ==============================================================================
# Development Commands
# ==============================================================================

.PHONY: dev
dev: build ## Build and run the CLI in development mode
	@echo "$(BLUE)Running CLI in development mode...$(RESET)"
	./$(BUILD_DIR)/$(BINARY_NAME) run --color -v

.PHONY: dev-watch
dev-watch: build ## Build and run the CLI in watch mode
	@echo "$(BLUE)Running CLI in watch mode...$(RESET)"
	./$(BUILD_DIR)/$(BINARY_NAME) run --watch --color -v

.PHONY: demo
demo: build ## Run CLI demo phases
	@echo "$(BLUE)Running CLI demos...$(RESET)"
	@for phase in 1 2 3 4 5 6 7; do \
		echo "$(CYAN)Running demo phase $$phase...$(RESET)"; \
		./$(BUILD_DIR)/$(BINARY_NAME) demo --phase=$$phase; \
		echo ""; \
	done

# ==============================================================================
# Tool Installation
# ==============================================================================

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(RESET)"
	# Install golangci-lint
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2; \
	fi
	# Install goimports
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	# Install gotestsum for better test output
	@if ! command -v gotestsum >/dev/null 2>&1; then \
		echo "Installing gotestsum..."; \
		go install gotest.tools/gotestsum@latest; \
	fi
	@echo "$(GREEN)✓ Development tools installed$(RESET)"

# ==============================================================================
# Cleanup Commands
# ==============================================================================

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f *.exe
	@echo "$(GREEN)✓ Clean complete$(RESET)"

.PHONY: clean-deps
clean-deps: ## Clean dependency cache
	@echo "$(BLUE)Cleaning dependency cache...$(RESET)"
	go clean -modcache
	@echo "$(GREEN)✓ Dependency cache cleaned$(RESET)"

.PHONY: clean-all
clean-all: clean clean-deps ## Clean everything

# ==============================================================================
# Release Commands
# ==============================================================================

.PHONY: version
version: ## Show version information
	@echo "$(CYAN)Version Information:$(RESET)"
	@echo "  Version: $(VERSION)"
	@echo "  Commit:  $(COMMIT)"
	@echo "  Build:   $(BUILD_TIME)"
	@echo "  Go:      $(shell go version)"

.PHONY: tag
tag: ## Create a new version tag (requires VERSION=x.y.z)
ifndef VERSION
	@echo "$(RED)ERROR: VERSION is required. Usage: make tag VERSION=1.0.0$(RESET)"
	@exit 1
endif
	@echo "$(BLUE)Creating tag $(VERSION)...$(RESET)"
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)
	@echo "$(GREEN)✓ Tag v$(VERSION) created and pushed$(RESET)"

.PHONY: release
release: clean cross-compile test-coverage lint ## Prepare a release build
	@echo "$(GREEN)✓ Release build complete$(RESET)"
	@echo "$(CYAN)Release artifacts:$(RESET)"
	@ls -la $(BUILD_DIR)/

# ==============================================================================
# Docker Commands (Future)
# ==============================================================================

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(RESET)"
	docker build -t go-sentinel-cli:$(VERSION) .
	docker build -t go-sentinel-cli:latest .
	@echo "$(GREEN)✓ Docker image built$(RESET)"

# ==============================================================================
# CI/CD Commands
# ==============================================================================

.PHONY: ci
ci: mod-tidy fmt vet lint test-coverage-ci ## Run CI pipeline
	@echo "$(GREEN)✓ CI pipeline complete$(RESET)"

.PHONY: local-ci
local-ci: install-tools ci ## Run full CI pipeline locally
	@echo "$(GREEN)✓ Local CI pipeline complete$(RESET)"

# ==============================================================================
# Information Commands
# ==============================================================================

.PHONY: info
info: ## Show project information
	@echo "$(CYAN)Go Sentinel CLI Project Information:$(RESET)"
	@echo "  Binary Name:    $(BINARY_NAME)"
	@echo "  Version:        $(VERSION)"
	@echo "  Commit:         $(COMMIT)"
	@echo "  Build Time:     $(BUILD_TIME)"
	@echo "  Go Version:     $(shell go version)"
	@echo "  GOOS:           $(GOOS)"
	@echo "  GOARCH:         $(GOARCH)"
	@echo "  GOPATH:         $(GOPATH)"
	@echo ""
	@echo "$(CYAN)Project Structure:$(RESET)"
	@echo "  Source:         ./cmd/go-sentinel-cli/"
	@echo "  Internal:       ./internal/cli/"
	@echo "  Build Dir:      $(BUILD_DIR)/"
	@echo "  Coverage Dir:   $(COVERAGE_DIR)/"

.PHONY: deps
deps: ## Show dependency information
	@echo "$(CYAN)Go Dependencies:$(RESET)"
	go list -m all

# ==============================================================================
# Quality Gate Commands
# ==============================================================================

.PHONY: quality-gate
quality-gate: ## Run complete quality gate pipeline
	@echo "$(CYAN)Running Quality Gate Pipeline...$(RESET)"
	@./scripts/quality-gate.sh

.PHONY: quality-gate-setup
quality-gate-setup: ## Set up quality gate dependencies
	@echo "$(CYAN)Setting up quality gate dependencies...$(RESET)"
	@mkdir -p scripts build coverage
	@chmod +x scripts/quality-gate.sh
	@echo "$(GREEN)✓ Quality gate setup complete$(RESET)"

.PHONY: test-coverage-check
test-coverage-check: ## Run tests and enforce coverage threshold
	@echo "$(BLUE)Running tests with coverage checking...$(RESET)"
	@mkdir -p $(COVERAGE_DIR)
	go test -race -covermode=atomic -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@echo "$(BLUE)Checking coverage threshold ($(COVERAGE_THRESHOLD)%)...$(RESET)"
	@COVERAGE=$$(go tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	if [ "$$(echo "$$COVERAGE < $(COVERAGE_THRESHOLD)" | bc -l)" -eq 1 ]; then \
		echo "$(RED)❌ Coverage $$COVERAGE% is below threshold $(COVERAGE_THRESHOLD)%$(RESET)"; \
		exit 1; \
	else \
		echo "$(GREEN)✅ Coverage $$COVERAGE% meets threshold $(COVERAGE_THRESHOLD)%$(RESET)"; \
	fi
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)✓ Coverage report: $(COVERAGE_DIR)/coverage.html$(RESET)"

.PHONY: security-scan
security-scan: ## Run security vulnerability scan
	@echo "$(BLUE)Running security scan...$(RESET)"
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	gosec -fmt json -out $(BUILD_DIR)/gosec-report.json ./...
	@echo "$(GREEN)✓ Security scan complete: $(BUILD_DIR)/gosec-report.json$(RESET)"

.PHONY: lint-fix
lint-fix: ## Auto-fix linting issues where possible
	@echo "$(BLUE)Auto-fixing linting issues...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix --config .golangci.yml ./...; \
		echo "$(GREEN)✓ Auto-fix complete$(RESET)"; \
	else \
		echo "$(RED)✗ golangci-lint not installed. Run: make install-tools$(RESET)"; \
		exit 1; \
	fi

.PHONY: pre-commit
pre-commit: fmt vet lint-fix test-short ## Run pre-commit checks
	@echo "$(GREEN)✓ Pre-commit checks passed$(RESET)"

.PHONY: ci-local
ci-local: quality-gate ## Run full CI pipeline locally
	@echo "$(GREEN)✓ Local CI pipeline complete$(RESET)"

# Default target
.DEFAULT_GOAL := help 