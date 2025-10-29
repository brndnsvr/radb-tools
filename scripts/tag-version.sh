#!/bin/bash
#
# Version Tagging Script
# Reads VERSION file, creates git tag, and updates references
#
# Usage: ./scripts/tag-version.sh [version]
#

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Get version from argument or VERSION file
if [ -n "$1" ]; then
    NEW_VERSION="$1"
    print_info "Using version from argument: $NEW_VERSION"
else
    if [ -f VERSION ]; then
        NEW_VERSION=$(cat VERSION)
        print_info "Using version from VERSION file: $NEW_VERSION"
    else
        print_error "No version specified and VERSION file not found"
        echo "Usage: $0 [version]"
        exit 1
    fi
fi

# Validate version format (semantic versioning)
if ! [[ "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]]; then
    print_error "Invalid version format: $NEW_VERSION"
    echo "Expected format: X.Y.Z or X.Y.Z-suffix (e.g., 1.0.0, 1.0.0-pre, 1.0.0-rc1)"
    exit 1
fi

print_info "Preparing to tag version: v$NEW_VERSION"

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_error "Not in a git repository"
    exit 1
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    print_warning "You have uncommitted changes"
    read -p "Continue anyway? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check if tag already exists
if git rev-parse "v$NEW_VERSION" >/dev/null 2>&1; then
    print_error "Tag v$NEW_VERSION already exists"
    read -p "Delete existing tag and recreate? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag -d "v$NEW_VERSION"
        print_info "Deleted existing tag"
    else
        exit 1
    fi
fi

# Update VERSION file if different
if [ -f VERSION ]; then
    CURRENT_VERSION=$(cat VERSION)
    if [ "$CURRENT_VERSION" != "$NEW_VERSION" ]; then
        print_info "Updating VERSION file: $CURRENT_VERSION -> $NEW_VERSION"
        echo "$NEW_VERSION" > VERSION
    fi
else
    print_info "Creating VERSION file"
    echo "$NEW_VERSION" > VERSION
fi

# Update version in internal/version/version.go
VERSION_FILE="internal/version/version.go"
if [ -f "$VERSION_FILE" ]; then
    print_info "Updating $VERSION_FILE"
    sed -i.bak "s/Version = \".*\"/Version = \"$NEW_VERSION\"/" "$VERSION_FILE"
    rm -f "${VERSION_FILE}.bak"
else
    print_warning "$VERSION_FILE not found, skipping update"
fi

# Show what changed
print_info "Changes to be committed:"
git diff VERSION internal/version/version.go 2>/dev/null || true

# Ask for confirmation
echo ""
read -p "Create git tag v$NEW_VERSION? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_warning "Aborted"
    exit 0
fi

# Commit version changes if there are any
if ! git diff-index --quiet HEAD -- VERSION internal/version/version.go 2>/dev/null; then
    print_info "Committing version changes..."
    git add VERSION internal/version/version.go
    git commit -m "Bump version to $NEW_VERSION

🤖 Generated with [Claude Code](https://claude.com/claude-code)
via [Happy](https://happy.engineering)

Co-Authored-By: Claude <noreply@anthropic.com>
Co-Authored-By: Happy <yesreply@happy.engineering>"
    print_success "Version changes committed"
fi

# Create annotated tag
print_info "Creating git tag v$NEW_VERSION..."

# Determine if this is a pre-release
if [[ "$NEW_VERSION" =~ -(pre|alpha|beta|rc) ]]; then
    TAG_MESSAGE="Pre-release v$NEW_VERSION"
else
    TAG_MESSAGE="Release v$NEW_VERSION"
fi

git tag -a "v$NEW_VERSION" -m "$TAG_MESSAGE"

print_success "Tag v$NEW_VERSION created successfully"

# Show tag info
echo ""
print_info "Tag information:"
git show "v$NEW_VERSION" --no-patch

echo ""
print_success "Version tagging complete!"
echo ""
echo "Next steps:"
echo "  1. Review the tag: git show v$NEW_VERSION"
echo "  2. Push to remote: git push origin main --tags"
echo "  3. Build release: ./scripts/build.sh"
echo ""
echo "To push everything:"
echo "  git push origin main && git push origin v$NEW_VERSION"
