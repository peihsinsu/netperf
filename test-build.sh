#!/bin/bash
cd /Users/cx009/project/claudews/netperf
echo "=== Building bandfetch ==="
go build -o bin/bandfetch ./cmd/bandfetch
if [ $? -eq 0 ]; then
    echo "✓ Build successful"
    echo ""
    echo "=== Running tests ==="
    go test ./...
    if [ $? -eq 0 ]; then
        echo "✓ All tests passed"
    else
        echo "✗ Tests failed"
        exit 1
    fi
else
    echo "✗ Build failed"
    exit 1
fi
