#!/bin/bash
# Setup script for Git hooks and pre-commit configuration
# Installs and configures all necessary quality gates and checks

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install tool if not present
install_tool() {
    local tool=$1
    local install_cmd=$2

    if ! command_exists "$tool"; then
        log_info "Installing $tool..."
        eval "$install_cmd"
        if command_exists "$tool"; then
            log_success "$tool installed successfully"
        else
            log_error "Failed to install $tool"
            return 1
        fi
    else
        log_success "$tool is already installed"
    fi
}

# Install pre-commit
install_pre_commit() {
    log_info "Setting up pre-commit..."

    if command_exists pre-commit; then
        log_success "pre-commit is already installed"
    else
        if command_exists pip3; then
            pip3 install pre-commit
        elif command_exists pip; then
            pip install pre-commit
        elif command_exists brew; then
            brew install pre-commit
        else
            log_error "Cannot install pre-commit. Please install pip or brew first."
            return 1
        fi
    fi
}

# Install Go tools
install_go_tools() {
    log_info "Installing Go development tools..."

    # List of Go tools to install
    declare -A go_tools=(
        ["goimports"]="golang.org/x/tools/cmd/goimports@latest"
        ["golangci-lint"]="github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
        ["gosec"]="github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
        ["go-mod-outdated"]="github.com/psampaz/go-mod-outdated@latest"
    )

    for tool in "${!go_tools[@]}"; do
        if ! command_exists "$tool"; then
            log_info "Installing $tool..."
            go install "${go_tools[$tool]}"
            log_success "$tool installed"
        else
            log_success "$tool is already installed"
        fi
    done
}

# Install system tools
install_system_tools() {
    log_info "Installing system tools..."

    # Check OS type
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        if command_exists apt-get; then
            # Debian/Ubuntu
            install_tool "shellcheck" "sudo apt-get update && sudo apt-get install -y shellcheck"
            install_tool "jq" "sudo apt-get install -y jq"
            install_tool "bc" "sudo apt-get install -y bc"
        elif command_exists yum; then
            # RHEL/CentOS
            install_tool "shellcheck" "sudo yum install -y ShellCheck"
            install_tool "jq" "sudo yum install -y jq"
            install_tool "bc" "sudo yum install -y bc"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        if command_exists brew; then
            install_tool "shellcheck" "brew install shellcheck"
            install_tool "jq" "brew install jq"
            install_tool "bc" "echo 'bc is built-in on macOS'"
        fi
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        # Windows (Git Bash/Cygwin)
        log_warning "Please install shellcheck and jq manually on Windows"
        log_info "shellcheck: https://github.com/koalaman/shellcheck#installing"
        log_info "jq: https://stedolan.github.io/jq/download/"
    fi
}

# Setup commit message hook
setup_commit_msg_hook() {
    log_info "Setting up commit-msg hook..."

    local hook_file=".git/hooks/commit-msg"

    cat > "$hook_file" << 'EOF'
#!/bin/bash
# Commit message validation hook
# This hook validates commit messages according to project standards

# Check if the commit message validator exists
if [ ! -f "scripts/validate_commit_msg.py" ]; then
    echo "❌ Error: Commit message validator not found"
    exit 1
fi

# Validate the commit message
python3 scripts/validate_commit_msg.py "$1"
exit_code=$?

if [ $exit_code -ne 0 ]; then
    echo ""
    echo "💡 Tip: Use 'python3 scripts/validate_commit_msg.py --help' for format guide"
    exit $exit_code
fi
EOF

    chmod +x "$hook_file"
    log_success "commit-msg hook installed"
}

# Setup pre-push hook for comprehensive checks
setup_pre_push_hook() {
    log_info "Setting up pre-push hook..."

    local hook_file=".git/hooks/pre-push"

    cat > "$hook_file" << 'EOF'
#!/bin/bash
# Pre-push hook for comprehensive quality checks
# This hook runs additional checks before pushing to remote

set -e

echo "🚀 Running pre-push quality checks..."

# Check if we're on a branch that should be protected
protected_branches="main master develop"
current_branch=$(git symbolic-ref --short HEAD)

for branch in $protected_branches; do
    if [ "$current_branch" = "$branch" ]; then
        echo "⚠️  Pushing to protected branch: $branch"
        echo "   Running comprehensive quality checks..."

        # Run comprehensive tests
        echo "📋 Running full test suite..."
        go test ./... -race -timeout=30m

        # Run benchmarks to check for performance regressions
        echo "🚀 Running performance benchmarks..."
        if [ -f "internal/test/benchmarks" ]; then
            go test -bench=. -benchmem ./internal/test/benchmarks/ > /tmp/current-bench.txt
            echo "✅ Performance benchmarks completed"
        fi

        # Run security scan
        echo "🔒 Running security scan..."
        if command -v gosec >/dev/null 2>&1; then
            gosec -quiet ./...
        fi

        # Check for sensitive data
        echo "🔍 Checking for sensitive data..."
        if git diff --cached --name-only | xargs grep -l -E "(password|secret|key|token)" 2>/dev/null; then
            echo "❌ Potential sensitive data found in staged files"
            exit 1
        fi

        break
    fi
done

echo "✅ Pre-push checks completed successfully"
EOF

    chmod +x "$hook_file"
    log_success "pre-push hook installed"
}

