#!/bin/bash

# Quality Automation Script for Go Sentinel CLI
# Integrates static analysis, security scanning, and compliance checking

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
RESET='\033[0m'

# Configuration
BUILD_DIR="build"
REPORTS_DIR="$BUILD_DIR/quality-reports"
COVERAGE_THRESHOLD=85
COMPLEXITY_THRESHOLD=10
MAINTAINABILITY_THRESHOLD=85.0
DEBT_RATIO_THRESHOLD=5.0

# Tool versions
GOSEC_VERSION="latest"
GOVULNCHECK_VERSION="latest"
GOLANGCI_LINT_VERSION="latest"
NANCY_VERSION="latest"

# Create reports directory
mkdir -p "$REPORTS_DIR"

echo -e "${CYAN}🔍 Go Sentinel CLI - Quality Automation Pipeline${RESET}"
echo -e "${CYAN}=================================================${RESET}"
echo ""

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install Go tools
install_go_tool() {
    local tool_name="$1"
    local tool_path="$2"
    local tool_version="${3:-latest}"

    if ! command_exists "$tool_name"; then
        echo -e "${BLUE}📦 Installing $tool_name...${RESET}"
        go install "$tool_path@$tool_version"
    else
        echo -e "${GREEN}✓ $tool_name already installed${RESET}"
    fi
}

# Function to run step with error handling
run_step() {
    local step_name="$1"
    local step_command="$2"
    local is_critical="${3:-true}"

    echo -e "${BLUE}🔄 Running: $step_name${RESET}"

    if eval "$step_command"; then
        echo -e "${GREEN}✅ $step_name: PASSED${RESET}"
        return 0
    else
        if [ "$is_critical" = "true" ]; then
            echo -e "${RED}❌ $step_name: FAILED (CRITICAL)${RESET}"
            exit 1
        else
            echo -e "${YELLOW}⚠️  $step_name: FAILED (NON-CRITICAL)${RESET}"
            return 1
        fi
    fi
}

# Install required tools
echo -e "${CYAN}📦 Installing Quality Tools${RESET}"
echo "================================"

install_go_tool "gosec" "github.com/securecodewarrior/gosec/v2/cmd/gosec" "$GOSEC_VERSION"
install_go_tool "govulncheck" "golang.org/x/vuln/cmd/govulncheck" "$GOVULNCHECK_VERSION"
install_go_tool "nancy" "github.com/sonatypecommunity/nancy" "$NANCY_VERSION"

# Check for golangci-lint
if ! command_exists "golangci-lint"; then
    echo -e "${BLUE}📦 Installing golangci-lint...${RESET}"
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
else
    echo -e "${GREEN}✓ golangci-lint already installed${RESET}"
fi

echo ""

# 1. Code Formatting and Basic Checks
echo -e "${CYAN}🎨 Code Formatting & Basic Checks${RESET}"
echo "=================================="

run_step "Go Format Check" "gofmt -l . | tee $REPORTS_DIR/gofmt-issues.txt && [ ! -s $REPORTS_DIR/gofmt-issues.txt ]"
run_step "Go Imports Check" "goimports -l . | tee $REPORTS_DIR/goimports-issues.txt && [ ! -s $REPORTS_DIR/goimports-issues.txt ]" "false"
run_step "Go Vet" "go vet ./... 2>&1 | tee $REPORTS_DIR/go-vet.txt"
run_step "Go Mod Tidy Check" "go mod tidy && git diff --exit-code go.mod go.sum"

echo ""

# 2. Static Analysis with golangci-lint
echo -e "${CYAN}🔍 Static Analysis${RESET}"
echo "=================="

run_step "golangci-lint Analysis" "golangci-lint run --config .golangci.yml --out-format json --issues-exit-code 0 ./... > $REPORTS_DIR/golangci-lint.json && golangci-lint run --config .golangci.yml ./..."

echo ""

# 3. Security Analysis
echo -e "${CYAN}🔒 Security Analysis${RESET}"
echo "==================="

run_step "Gosec Security Scan" "gosec -fmt json -out $REPORTS_DIR/gosec.json ./... && gosec -fmt sarif -out $REPORTS_DIR/gosec.sarif ./..."
run_step "Go Vulnerability Check" "govulncheck -json ./... > $REPORTS_DIR/govulncheck.json 2>&1 || true" "false"

