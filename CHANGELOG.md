# Changelog

## [Unreleased]

### Added
- **Graceful shutdown on Ctrl+C (SIGINT/SIGTERM)**: Program now handles interrupt signals gracefully
- **Summary report**: Displays detailed statistics at the end of download session
  - Total downloaded size
  - Elapsed time
  - Average bandwidth
  - Peak bandwidth
- **Peak bandwidth tracking**: Monitors and records the highest bandwidth achieved during the session
- **Cross-platform builds**: Support for building on multiple platforms
  - Linux (AMD64, ARM64)
  - Windows (AMD64, ARM64)
  - macOS (AMD64, ARM64/Apple Silicon)
- **Build scripts**: Automated release build scripts
  - `build-release.sh` for Unix/macOS/Linux
  - `build-release.ps1` for Windows PowerShell
- **Makefile targets**: New build targets for all platforms
  - `make build-all` - Build for all platforms
  - `make build-linux`, `build-windows`, `build-darwin`
  - Platform-specific targets (e.g., `make build-linux-amd64`)
- Beautiful formatted summary output with box-drawing characters
- Comprehensive build documentation in `BUILD.md`

### Changed
- Main loop now runs in a goroutine to allow signal handling
- Exit code 130 for interrupted downloads (standard Ctrl+C exit code)
- Summary output replaces simple one-line summary
- Updated `.gitignore` to exclude release artifacts
- Enhanced README with installation and build instructions

### Technical Details
- Added `PeakBps()` and `UpdatePeakBps()` methods to `metrics.Aggregator`
- Added `GetSummary()` method returning structured `Summary` type
- Added `FormatSummary()` for pretty-printed report output
- Signal handling via `os/signal` package
- Atomic float64 operations for thread-safe peak tracking using `math.Float64bits`
- LDFLAGS optimization for smaller binaries (`-s -w`)
- Version information embedded via `-X main.version`

## [0.1.0] - 2025-10-01

### Added
- Initial release of bandfetch
- Concurrent download with configurable worker pool
- Real-time bandwidth monitoring with EWMA smoothing
- File saving or discard mode (bandwidth testing only)
- Automatic retry with exponential backoff
- Configurable timeout and retry count
- URL list parsing with comment support
- Comprehensive test coverage
