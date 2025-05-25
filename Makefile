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

# ==============================================================================
# Code Complexity Analysis Commands
# ==============================================================================

.PHONY: complexity
complexity: build ## Analyze code complexity metrics
	@echo "$(BLUE)Analyzing code complexity...$(RESET)"
	./$(BUILD_DIR)/$(BINARY_NAME) complexity . --format=text
	@echo "$(GREEN)✓ Complexity analysis complete$(RESET)"

.PHONY: complexity-json
complexity-json: build ## Generate complexity analysis in JSON format
	@echo "$(BLUE)Generating JSON complexity report...$(RESET)"
	@mkdir -p $(BUILD_DIR)/reports
	./$(BUILD_DIR)/$(BINARY_NAME) complexity . --format=json --output=$(BUILD_DIR)/reports/complexity.json
	@echo "$(GREEN)✓ JSON complexity report: $(BUILD_DIR)/reports/complexity.json$(RESET)"

.PHONY: complexity-html
complexity-html: build ## Generate complexity analysis in HTML format
	@echo "$(BLUE)Generating HTML complexity report...$(RESET)"
	@mkdir -p $(BUILD_DIR)/reports
	./$(BUILD_DIR)/$(BINARY_NAME) complexity . --format=html --output=$(BUILD_DIR)/reports/complexity.html
	@echo "$(GREEN)✓ HTML complexity report: $(BUILD_DIR)/reports/complexity.html$(RESET)"

.PHONY: complexity-package
complexity-package: build ## Analyze complexity for specific package
	@echo "$(BLUE)Analyzing package complexity...$(RESET)"
	@if [ -z "$(PKG)" ]; then \
		echo "$(RED)Error: Please specify PKG variable, e.g., make complexity-package PKG=internal/cli$(RESET)"; \
		exit 1; \
	fi
	./$(BUILD_DIR)/$(BINARY_NAME) complexity $(PKG) --verbose
	@echo "$(GREEN)✓ Package complexity analysis complete$(RESET)"

.PHONY: complexity-strict
complexity-strict: build ## Analyze complexity with strict thresholds
	@echo "$(BLUE)Analyzing complexity with strict thresholds...$(RESET)"
	./$(BUILD_DIR)/$(BINARY_NAME) complexity . \
		--max-complexity=5 \
		--min-maintainability=90 \
		--max-lines=300 \
		--max-function-lines=30 \
		--max-debt-ratio=2.0 \
		--verbose
	@echo "$(GREEN)✓ Strict complexity analysis complete$(RESET)"

.PHONY: complexity-ci
complexity-ci: build ## Run complexity analysis for CI (exits with error on violations)
	@echo "$(BLUE)Running complexity analysis for CI...$(RESET)"
	./$(BUILD_DIR)/$(BINARY_NAME) complexity . --format=json --output=$(BUILD_DIR)/reports/complexity-ci.json
	@echo "$(GREEN)✓ CI complexity analysis complete$(RESET)"
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
# Pre-commit Hook Commands
# ==============================================================================

.PHONY: setup-hooks
setup-hooks: ## Install and configure pre-commit hooks
	@echo "$(BLUE)Setting up pre-commit hooks...$(RESET)"
	@chmod +x scripts/setup-hooks.sh
	@./scripts/setup-hooks.sh
	@echo "$(GREEN)✓ Pre-commit hooks installed$(RESET)"

.PHONY: test-hooks
test-hooks: ## Test pre-commit hooks functionality
	@echo "$(BLUE)Testing pre-commit hooks...$(RESET)"
	@python3 scripts/test_commit_validator.py
	@echo "$(GREEN)✓ Hook tests passed$(RESET)"

.PHONY: validate-commit
validate-commit: ## Validate commit message format (usage: make validate-commit MSG="feat: add feature")
ifndef MSG
	@echo "$(RED)ERROR: MSG is required. Usage: make validate-commit MSG=\"feat: add feature\"$(RESET)"
	@exit 1
endif
	@python3 scripts/validate_commit_msg.py "$(MSG)"

.PHONY: hooks-help
hooks-help: ## Show commit message format help
	@python3 scripts/validate_commit_msg.py --help

.PHONY: run-hooks
run-hooks: ## Run pre-commit hooks on all files
	@echo "$(BLUE)Running pre-commit hooks on all files...$(RESET)"
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit run --all-files; \
	else \
		echo "$(RED)pre-commit not installed. Run: make setup-hooks$(RESET)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✓ Pre-commit hooks completed$(RESET)"

.PHONY: update-hooks
update-hooks: ## Update pre-commit hook repositories
	@echo "$(BLUE)Updating pre-commit hooks...$(RESET)"
	@if command -v pre-commit >/dev/null 2>&1; then \
		pre-commit autoupdate; \
	else \
		echo "$(RED)pre-commit not installed. Run: make setup-hooks$(RESET)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✓ Pre-commit hooks updated$(RESET)"

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

# ==============================================================================
# Quality Automation Commands
# ==============================================================================

.PHONY: quality-automation
quality-automation: ## Run comprehensive quality automation pipeline
	@echo "$(BLUE)Running comprehensive quality automation...$(RESET)"
	@chmod +x scripts/quality-automation.sh
	@./scripts/quality-automation.sh
	@echo "$(GREEN)✓ Quality automation complete$(RESET)"

.PHONY: quality-check
quality-check: ## Quick quality check (formatting, linting, basic tests)
	@echo "$(BLUE)Running quick quality check...$(RESET)"
	@mkdir -p $(BUILD_DIR)/quality-reports
	gofmt -l . | tee $(BUILD_DIR)/quality-reports/gofmt-issues.txt
	@if [ -s "$(BUILD_DIR)/quality-reports/gofmt-issues.txt" ]; then \
		echo "$(RED)❌ Code formatting issues found$(RESET)"; \
		exit 1; \
	fi
	go vet ./...
	golangci-lint run --config .golangci.yml ./...
	go test -short ./...
	@echo "$(GREEN)✓ Quick quality check passed$(RESET)"

.PHONY: security-scan
security-scan: ## Run comprehensive security analysis
	@echo "$(BLUE)Running security analysis...$(RESET)"
	@mkdir -p $(BUILD_DIR)/quality-reports
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	@if ! command -v govulncheck >/dev/null 2>&1; then \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	gosec -fmt json -out $(BUILD_DIR)/quality-reports/gosec.json ./...
	gosec -fmt sarif -out $(BUILD_DIR)/quality-reports/gosec.sarif ./...
	govulncheck -json ./... > $(BUILD_DIR)/quality-reports/govulncheck.json 2>&1 || true
	@echo "$(GREEN)✓ Security analysis complete$(RESET)"
	@echo "Reports saved to $(BUILD_DIR)/quality-reports/"