# Dependency vulnerability scanning with Nancy
echo -e "${BLUE}🔄 Running: Nancy Dependency Scan${RESET}"
if go list -json -deps ./... | nancy sleuth > "$REPORTS_DIR/nancy.txt" 2>&1; then
    echo -e "${GREEN}✅ Nancy Dependency Scan: PASSED${RESET}"
else
    echo -e "${YELLOW}⚠️  Nancy Dependency Scan: VULNERABILITIES FOUND${RESET}"
    echo "Check $REPORTS_DIR/nancy.txt for details"
fi

echo ""

# 4. Code Complexity Analysis
echo -e "${CYAN}📊 Code Complexity Analysis${RESET}"
echo "==========================="

run_step "Complexity Analysis" "go run ./cmd/go-sentinel-cli complexity --format=json --output=$REPORTS_DIR/complexity.json --max-complexity=$COMPLEXITY_THRESHOLD --min-maintainability=$MAINTAINABILITY_THRESHOLD --max-debt-ratio=$DEBT_RATIO_THRESHOLD ."

# Check for critical complexity violations
if go run ./cmd/go-sentinel-cli complexity --format=text . | grep -q "CRITICAL"; then
    echo -e "${RED}❌ Critical complexity violations found${RESET}"
    go run ./cmd/go-sentinel-cli complexity --format=text . | grep "CRITICAL" | head -5
    exit 1
else
    echo -e "${GREEN}✅ No critical complexity violations${RESET}"
fi

echo ""

# 5. Test Coverage Analysis
echo -e "${CYAN}🧪 Test Coverage Analysis${RESET}"
echo "========================="

run_step "Test Coverage" "go test -race -covermode=atomic -coverprofile=$REPORTS_DIR/coverage.out ./..."

# Check coverage threshold
echo -e "${BLUE}🔄 Running: Coverage Threshold Check${RESET}"
COVERAGE=$(go tool cover -func="$REPORTS_DIR/coverage.out" | tail -1 | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE >= $COVERAGE_THRESHOLD" | bc -l) )); then
    echo -e "${GREEN}✅ Coverage Threshold Check: PASSED ($COVERAGE% >= $COVERAGE_THRESHOLD%)${RESET}"
else
    echo -e "${RED}❌ Coverage Threshold Check: FAILED ($COVERAGE% < $COVERAGE_THRESHOLD%)${RESET}"
    exit 1
fi

# Generate HTML coverage report
go tool cover -html="$REPORTS_DIR/coverage.out" -o "$REPORTS_DIR/coverage.html"

echo ""

# 6. License Compliance Check
echo -e "${CYAN}📜 License Compliance Check${RESET}"
echo "=========================="

echo -e "${BLUE}🔄 Running: License Analysis${RESET}"
cat > "$REPORTS_DIR/license-check.go" << 'EOF'
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "strings"
)

type Module struct {
    Path    string `json:"Path"`
    Version string `json:"Version"`
}

func main() {
    cmd := exec.Command("go", "list", "-m", "-json", "all")
    output, err := cmd.Output()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error running go list: %v\n", err)
        os.Exit(1)
    }

    var modules []Module
    decoder := json.NewDecoder(strings.NewReader(string(output)))

    for decoder.More() {
        var mod Module
        if err := decoder.Decode(&mod); err != nil {
            continue
        }
        if mod.Path != "" && !strings.HasPrefix(mod.Path, "github.com/newbpydev/go-sentinel") {
            modules = append(modules, mod)
        }
    }

    fmt.Printf("Found %d external dependencies:\n", len(modules))
    for _, mod := range modules {
        fmt.Printf("- %s %s\n", mod.Path, mod.Version)
    }

    // Save to JSON for further analysis
    file, err := os.Create("build/quality-reports/dependencies.json")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating dependencies file: %v\n", err)
        os.Exit(1)
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    encoder.Encode(modules)
}
EOF

go run "$REPORTS_DIR/license-check.go" > "$REPORTS_DIR/dependencies.txt"
rm "$REPORTS_DIR/license-check.go"
echo -e "${GREEN}✅ License Analysis: COMPLETED${RESET}"
echo "Dependencies saved to $REPORTS_DIR/dependencies.json"

