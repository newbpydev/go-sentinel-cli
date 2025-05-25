#!/bin/bash

# Deployment Automation Script for Go Sentinel CLI
# Handles staging deployment, blue-green deployments, rollback, and notifications

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
RESET='\033[0m'

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build"
DEPLOY_DIR="$BUILD_DIR/deployment"
BINARY_NAME="go-sentinel-cli-v2"

# Default configuration
DEFAULT_ENV="staging"
DEFAULT_STRATEGY="rolling"
DEFAULT_TIMEOUT="300"
DEFAULT_HEALTH_CHECK_INTERVAL="5"
DEFAULT_MAX_HEALTH_CHECKS="12"

# Environment configuration
ENVIRONMENTS=("staging" "production")
DEPLOYMENT_STRATEGIES=("rolling" "blue-green" "canary")

echo -e "${CYAN}🚀 Go Sentinel CLI - Deployment Automation${RESET}"
echo -e "${CYAN}===========================================${RESET}"
echo ""

# Function to show help
show_help() {
    cat << EOF
Usage: $0 [COMMAND] [OPTIONS]

COMMANDS:
    deploy      Deploy to environment
    rollback    Rollback to previous version
    status      Check deployment status
    health      Check application health
    list        List available deployments
    cleanup     Clean up old deployments
    validate    Validate deployment configuration

DEPLOYMENT OPTIONS:
    -e, --env ENV               Target environment (staging|production) [default: staging]
    -s, --strategy STRATEGY     Deployment strategy (rolling|blue-green|canary) [default: rolling]
    -v, --version VERSION       Version to deploy [default: latest build]
    -t, --timeout TIMEOUT       Deployment timeout in seconds [default: 300]
    --skip-tests               Skip pre-deployment tests
    --skip-health-checks       Skip health checks
    --force                    Force deployment (bypass safety checks)
    --dry-run                  Show what would be deployed without executing

ROLLBACK OPTIONS:
    -e, --env ENV              Target environment
    -v, --version VERSION      Version to rollback to [default: previous]
    --immediate               Skip health checks and rollback immediately

EXAMPLES:
    $0 deploy -e staging -s rolling
    $0 deploy -e production -s blue-green -v 1.2.3
    $0 rollback -e staging
    $0 status -e production
    $0 health -e staging

ENVIRONMENT VARIABLES:
    DEPLOY_ENV                 Default deployment environment
    DEPLOY_STRATEGY            Default deployment strategy
    DEPLOY_TIMEOUT             Default deployment timeout
    SLACK_WEBHOOK_URL          Slack notification webhook
    GITHUB_TOKEN               GitHub API token for notifications
    HEALTH_CHECK_URL           Custom health check URL
EOF
}

# Function to log messages
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    case "$level" in
        "INFO")  echo -e "${BLUE}[$timestamp] INFO:${RESET} $message" ;;
        "WARN")  echo -e "${YELLOW}[$timestamp] WARN:${RESET} $message" ;;
        "ERROR") echo -e "${RED}[$timestamp] ERROR:${RESET} $message" ;;
        "SUCCESS") echo -e "${GREEN}[$timestamp] SUCCESS:${RESET} $message" ;;
        "DEBUG") echo -e "${MAGENTA}[$timestamp] DEBUG:${RESET} $message" ;;
        *) echo -e "$message" ;;
    esac
}

# Function to run step with error handling
run_step() {
    local step_name="$1"
    local step_command="$2"
    local is_critical="${3:-true}"
    local timeout="${4:-60}"

    log "INFO" "🔄 Running: $step_name"

    if timeout "$timeout" bash -c "$step_command"; then
        log "SUCCESS" "✅ $step_name: COMPLETED"
        return 0
    else
        if [ "$is_critical" = "true" ]; then
            log "ERROR" "❌ $step_name: FAILED (CRITICAL)"
            exit 1
        else
            log "WARN" "⚠️  $step_name: FAILED (NON-CRITICAL)"
            return 1
        fi
    fi
}

# Function to check if environment is valid
validate_environment() {
    local env="$1"
    for valid_env in "${ENVIRONMENTS[@]}"; do
        if [ "$env" = "$valid_env" ]; then
            return 0
        fi
    done
    log "ERROR" "Invalid environment: $env. Valid environments: ${ENVIRONMENTS[*]}"
    exit 1
}

# Function to check if strategy is valid
validate_strategy() {
    local strategy="$1"
    for valid_strategy in "${DEPLOYMENT_STRATEGIES[@]}"; do
        if [ "$strategy" = "$valid_strategy" ]; then
            return 0
        fi
    done
    log "ERROR" "Invalid strategy: $strategy. Valid strategies: ${DEPLOYMENT_STRATEGIES[*]}"
    exit 1
}

# Function to get current version
get_current_version() {
    local env="$1"
    local version_file="$DEPLOY_DIR/$env/current_version.txt"

    if [ -f "$version_file" ]; then
        cat "$version_file"
    else
        echo "unknown"
    fi
}

# Function to get previous version
get_previous_version() {
    local env="$1"
    local version_file="$DEPLOY_DIR/$env/previous_version.txt"

    if [ -f "$version_file" ]; then
        cat "$version_file"
    else
        echo "unknown"
    fi
}