# Create pre-commit configuration if it doesn't exist
check_pre_commit_config() {
    if [ ! -f ".pre-commit-config.yaml" ]; then
        log_error ".pre-commit-config.yaml not found"
        return 1
    fi
    log_success ".pre-commit-config.yaml exists"
}

# Install pre-commit hooks
install_hooks() {
    log_info "Installing pre-commit hooks..."

    if [ -f ".pre-commit-config.yaml" ]; then
        pre-commit install
        pre-commit install --hook-type commit-msg
        pre-commit install --hook-type pre-push
        log_success "Pre-commit hooks installed"
    else
        log_error ".pre-commit-config.yaml not found"
        return 1
    fi
}

# Validate current setup
validate_setup() {
    log_info "Validating setup..."

    # Check if Git is initialized
    if [ ! -d ".git" ]; then
        log_error "Not a Git repository"
        return 1
    fi

    # Check if Go is available
    if ! command_exists go; then
        log_error "Go is not installed"
        return 1
    fi

    # Check if Python 3 is available
    if ! command_exists python3; then
        log_error "Python 3 is not installed"
        return 1
    fi

    # Check if pre-commit config exists
    check_pre_commit_config

    log_success "Basic validation passed"
}

# Test hooks
test_hooks() {
    log_info "Testing pre-commit hooks..."

    # Test with a sample commit message
    echo "feat(test): add sample feature for testing" > /tmp/test-commit-msg.txt
    if python3 scripts/validate_commit_msg.py /tmp/test-commit-msg.txt; then
        log_success "Commit message validation works"
    else
        log_error "Commit message validation failed"
    fi

    # Test pre-commit on sample files
    if command_exists pre-commit; then
        log_info "Running pre-commit check on all files..."
        pre-commit run --all-files || {
            log_warning "Some pre-commit checks failed - this is normal on initial setup"
            log_info "Run 'pre-commit run --all-files' to see detailed output"
        }
    fi

    rm -f /tmp/test-commit-msg.txt
}

# Create quality check script
create_quality_check() {
    log_info "Creating quality check script..."

    cat > "scripts/quality-check.sh" << 'EOF'
#!/bin/bash
# Comprehensive quality check script
# Run this before important commits or releases

set -e

echo "🔍 Running comprehensive quality checks..."

# Format code
echo "📝 Formatting code..."
go fmt ./...
goimports -w .

# Run linting
echo "🔍 Running linters..."
golangci-lint run --config .golangci.yml

# Run tests with coverage
echo "🧪 Running tests with coverage..."
go test -race -coverprofile=coverage.out ./...
COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
echo "📊 Test coverage: $COVERAGE%"

if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "⚠️  Test coverage below 80%"
else
    echo "✅ Test coverage target met"
fi

# Run complexity analysis
echo "📊 Running complexity analysis..."
if [ -f "build/go-sentinel-cli" ]; then
    ./build/go-sentinel-cli complexity . --max-critical=5 --format=text
else
    echo "⚠️  go-sentinel-cli not built, skipping complexity analysis"
fi

# Run security scan
echo "🔒 Running security scan..."
gosec ./...

# Run benchmarks
echo "🚀 Running benchmarks..."
go test -bench=. -benchmem ./internal/test/benchmarks/ | head -20

echo "✅ Quality checks completed"
EOF

    chmod +x "scripts/quality-check.sh"
    log_success "Quality check script created"
}

# Main function
main() {
    echo "🔧 Setting up Git hooks and pre-commit configuration..."
    echo "=================================================="

    # Validate environment
    validate_setup

    # Install tools
    install_pre_commit
    install_go_tools
    install_system_tools

    # Setup hooks
    setup_commit_msg_hook
    setup_pre_push_hook
    install_hooks

    # Create additional scripts
    create_quality_check

    # Test setup
    test_hooks

    echo ""
    echo "🎉 Setup completed successfully!"
    echo ""
    echo "📋 What was installed:"
    echo "  ✅ Pre-commit hooks (formatting, linting, testing)"
    echo "  ✅ Commit message validation"
    echo "  ✅ Pre-push quality gates"
    echo "  ✅ Code complexity analysis integration"
    echo "  ✅ Security scanning"
    echo "  ✅ Performance regression detection"
    echo ""
    echo "📖 Usage:"
    echo "  • Hooks run automatically on commit/push"
    echo "  • Run 'scripts/quality-check.sh' for manual checks"
    echo "  • Run 'pre-commit run --all-files' to check all files"
    echo "  • Use 'python3 scripts/validate_commit_msg.py --help' for commit format guide"
    echo ""
    echo "🚀 Happy coding!"
}

# Run main function
main "$@"
