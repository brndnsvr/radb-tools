#!/bin/bash
# Build RADb client for multiple platforms

set -e

VERSION="${VERSION:-dev}"
OUTPUT_DIR="${OUTPUT_DIR:-./bin}"

echo "Building RADb client v${VERSION}..."
echo "===================================="

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

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r -a parts <<< "$platform"
    GOOS="${parts[0]}"
    GOARCH="${parts[1]}"

    output_name="radb-client-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi

    echo ""
    echo "Building for ${GOOS}/${GOARCH}..."

    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-s -w -X main.version=${VERSION}" \
        -o "${OUTPUT_DIR}/${output_name}" \
        ./cmd/radb-client

    echo "Created: ${OUTPUT_DIR}/${output_name}"
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