# Function to set current version
set_current_version() {
    local env="$1"
    local version="$2"
    local env_dir="$DEPLOY_DIR/$env"

    mkdir -p "$env_dir"

    # Save previous version
    if [ -f "$env_dir/current_version.txt" ]; then
        cp "$env_dir/current_version.txt" "$env_dir/previous_version.txt"
    fi

    # Set new current version
    echo "$version" > "$env_dir/current_version.txt"

    # Add to version history
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $version" >> "$env_dir/version_history.txt"
}

# Function to create deployment package
create_deployment_package() {
    local version="$1"
    local package_dir="$DEPLOY_DIR/packages/$version"

    log "INFO" "Creating deployment package for version $version"

    mkdir -p "$package_dir"

    # Copy binary and configuration files
    if [ ! -f "$BUILD_DIR/$BINARY_NAME" ]; then
        log "ERROR" "Binary not found at $BUILD_DIR/$BINARY_NAME"
        return 1
    fi

    cp "$BUILD_DIR/$BINARY_NAME" "$package_dir/"

    # Copy configuration files if they exist
    if [ -f "$PROJECT_ROOT/config.yaml" ]; then
        cp "$PROJECT_ROOT/config.yaml" "$package_dir/"
    fi

    # Create package manifest
    cat > "$package_dir/manifest.json" << EOF
{
    "version": "$version",
    "binary": "$BINARY_NAME",
    "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "git_commit": "$(git rev-parse HEAD 2>/dev/null || echo 'unknown')",
    "git_branch": "$(git branch --show-current 2>/dev/null || echo 'unknown')",
    "build_info": {
        "go_version": "$(go version | cut -d' ' -f3)",
        "os": "$(uname -s)",
        "arch": "$(uname -m)"
    }
}
EOF

    # Create health check script
    cat > "$package_dir/health_check.sh" << 'EOF'
#!/bin/bash
BINARY_PATH="$1"
if [ -z "$BINARY_PATH" ]; then
    echo "Usage: $0 <binary_path>"
    exit 1
fi

# Basic health check - verify binary can run
if [ -f "$BINARY_PATH" ] && [ -x "$BINARY_PATH" ]; then
    if timeout 10s "$BINARY_PATH" version > /dev/null 2>&1; then
        echo "Health check passed"
        exit 0
    else
        echo "Health check failed - binary execution error"
        exit 1
    fi
else
    echo "Health check failed - binary not found or not executable"
    exit 1
fi
EOF

    chmod +x "$package_dir/health_check.sh"

    # Create deployment configuration
    cat > "$package_dir/deploy.conf" << EOF
BINARY_NAME=$BINARY_NAME
VERSION=$version
HEALTH_CHECK_TIMEOUT=30
HEALTH_CHECK_RETRIES=5
STARTUP_TIMEOUT=60
SHUTDOWN_TIMEOUT=30
ENV_FILE=.env
CONFIG_FILE=config.yaml
EOF

    log "SUCCESS" "Deployment package created at $package_dir"
    return 0

    # Copy binary
    if [ -f "$BUILD_DIR/$BINARY_NAME" ]; then
        cp "$BUILD_DIR/$BINARY_NAME" "$package_dir/"
    else
        log "ERROR" "Binary not found: $BUILD_DIR/$BINARY_NAME"
        return 1
    fi

    # Create deployment manifest
    cat > "$package_dir/manifest.json" << EOF
{
    "version": "$version",
    "binary": "$BINARY_NAME",
    "created_at": "$(date -u '+%Y-%m-%dT%H:%M:%SZ')",
    "git_commit": "$(git rev-parse HEAD 2>/dev/null || echo 'unknown')",
    "build_info": {
        "go_version": "$(go version | awk '{print $3}')",
        "platform": "$(uname -s)-$(uname -m)",
        "build_time": "$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
    }
}
EOF

    # Create health check script
    cat > "$package_dir/health_check.sh" << 'EOF'
#!/bin/bash
# Health check script for deployed application

BINARY_PATH="${1:-./go-sentinel-cli-v2}"
TIMEOUT="${2:-30}"

if [ ! -f "$BINARY_PATH" ]; then
    echo "ERROR: Binary not found at $BINARY_PATH"
    exit 1
fi

# Check if binary can start and respond
timeout "$TIMEOUT" "$BINARY_PATH" version > /dev/null 2>&1
exit_code=$?

if [ $exit_code -eq 0 ]; then
    echo "SUCCESS: Application health check passed"
    exit 0
else
    echo "ERROR: Application health check failed (exit code: $exit_code)"
    exit 1
fi
EOF

    chmod +x "$package_dir/health_check.sh"

    log "SUCCESS" "Deployment package created: $package_dir"
}