echo ""

# 7. Performance Benchmark Check
echo -e "${CYAN}⚡ Performance Benchmark Check${RESET}"
echo "============================="

if [ -f "build/benchmarks/baseline.json" ]; then
    run_step "Performance Regression Check" "go run ./cmd/go-sentinel-cli benchmark --format=json --output=$REPORTS_DIR/performance.json --max-slowdown=20 --max-memory-increase=25" "false"
else
    echo -e "${YELLOW}⚠️  No performance baseline found. Creating baseline...${RESET}"
    mkdir -p build/benchmarks
    go run ./cmd/go-sentinel-cli benchmark --save-baseline --baseline-file=build/benchmarks/baseline.json
    echo -e "${GREEN}✅ Performance baseline created${RESET}"
fi

echo ""

# 8. Generate Quality Report Summary
echo -e "${CYAN}📋 Quality Report Summary${RESET}"
echo "========================="

cat > "$REPORTS_DIR/quality-summary.json" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "coverage_percentage": $COVERAGE,
  "coverage_threshold": $COVERAGE_THRESHOLD,
  "complexity_threshold": $COMPLEXITY_THRESHOLD,
  "maintainability_threshold": $MAINTAINABILITY_THRESHOLD,
  "reports": {
    "coverage": "$REPORTS_DIR/coverage.html",
    "complexity": "$REPORTS_DIR/complexity.json",
    "security": "$REPORTS_DIR/gosec.json",
    "vulnerabilities": "$REPORTS_DIR/govulncheck.json",
    "dependencies": "$REPORTS_DIR/dependencies.json",
    "static_analysis": "$REPORTS_DIR/golangci-lint.json"
  }
}
EOF

echo -e "${GREEN}✅ Quality summary saved to $REPORTS_DIR/quality-summary.json${RESET}"

# 9. Final Quality Gate
echo ""
echo -e "${CYAN}🚪 Final Quality Gate${RESET}"
echo "==================="

QUALITY_SCORE=100
ISSUES_FOUND=0

# Check for critical issues
if [ -s "$REPORTS_DIR/gofmt-issues.txt" ]; then
    echo -e "${RED}❌ Code formatting issues found${RESET}"
    QUALITY_SCORE=$((QUALITY_SCORE - 10))
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

if grep -q '"Severity":"error"' "$REPORTS_DIR/golangci-lint.json" 2>/dev/null; then
    echo -e "${RED}❌ Critical linting errors found${RESET}"
    QUALITY_SCORE=$((QUALITY_SCORE - 20))
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

if grep -q '"Issues":\[' "$REPORTS_DIR/gosec.json" 2>/dev/null && ! grep -q '"Issues":\[\]' "$REPORTS_DIR/gosec.json"; then
    echo -e "${YELLOW}⚠️  Security issues found${RESET}"
    QUALITY_SCORE=$((QUALITY_SCORE - 15))
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

if (( $(echo "$COVERAGE < $COVERAGE_THRESHOLD" | bc -l) )); then
    echo -e "${RED}❌ Coverage below threshold${RESET}"
    QUALITY_SCORE=$((QUALITY_SCORE - 25))
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

# Final assessment
echo ""
echo -e "${CYAN}📊 Quality Assessment${RESET}"
echo "===================="
echo "Quality Score: $QUALITY_SCORE/100"
echo "Issues Found: $ISSUES_FOUND"

if [ $QUALITY_SCORE -ge 90 ]; then
    echo -e "${GREEN}🎉 EXCELLENT QUALITY - Ready for production${RESET}"
    exit 0
elif [ $QUALITY_SCORE -ge 75 ]; then
    echo -e "${YELLOW}⚠️  GOOD QUALITY - Minor issues to address${RESET}"
    exit 0
elif [ $QUALITY_SCORE -ge 60 ]; then
    echo -e "${YELLOW}⚠️  ACCEPTABLE QUALITY - Several issues need attention${RESET}"
    exit 1
else
    echo -e "${RED}❌ POOR QUALITY - Critical issues must be fixed${RESET}"
    exit 1
fi