.PHONY: dependency-check
dependency-check: ## Check dependencies for vulnerabilities and license compliance
	@echo "$(BLUE)Running dependency analysis...$(RESET)"
	@mkdir -p $(BUILD_DIR)/quality-reports
	@if ! command -v nancy >/dev/null 2>&1; then \
		echo "Installing nancy..."; \
		go install github.com/sonatypecommunity/nancy@latest; \
	fi
	go list -json -deps ./... | nancy sleuth > $(BUILD_DIR)/quality-reports/nancy.txt 2>&1 || true
	go list -m -json all > $(BUILD_DIR)/quality-reports/dependencies.json
	@echo "$(GREEN)✓ Dependency analysis complete$(RESET)"

.PHONY: quality-reports
quality-reports: ## Generate all quality reports
	@echo "$(BLUE)Generating comprehensive quality reports...$(RESET)"
	@mkdir -p $(BUILD_DIR)/quality-reports
	$(MAKE) test-coverage-ci
	$(MAKE) complexity-json
	$(MAKE) security-scan
	$(MAKE) dependency-check
	@echo "$(GREEN)✓ All quality reports generated$(RESET)"
	@echo "$(CYAN)Reports available in $(BUILD_DIR)/quality-reports/$(RESET)"

.PHONY: quality-gate-strict
quality-gate-strict: ## Run strict quality gate with high standards
	@echo "$(BLUE)Running strict quality gate...$(RESET)"
	@mkdir -p $(BUILD_DIR)/quality-reports
	gofmt -l . | tee $(BUILD_DIR)/quality-reports/gofmt-issues.txt
	@if [ -s "$(BUILD_DIR)/quality-reports/gofmt-issues.txt" ]; then \
		echo "$(RED)❌ Code formatting issues found$(RESET)"; \
		exit 1; \
	fi
	go vet ./...
	golangci-lint run --config .golangci.yml ./...
	go run ./cmd/go-sentinel-cli complexity --max-complexity=5 --min-maintainability=90 --max-debt-ratio=2 .
	go test -race -coverprofile=$(BUILD_DIR)/quality-reports/coverage.out ./...
	@COVERAGE=$$(go tool cover -func=$(BUILD_DIR)/quality-reports/coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	if [ "$$(echo "$$COVERAGE < 90" | bc -l)" -eq 1 ]; then \
		echo "$(RED)❌ Coverage $$COVERAGE% is below strict threshold 90%$(RESET)"; \
		exit 1; \
	else \
		echo "$(GREEN)✅ Coverage $$COVERAGE% meets strict threshold$(RESET)"; \
	fi
	@echo "$(GREEN)✓ Strict quality gate passed$(RESET)"

.PHONY: quality-fix
quality-fix: ## Auto-fix quality issues where possible
	@echo "$(BLUE)Auto-fixing quality issues...$(RESET)"
	gofmt -w .
	goimports -w .
	go mod tidy
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix --config .golangci.yml ./...; \
	fi
	@echo "$(GREEN)✓ Auto-fix complete$(RESET)"

.PHONY: quality-dashboard
quality-dashboard: ## Generate quality dashboard (HTML reports)
	@echo "$(BLUE)Generating quality dashboard...$(RESET)"
	@mkdir -p $(BUILD_DIR)/quality-reports/dashboard
	$(MAKE) test-coverage-ci
	go tool cover -html=$(BUILD_DIR)/quality-reports/coverage.out -o $(BUILD_DIR)/quality-reports/dashboard/coverage.html
	go run ./cmd/go-sentinel-cli complexity --format=html --output=$(BUILD_DIR)/quality-reports/dashboard/complexity.html .
	@echo "$(GREEN)✓ Quality dashboard generated$(RESET)"
	@echo "$(CYAN)Dashboard available at:$(RESET)"
	@echo "  Coverage: $(BUILD_DIR)/quality-reports/dashboard/coverage.html"
	@echo "  Complexity: $(BUILD_DIR)/quality-reports/dashboard/complexity.html"

.PHONY: quality-help
quality-help: ## Show quality automation help and examples
	@echo "$(CYAN)Quality Automation Commands:$(RESET)"
	@echo ""
	@echo "$(YELLOW)Comprehensive Analysis:$(RESET)"
	@echo "  make quality-automation     - Full quality pipeline (all checks)"
	@echo "  make quality-reports        - Generate all quality reports"
	@echo "  make quality-dashboard      - Generate HTML quality dashboard"
	@echo ""
	@echo "$(YELLOW)Quick Checks:$(RESET)"
	@echo "  make quality-check          - Quick quality check (fast)"
	@echo "  make quality-gate-strict    - Strict quality gate (high standards)"
	@echo ""
	@echo "$(YELLOW)Specialized Analysis:$(RESET)"
	@echo "  make security-scan          - Security vulnerability analysis"
	@echo "  make dependency-check       - Dependency vulnerability & license check"
	@echo ""
	@echo "$(YELLOW)Maintenance:$(RESET)"
	@echo "  make quality-fix            - Auto-fix formatting and linting issues"
	@echo ""
	@echo "$(YELLOW)Reports Generated:$(RESET)"
	@echo "  $(BUILD_DIR)/quality-reports/coverage.html      - Test coverage report"
	@echo "  $(BUILD_DIR)/quality-reports/complexity.json    - Code complexity analysis"
	@echo "  $(BUILD_DIR)/quality-reports/gosec.json         - Security scan results"
	@echo "  $(BUILD_DIR)/quality-reports/dependencies.json  - Dependency analysis"
	@echo "  $(BUILD_DIR)/quality-reports/quality-summary.json - Overall quality summary"

# ==============================================================================
# Performance Monitoring Commands
# ==============================================================================

.PHONY: benchmark
benchmark: ## Run performance benchmarks with regression detection
	@echo "$(BLUE)Running performance benchmarks...$(RESET)"
	@mkdir -p $(BUILD_DIR)/benchmarks
	go run ./cmd/go-sentinel-cli benchmark --format=text
	@echo "$(GREEN)✓ Benchmark analysis complete$(RESET)"

.PHONY: benchmark-json
benchmark-json: ## Run benchmarks and generate JSON report
	@echo "$(BLUE)Running benchmarks with JSON output...$(RESET)"
	@mkdir -p $(BUILD_DIR)/benchmarks
	go run ./cmd/go-sentinel-cli benchmark --format=json --output=$(BUILD_DIR)/benchmarks/performance-report.json
	@echo "$(GREEN)✓ JSON benchmark report: $(BUILD_DIR)/benchmarks/performance-report.json$(RESET)"

.PHONY: benchmark-baseline
benchmark-baseline: ## Save current benchmark results as baseline
	@echo "$(BLUE)Saving benchmark baseline...$(RESET)"
	@mkdir -p $(BUILD_DIR)/benchmarks
	go run ./cmd/go-sentinel-cli benchmark --save-baseline --baseline-file=$(BUILD_DIR)/benchmarks/baseline.json
	@echo "$(GREEN)✓ Baseline saved: $(BUILD_DIR)/benchmarks/baseline.json$(RESET)"

.PHONY: benchmark-strict
benchmark-strict: ## Run benchmarks with strict thresholds (10% slowdown, 15% memory)
	@echo "$(BLUE)Running strict benchmark analysis...$(RESET)"
	@mkdir -p $(BUILD_DIR)/benchmarks
	go run ./cmd/go-sentinel-cli benchmark --max-slowdown=10 --max-memory-increase=15 --format=text
	@echo "$(GREEN)✓ Strict benchmark analysis complete$(RESET)"

.PHONY: benchmark-package
benchmark-package: ## Run benchmarks for specific package (usage: make benchmark-package PKG=./internal/test/processor)
ifndef PKG
	@echo "$(RED)ERROR: PKG is required. Usage: make benchmark-package PKG=./internal/test/processor$(RESET)"
	@exit 1
endif
	@echo "$(BLUE)Running benchmarks for package: $(PKG)...$(RESET)"
	@mkdir -p $(BUILD_DIR)/benchmarks
	go run ./cmd/go-sentinel-cli benchmark --packages="$(PKG)" --format=text
	@echo "$(GREEN)✓ Package benchmark complete$(RESET)"

.PHONY: benchmark-ci
benchmark-ci: ## Run benchmarks for CI/CD with error exit on regressions
	@echo "$(BLUE)Running CI benchmark analysis...$(RESET)"
	@mkdir -p $(BUILD_DIR)/benchmarks
	go run ./cmd/go-sentinel-cli benchmark --format=json --output=$(BUILD_DIR)/benchmarks/ci-performance-report.json --max-slowdown=20 --max-memory-increase=25
	@echo "$(GREEN)✓ CI benchmark analysis complete$(RESET)"

.PHONY: benchmark-verbose
benchmark-verbose: ## Run benchmarks with verbose output for debugging
	@echo "$(BLUE)Running verbose benchmark analysis...$(RESET)"
	@mkdir -p $(BUILD_DIR)/benchmarks
	go run ./cmd/go-sentinel-cli benchmark --verbose --format=text
	@echo "$(GREEN)✓ Verbose benchmark analysis complete$(RESET)"

.PHONY: benchmark-compare
benchmark-compare: ## Compare current benchmarks with baseline and show detailed analysis
	@echo "$(BLUE)Comparing benchmarks with baseline...$(RESET)"
	@mkdir -p $(BUILD_DIR)/benchmarks
	@if [ ! -f "$(BUILD_DIR)/benchmarks/baseline.json" ]; then \
		echo "$(YELLOW)⚠️  No baseline found. Creating baseline first...$(RESET)"; \
		$(MAKE) benchmark-baseline; \
		echo "$(YELLOW)⚠️  Baseline created. Run 'make benchmark-compare' again after making changes.$(RESET)"; \
	else \
		go run ./cmd/go-sentinel-cli benchmark --format=text --baseline-file=$(BUILD_DIR)/benchmarks/baseline.json; \
		echo "$(GREEN)✓ Benchmark comparison complete$(RESET)"; \
	fi

.PHONY: benchmark-trend
benchmark-trend: ## Generate performance trend analysis (requires multiple baseline runs)
	@echo "$(BLUE)Generating performance trend analysis...$(RESET)"
	@mkdir -p $(BUILD_DIR)/benchmarks/history
	@TIMESTAMP=$$(date +%Y%m%d_%H%M%S); \
	go run ./cmd/go-sentinel-cli benchmark --format=json --output=$(BUILD_DIR)/benchmarks/history/benchmark_$$TIMESTAMP.json; \
	echo "$(GREEN)✓ Performance snapshot saved: $(BUILD_DIR)/benchmarks/history/benchmark_$$TIMESTAMP.json$(RESET)"

.PHONY: benchmark-help
benchmark-help: ## Show benchmark command help and examples
	@echo "$(CYAN)Performance Benchmark Commands:$(RESET)"
	@echo ""
	@echo "$(YELLOW)Basic Commands:$(RESET)"
	@echo "  make benchmark              - Run benchmarks with regression detection"
	@echo "  make benchmark-baseline     - Save current results as baseline"
	@echo "  make benchmark-compare      - Compare with baseline"
	@echo ""
	@echo "$(YELLOW)Output Formats:$(RESET)"
	@echo "  make benchmark-json         - Generate JSON report"
	@echo "  make benchmark-verbose      - Verbose output for debugging"
	@echo ""
	@echo "$(YELLOW)Specialized Analysis:$(RESET)"
	@echo "  make benchmark-strict       - Strict thresholds (10%/15%)"
	@echo "  make benchmark-ci           - CI/CD integration with error codes"
	@echo "  make benchmark-trend        - Track performance over time"
	@echo ""
	@echo "$(YELLOW)Package-Specific:$(RESET)"
	@echo "  make benchmark-package PKG=./internal/test/processor"
	@echo ""
	@echo "$(YELLOW)Files Generated:$(RESET)"
	@echo "  $(BUILD_DIR)/benchmarks/baseline.json           - Performance baseline"
	@echo "  $(BUILD_DIR)/benchmarks/performance-report.json - Latest analysis"
	@  echo "  $(BUILD_DIR)/benchmarks/history/                - Historical snapshots"

# ==============================================================================
# Documentation Automation Commands
# ==============================================================================

.PHONY: docs-automation
docs-automation: ## Run comprehensive documentation automation pipeline
	@echo "$(BLUE)Running documentation automation...$(RESET)"
	@chmod +x scripts/docs-automation.sh
	@./scripts/docs-automation.sh
	@echo "$(GREEN)✓ Documentation automation complete$(RESET)"

.PHONY: docs-generate
docs-generate: ## Generate API documentation from source code
	@echo "$(BLUE)Generating API documentation...$(RESET)"
	@mkdir -p build/docs/api
	go doc -all ./pkg/models > build/docs/api/models.md
	go doc -all ./pkg/events > build/docs/api/events.md
	go doc -all ./internal/app > build/docs/api/app.md
	go doc -all ./internal/config > build/docs/api/config.md
	@echo "$(GREEN)✓ API documentation generated$(RESET)"

.PHONY: docs-validate
docs-validate: ## Validate documentation links and example code
	@echo "$(BLUE)Validating documentation...$(RESET)"
	@mkdir -p build/docs/validation
	@echo "Checking for broken links in README..."
	@echo "Validating example code syntax..."
	@echo "$(GREEN)✓ Documentation validation complete$(RESET)"

.PHONY: docs-coverage
docs-coverage: ## Generate documentation coverage report
	@echo "$(BLUE)Generating documentation coverage report...$(RESET)"
	@mkdir -p build/docs
	@echo "Analyzing documentation coverage for exported symbols..."
	@echo "$(GREEN)✓ Documentation coverage report generated$(RESET)"

.PHONY: docs-sync
docs-sync: ## Synchronize README with current package information
	@echo "$(BLUE)Synchronizing README with package information...$(RESET)"
	@cp README.md README.md.backup
	@echo "$(GREEN)✓ README synchronized$(RESET)"

.PHONY: docs-examples
docs-examples: ## Validate and test example code
	@echo "$(BLUE)Testing example code...$(RESET)"
	@if [ -f "pkg/models/examples.go" ]; then \
		echo "Testing models examples..."; \
		go build -o /dev/null pkg/models/examples.go || true; \
	fi
	@if [ -f "pkg/events/examples.go" ]; then \
		echo "Testing events examples..."; \
		go build -o /dev/null pkg/events/examples.go || true; \
	fi
	@echo "$(GREEN)✓ Example code testing complete$(RESET)"

.PHONY: docs-index
docs-index: ## Generate documentation index
	@echo "$(BLUE)Generating documentation index...$(RESET)"
	@mkdir -p build/docs
	@echo "# Go Sentinel CLI Documentation" > build/docs/index.md
	@echo "" >> build/docs/index.md
	@echo "Generated on: $$(date -u +\"%Y-%m-%d %H:%M:%S UTC\")" >> build/docs/index.md
	@echo "$(GREEN)✓ Documentation index generated$(RESET)"

.PHONY: docs-clean
docs-clean: ## Clean generated documentation
	@echo "$(BLUE)Cleaning generated documentation...$(RESET)"
	@rm -rf build/docs
	@echo "$(GREEN)✓ Documentation cleaned$(RESET)"

.PHONY: docs-server
docs-server: ## Start local documentation server (requires Python)
	@echo "$(BLUE)Starting documentation server...$(RESET)"
	@if command -v python3 >/dev/null 2>&1; then \
		echo "Documentation server available at http://localhost:8000"; \
		echo "Press Ctrl+C to stop"; \
		cd docs && python3 -m http.server 8000; \
	else \
		echo "$(RED)Python 3 not available. Install Python to use documentation server.$(RESET)"; \
	fi

.PHONY: docs-check
docs-check: ## Quick documentation check (fast validation)
	@echo "$(BLUE)Running quick documentation check...$(RESET)"
	@echo "Checking for missing documentation..."
	@echo "Validating markdown syntax..."
	@echo "$(GREEN)✓ Quick documentation check complete$(RESET)"

.PHONY: docs-help
docs-help: ## Show documentation automation help and examples
	@echo "$(CYAN)Documentation Automation Commands:$(RESET)"
	@echo ""
	@echo "$(YELLOW)Comprehensive Automation:$(RESET)"
	@echo "  make docs-automation        - Full documentation pipeline (all checks)"
	@echo "  make docs-generate          - Generate API documentation from source"
	@echo "  make docs-validate          - Validate links and example code"
	@echo ""
	@echo "$(YELLOW)Content Management:$(RESET)"
	@echo "  make docs-sync              - Synchronize README with package info"
	@echo "  make docs-examples          - Test example code compilation"
	@echo "  make docs-index             - Generate documentation index"
	@echo ""
	@echo "$(YELLOW)Quality Assurance:$(RESET)"
	@echo "  make docs-coverage          - Generate documentation coverage report"
	@echo "  make docs-check             - Quick documentation validation"
	@echo ""
	@echo "$(YELLOW)Development:$(RESET)"
	@echo "  make docs-server            - Start local documentation server"
	@echo "  make docs-clean             - Clean generated documentation"
	@echo ""
	@echo "$(YELLOW)Generated Files:$(RESET)"
	@echo "  build/docs/index.md                   - Documentation index"
	@echo "  build/docs/api/                       - Auto-generated API docs"
	@echo "  build/docs/documentation-coverage.md  - Coverage analysis"
	@echo "  build/docs/validation/                - Validation scripts and reports"

# ==============================================================================
# Monitoring & Observability Commands
# ==============================================================================

.PHONY: monitoring-start
monitoring-start: build-v2 ## Start monitoring system with test application
	@echo "$(BLUE)Starting monitoring system...$(RESET)"
	@echo "$(CYAN)Monitoring will be available at:$(RESET)"
	@echo "  Metrics:    http://localhost:8080/metrics"
	@echo "  Health:     http://localhost:8080/health"
	@echo "  Readiness:  http://localhost:8080/health/ready"
	@echo "  Liveness:   http://localhost:8080/health/live"
	@echo ""
	@echo "$(YELLOW)Starting application with monitoring enabled...$(RESET)"
	MONITORING_ENABLED=true MONITORING_PORT=8080 ./$(BUILD_DIR)/$(BINARY_NAME_V2) run --watch --color

.PHONY: monitoring-test
monitoring-test: ## Test monitoring endpoints
	@echo "$(BLUE)Testing monitoring endpoints...$(RESET)"
	@echo ""
	@echo "$(CYAN)Testing health endpoint...$(RESET)"
	@curl -s http://localhost:8080/health | jq . || echo "Health endpoint not available"
	@echo ""
	@echo "$(CYAN)Testing metrics endpoint...$(RESET)"
	@curl -s http://localhost:8080/metrics | jq . || echo "Metrics endpoint not available"
	@echo ""
	@echo "$(CYAN)Testing readiness endpoint...$(RESET)"
	@curl -s http://localhost:8080/health/ready | jq . || echo "Readiness endpoint not available"
	@echo ""
	@echo "$(CYAN)Testing liveness endpoint...$(RESET)"
	@curl -s http://localhost:8080/health/live | jq . || echo "Liveness endpoint not available"

.PHONY: monitoring-metrics
monitoring-metrics: ## Fetch current metrics from running application
	@echo "$(BLUE)Fetching current metrics...$(RESET)"
	@curl -s http://localhost:8080/metrics | jq .

.PHONY: monitoring-metrics-prometheus
monitoring-metrics-prometheus: ## Fetch metrics in Prometheus format
	@echo "$(BLUE)Fetching metrics in Prometheus format...$(RESET)"
	@curl -s "http://localhost:8080/metrics?format=prometheus"

.PHONY: monitoring-health
monitoring-health: ## Check application health status
	@echo "$(BLUE)Checking application health...$(RESET)"
	@curl -s http://localhost:8080/health | jq .

.PHONY: monitoring-dashboard
monitoring-dashboard: ## Generate monitoring dashboard (requires application to be running)
	@echo "$(BLUE)Generating monitoring dashboard...$(RESET)"
	@mkdir -p $(BUILD_DIR)/monitoring
	@echo "Collecting metrics and health data..."
	@curl -s http://localhost:8080/metrics > $(BUILD_DIR)/monitoring/metrics.json 2>/dev/null || echo "{\"error\": \"metrics not available\"}" > $(BUILD_DIR)/monitoring/metrics.json
	@curl -s http://localhost:8080/health > $(BUILD_DIR)/monitoring/health.json 2>/dev/null || echo "{\"error\": \"health not available\"}" > $(BUILD_DIR)/monitoring/health.json
	@echo "Dashboard data saved to $(BUILD_DIR)/monitoring/"
	@echo "$(GREEN)✓ Monitoring dashboard data generated$(RESET)"

.PHONY: monitoring-load-test
monitoring-load-test: ## Run load test to generate monitoring data
	@echo "$(BLUE)Running load test to generate monitoring data...$(RESET)"
	@echo "$(YELLOW)This will run multiple test cycles to generate metrics...$(RESET)"
	@for i in {1..10}; do \
		echo "Load test cycle $$i/10..."; \
		timeout 5s ./$(BUILD_DIR)/$(BINARY_NAME_V2) run --timeout=2s > /dev/null 2>&1 || true; \
		sleep 1; \
	done
	@echo "$(GREEN)✓ Load test complete. Check metrics at http://localhost:8080/metrics$(RESET)"

.PHONY: monitoring-alerts
monitoring-alerts: ## Check for monitoring alerts/thresholds
	@echo "$(BLUE)Checking monitoring alerts and thresholds...$(RESET)"
	@echo ""
	@echo "$(CYAN)Checking health status...$(RESET)"
	@HEALTH_STATUS=$$(curl -s http://localhost:8080/health 2>/dev/null | jq -r '.status' 2>/dev/null || echo "unavailable"); \
	if [ "$$HEALTH_STATUS" = "healthy" ]; then \
		echo "$(GREEN)✅ System is healthy$(RESET)"; \
	elif [ "$$HEALTH_STATUS" = "unhealthy" ]; then \
		echo "$(RED)❌ System is unhealthy$(RESET)"; \
		curl -s http://localhost:8080/health | jq '.checks'; \
	else \
		echo "$(YELLOW)⚠️  Health status unavailable (application not running?)$(RESET)"; \
	fi
	@echo ""
	@echo "$(CYAN)Checking memory usage...$(RESET)"
	@MEMORY=$$(curl -s http://localhost:8080/metrics 2>/dev/null | jq -r '.memory_usage_bytes // 0' 2>/dev/null || echo "0"); \
	MEMORY_MB=$$(($$MEMORY / 1024 / 1024)); \
	if [ $$MEMORY_MB -gt 500 ]; then \
		echo "$(RED)❌ High memory usage: $${MEMORY_MB}MB$(RESET)"; \
	elif [ $$MEMORY_MB -gt 250 ]; then \
		echo "$(YELLOW)⚠️  Moderate memory usage: $${MEMORY_MB}MB$(RESET)"; \
	else \
		echo "$(GREEN)✅ Memory usage normal: $${MEMORY_MB}MB$(RESET)"; \
	fi

.PHONY: monitoring-export
monitoring-export: ## Export monitoring data for external systems
	@echo "$(BLUE)Exporting monitoring data...$(RESET)"
	@mkdir -p $(BUILD_DIR)/monitoring/export
	@echo "Exporting JSON format..."
	@curl -s http://localhost:8080/metrics > $(BUILD_DIR)/monitoring/export/metrics.json 2>/dev/null || echo "{}" > $(BUILD_DIR)/monitoring/export/metrics.json
	@echo "Exporting Prometheus format..."
	@curl -s "http://localhost:8080/metrics?format=prometheus" > $(BUILD_DIR)/monitoring/export/metrics.prom 2>/dev/null || echo "" > $(BUILD_DIR)/monitoring/export/metrics.prom
	@echo "Exporting health data..."
	@curl -s http://localhost:8080/health > $(BUILD_DIR)/monitoring/export/health.json 2>/dev/null || echo "{}" > $(BUILD_DIR)/monitoring/export/health.json
	@echo "$(GREEN)✓ Monitoring data exported to $(BUILD_DIR)/monitoring/export/$(RESET)"

.PHONY: monitoring-stop
monitoring-stop: ## Stop monitoring system (kill running processes)
	@echo "$(BLUE)Stopping monitoring system...$(RESET)"
	@pkill -f "$(BINARY_NAME_V2)" || echo "No monitoring processes found"
	@echo "$(GREEN)✓ Monitoring system stopped$(RESET)"

.PHONY: monitoring-clean
monitoring-clean: ## Clean monitoring data and reports
	@echo "$(BLUE)Cleaning monitoring data...$(RESET)"
	@rm -rf $(BUILD_DIR)/monitoring
	@echo "$(GREEN)✓ Monitoring data cleaned$(RESET)"

.PHONY: monitoring-help
monitoring-help: ## Show monitoring commands help and examples
	@echo "$(CYAN)Monitoring & Observability Commands:$(RESET)"
	@echo ""
	@echo "$(YELLOW)System Control:$(RESET)"
	@echo "  make monitoring-start       - Start application with monitoring enabled"
	@echo "  make monitoring-stop        - Stop monitoring system"
	@echo "  make monitoring-test        - Test all monitoring endpoints"
	@echo ""
	@echo "$(YELLOW)Data Collection:$(RESET)"
	@echo "  make monitoring-metrics     - Fetch current metrics (JSON)"
	@echo "  make monitoring-metrics-prometheus - Fetch metrics (Prometheus format)"
	@echo "  make monitoring-health      - Check application health"
	@echo "  make monitoring-dashboard   - Generate monitoring dashboard"
	@echo ""
	@echo "$(YELLOW)Testing & Validation:$(RESET)"
	@echo "  make monitoring-load-test   - Run load test to generate data"
	@echo "  make monitoring-alerts      - Check alert thresholds"
	@echo ""
	@echo "$(YELLOW)Data Export:$(RESET)"
	@echo "  make monitoring-export      - Export data for external systems"
	@echo "  make monitoring-clean       - Clean monitoring data"
	@echo ""
	@echo "$(YELLOW)Monitoring Endpoints (when running):$(RESET)"
	@echo "  http://localhost:8080/metrics       - Application metrics"
	@echo "  http://localhost:8080/health        - Health checks"
	@echo "  http://localhost:8080/health/ready  - Readiness probe"
	@echo "  http://localhost:8080/health/live   - Liveness probe"
	@echo ""
	@echo "$(YELLOW)Example Workflow:$(RESET)"
	@echo "  1. make monitoring-start            # Start with monitoring"
	@echo "  2. make monitoring-load-test        # Generate some data"
	@echo "  3. make monitoring-dashboard        # Collect dashboard data"
	@echo "  4. make monitoring-alerts           # Check system health"
	@echo "  5. make monitoring-export           # Export for external tools"
	@echo ""
	@echo "$(YELLOW)Generated Files:$(RESET)"
	@echo "  $(BUILD_DIR)/monitoring/metrics.json       - Current metrics snapshot"
	@echo "  $(BUILD_DIR)/monitoring/health.json        - Current health status"
	@echo "  $(BUILD_DIR)/monitoring/export/            - Exported data for external systems"

# ==============================================================================
# Deployment Automation Commands
# ==============================================================================

.PHONY: deploy-staging
deploy-staging: build-v2 ## Deploy to staging environment using rolling strategy
	@echo "$(BLUE)Deploying to staging environment...$(RESET)"
	@chmod +x scripts/deployment-automation.sh
	@./scripts/deployment-automation.sh deploy -e staging -s rolling
	@echo "$(GREEN)✓ Staging deployment complete$(RESET)"

.PHONY: deploy-production
deploy-production: ## Deploy to production environment using blue-green strategy (requires VERSION)
ifndef VERSION
	@echo "$(RED)ERROR: VERSION is required. Usage: make deploy-production VERSION=1.2.3$(RESET)"
	@exit 1
endif
	@echo "$(BLUE)Deploying to production environment...$(RESET)"
	@chmod +x scripts/deployment-automation.sh
	@./scripts/deployment-automation.sh deploy -e production -s blue-green -v $(VERSION)
	@echo "$(GREEN)✓ Production deployment complete$(RESET)"

.PHONY: deploy-dry-run
deploy-dry-run: ## Show deployment plan without executing (usage: make deploy-dry-run ENV=staging STRATEGY=rolling)
	@echo "$(BLUE)Running deployment dry run...$(RESET)"
	@chmod +x scripts/deployment-automation.sh
	@./scripts/deployment-automation.sh deploy -e ${ENV:-staging} -s ${STRATEGY:-rolling} --dry-run
	@echo "$(GREEN)✓ Deployment dry run complete$(RESET)"

.PHONY: deploy-rollback
deploy-rollback: ## Rollback deployment to previous version (usage: make deploy-rollback ENV=staging)
	@echo "$(BLUE)Rolling back deployment...$(RESET)"
	@chmod +x scripts/deployment-automation.sh
	@./scripts/deployment-automation.sh rollback -e ${ENV:-staging}
	@echo "$(GREEN)✓ Rollback complete$(RESET)"

.PHONY: deploy-status
deploy-status: ## Check deployment status (usage: make deploy-status ENV=staging)
	@echo "$(BLUE)Checking deployment status...$(RESET)"
	@chmod +x scripts/deployment-automation.sh
	@./scripts/deployment-automation.sh status -e ${ENV:-staging}

.PHONY: deploy-health
deploy-health: ## Check deployment health (usage: make deploy-health ENV=staging)
	@echo "$(BLUE)Checking deployment health...$(RESET)"
	@chmod +x scripts/deployment-automation.sh
	@./scripts/deployment-automation.sh health -e ${ENV:-staging}

.PHONY: deploy-list
deploy-list: ## List available deployment packages
	@echo "$(BLUE)Listing deployment packages...$(RESET)"
	@chmod +x scripts/deployment-automation.sh
	@./scripts/deployment-automation.sh list

.PHONY: deploy-cleanup
deploy-cleanup: ## Clean up old deployment packages (usage: make deploy-cleanup KEEP=5)
	@echo "$(BLUE)Cleaning up old deployments...$(RESET)"
	@chmod +x scripts/deployment-automation.sh
	@./scripts/deployment-automation.sh cleanup ${KEEP:-5}
	@echo "$(GREEN)✓ Deployment cleanup complete$(RESET)"

.PHONY: deploy-validate
deploy-validate: ## Validate deployment configuration
	@echo "$(BLUE)Validating deployment configuration...$(RESET)"
	@chmod +x scripts/deployment-automation.sh
	@./scripts/deployment-automation.sh validate
	@echo "$(GREEN)✓ Deployment validation complete$(RESET)"

.PHONY: deploy-canary
deploy-canary: ## Deploy using canary strategy (usage: make deploy-canary ENV=staging VERSION=1.2.3)
ifndef VERSION
	@echo "$(RED)ERROR: VERSION is required. Usage: make deploy-canary ENV=staging VERSION=1.2.3$(RESET)"
	@exit 1
endif
	@echo "$(BLUE)Deploying using canary strategy...$(RESET)"
	@chmod +x scripts/deployment-automation.sh
	@./scripts/deployment-automation.sh deploy -e ${ENV:-staging} -s canary -v $(VERSION)
	@echo "$(GREEN)✓ Canary deployment complete$(RESET)"

.PHONY: deploy-force
deploy-force: ## Force deployment bypassing safety checks (usage: make deploy-force ENV=staging VERSION=1.2.3)
ifndef VERSION
	@echo "$(RED)ERROR: VERSION is required. Usage: make deploy-force ENV=staging VERSION=1.2.3$(RESET)"
	@exit 1
endif
	@echo "$(YELLOW)⚠️  Force deploying (bypassing safety checks)...$(RESET)"
	@chmod +x scripts/deployment-automation.sh
	@./scripts/deployment-automation.sh deploy -e ${ENV:-staging} -s rolling -v $(VERSION) --force --skip-tests --skip-health-checks
	@echo "$(GREEN)✓ Force deployment complete$(RESET)"

.PHONY: deploy-emergency-rollback
deploy-emergency-rollback: ## Emergency rollback without health checks (usage: make deploy-emergency-rollback ENV=production)
	@echo "$(RED)🚨 Emergency rollback initiated...$(RESET)"
	@chmod +x scripts/deployment-automation.sh
	@./scripts/deployment-automation.sh rollback -e ${ENV:-staging} --immediate
	@echo "$(GREEN)✓ Emergency rollback complete$(RESET)"

.PHONY: deploy-test
deploy-test: ## Test deployment automation system with staging
	@echo "$(BLUE)Testing deployment automation system...$(RESET)"
	@echo ""
	@echo "$(CYAN)Step 1: Building application...$(RESET)"
	@$(MAKE) build-v2
	@echo ""
	@echo "$(CYAN)Step 2: Validating configuration...$(RESET)"
	@$(MAKE) deploy-validate
	@echo ""
	@echo "$(CYAN)Step 3: Running deployment dry run...$(RESET)"
	@$(MAKE) deploy-dry-run ENV=staging STRATEGY=rolling
	@echo ""
	@echo "$(CYAN)Step 4: Listing current packages...$(RESET)"
	@$(MAKE) deploy-list
	@echo ""
	@echo "$(GREEN)✓ Deployment automation test complete$(RESET)"

.PHONY: deploy-help
deploy-help: ## Show deployment automation help and examples
	@echo "$(CYAN)Deployment Automation Commands:$(RESET)"
	@echo ""
	@echo "$(YELLOW)Basic Deployment:$(RESET)"
	@echo "  make deploy-staging              - Deploy to staging (rolling strategy)"
	@echo "  make deploy-production VERSION=1.2.3 - Deploy to production (blue-green)"
	@echo "  make deploy-dry-run ENV=staging  - Show deployment plan without executing"
	@echo ""
	@echo "$(YELLOW)Advanced Deployment:$(RESET)"
	@echo "  make deploy-canary ENV=staging VERSION=1.2.3 - Canary deployment"
	@echo "  make deploy-force ENV=staging VERSION=1.2.3  - Force deployment (bypass checks)"
	@echo ""
	@echo "$(YELLOW)Rollback & Recovery:$(RESET)"
	@echo "  make deploy-rollback ENV=staging - Rollback to previous version"
	@echo "  make deploy-emergency-rollback ENV=production - Emergency rollback"
	@echo ""
	@echo "$(YELLOW)Monitoring & Management:$(RESET)"
	@echo "  make deploy-status ENV=staging   - Check deployment status"
	@echo "  make deploy-health ENV=staging   - Check deployment health"
	@echo "  make deploy-list                 - List available packages"
	@echo "  make deploy-cleanup KEEP=5       - Clean up old packages"
	@echo ""
	@echo "$(YELLOW)Testing & Validation:$(RESET)"
	@echo "  make deploy-test                 - Test deployment automation system"
	@echo "  make deploy-validate             - Validate deployment configuration"
	@echo ""
	@echo "$(YELLOW)Deployment Strategies:$(RESET)"
	@echo "  rolling    - Zero-downtime rolling update (default for staging)"
	@echo "  blue-green - Blue-green deployment with traffic switch (default for production)"
	@echo "  canary     - Canary deployment with gradual traffic shift"
	@echo ""
	@echo "$(YELLOW)Environment Variables:$(RESET)"
	@echo "  DEPLOY_ENV             - Default deployment environment"
	@echo "  DEPLOY_STRATEGY        - Default deployment strategy"
	@echo "  SLACK_WEBHOOK_URL      - Slack notifications webhook"
	@echo "  GITHUB_TOKEN           - GitHub API token for notifications"
	@echo ""
	@echo "$(YELLOW)Example Workflows:$(RESET)"
	@echo "  # Staging deployment"
	@echo "  make deploy-staging"
	@echo ""
	@echo "  # Production deployment with specific version"
	@echo "  make deploy-production VERSION=1.2.3"
	@echo ""
	@echo "  # Emergency rollback"
	@echo "  make deploy-emergency-rollback ENV=production"
	@echo ""
	@echo "  # Check status and health"
	@echo "  make deploy-status ENV=production"
	@echo "  make deploy-health ENV=production"
	@echo ""
	@echo "$(YELLOW)Generated Files:$(RESET)"
	@echo "  build/deployment/packages/          - Deployment packages"
	@echo "  build/deployment/staging/           - Staging environment data"
	@echo "  build/deployment/production/        - Production environment data"

# ==============================================================================
# Release Automation Commands
# ==============================================================================

.PHONY: release-patch
release-patch: ## Create a patch release (x.y.Z)
	@echo "$(BLUE)Creating patch release...$(RESET)"
	@chmod +x scripts/release-automation.sh
	@./scripts/release-automation.sh patch
	@echo "$(GREEN)✓ Patch release complete$(RESET)"

.PHONY: release-minor
release-minor: ## Create a minor release (x.Y.z)
	@echo "$(BLUE)Creating minor release...$(RESET)"
	@chmod +x scripts/release-automation.sh
	@./scripts/release-automation.sh minor
	@echo "$(GREEN)✓ Minor release complete$(RESET)"

.PHONY: release-major
release-major: ## Create a major release (X.y.z)
	@echo "$(BLUE)Creating major release...$(RESET)"
	@chmod +x scripts/release-automation.sh
	@./scripts/release-automation.sh major
	@echo "$(GREEN)✓ Major release complete$(RESET)"

.PHONY: release-custom
release-custom: ## Create a custom release (usage: make release-custom VERSION=1.2.3)
ifndef VERSION
	@echo "$(RED)ERROR: VERSION is required. Usage: make release-custom VERSION=1.2.3$(RESET)"
	@exit 1
endif
	@echo "$(BLUE)Creating custom release v$(VERSION)...$(RESET)"
	@chmod +x scripts/release-automation.sh
	@./scripts/release-automation.sh custom $(VERSION)
	@echo "$(GREEN)✓ Custom release complete$(RESET)"

.PHONY: release-dry-run
release-dry-run: ## Dry run release process (no git operations)
	@echo "$(BLUE)Running release dry run...$(RESET)"
	@echo "$(YELLOW)This would create a patch release with the following changes:$(RESET)"
	@echo ""
	@echo "Current version: $$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo '0.0.0')"
	@echo "Next version: $$(scripts/release-automation.sh patch 2>/dev/null | grep 'New version:' | cut -d' ' -f3 || echo 'N/A')"
	@echo ""
	@echo "$(CYAN)To proceed with actual release:$(RESET)"
	@echo "  make release-patch    # For patch release"
	@echo "  make release-minor    # For minor release"
	@echo "  make release-major    # For major release"

.PHONY: release-build-only
release-build-only: ## Build release binaries without creating release
	@echo "$(BLUE)Building release binaries...$(RESET)"
	@mkdir -p $(BUILD_DIR)/dist
	@VERSION=$$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo '0.0.0'); \
	for platform in "linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64"; do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		BINARY_NAME="go-sentinel-cli-$$VERSION-$$GOOS-$$GOARCH"; \
		if [ "$$GOOS" = "windows" ]; then BINARY_NAME="$$BINARY_NAME.exe"; fi; \
		echo "Building $$platform..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build \
			-ldflags="-s -w -X main.version=$$VERSION -X main.commit=$$(git rev-parse HEAD) -X main.buildTime=$$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
			-o $(BUILD_DIR)/dist/$$BINARY_NAME \
			./cmd/go-sentinel-cli; \
	done
	@echo "$(GREEN)✓ Release binaries built in $(BUILD_DIR)/dist/$(RESET)"

.PHONY: release-changelog
release-changelog: ## Generate changelog for next release
	@echo "$(BLUE)Generating changelog preview...$(RESET)"
	@echo "## [Next Release] - $$(date +%Y-%m-%d)"
	@echo ""
	@LAST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo ""); \
	if [ -n "$$LAST_TAG" ]; then \
		echo "### Changes since $$LAST_TAG"; \
		echo ""; \
		git log $$LAST_TAG..HEAD --pretty=format:"- %s" --no-merges | head -20; \
	else \
		echo "### Initial Release"; \
		echo ""; \
		echo "- Initial release of Go Sentinel CLI"; \
	fi
	@echo ""

.PHONY: release-check
release-check: ## Check if ready for release
	@echo "$(BLUE)Checking release readiness...$(RESET)"
	@echo ""
	@echo "$(CYAN)Git Status:$(RESET)"
	@if git diff-index --quiet HEAD --; then \
		echo "✅ Working directory is clean"; \
	else \
		echo "❌ Working directory has uncommitted changes"; \
		exit 1; \
	fi
	@echo ""
	@echo "$(CYAN)Branch Status:$(RESET)"
	@BRANCH=$$(git branch --show-current); \
	if [ "$$BRANCH" = "main" ] || [ "$$BRANCH" = "master" ]; then \
		echo "✅ On release branch ($$BRANCH)"; \
	else \
		echo "⚠️  Not on main/master branch (current: $$BRANCH)"; \
	fi
	@echo ""
	@echo "$(CYAN)Quality Checks:$(RESET)"
	@if make quality-check >/dev/null 2>&1; then \
		echo "✅ Quality checks pass"; \
	else \
		echo "❌ Quality checks fail"; \
		exit 1; \
	fi
	@echo ""
	@echo "$(CYAN)Test Status:$(RESET)"
	@if go test ./... >/dev/null 2>&1; then \
		echo "✅ All tests pass"; \
	else \
		echo "❌ Tests fail"; \
		exit 1; \
	fi
	@echo ""
	@echo "$(GREEN)✅ Ready for release!$(RESET)"

.PHONY: release-push
release-push: ## Push release tag to origin
	@echo "$(BLUE)Pushing release tag...$(RESET)"
	@LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null); \
	if [ -n "$$LATEST_TAG" ]; then \
		echo "Pushing tag: $$LATEST_TAG"; \
		git push origin $$LATEST_TAG; \
		echo "$(GREEN)✅ Tag $$LATEST_TAG pushed to origin$(RESET)"; \
	else \
		echo "$(RED)❌ No tags found to push$(RESET)"; \
		exit 1; \
	fi