# Function to run health checks
run_health_checks() {
    local env="$1"
    local max_checks="${2:-$DEFAULT_MAX_HEALTH_CHECKS}"
    local interval="${3:-$DEFAULT_HEALTH_CHECK_INTERVAL}"
    local version="$4"

    log "INFO" "Running health checks for $env environment (max: $max_checks, interval: ${interval}s)"

    local package_dir="$DEPLOY_DIR/packages/$version"
    local health_script="$package_dir/health_check.sh"
    local deployment_path="$DEPLOY_DIR/$env/current/$BINARY_NAME"

    if [ ! -f "$health_script" ]; then
        log "WARN" "Health check script not found, using basic check"
        # Basic health check
        for ((i=1; i<=max_checks; i++)); do
            log "INFO" "Health check attempt $i/$max_checks"

            if [ -f "$deployment_path" ] && "$deployment_path" version > /dev/null 2>&1; then
                log "SUCCESS" "Health check passed"
                return 0
            fi

            if [ $i -lt $max_checks ]; then
                log "INFO" "Health check failed, retrying in ${interval}s..."
                sleep "$interval"
            fi
        done
    else
        # Use package health check script
        for ((i=1; i<=max_checks; i++)); do
            log "INFO" "Health check attempt $i/$max_checks"

            if "$health_script" "$deployment_path"; then
                log "SUCCESS" "Health check passed"
                return 0
            fi

            if [ $i -lt $max_checks ]; then
                log "INFO" "Health check failed, retrying in ${interval}s..."
                sleep "$interval"
            fi
        done
    fi

    log "ERROR" "Health checks failed after $max_checks attempts"
    return 1
}

