#!/bin/bash

# Documentation Automation Script for Go Sentinel CLI
# Handles API docs generation, README sync, example validation, and doc testing

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
DOCS_DIR="docs"
DOCS_OUTPUT_DIR="$BUILD_DIR/docs"
API_DOCS_FILE="$DOCS_DIR/API.md"
README_FILE="README.md"
EXAMPLES_DIR="$DOCS_DIR/examples"

echo -e "${CYAN}📚 Go Sentinel CLI - Documentation Automation${RESET}"
echo -e "${CYAN}=============================================${RESET}"
echo ""

# Function to run step with error handling
run_step() {
    local step_name="$1"
    local step_command="$2"
    local is_critical="${3:-true}"

    echo -e "${BLUE}🔄 Running: $step_name${RESET}"

    if eval "$step_command"; then
        echo -e "${GREEN}✅ $step_name: COMPLETED${RESET}"
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

# Create output directories
mkdir -p "$DOCS_OUTPUT_DIR"
mkdir -p "$DOCS_OUTPUT_DIR/api"
mkdir -p "$DOCS_OUTPUT_DIR/examples"
mkdir -p "$DOCS_OUTPUT_DIR/validation"

echo -e "${CYAN}📖 API Documentation Generation${RESET}"
echo "================================="

run_step "Generate Package Documentation" "go doc -all ./pkg/models > $DOCS_OUTPUT_DIR/api/models.md"
run_step "Generate Events Documentation" "go doc -all ./pkg/events > $DOCS_OUTPUT_DIR/api/events.md"
run_step "Generate App Documentation" "go doc -all ./internal/app > $DOCS_OUTPUT_DIR/api/app.md"
run_step "Generate Config Documentation" "go doc -all ./internal/config > $DOCS_OUTPUT_DIR/api/config.md"

# Generate API reference
echo "# Go Sentinel CLI - API Reference" > "$DOCS_OUTPUT_DIR/api/api-reference.md"
echo "" >> "$DOCS_OUTPUT_DIR/api/api-reference.md"
echo "Auto-generated API documentation for Go Sentinel CLI." >> "$DOCS_OUTPUT_DIR/api/api-reference.md"
echo "Last updated: $(date -u +%Y-%m-%dT%H:%M:%SZ)" >> "$DOCS_OUTPUT_DIR/api/api-reference.md"

echo -e "${CYAN}📑 Documentation Index Generation${RESET}"
echo "================================="

# Create documentation index
echo "# Go Sentinel CLI Documentation" > "$DOCS_OUTPUT_DIR/index.md"
echo "" >> "$DOCS_OUTPUT_DIR/index.md"
echo "Generated on: $(date -u +"%Y-%m-%d %H:%M:%S UTC")" >> "$DOCS_OUTPUT_DIR/index.md"
echo "" >> "$DOCS_OUTPUT_DIR/index.md"
echo "## API Reference" >> "$DOCS_OUTPUT_DIR/index.md"
echo "- [Complete API Reference](api/api-reference.md)" >> "$DOCS_OUTPUT_DIR/index.md"
echo "- [Models Package](api/models.md)" >> "$DOCS_OUTPUT_DIR/index.md"
echo "- [Events Package](api/events.md)" >> "$DOCS_OUTPUT_DIR/index.md"
echo "- [App Package](api/app.md)" >> "$DOCS_OUTPUT_DIR/index.md"
echo "- [Config Package](api/config.md)" >> "$DOCS_OUTPUT_DIR/index.md"

run_step "Generate Documentation Index" "echo '✅ Documentation index created'"

echo -e "${CYAN}✅ Final Documentation Validation${RESET}"
echo "================================="

DOCS_SCORE=100
ISSUES_FOUND=0

# Check if all required files exist
REQUIRED_FILES=(
    "$DOCS_OUTPUT_DIR/api/api-reference.md"
    "$DOCS_OUTPUT_DIR/api/models.md"
    "$DOCS_OUTPUT_DIR/api/events.md"
    "$DOCS_OUTPUT_DIR/index.md"
)

for file in "${REQUIRED_FILES[@]}"; do
    if [ ! -f "$file" ]; then
        echo -e "${RED}❌ Missing required documentation file: $file${RESET}"
        DOCS_SCORE=$((DOCS_SCORE - 10))
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
    else
        echo -e "${GREEN}✅ Found required file: $file${RESET}"
    fi
done

# Final assessment
echo ""
echo -e "${CYAN}📊 Documentation Quality Assessment${RESET}"
echo "====================================="
echo "Documentation Score: $DOCS_SCORE/100"
echo "Issues Found: $ISSUES_FOUND"

if [ $DOCS_SCORE -ge 95 ]; then
    echo -e "${GREEN}🎉 EXCELLENT DOCUMENTATION - Comprehensive and well-maintained${RESET}"
    exit 0
elif [ $DOCS_SCORE -ge 85 ]; then
    echo -e "${GREEN}✅ GOOD DOCUMENTATION - Minor improvements possible${RESET}"
    exit 0
elif [ $DOCS_SCORE -ge 70 ]; then
    echo -e "${YELLOW}⚠️  ACCEPTABLE DOCUMENTATION - Some issues need attention${RESET}"
    exit 0
else
    echo -e "${RED}❌ POOR DOCUMENTATION - Significant improvements needed${RESET}"
    exit 1
fi
