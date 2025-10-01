#!/bin/bash
# Cross-platform build script for bandfetch releases

set -e

VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BINARY="bandfetch"
BUILD_DIR="bin"
RELEASE_DIR="releases/${VERSION}"

echo "╔════════════════════════════════════════════════════════╗"
echo "║     bandfetch Cross-Platform Build Script             ║"
echo "║     Version: ${VERSION}                                      "
echo "╚════════════════════════════════════════════════════════╝"
echo ""

# Clean and create directories
echo "→ Preparing directories..."
rm -rf "${BUILD_DIR}"
rm -rf "${RELEASE_DIR}"
mkdir -p "${BUILD_DIR}"
mkdir -p "${RELEASE_DIR}"
echo "  ✓ Directories ready"
echo ""

# Build for all platforms
echo "→ Building for all platforms..."
echo ""

# Linux AMD64
echo "  [1/6] Linux (AMD64)..."
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=${VERSION}" \
    -o "${BUILD_DIR}/${BINARY}-linux-amd64" ./cmd/bandfetch
echo "        ✓ Complete"

# Linux ARM64
echo "  [2/6] Linux (ARM64)..."
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.version=${VERSION}" \
    -o "${BUILD_DIR}/${BINARY}-linux-arm64" ./cmd/bandfetch
echo "        ✓ Complete"

# Windows AMD64
echo "  [3/6] Windows (AMD64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=${VERSION}" \
    -o "${BUILD_DIR}/${BINARY}-windows-amd64.exe" ./cmd/bandfetch
echo "        ✓ Complete"

# Windows ARM64
echo "  [4/6] Windows (ARM64)..."
GOOS=windows GOARCH=arm64 go build -ldflags "-s -w -X main.version=${VERSION}" \
    -o "${BUILD_DIR}/${BINARY}-windows-arm64.exe" ./cmd/bandfetch
echo "        ✓ Complete"

# macOS AMD64
echo "  [5/6] macOS (AMD64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.version=${VERSION}" \
    -o "${BUILD_DIR}/${BINARY}-darwin-amd64" ./cmd/bandfetch
echo "        ✓ Complete"

# macOS ARM64 (Apple Silicon)
echo "  [6/6] macOS (ARM64/Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.version=${VERSION}" \
    -o "${BUILD_DIR}/${BINARY}-darwin-arm64" ./cmd/bandfetch
echo "        ✓ Complete"

echo ""
echo "→ Creating release archives..."
echo ""

cd "${BUILD_DIR}"

# Create archives for each platform
create_archive() {
    local platform=$1
    local binary=$2
    local archive_name="${BINARY}-${VERSION}-${platform}"

    if [[ "$binary" == *.exe ]]; then
        # Windows: zip archive
        zip -q "${archive_name}.zip" "$binary"
        echo "  ✓ ${archive_name}.zip"
    else
        # Unix: tar.gz archive
        tar -czf "${archive_name}.tar.gz" "$binary"
        echo "  ✓ ${archive_name}.tar.gz"
    fi

    # Move to release directory
    mv "${archive_name}".* "../${RELEASE_DIR}/"
}

create_archive "linux-amd64" "${BINARY}-linux-amd64"
create_archive "linux-arm64" "${BINARY}-linux-arm64"
create_archive "windows-amd64" "${BINARY}-windows-amd64.exe"
create_archive "windows-arm64" "${BINARY}-windows-arm64.exe"
create_archive "darwin-amd64" "${BINARY}-darwin-amd64"
create_archive "darwin-arm64" "${BINARY}-darwin-arm64"

cd ..

echo ""
echo "→ Generating checksums..."
cd "${RELEASE_DIR}"
shasum -a 256 * > SHA256SUMS
echo "  ✓ SHA256SUMS created"
cd ../..

echo ""
echo "╔════════════════════════════════════════════════════════╗"
echo "║              Build Complete!                           ║"
echo "╚════════════════════════════════════════════════════════╝"
echo ""
echo "Release artifacts:"
ls -lh "${RELEASE_DIR}"
echo ""
echo "Total size: $(du -sh ${RELEASE_DIR} | cut -f1)"
echo ""
echo "Archives ready for distribution in: ${RELEASE_DIR}/"
