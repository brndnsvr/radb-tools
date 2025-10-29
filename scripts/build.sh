#!/bin/bash
# Build RADb client for multiple platforms with version injection

set -e

# Read version from VERSION file if not set
if [ -z "$VERSION" ]; then
    if [ -f VERSION ]; then
        VERSION=$(cat VERSION)
    else
        VERSION="dev"
    fi
fi

OUTPUT_DIR="${OUTPUT_DIR:-./dist}"

echo "Building RADb client v${VERSION}..."
echo "===================================="

# Get git information
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S_UTC')

echo "Version:     $VERSION"
echo "Git Commit:  $GIT_COMMIT"
echo "Git Branch:  $GIT_BRANCH"
echo "Build Date:  $BUILD_DATE"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build for multiple platforms
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# Build flags with version information
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X 'github.com/bss/radb-client/internal/version.Version=$VERSION'"
LDFLAGS="$LDFLAGS -X 'github.com/bss/radb-client/internal/version.GitCommit=$GIT_COMMIT'"
LDFLAGS="$LDFLAGS -X 'github.com/bss/radb-client/internal/version.GitBranch=$GIT_BRANCH'"
LDFLAGS="$LDFLAGS -X 'github.com/bss/radb-client/internal/version.BuildDate=$BUILD_DATE'"

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r -a parts <<< "$platform"
    GOOS="${parts[0]}"
    GOARCH="${parts[1]}"

    output_name="radb-client-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi

    echo "Building for ${GOOS}/${GOARCH}..."

    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "$LDFLAGS" \
        -o "${OUTPUT_DIR}/${output_name}" \
        ./cmd/radb-client

    echo "  Created: ${OUTPUT_DIR}/${output_name}"
done

# Generate checksums
echo ""
echo "Generating checksums..."
cd "$OUTPUT_DIR"
sha256sum radb-client-* > checksums.txt
echo "Created: ${OUTPUT_DIR}/checksums.txt"

cd - > /dev/null

echo ""
echo "===================================="
echo "Build complete! Binaries are in ${OUTPUT_DIR}/"
ls -lh "$OUTPUT_DIR"