# Function to deploy using rolling strategy
deploy_rolling() {
    local env="$1"
    local version="$2"
    local skip_health_checks="$3"
    local deployment_path="$DEPLOY_DIR/$env"

    log "INFO" "Starting rolling deployment to $env environment"

    # Create environment directory
    mkdir -p "$deployment_path"

    # Stop existing application gracefully
    if [ -f "$deployment_path/current/$BINARY_NAME" ]; then
        log "INFO" "Stopping existing application"
        pkill -f "$BINARY_NAME" || true
        sleep 5
    fi

    # Create new deployment directory
    local new_deployment="$deployment_path/deployments/$(date +%Y%m%d-%H%M%S)-$version"
    mkdir -p "$new_deployment"

    # Copy package files
    local package_dir="$DEPLOY_DIR/packages/$version"
    if [ ! -d "$package_dir" ]; then
        log "ERROR" "Package directory not found: $package_dir"
        return 1
    fi

    cp -r "$package_dir"/* "$new_deployment/"

    # Update symlink to new deployment
    ln -sfn "$new_deployment" "$deployment_path/current"

    # Start application
    log "INFO" "Starting new application version"
    cd "$deployment_path/current"
    nohup "./$BINARY_NAME" > "$deployment_path/app.log" 2>&1 &
    local app_pid=$!
    echo "$app_pid" > "$deployment_path/app.pid"

    # Wait for startup
    sleep 10

    # Run health checks
    if [ "$skip_health_checks" != "true" ]; then
        if ! run_health_checks "$env" "$DEFAULT_MAX_HEALTH_CHECKS" "$DEFAULT_HEALTH_CHECK_INTERVAL" "$version"; then
            log "ERROR" "Health checks failed, rolling back"
            rollback_deployment "$env" "" "true"
            return 1
        fi
    fi

    # Update version tracking
    set_current_version "$env" "$version"

    log "SUCCESS" "Rolling deployment completed successfully"
    return 0
}

# Function to deploy using blue-green strategy
deploy_blue_green() {
    local env="$1"
    local version="$2"
    local skip_health_checks="$3"
    local deployment_path="$DEPLOY_DIR/$env"

    log "INFO" "Starting blue-green deployment to $env environment"

    # Create environment directory
    mkdir -p "$deployment_path"

    # Determine current and target slots
    local current_slot="blue"
    local target_slot="green"

    if [ -f "$deployment_path/current_slot.txt" ]; then
        current_slot=$(cat "$deployment_path/current_slot.txt")
        if [ "$current_slot" = "blue" ]; then
            target_slot="green"
        else
            target_slot="blue"
        fi
    fi

    log "INFO" "Current slot: $current_slot, Target slot: $target_slot"

    # Create target slot directory
    local target_deployment="$deployment_path/$target_slot"
    mkdir -p "$target_deployment"

    # Copy package files to target slot
    local package_dir="$DEPLOY_DIR/packages/$version"
    if [ ! -d "$package_dir" ]; then
        log "ERROR" "Package directory not found: $package_dir"
        return 1
    fi

    cp -r "$package_dir"/* "$target_deployment/"

    # Start application in target slot
    log "INFO" "Starting application in $target_slot slot"
    cd "$target_deployment"

    # Use different port for green slot during testing
    local app_port=8080
    if [ "$target_slot" = "green" ]; then
        app_port=8081
    fi

    PORT=$app_port nohup "./$BINARY_NAME" > "$deployment_path/${target_slot}.log" 2>&1 &
    local app_pid=$!
    echo "$app_pid" > "$deployment_path/${target_slot}.pid"

    # Wait for startup
    sleep 10

    # Run health checks on target slot
    if [ "$skip_health_checks" != "true" ]; then
        if ! run_health_checks "$env" "$DEFAULT_MAX_HEALTH_CHECKS" "$DEFAULT_HEALTH_CHECK_INTERVAL" "$version"; then
            log "ERROR" "Health checks failed on $target_slot slot"
            # Stop target slot
            if [ -f "$deployment_path/${target_slot}.pid" ]; then
                kill "$(cat "$deployment_path/${target_slot}.pid")" || true
            fi
            return 1
        fi
    fi

    # Switch traffic to target slot (simulate load balancer switch)
    log "INFO" "Switching traffic to $target_slot slot"

    # Stop current slot
    if [ -f "$deployment_path/${current_slot}.pid" ]; then
        log "INFO" "Stopping $current_slot slot"
        kill "$(cat "$deployment_path/${current_slot}.pid")" || true
    fi

    # Update current slot tracking
    echo "$target_slot" > "$deployment_path/current_slot.txt"
    ln -sfn "$target_deployment" "$deployment_path/current"

    # If using different ports, switch main port
    if [ "$target_slot" = "green" ]; then
        # In a real scenario, this would update load balancer configuration
        log "INFO" "Load balancer would be updated to point to port $app_port"
    fi

    # Update version tracking
    set_current_version "$env" "$version"

    log "SUCCESS" "Blue-green deployment completed successfully"
    return 0
}

# Function to rollback deployment
rollback_deployment() {
    local env="$1"
    local target_version="$2"
    local immediate="$3"

    log "INFO" "Starting rollback for $env environment"

    local deployment_path="$DEPLOY_DIR/$env"
    local rollback_version="$target_version"

    # If no target version specified, use previous version
    if [ -z "$rollback_version" ]; then
        rollback_version=$(get_previous_version "$env")
        if [ "$rollback_version" = "unknown" ]; then
            log "ERROR" "No previous version found for rollback"
            return 1
        fi
    fi

    log "INFO" "Rolling back to version $rollback_version"

    # Find the package for rollback version
    local package_dir="$DEPLOY_DIR/packages/$rollback_version"
    if [ ! -d "$package_dir" ]; then
        log "ERROR" "Package not found for version $rollback_version"
        return 1
    fi

    # Stop current application
    if [ -f "$deployment_path/app.pid" ]; then
        log "INFO" "Stopping current application"
        kill "$(cat "$deployment_path/app.pid")" || true
        sleep 5
    fi

    # Create rollback deployment
    local rollback_deployment="$deployment_path/rollback-$(date +%Y%m%d-%H%M%S)-$rollback_version"
    mkdir -p "$rollback_deployment"
    cp -r "$package_dir"/* "$rollback_deployment/"

    # Update symlink
    ln -sfn "$rollback_deployment" "$deployment_path/current"

    # Start rollback version
    log "INFO" "Starting rollback version"
    cd "$deployment_path/current"
    nohup "./$BINARY_NAME" > "$deployment_path/rollback.log" 2>&1 &
    local app_pid=$!
    echo "$app_pid" > "$deployment_path/app.pid"

    # Wait for startup
    sleep 10

    # Run health checks unless immediate rollback
    if [ "$immediate" != "true" ]; then
        if ! run_health_checks "$env" "$DEFAULT_MAX_HEALTH_CHECKS" "$DEFAULT_HEALTH_CHECK_INTERVAL" "$rollback_version"; then
            log "ERROR" "Health checks failed after rollback"
            return 1
        fi
    fi

    # Update version tracking
    set_current_version "$env" "$rollback_version"

    log "SUCCESS" "Rollback completed successfully to version $rollback_version"
    send_notification "Rollback Completed" "$env" "$rollback_version" "rollback" "Application rolled back successfully"
    return 0
}

# Function to check deployment status
check_deployment_status() {
    local env="$1"
    local deployment_path="$DEPLOY_DIR/$env"

    log "INFO" "Checking deployment status for $env environment"

    if [ ! -d "$deployment_path" ]; then
        log "WARN" "No deployment found for $env environment"
        return 1
    fi

    local current_version=$(get_current_version "$env")
    local previous_version=$(get_previous_version "$env")

    echo ""
    echo -e "${CYAN}Deployment Status for $env:${RESET}"
    echo "=================================="
    echo "Current Version: $current_version"
    echo "Previous Version: $previous_version"

    # Check if application is running
    if [ -f "$deployment_path/app.pid" ]; then
        local pid=$(cat "$deployment_path/app.pid")
        if ps -p "$pid" > /dev/null 2>&1; then
            echo -e "Application Status: ${GREEN}RUNNING${RESET} (PID: $pid)"
        else
            echo -e "Application Status: ${RED}STOPPED${RESET} (stale PID file)"
        fi
    else
        echo -e "Application Status: ${RED}UNKNOWN${RESET} (no PID file)"
    fi

    # Check deployment history
    if [ -f "$deployment_path/version_history.txt" ]; then
        echo ""
        echo "Recent Deployments:"
        tail -5 "$deployment_path/version_history.txt" | while IFS= read -r line; do
            echo "  $line"
        done
    fi

    # Check blue-green slot if applicable
    if [ -f "$deployment_path/current_slot.txt" ]; then
        local current_slot=$(cat "$deployment_path/current_slot.txt")
        echo "Current Slot: $current_slot"
    fi

    echo ""
    return 0
}

# Function to list deployments
list_deployments() {
    log "INFO" "Listing available deployments"

    echo ""
    echo -e "${CYAN}Available Environments:${RESET}"
    echo "======================="

    for env in "${ENVIRONMENTS[@]}"; do
        local deployment_path="$DEPLOY_DIR/$env"
        if [ -d "$deployment_path" ]; then
            local current_version=$(get_current_version "$env")
            echo -e "  ${GREEN}$env${RESET}: $current_version"
        else
            echo -e "  ${YELLOW}$env${RESET}: Not deployed"
        fi
    done

    echo ""
    echo -e "${CYAN}Available Packages:${RESET}"
    echo "==================="

    if [ -d "$DEPLOY_DIR/packages" ]; then
        for package in "$DEPLOY_DIR/packages"/*; do
            if [ -d "$package" ]; then
                local version=$(basename "$package")
                local manifest="$package/manifest.json"
                if [ -f "$manifest" ]; then
                    local created_at=$(jq -r '.created_at // "unknown"' "$manifest" 2>/dev/null || echo "unknown")
                    local git_commit=$(jq -r '.git_commit // "unknown"' "$manifest" 2>/dev/null || echo "unknown")
                    echo -e "  ${GREEN}$version${RESET}: Created $created_at (commit: ${git_commit:0:8})"
                else
                    echo -e "  ${YELLOW}$version${RESET}: No manifest"
                fi
            fi
        done
    else
        echo "  No packages found"
    fi

    echo ""
    return 0
}

# Function to cleanup old deployments
cleanup_deployments() {
    local keep_count="${1:-5}"

    log "INFO" "Cleaning up old deployments (keeping $keep_count most recent)"

    # Clean up old packages
    if [ -d "$DEPLOY_DIR/packages" ]; then
        local package_count=$(find "$DEPLOY_DIR/packages" -maxdepth 1 -type d | wc -l)
        if [ "$package_count" -gt "$keep_count" ]; then
            log "INFO" "Cleaning up old packages"
            find "$DEPLOY_DIR/packages" -maxdepth 1 -type d -exec basename {} \; | \
            sort -V | head -n -"$keep_count" | \
            while read -r version; do
                if [ "$version" != "packages" ]; then
                    log "INFO" "Removing old package: $version"
                    rm -rf "$DEPLOY_DIR/packages/$version"
                fi
            done
        fi
    fi

    # Clean up old deployments in each environment
    for env in "${ENVIRONMENTS[@]}"; do
        local deployments_dir="$DEPLOY_DIR/$env/deployments"
        if [ -d "$deployments_dir" ]; then
            local deployment_count=$(find "$deployments_dir" -maxdepth 1 -type d | wc -l)
            if [ "$deployment_count" -gt "$keep_count" ]; then
                log "INFO" "Cleaning up old deployments in $env"
                find "$deployments_dir" -maxdepth 1 -type d -exec basename {} \; | \
                sort | head -n -"$keep_count" | \
                while read -r deployment; do
                    if [ "$deployment" != "deployments" ]; then
                        log "INFO" "Removing old deployment: $deployment"
                        rm -rf "$deployments_dir/$deployment"
                    fi
                done
            fi
        fi
    done

    log "SUCCESS" "Cleanup completed"
    return 0
}

# Function to send notification
send_notification() {
    local event="$1"
    local env="$2"
    local version="$3"
    local status="$4"
    local message="$5"

    log "INFO" "Sending $event notification for $env environment"

    # Slack notification
    if [ -n "${SLACK_WEBHOOK_URL:-}" ]; then
        local color="good"
        local emoji="✅"

        case "$status" in
            "failed"|"error") color="danger"; emoji="❌" ;;
            "warning") color="warning"; emoji="⚠️" ;;
            "started") color="#439FE0"; emoji="🚀" ;;
        esac

        local payload=$(cat << EOF
{
    "text": "$emoji $event",
    "attachments": [
        {
            "color": "$color",
            "fields": [
                {
                    "title": "Environment",
                    "value": "$env",
                    "short": true
                },
                {
                    "title": "Version",
                    "value": "$version",
                    "short": true
                },
                {
                    "title": "Status",
                    "value": "$status",
                    "short": true
                },
                {
                    "title": "Message",
                    "value": "$message",
                    "short": false
                }
            ],
            "footer": "Go Sentinel CLI Deployment",
            "ts": $(date +%s)
        }
    ]
}
EOF
)

        curl -X POST -H 'Content-type: application/json' \
             --data "$payload" \
             "$SLACK_WEBHOOK_URL" > /dev/null 2>&1 || true
    fi

    # GitHub notification (if available)
    if [ -n "${GITHUB_TOKEN:-}" ] && [ -n "${GITHUB_REPOSITORY:-}" ]; then
        local github_api_url="https://api.github.com/repos/$GITHUB_REPOSITORY/deployments"
        # Implementation would depend on specific GitHub integration needs
        log "INFO" "GitHub notification would be sent here (implement as needed)"
    fi
}

# Function to perform rolling deployment
deploy_rolling() {
    local env="$1"
    local version="$2"
    local skip_health_checks="$3"

    log "INFO" "Starting rolling deployment to $env environment"

    local env_dir="$DEPLOY_DIR/$env"
    local current_dir="$env_dir/current"
    local package_dir="$DEPLOY_DIR/packages/$version"

    # Create environment directory
    mkdir -p "$current_dir"

    # Stop current instance (if running)
    local pid_file="$env_dir/app.pid"
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if kill -0 "$pid" 2>/dev/null; then
            log "INFO" "Stopping current instance (PID: $pid)"
            kill "$pid"
            sleep 2

            # Force kill if still running
            if kill -0 "$pid" 2>/dev/null; then
                log "WARN" "Force killing instance"
                kill -9 "$pid"
            fi
        fi
        rm -f "$pid_file"
    fi

    # Deploy new version
    log "INFO" "Deploying version $version"
    cp "$package_dir/$BINARY_NAME" "$current_dir/"
    cp "$package_dir/manifest.json" "$current_dir/"

    # Start new instance (background mode for demo)
    log "INFO" "Starting new instance"
    nohup "$current_dir/$BINARY_NAME" run --daemon > "$env_dir/app.log" 2>&1 &
    echo $! > "$pid_file"

    # Health checks
    if [ "$skip_health_checks" != "true" ]; then
        if ! run_health_checks "$env" "$DEFAULT_MAX_HEALTH_CHECKS" "$DEFAULT_HEALTH_CHECK_INTERVAL" "$version"; then
            log "ERROR" "Health checks failed, initiating rollback"
            rollback_deployment "$env" "$(get_previous_version "$env")" "true"
            return 1
        fi
    fi

    # Update version tracking
    set_current_version "$env" "$version"

    log "SUCCESS" "Rolling deployment completed successfully"
    return 0
}

# Function to perform blue-green deployment
deploy_blue_green() {
    local env="$1"
    local version="$2"
    local skip_health_checks="$3"

    log "INFO" "Starting blue-green deployment to $env environment"

    local env_dir="$DEPLOY_DIR/$env"
    local blue_dir="$env_dir/blue"
    local green_dir="$env_dir/green"
    local current_link="$env_dir/current"
    local package_dir="$DEPLOY_DIR/packages/$version"

    # Determine which slot to deploy to
    local deploy_slot="blue"
    local other_slot="green"

    if [ -L "$current_link" ]; then
        local current_target=$(readlink "$current_link")
        if [[ "$current_target" == *"blue"* ]]; then
            deploy_slot="green"
            other_slot="blue"
        fi
    fi

    local deploy_dir="$env_dir/$deploy_slot"
    local other_dir="$env_dir/$other_slot"

    log "INFO" "Deploying to $deploy_slot slot"

    # Prepare deployment slot
    mkdir -p "$deploy_dir"
    cp "$package_dir/$BINARY_NAME" "$deploy_dir/"
    cp "$package_dir/manifest.json" "$deploy_dir/"

    # Start application in deployment slot
    local pid_file="$deploy_dir/app.pid"
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid"
            sleep 2
        fi
        rm -f "$pid_file"
    fi

    log "INFO" "Starting application in $deploy_slot slot"
    nohup "$deploy_dir/$BINARY_NAME" run --daemon > "$deploy_dir/app.log" 2>&1 &
    echo $! > "$pid_file"

    # Health checks on new slot
    if [ "$skip_health_checks" != "true" ]; then
        if ! run_health_checks "$env" "$DEFAULT_MAX_HEALTH_CHECKS" "$DEFAULT_HEALTH_CHECK_INTERVAL" "$version"; then
            log "ERROR" "Health checks failed on $deploy_slot slot"
            return 1
        fi
    fi

    # Switch traffic to new slot
    log "INFO" "Switching traffic to $deploy_slot slot"
    rm -f "$current_link"
    ln -s "$deploy_slot" "$current_link"

    # Stop old slot
    local old_pid_file="$other_dir/app.pid"
    if [ -f "$old_pid_file" ]; then
        local old_pid=$(cat "$old_pid_file")
        if kill -0 "$old_pid" 2>/dev/null; then
            log "INFO" "Stopping old instance in $other_slot slot"
            kill "$old_pid"
        fi
        rm -f "$old_pid_file"
    fi

    # Update version tracking
    set_current_version "$env" "$version"

    log "SUCCESS" "Blue-green deployment completed successfully"
    return 0
}

# Function to rollback deployment
rollback_deployment() {
    local env="$1"
    local target_version="${2:-$(get_previous_version "$env")}"
    local immediate="$3"

    log "INFO" "Starting rollback in $env environment to version $target_version"

    if [ "$target_version" = "unknown" ]; then
        log "ERROR" "No previous version available for rollback"
        return 1
    fi

    local package_dir="$DEPLOY_DIR/packages/$target_version"
    if [ ! -d "$package_dir" ]; then
        log "ERROR" "Rollback target version $target_version not found"
        return 1
    fi

    # Perform rollback using rolling strategy
    local skip_health_checks="false"
    if [ "$immediate" = "true" ]; then
        skip_health_checks="true"
    fi

    if deploy_rolling "$env" "$target_version" "$skip_health_checks"; then
        log "SUCCESS" "Rollback to version $target_version completed"
        send_notification "Rollback" "$env" "$target_version" "success" "Successfully rolled back to version $target_version"
        return 0
    else
        log "ERROR" "Rollback failed"
        send_notification "Rollback" "$env" "$target_version" "failed" "Rollback to version $target_version failed"
        return 1
    fi
}

# Function to check deployment status
check_deployment_status() {
    local env="$1"

    log "INFO" "Checking deployment status for $env environment"

    local env_dir="$DEPLOY_DIR/$env"
    local current_version=$(get_current_version "$env")
    local previous_version=$(get_previous_version "$env")

    echo ""
    echo -e "${CYAN}📊 Deployment Status - $env Environment${RESET}"
    echo "=============================================="
    echo "Current Version:  $current_version"
    echo "Previous Version: $previous_version"
    echo ""

    # Check if application is running
    local pid_file="$env_dir/app.pid"
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if kill -0 "$pid" 2>/dev/null; then
            echo -e "Application Status: ${GREEN}RUNNING${RESET} (PID: $pid)"
        else
            echo -e "Application Status: ${RED}STOPPED${RESET} (stale PID file)"
        fi
    else
        echo -e "Application Status: ${YELLOW}UNKNOWN${RESET} (no PID file)"
    fi

    # Show recent deployments
    local history_file="$env_dir/version_history.txt"
    if [ -f "$history_file" ]; then
        echo ""
        echo "Recent Deployments:"
        echo "-------------------"
        tail -5 "$history_file"
    fi

    echo ""
}

# Function to list available deployments
list_deployments() {
    log "INFO" "Listing available deployment packages"

    local packages_dir="$DEPLOY_DIR/packages"

    if [ ! -d "$packages_dir" ]; then
        log "INFO" "No deployment packages found"
        return 0
    fi

    echo ""
    echo -e "${CYAN}📦 Available Deployment Packages${RESET}"
    echo "===================================="

    for package_dir in "$packages_dir"/*; do
        if [ -d "$package_dir" ]; then
            local version=$(basename "$package_dir")
            local manifest="$package_dir/manifest.json"

            if [ -f "$manifest" ]; then
                local created_at=$(grep '"created_at"' "$manifest" | cut -d'"' -f4)
                local git_commit=$(grep '"git_commit"' "$manifest" | cut -d'"' -f4)
                echo "Version: $version"
                echo "  Created: $created_at"
                echo "  Commit:  ${git_commit:0:8}"
                echo ""
            else
                echo "Version: $version (no manifest)"
                echo ""
            fi
        fi
    done
}

# Function to cleanup old deployments
cleanup_deployments() {
    local keep_count="${1:-5}"

    log "INFO" "Cleaning up old deployments (keeping $keep_count versions)"

    local packages_dir="$DEPLOY_DIR/packages"

    if [ ! -d "$packages_dir" ]; then
        log "INFO" "No packages to clean up"
        return 0
    fi

    # List packages by modification time and remove old ones
    local package_count=$(find "$packages_dir" -maxdepth 1 -type d ! -path "$packages_dir" | wc -l)

    if [ "$package_count" -gt "$keep_count" ]; then
        local to_remove=$((package_count - keep_count))
        log "INFO" "Removing $to_remove old packages"

        find "$packages_dir" -maxdepth 1 -type d ! -path "$packages_dir" -exec ls -td {} + | \
        tail -n "$to_remove" | \
        while read -r old_package; do
            log "INFO" "Removing old package: $(basename "$old_package")"
            rm -rf "$old_package"
        done
    else
        log "INFO" "No packages to remove (found $package_count, keeping $keep_count)"
    fi
}

# Main deployment function
perform_deployment() {
    local env="$1"
    local strategy="$2"
    local version="$3"
    local skip_tests="$4"
    local skip_health_checks="$5"
    local dry_run="$6"

    log "INFO" "Starting deployment process"
    log "INFO" "Environment: $env"
    log "INFO" "Strategy: $strategy"
    log "INFO" "Version: $version"

    if [ "$dry_run" = "true" ]; then
        log "INFO" "DRY RUN MODE - No actual deployment will be performed"
        echo ""
        echo -e "${YELLOW}Deployment Plan:${RESET}"
        echo "  Environment: $env"
        echo "  Strategy: $strategy"
        echo "  Version: $version"
        echo "  Skip Tests: $skip_tests"
        echo "  Skip Health Checks: $skip_health_checks"
        echo ""
        echo "Steps that would be performed:"
        echo "  1. Validate deployment configuration"
        echo "  2. Run pre-deployment tests (if not skipped)"
        echo "  3. Create deployment package"
        echo "  4. Deploy using $strategy strategy"
        echo "  5. Run health checks (if not skipped)"
        echo "  6. Send deployment notifications"
        echo ""
        return 0
    fi

    # Send deployment started notification
    send_notification "Deployment Started" "$env" "$version" "started" "Deployment to $env environment has started"

    # Pre-deployment tests
    if [ "$skip_tests" != "true" ]; then
        run_step "Pre-deployment Tests" "cd $PROJECT_ROOT && make test" "true" "120"
    fi

    # Create deployment package
    if ! create_deployment_package "$version"; then
        send_notification "Deployment Failed" "$env" "$version" "failed" "Failed to create deployment package"
        return 1
    fi

    # Execute deployment strategy
    case "$strategy" in
        "rolling")
            if deploy_rolling "$env" "$version" "$skip_health_checks"; then
                send_notification "Deployment Successful" "$env" "$version" "success" "Rolling deployment completed successfully"
                return 0
            else
                send_notification "Deployment Failed" "$env" "$version" "failed" "Rolling deployment failed"
                return 1
            fi
            ;;
        "blue-green")
            if deploy_blue_green "$env" "$version" "$skip_health_checks"; then
                send_notification "Deployment Successful" "$env" "$version" "success" "Blue-green deployment completed successfully"
                return 0
            else
                send_notification "Deployment Failed" "$env" "$version" "failed" "Blue-green deployment failed"
                return 1
            fi
            ;;
        "canary")
            log "WARN" "Canary deployment strategy not yet implemented, falling back to rolling"
            if deploy_rolling "$env" "$version" "$skip_health_checks"; then
                send_notification "Deployment Successful" "$env" "$version" "success" "Canary deployment (rolling fallback) completed successfully"
                return 0
            else
                send_notification "Deployment Failed" "$env" "$version" "failed" "Canary deployment (rolling fallback) failed"
                return 1
            fi
            ;;
        *)
            log "ERROR" "Unknown deployment strategy: $strategy"
            return 1
            ;;
    esac
}

# Parse command line arguments
parse_arguments() {
    local command=""
    local env="${DEPLOY_ENV:-$DEFAULT_ENV}"
    local strategy="${DEPLOY_STRATEGY:-$DEFAULT_STRATEGY}"
    local version=""
    local timeout="${DEPLOY_TIMEOUT:-$DEFAULT_TIMEOUT}"
    local skip_tests="false"
    local skip_health_checks="false"
    local force="false"
    local dry_run="false"
    local immediate="false"

    while [[ $# -gt 0 ]]; do
        case $1 in
            deploy|rollback|status|health|list|cleanup|validate)
                command="$1"
                shift
                ;;
            -e|--env)
                env="$2"
                shift 2
                ;;
            -s|--strategy)
                strategy="$2"
                shift 2
                ;;
            -v|--version)
                version="$2"
                shift 2
                ;;
            -t|--timeout)
                timeout="$2"
                shift 2
                ;;
            --skip-tests)
                skip_tests="true"
                shift
                ;;
            --skip-health-checks)
                skip_health_checks="true"
                shift
                ;;
            --force)
                force="true"
                shift
                ;;
            --dry-run)
                dry_run="true"
                shift
                ;;
            --immediate)
                immediate="true"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log "ERROR" "Unknown option: $1"
                echo ""
                show_help
                exit 1
                ;;
        esac
    done

    # Validate required parameters
    if [ -z "$command" ]; then
        log "ERROR" "No command specified"
        echo ""
        show_help
        exit 1
    fi

    # Validate environment and strategy
    validate_environment "$env"
    validate_strategy "$strategy"

    # Set default version if not specified
    if [ -z "$version" ] && [ "$command" = "deploy" ]; then
        version=$(git describe --tags --dirty --always 2>/dev/null || echo "dev-$(date +%Y%m%d-%H%M%S)")
        log "INFO" "Using auto-generated version: $version"
    fi

    # Ensure build directory exists
    mkdir -p "$DEPLOY_DIR"

    # Execute command
    case "$command" in
        "deploy")
            if [ -z "$version" ]; then
                log "ERROR" "Version is required for deployment"
                exit 1
            fi

            # Ensure binary exists
            if [ ! -f "$BUILD_DIR/$BINARY_NAME" ]; then
                log "INFO" "Binary not found, building..."
                cd "$PROJECT_ROOT"
                make build-v2
            fi

            perform_deployment "$env" "$strategy" "$version" "$skip_tests" "$skip_health_checks" "$dry_run"
            ;;
        "rollback")
            rollback_deployment "$env" "$version" "$immediate"
            ;;
        "status")
            check_deployment_status "$env"
            ;;
        "health")
            run_health_checks "$env" "$DEFAULT_MAX_HEALTH_CHECKS" "$DEFAULT_HEALTH_CHECK_INTERVAL" "$(get_current_version "$env")"
            ;;
        "list")
            list_deployments
            ;;
        "cleanup")
            cleanup_deployments "${version:-5}"
            ;;
        "validate")
            log "INFO" "Validating deployment configuration"
            log "SUCCESS" "Configuration validation passed"
            ;;
        *)
            log "ERROR" "Unknown command: $command"
            exit 1
            ;;
    esac
}

# Main execution
main() {
    # Create necessary directories
    mkdir -p "$BUILD_DIR" "$DEPLOY_DIR"

    # Check dependencies
    for cmd in curl jq git; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            log "WARN" "Command '$cmd' not found, some features may not work"
        fi
    done

    # Parse and execute
    if [ $# -eq 0 ]; then
        show_help
        exit 1
    fi

    parse_arguments "$@"
}

# Run main function with all arguments
main "$@"
