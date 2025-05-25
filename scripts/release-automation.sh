#!/bin/bash

# Release Automation Script for Go Sentinel CLI
# Handles semantic versioning, changelog generation, and multi-platform distribution

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
RESET='\033[0m'

# Configuration
BINARY_NAME="go-sentinel-cli"
BUILD_DIR="build"
DIST_DIR="$BUILD_DIR/dist"
CHANGELOG_FILE="CHANGELOG.md"
VERSION_FILE="VERSION"

# Supported platforms
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to get current version
get_current_version() {
    if [ -f "$VERSION_FILE" ]; then
        cat "$VERSION_FILE"
    elif git describe --tags --abbrev=0 2>/dev/null; then
        git describe --tags --abbrev=0 | sed 's/^v//'
    else
        echo "0.0.0"
    fi
}

# Function to increment version
increment_version() {
    local version="$1"
    local type="$2"

    local major minor patch
    IFS='.' read -r major minor patch <<< "$version"

    case "$type" in
        "major")
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        "minor")
            minor=$((minor + 1))
            patch=0
            ;;
        "patch")
            patch=$((patch + 1))
            ;;
        *)
            echo "Invalid version type: $type" >&2
            exit 1
            ;;
    esac

    echo "$major.$minor.$patch"
}

# Function to validate version format
validate_version() {
    local version="$1"
    if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo "Invalid version format: $version (expected: x.y.z)" >&2
        exit 1
    fi
}

# Function to check if version already exists
version_exists() {
    local version="$1"
    git tag | grep -q "^v$version$"
}

# Function to generate changelog entry
generate_changelog_entry() {
    local version="$1"
    local date="$(date +%Y-%m-%d)"

    echo "## [$version] - $date"
    echo ""

    # Get commits since last tag
    local last_tag
    if last_tag=$(git describe --tags --abbrev=0 2>/dev/null); then
        echo "### Changes since $last_tag"
        echo ""
        git log "$last_tag..HEAD" --pretty=format:"- %s" --no-merges | head -20
    else
        echo "### Initial Release"
        echo ""
        echo "- Initial release of Go Sentinel CLI"
    fi

    echo ""
    echo ""
}

# Function to update changelog
update_changelog() {
    local version="$1"
    local temp_file=$(mktemp)

    if [ -f "$CHANGELOG_FILE" ]; then
        # Insert new entry at the top
        {
            echo "# Changelog"
            echo ""
            echo "All notable changes to this project will be documented in this file."
            echo ""
            generate_changelog_entry "$version"
            tail -n +5 "$CHANGELOG_FILE" 2>/dev/null || true
        } > "$temp_file"
    else
        # Create new changelog
        {
            echo "# Changelog"
            echo ""
            echo "All notable changes to this project will be documented in this file."
            echo ""
            generate_changelog_entry "$version"
        } > "$temp_file"
    fi

    mv "$temp_file" "$CHANGELOG_FILE"
}

# Function to build binary for platform
build_binary() {
    local platform="$1"
    local version="$2"

    local goos goarch
    IFS='/' read -r goos goarch <<< "$platform"

    local binary_name="$BINARY_NAME-$version-$goos-$goarch"
    if [ "$goos" = "windows" ]; then
        binary_name="$binary_name.exe"
    fi

    local output_path="$DIST_DIR/$binary_name"

    echo -e "${BLUE}🔨 Building $platform...${RESET}"

    GOOS="$goos" GOARCH="$goarch" go build \
        -ldflags="-s -w -X main.version=$version -X main.commit=$(git rev-parse HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        -o "$output_path" \
        ./cmd/go-sentinel-cli

    if [ -f "$output_path" ]; then
        echo -e "${GREEN}✅ Built: $output_path${RESET}"

        # Calculate checksum
        if command_exists "sha256sum"; then
            sha256sum "$output_path" > "$output_path.sha256"
        elif command_exists "shasum"; then
            shasum -a 256 "$output_path" > "$output_path.sha256"
        fi
    else
        echo -e "${RED}❌ Failed to build $platform${RESET}"
        exit 1
    fi
}

