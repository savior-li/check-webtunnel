#!/bin/bash

set -e

BUILD_DIR="./build"
mkdir -p "$BUILD_DIR"

PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "darwin/amd64"
    "darwin/arm64"
)

echo "Building tor-bridge-collector..."

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT="tor-bridge-collector-${GOOS}-${GOARCH}"
    
    if [ "$GOOS" == "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi
    
    echo "Building $OUTPUT..."
    
    if [ "$GOOS" == "windows" ]; then
        GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build \
            -ldflags="-s -w" \
            -o "$BUILD_DIR/$OUTPUT" \
            ./cmd/server
    else
        GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=1 go build \
            -ldflags="-s -w" \
            -o "$BUILD_DIR/$OUTPUT" \
            ./cmd/server
    fi
    
    if [ -f "$BUILD_DIR/$OUTPUT" ]; then
        SIZE=$(du -h "$BUILD_DIR/$OUTPUT" | cut -f1)
        echo "  -> $BUILD_DIR/$OUTPUT ($SIZE)"
    fi
done

echo ""
echo "Build complete! Binaries in $BUILD_DIR/"
ls -lh "$BUILD_DIR"
