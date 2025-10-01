# netperf - Network Performance Testing Tool

A Go CLI tool for bandwidth testing with concurrent HTTP downloads.

[中文文件](README.zh-TW.md)

## Quick Start

```bash
# Build
make build

# Run bandwidth test (no file saving)
./bin/bandfetch -list urls.txt -workers 16

# Download and save files
./bin/bandfetch -list urls.txt -save -out downloads

# Run tests
make test
```

## Features

- **Concurrent Downloads**: Worker pool architecture for parallel downloads
- **Real-time Bandwidth Monitoring**: Live bandwidth stats (current, EWMA, average)
- **Flexible Storage**: Save files or discard (bandwidth test only)
- **Reliability**: Auto-retry with exponential backoff, timeout control
- **Optimized HTTP Client**: High connection limits, HTTP/2 support
- **Graceful Shutdown**: Ctrl+C handling with complete summary report
- **Summary Report**: Detailed statistics including peak and average bandwidth
- **Full Test Coverage**: Unit and integration tests

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [Releases](https://github.com/cx009/netperf/releases) page.

**Supported Platforms:**
- Linux (AMD64, ARM64)
- Windows (AMD64, ARM64)
- macOS (AMD64, ARM64/Apple Silicon)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/cx009/netperf.git
cd netperf

# Build for your platform
make build

# Or build for all platforms
make build-all

# Or run directly
go run ./cmd/bandfetch -list urls.txt
```

See [BUILD.md](BUILD.md) for detailed build instructions.

## Usage

### Command Line Options

```bash
bandfetch -list <file> [options]

Options:
  -list string
        Path to URL list file (required)
  -save
        Save downloaded files to disk
  -out string
        Output directory (default "downloads", implies -save)
  -workers int
        Number of concurrent workers (default: CPU*2, max 64)
  -timeout duration
        Request timeout (default 60s)
  -retries int
        Number of retry attempts (default 3)
  -progress
        Show live bandwidth output (default true)
```

### Examples

```bash
# Basic bandwidth test (no file saving)
./bin/bandfetch -list urls.txt

# Save files to default directory
./bin/bandfetch -list urls.txt -save

# Custom output directory and workers
./bin/bandfetch -list urls.txt -out ./downloads -workers 24

# Increase timeout and retries
./bin/bandfetch -list urls.txt -timeout 120s -retries 5

# Quiet mode for scripting
./bin/bandfetch -list urls.txt -progress=false
```

### URL List Format

Create a text file (e.g., `urls.txt`) with one URL per line:

```
https://example.com/file1.bin
https://example.com/file2.bin
# Comments start with #

https://example.com/file3.bin
```

- Blank lines are ignored
- Lines starting with `#` are comments

## Project Structure

```
netperf/
├── cmd/
│   └── bandfetch/          # CLI entry point
├── internal/
│   ├── config/             # Flag parsing and validation
│   ├── downloader/         # Download manager and implementations
│   ├── metrics/            # Bandwidth tracking and reporting
│   └── urls/               # URL list parsing
├── prd/                    # Design documents
├── samples/                # Sample prototype code
└── Makefile
```

## Output Example

### During Download
```
[OK] https://example.com/file1.bin (discarded)
[OK] https://example.com/file2.bin -> downloads/file2.bin
[BW] now=125.43 Mbit/s  ewma=118.76 Mbit/s  avg=115.22 Mbit/s  total=1.23 GiB
[BW] now=132.18 Mbit/s  ewma=122.12 Mbit/s  avg=117.45 Mbit/s  total=1.39 GiB
[FAIL] https://example.com/timeout.bin -> Get: context deadline exceeded
```

### Summary Report (on completion or Ctrl+C)
```
╔══════════════════════════════════════════════════════╗
║              Download Summary Report                 ║
╠══════════════════════════════════════════════════════╣
║  Total Downloaded : 2.45 GiB                         ║
║  Elapsed Time     : 2m 15s                           ║
║  Average Speed    : 145.67 Mbit/s                    ║
║  Peak Speed       : 182.33 Mbit/s                    ║
╚══════════════════════════════════════════════════════╝
```

### Graceful Shutdown
Press **Ctrl+C** at any time to stop downloads gracefully and see the summary report:
```
^C
[INTERRUPT] Received signal interrupt, shutting down gracefully...
(Summary report shown above)
```

## Architecture

- **Manager**: Orchestrates worker pool and job queue
- **Downloader**: HTTP client with retry logic and exponential backoff
- **Sink**: Two implementations:
  - `FileSink`: Writes to `.part` temp file, renames on success
  - `DiscardSink`: Discards data, only tracks bandwidth
- **Metrics**: Atomic counters for bandwidth tracking, EWMA calculation

## Performance Tips

To maximize bandwidth utilization:

1. **Increase workers**: `-workers 24` or higher
2. **Diversify sources**: Download from different hosts to avoid per-host limits
3. **Use SSD/tmpfs**: Avoid disk I/O bottleneck
4. **HTTP/2**: Automatically enabled for HTTPS connections
5. **Test without saving**: Use `-save=false` to eliminate disk I/O

## Development

```bash
# Run tests
make test

# Format code
make fmt

# Build binary
make build

# Build for specific platform
make build-linux-amd64
make build-windows-amd64
make build-darwin-arm64

# Build all platforms
make build-all

# Create release packages
bash build-release.sh

# Clean build artifacts
make clean
```

### Running with Make

```bash
# Basic run
make run LIST=urls.txt

# With options
make run LIST=urls.txt SAVE=1 OUT=downloads WORKERS=16
```

### Cross-Platform Build

See [BUILD.md](BUILD.md) for comprehensive build instructions including:
- Building for specific platforms
- Creating release packages
- Manual cross-compilation
- CI/CD integration

## Technical Details

- **Language**: Go 1.22+
- **Dependencies**: Standard library only
- **Concurrency**: Worker pool with buffered channels
- **HTTP Client**: Custom `http.Transport` with:
  - High connection limits
  - HTTP/2 enabled
  - 1MiB buffer for efficient copying
  - Keep-alive connections

## Testing

The project includes comprehensive tests:

- Unit tests for all components
- Integration tests with `httptest.Server`
- Test coverage for:
  - Configuration parsing
  - Download logic with retries
  - Sink implementations
  - Metrics aggregation

Run tests with:
```bash
make test
```

## Documentation

- [Design Document](prd/design.md) - Detailed architecture and design decisions
- [Sample Prototype](samples/main.go) - Original prototype implementation

## License

Personal learning and testing project.

## Contributing

This is a personal project for learning purposes. Feel free to fork and experiment!
