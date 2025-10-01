# Cross-Platform Compilation Summary

## Quick Reference

### Build Commands

| Command | Description |
|---------|-------------|
| `make build` | Build for current platform |
| `make build-all` | Build for all platforms (6 binaries) |
| `make build-linux` | Build for Linux (AMD64 + ARM64) |
| `make build-windows` | Build for Windows (AMD64 + ARM64) |
| `make build-darwin` | Build for macOS (AMD64 + ARM64) |
| `make build-linux-amd64` | Build for Linux AMD64 only |
| `make build-windows-amd64` | Build for Windows AMD64 only |
| `make build-darwin-arm64` | Build for macOS ARM64 only |

### Release Scripts

| Platform | Command | Output |
|----------|---------|--------|
| Unix/macOS/Linux | `bash build-release.sh` | Archives in `releases/<version>/` |
| Windows | `.\build-release.ps1` | Archives in `releases\<version>\` |

## Supported Platforms

### ✅ Linux
- **AMD64** (x86_64): Most common, servers, desktops
- **ARM64** (aarch64): Raspberry Pi 4+, AWS Graviton, ARM servers

### ✅ Windows
- **AMD64** (x86-64): Most common, desktops, laptops
- **ARM64** (ARM64): Surface Pro X, Windows on ARM devices

### ✅ macOS
- **AMD64** (x86_64): Intel Macs (2006-2020)
- **ARM64** (arm64): Apple Silicon M1/M2/M3 (2020+)

## Binary Naming Convention

```
bandfetch-<os>-<arch>[.exe]
```

Examples:
- `bandfetch-linux-amd64`
- `bandfetch-windows-amd64.exe`
- `bandfetch-darwin-arm64`

## Release Archive Structure

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

## Build Options

### Version Information
```bash
# Automatic (from git tags)
make build

# Manual
VERSION=v1.2.3 make build
```

### Optimization Flags
Already included in Makefile:
- `-s`: Strip symbol table
- `-w`: Strip DWARF debug info
- `-X main.version=$(VERSION)`: Embed version

### CGO
Disabled by default for maximum portability.

## Testing Builds

```bash
# Build all platforms
make build-all

# Test each binary
bin/bandfetch-linux-amd64 -version
bin/bandfetch-windows-amd64.exe -version
bin/bandfetch-darwin-arm64 -version
```

## File Sizes (Approximate)

| Platform | Uncompressed | Compressed (.tar.gz/.zip) |
|----------|--------------|---------------------------|
| Linux AMD64 | ~8.5 MB | ~3.5 MB |
| Linux ARM64 | ~8.2 MB | ~3.4 MB |
| Windows AMD64 | ~8.8 MB | ~3.6 MB |
| Windows ARM64 | ~8.5 MB | ~3.5 MB |
| macOS AMD64 | ~8.7 MB | ~3.6 MB |
| macOS ARM64 | ~8.4 MB | ~3.5 MB |

## Quick Start for Each Platform

### Linux (AMD64)
```bash
make build-linux-amd64
./bin/bandfetch-linux-amd64 -list urls.txt
```

### Windows (AMD64)
```powershell
make build-windows-amd64
.\bin\bandfetch-windows-amd64.exe -list urls.txt
```

### macOS (Apple Silicon)
```bash
make build-darwin-arm64
./bin/bandfetch-darwin-arm64 -list urls.txt
```

## Common Use Cases

### 1. Development on macOS, Deploy to Linux Server
```bash
# On macOS
make build-linux-amd64

# Copy to server
scp bin/bandfetch-linux-amd64 user@server:/usr/local/bin/bandfetch

# On server
chmod +x /usr/local/bin/bandfetch
```

### 2. Build Everything for Release
```bash
# Create all binaries + archives
bash build-release.sh

# Upload to GitHub releases
gh release create v1.0.0 releases/v1.0.0/*
```

### 3. Windows Developer Building for Linux
```powershell
# Using PowerShell
.\build-release.ps1

# Or using WSL
wsl bash build-release.sh
```

## Troubleshooting

### Build fails with "unsupported GOOS/GOARCH"
- Update Go to 1.22+: `go version`

### Binary too large
- Already optimized with `-s -w` flags
- Use `upx` for additional compression (optional):
  ```bash
  upx --best bin/bandfetch-linux-amd64
  ```

### Can't execute on target platform
- Check architecture matches:
  ```bash
  # Linux
  uname -m
  # x86_64 = AMD64
  # aarch64 = ARM64

  # macOS
  uname -m
  # x86_64 = AMD64
  # arm64 = ARM64
  ```

### Windows "file not found" error
- Use full path or add `.exe` extension:
  ```cmd
  .\bandfetch-windows-amd64.exe -list urls.txt
  ```

## Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `GOOS` | Target operating system | Current OS |
| `GOARCH` | Target architecture | Current arch |
| `CGO_ENABLED` | Enable CGO | `0` (disabled) |
| `VERSION` | Binary version | From git or "dev" |

## Manual Cross-Compilation Examples

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o bin/bandfetch-linux-amd64 ./cmd/bandfetch

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o bin/bandfetch-windows-amd64.exe ./cmd/bandfetch

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o bin/bandfetch-darwin-arm64 ./cmd/bandfetch

# All with optimization
GOOS=linux GOARCH=amd64 go build \
  -ldflags "-s -w -X main.version=v1.0.0" \
  -o bin/bandfetch-linux-amd64 \
  ./cmd/bandfetch
```

## Verification

### Verify binary architecture
```bash
# Linux
file bin/bandfetch-linux-amd64
# Output: ELF 64-bit LSB executable, x86-64

# macOS
file bin/bandfetch-darwin-arm64
# Output: Mach-O 64-bit executable arm64

# Windows (using WSL or Linux)
file bin/bandfetch-windows-amd64.exe
# Output: PE32+ executable (console) x86-64
```

### Verify checksums
```bash
cd releases/v1.0.0/
sha256sum -c SHA256SUMS
```

## CI/CD Integration

### GitHub Actions
```yaml
- name: Build all platforms
  run: make build-all

- name: Create release
  run: bash build-release.sh
```

### GitLab CI
```yaml
build:
  script:
    - make build-all
  artifacts:
    paths:
      - bin/
```

## Additional Resources

- **Full build guide**: [BUILD.md](BUILD.md)
- **Platform-specific guide**: [PLATFORM-GUIDE.md](PLATFORM-GUIDE.md)
- **Main documentation**: [README.md](README.md)
- **Go cross-compilation**: https://golang.org/doc/install/source#environment

## Summary

✅ **6 platforms supported** (Linux, Windows, macOS × AMD64/ARM64)
✅ **Simple commands** (`make build-all` or `bash build-release.sh`)
✅ **Optimized binaries** (~3.5 MB compressed)
✅ **No external dependencies** (static linking)
✅ **Automated release process** (archives + checksums)

Build once, run anywhere! 🚀
