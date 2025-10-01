#!/bin/bash
set -e

echo "╔════════════════════════════════════════════╗"
echo "║   netperf Build & Test Verification       ║"
echo "╚════════════════════════════════════════════╝"
echo ""

cd "$(dirname "$0")"

echo "→ Step 1: Formatting code..."
go fmt ./...
echo "  ✓ Code formatted"
echo ""

echo "→ Step 2: Running tests..."
go test -v ./...
if [ $? -eq 0 ]; then
    echo "  ✓ All tests passed"
else
    echo "  ✗ Tests failed"
    exit 1
fi
echo ""

echo "→ Step 3: Building binary..."
mkdir -p bin
go build -o bin/bandfetch ./cmd/bandfetch
if [ $? -eq 0 ]; then
    echo "  ✓ Build successful"
    echo "  Binary location: bin/bandfetch"
else
    echo "  ✗ Build failed"
    exit 1
fi
echo ""

echo "→ Step 4: Checking binary..."
if [ -f bin/bandfetch ]; then
    SIZE=$(du -h bin/bandfetch | cut -f1)
    echo "  ✓ Binary exists (size: $SIZE)"
else
    echo "  ✗ Binary not found"
    exit 1
fi
echo ""

echo "╔════════════════════════════════════════════╗"
echo "║          Verification Complete!            ║"
echo "╚════════════════════════════════════════════╝"
echo ""
echo "Usage examples:"
echo "  ./bin/bandfetch -list urls.example.txt"
echo "  ./bin/bandfetch -list urls.example.txt -save -workers 16"
echo ""
echo "To test Ctrl+C handling:"
echo "  ./bin/bandfetch -list urls.example.txt"
echo "  (Press Ctrl+C to see the summary report)"
