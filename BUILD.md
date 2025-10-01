# Build Guide

This document describes how to build bandfetch for various platforms.

## Quick Start

### Build for Current Platform

```bash
# Using Make
make build

# Or directly with Go
go build -o bin/bandfetch ./cmd/bandfetch
```

The binary will be in `bin/bandfetch` (or `bin/bandfetch.exe` on Windows).

## Cross-Platform Builds

### Build All Platforms

Build binaries for Linux, Windows, and macOS in one command:

```bash
# Using Make
make build-all

# Or using the build script (Unix/macOS/Linux)
bash build-release.sh

# Or using PowerShell (Windows)
.\build-release.ps1
```

This will create binaries for:
- Linux (AMD64, ARM64)
- Windows (AMD64, ARM64)
- macOS (AMD64, ARM64/Apple Silicon)

### Build Specific Platform

#### Linux

```bash
# AMD64 (most common)
make build-linux-amd64

# ARM64 (Raspberry Pi, ARM servers)
make build-linux-arm64

# Both
make build-linux
```

#### Windows

```bash
# AMD64 (most common)
make build-windows-amd64

# ARM64 (Surface Pro X, etc.)
make build-windows-arm64

# Both
make build-windows
```

#### macOS

```bash
# AMD64 (Intel Macs)
make build-darwin-amd64

# ARM64 (Apple Silicon M1/M2/M3)
make build-darwin-arm64

# Both
make build-darwin
```

## Manual Cross-Compilation

You can also build manually with environment variables:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o bin/bandfetch-linux-amd64 ./cmd/bandfetch

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o bin/bandfetch-windows-amd64.exe ./cmd/bandfetch

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o bin/bandfetch-darwin-arm64 ./cmd/bandfetch
```

## Build Flags

### Version Information

Add version information to the binary:

```bash
VERSION=v1.0.0 make build
```

Or manually:

```bash
go build -ldflags "-X main.version=v1.0.0" -o bin/bandfetch ./cmd/bandfetch
```

### Optimization Flags

The default build uses these optimization flags:

```bash
-ldflags "-s -w"
```

- `-s`: Strip symbol table
- `-w`: Strip DWARF debug information

This reduces binary size by ~30%.

### CGO

By default, CGO is disabled for better portability:

```bash
CGO_ENABLED=0 go build ./cmd/bandfetch
```

## Release Build

To create a complete release with archives and checksums:

### Unix/macOS/Linux

```bash
bash build-release.sh
```

### Windows

```powershell
.\build-release.ps1
```

This will:
1. Build binaries for all platforms
2. Create `.tar.gz` archives (Unix) or `.zip` archives (Windows)
3. Generate SHA256 checksums
4. Place everything in `releases/<version>/`

### Release Artifacts

After running the release build:

```
releases/v1.0.0/
├── bandfetch-v1.0.0-linux-amd64.tar.gz
├── bandfetch-v1.0.0-linux-arm64.tar.gz
├── bandfetch-v1.0.0-windows-amd64.zip
├── bandfetch-v1.0.0-windows-arm64.zip
├── bandfetch-v1.0.0-darwin-amd64.tar.gz
├── bandfetch-v1.0.0-darwin-arm64.tar.gz
└── SHA256SUMS
```

## Supported Platforms

| OS      | Architecture | Binary Name                    | Notes                    |
|---------|--------------|--------------------------------|--------------------------|
| Linux   | AMD64        | bandfetch-linux-amd64          | Most common              |
| Linux   | ARM64        | bandfetch-linux-arm64          | Raspberry Pi, ARM servers|
| Windows | AMD64        | bandfetch-windows-amd64.exe    | Most common              |
| Windows | ARM64        | bandfetch-windows-arm64.exe    | Surface Pro X, etc.      |
| macOS   | AMD64        | bandfetch-darwin-amd64         | Intel Macs               |
| macOS   | ARM64        | bandfetch-darwin-arm64         | Apple Silicon M1/M2/M3   |

## Build Requirements

- **Go**: 1.22 or later
- **Make**: GNU Make (optional, for using Makefile)
- **Git**: For version information (optional)
- **tar**: For creating Unix archives (usually pre-installed)
- **zip**: For creating Windows archives (usually pre-installed)

### Installing Go

- **macOS**: `brew install go`
- **Linux**: Download from https://golang.org/dl/
- **Windows**: Download installer from https://golang.org/dl/

### Verifying Installation

```bash
go version
# Should output: go version go1.22.x ...
```

## Troubleshooting

### "Command not found: make"

**macOS**: Install Xcode Command Line Tools:
```bash
xcode-select --install
```

**Linux (Debian/Ubuntu)**:
```bash
sudo apt install build-essential
```

**Windows**: Use PowerShell script instead:
```powershell
.\build-release.ps1
```

### "GOOS not supported"

Make sure you're using Go 1.22+:
```bash
go version
```

### Binary too large

Use optimization flags:
```bash
go build -ldflags "-s -w" ./cmd/bandfetch
```

Or use `make build` which includes these flags by default.

### Cross-compilation errors

Ensure CGO is disabled for cross-compilation:
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./cmd/bandfetch
```

## Testing Builds

After building, test the binary:

```bash
# Check version
./bin/bandfetch -version

# Basic functionality test
./bin/bandfetch -list urls.example.txt
```

## Clean Build Artifacts

```bash
# Clean binaries
make clean

# Remove all build and release artifacts
rm -rf bin/ releases/
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      - name: Build all platforms
        run: bash build-release.sh
      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: releases
          path: releases/
```

## Additional Resources

- [Go Cross Compilation](https://golang.org/doc/install/source#environment)
- [Go Build Constraints](https://golang.org/cmd/go/#hdr-Build_constraints)
- [Project README](README.md)