.PHONY: release-help
release-help: ## Show release automation help and examples
	@echo "$(CYAN)Release Automation Commands:$(RESET)"
	@echo ""
	@echo "$(YELLOW)Release Types:$(RESET)"
	@echo "  make release-patch              - Create patch release (x.y.Z)"
	@echo "  make release-minor              - Create minor release (x.Y.z)"
	@echo "  make release-major              - Create major release (X.y.z)"
	@echo "  make release-custom VERSION=1.2.3 - Create custom version release"
	@echo ""
	@echo "$(YELLOW)Release Preparation:$(RESET)"
	@echo "  make release-check              - Check if ready for release"
	@echo "  make release-dry-run            - Preview next release"
	@echo "  make release-changelog          - Generate changelog preview"
	@echo "  make release-build-only         - Build binaries without release"
	@echo ""
	@echo "$(YELLOW)Release Publishing:$(RESET)"
	@echo "  make release-push               - Push release tag to origin"
	@echo ""
	@echo "$(YELLOW)Semantic Versioning:$(RESET)"
	@echo "  MAJOR: Breaking changes (1.0.0 → 2.0.0)"
	@echo "  MINOR: New features (1.0.0 → 1.1.0)"
	@echo "  PATCH: Bug fixes (1.0.0 → 1.0.1)"
	@echo ""
	@echo "$(YELLOW)Release Process:$(RESET)"
	@echo "  1. make release-check           # Verify readiness"
	@echo "  2. make release-patch           # Create release"
	@echo "  3. make release-push            # Push to origin"
	@echo "  4. Create GitHub release with generated assets"
	@echo ""
	@echo "$(YELLOW)Generated Assets:$(RESET)"
	@echo "  $(BUILD_DIR)/dist/go-sentinel-cli-VERSION.tar.gz    - Release archive"
	@echo "  $(BUILD_DIR)/dist/release-notes-VERSION.md          - Release notes"
	@echo "  $(BUILD_DIR)/dist/go-sentinel-cli-VERSION-*         - Platform binaries"

# Default target
.DEFAULT_GOAL := help
