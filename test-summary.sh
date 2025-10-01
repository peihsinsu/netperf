#!/bin/bash
# Quick test script to verify summary report functionality

echo "Testing bandfetch with summary report..."
echo ""
echo "Building..."
go build -o bin/bandfetch ./cmd/bandfetch || exit 1
echo "Build successful!"
echo ""

# Create a test URL list with small files
cat > test-urls.txt <<EOF
# Test URLs with small files
https://speed.hetzner.de/10MB.bin
https://speed.hetzner.de/10MB.bin
EOF

echo "Starting download test..."
echo "Press Ctrl+C after a few seconds to see the summary report"
echo ""

./bin/bandfetch -list test-urls.txt -workers 4

# Cleanup
rm -f test-urls.txt

echo ""
echo "Test complete!"