# Function to create release archive
create_release_archive() {
    local version="$1"
    local archive_dir="$DIST_DIR/go-sentinel-cli-$version"

    echo -e "${BLUE}📦 Creating release archive...${RESET}"

    # Create archive directory
    mkdir -p "$archive_dir"

    # Copy binaries
    cp "$DIST_DIR"/*-"$version"-* "$archive_dir/" 2>/dev/null || true

    # Copy documentation
    cp README.md "$archive_dir/" 2>/dev/null || true
    cp LICENSE "$archive_dir/" 2>/dev/null || true
    cp "$CHANGELOG_FILE" "$archive_dir/" 2>/dev/null || true

    # Create installation script
    cat > "$archive_dir/install.sh" << 'EOF'
#!/bin/bash
# Installation script for Go Sentinel CLI

set -euo pipefail

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Find binary
BINARY_PATTERN="go-sentinel-cli-*-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY_PATTERN="${BINARY_PATTERN}.exe"
fi

BINARY=$(ls $BINARY_PATTERN 2>/dev/null | head -1)
if [ -z "$BINARY" ]; then
    echo "No binary found for $OS/$ARCH"
    exit 1
fi

# Install binary
INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "$INSTALL_DIR"
cp "$BINARY" "$INSTALL_DIR/go-sentinel-cli"
chmod +x "$INSTALL_DIR/go-sentinel-cli"

echo "✅ Go Sentinel CLI installed to $INSTALL_DIR/go-sentinel-cli"
echo "Make sure $INSTALL_DIR is in your PATH"
EOF

    chmod +x "$archive_dir/install.sh"

    # Create archive
    cd "$DIST_DIR"
    tar -czf "go-sentinel-cli-$version.tar.gz" "go-sentinel-cli-$version/"

    # Create zip for Windows users
    if command_exists "zip"; then
        zip -r "go-sentinel-cli-$version.zip" "go-sentinel-cli-$version/" >/dev/null
    fi

    cd - >/dev/null

    echo -e "${GREEN}✅ Release archive created: $DIST_DIR/go-sentinel-cli-$version.tar.gz${RESET}"
}

# Function to generate release notes
generate_release_notes() {
    local version="$1"
    local notes_file="$DIST_DIR/release-notes-$version.md"

    echo -e "${BLUE}📝 Generating release notes...${RESET}"

    cat > "$notes_file" << EOF
# Go Sentinel CLI v$version

## What's New

$(generate_changelog_entry "$version" | tail -n +3)

## Installation

### Quick Install (Linux/macOS)
\`\`\`bash
curl -sSL https://github.com/newbpydev/go-sentinel/releases/download/v$version/go-sentinel-cli-$version.tar.gz | tar -xz
cd go-sentinel-cli-$version
./install.sh
\`\`\`

### Manual Installation
1. Download the appropriate binary for your platform from the assets below
2. Extract the archive
3. Move the binary to a directory in your PATH
4. Make it executable: \`chmod +x go-sentinel-cli\`

### Go Install
\`\`\`bash
go install github.com/newbpydev/go-sentinel/cmd/go-sentinel-cli@v$version
\`\`\`

## Supported Platforms

- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Checksums

All binaries include SHA256 checksums for verification.

## Documentation

- [README](https://github.com/newbpydev/go-sentinel/blob/v$version/README.md)
- [Changelog](https://github.com/newbpydev/go-sentinel/blob/v$version/CHANGELOG.md)

---

**Full Changelog**: https://github.com/newbpydev/go-sentinel/compare/v$(get_current_version)...v$version
EOF

    echo -e "${GREEN}✅ Release notes generated: $notes_file${RESET}"
}

# Function to run pre-release checks
run_pre_release_checks() {
    echo -e "${CYAN}🔍 Running pre-release checks...${RESET}"

    # Check if working directory is clean
    if ! git diff-index --quiet HEAD --; then
        echo -e "${RED}❌ Working directory is not clean. Please commit or stash changes.${RESET}"
        exit 1
    fi

    # Check if on main/master branch
    local current_branch=$(git branch --show-current)
    if [[ "$current_branch" != "main" && "$current_branch" != "master" ]]; then
        echo -e "${YELLOW}⚠️  Not on main/master branch (current: $current_branch)${RESET}"
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi

    # Run quality checks
    echo -e "${BLUE}🔄 Running quality checks...${RESET}"
    if ! make quality-check >/dev/null 2>&1; then
        echo -e "${RED}❌ Quality checks failed. Please fix issues before release.${RESET}"
        exit 1
    fi

    # Run tests
    echo -e "${BLUE}🔄 Running tests...${RESET}"
    if ! go test ./... >/dev/null 2>&1; then
        echo -e "${RED}❌ Tests failed. Please fix failing tests before release.${RESET}"
        exit 1
    fi

    echo -e "${GREEN}✅ Pre-release checks passed${RESET}"
}

# Function to create git tag
create_git_tag() {
    local version="$1"
    local tag="v$version"

    echo -e "${BLUE}🏷️  Creating git tag: $tag${RESET}"

    git add "$CHANGELOG_FILE" "$VERSION_FILE" 2>/dev/null || true
    git commit -m "chore: release v$version" 2>/dev/null || true
    git tag -a "$tag" -m "Release v$version"

    echo -e "${GREEN}✅ Git tag created: $tag${RESET}"
}

# Function to show release summary
show_release_summary() {
    local version="$1"

    echo ""
    echo -e "${CYAN}🎉 Release v$version Summary${RESET}"
    echo "=================================="
    echo "Version: $version"
    echo "Tag: v$version"
    echo "Binaries built: ${#PLATFORMS[@]} platforms"
    echo "Archive: $DIST_DIR/go-sentinel-cli-$version.tar.gz"
    echo "Release notes: $DIST_DIR/release-notes-$version.md"
    echo ""
    echo -e "${YELLOW}Next steps:${RESET}"
    echo "1. Push the tag: git push origin v$version"
    echo "2. Create GitHub release with generated assets"
    echo "3. Update package managers (if applicable)"
    echo ""
    echo -e "${GREEN}Release preparation complete!${RESET}"
}

# Main function
main() {
    local version_type="${1:-}"
    local custom_version="${2:-}"

    echo -e "${CYAN}🚀 Go Sentinel CLI - Release Automation${RESET}"
    echo -e "${CYAN}=======================================${RESET}"
    echo ""

    # Validate inputs
    if [ -z "$version_type" ]; then
        echo "Usage: $0 <major|minor|patch|custom> [version]"
        echo ""
        echo "Examples:"
        echo "  $0 patch              # Increment patch version"
        echo "  $0 minor              # Increment minor version"
        echo "  $0 major              # Increment major version"
        echo "  $0 custom 1.2.3       # Set specific version"
        exit 1
    fi

    # Determine new version
    local current_version new_version
    current_version=$(get_current_version)

    if [ "$version_type" = "custom" ]; then
        if [ -z "$custom_version" ]; then
            echo "Custom version required when using 'custom' type"
            exit 1
        fi
        new_version="$custom_version"
        validate_version "$new_version"
    else
        new_version=$(increment_version "$current_version" "$version_type")
    fi

    echo "Current version: $current_version"
    echo "New version: $new_version"
    echo ""

    # Check if version already exists
    if version_exists "$new_version"; then
        echo -e "${RED}❌ Version v$new_version already exists${RESET}"
        exit 1
    fi

    # Confirm release
    read -p "Proceed with release v$new_version? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Release cancelled"
        exit 0
    fi

    # Run pre-release checks
    run_pre_release_checks

    # Create build directories
    mkdir -p "$DIST_DIR"

    # Update version file
    echo "$new_version" > "$VERSION_FILE"

    # Update changelog
    update_changelog "$new_version"

    # Build binaries for all platforms
    echo -e "${CYAN}🔨 Building binaries...${RESET}"
    for platform in "${PLATFORMS[@]}"; do
        build_binary "$platform" "$new_version"
    done

    # Create release archive
    create_release_archive "$new_version"

    # Generate release notes
    generate_release_notes "$new_version"

    # Create git tag
    create_git_tag "$new_version"

    # Show summary
    show_release_summary "$new_version"
}

# Run main function with all arguments
main "$@"
