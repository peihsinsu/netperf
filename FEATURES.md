# New Feature: Graceful Shutdown with Summary Report

## Overview

The bandfetch tool now supports graceful shutdown via **Ctrl+C (SIGINT)** or **SIGTERM** signals, and displays a comprehensive summary report showing detailed download statistics.

## What's New

### 1. Signal Handling
- Press **Ctrl+C** at any time during download
- Program will gracefully shut down all workers
- No data corruption or incomplete files
- Exit code 130 for interrupted downloads (standard)

### 2. Summary Report
At the end of every download session (whether completed or interrupted), you'll see:

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

### 3. Metrics Tracked

#### Total Downloaded
- Shows total bytes successfully downloaded
- Formatted in human-readable units (B, KiB, MiB, GiB)

#### Elapsed Time
- Time from start to finish
- Formatted as:
  - `5.0s` for less than 1 minute
  - `1m 30s` for less than 1 hour
  - `1h 1m 5s` for longer sessions

#### Average Speed
- Overall average bandwidth throughout the session
- Calculated as: `(total bytes × 8) / elapsed seconds`
- Formatted in Kbit/s, Mbit/s, or Gbit/s

#### Peak Speed
- Highest bandwidth achieved during any 1-second interval
- Tracked atomically for thread-safety
- Updated in real-time as downloads progress

## Implementation Details

### Architecture Changes

1. **Signal Channel**
   ```go
   sigChan := make(chan os.Signal, 1)
   signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
   ```

2. **Async Download Manager**
   - Manager runs in goroutine
   - Main thread waits on either completion or signal
   - Clean cancellation propagated via `context.Context`

3. **Peak Tracking**
   - Uses atomic `uint64` to store `float64` bits
   - Compare-and-swap for thread-safe updates
   - Updated every second during bandwidth measurement

4. **Summary Generation**
   ```go
   type Summary struct {
       TotalBytes   int64
       Elapsed      time.Duration
       AverageBps   float64
       PeakBps      float64
       // ... formatted strings
   }
   ```

### Thread Safety

- **Peak bandwidth**: Atomic CAS operations
- **Total bytes**: Already atomic in original design
- **Elapsed time**: Calculated from immutable start time
- No locks required!

### Testing

New tests added to `internal/metrics/aggregator_test.go`:

- `TestPeakBps`: Verifies peak tracking logic
- `TestGetSummary`: Validates summary generation
- `TestFormatSummary`: Checks formatted output
- `TestFormatDuration`: Tests duration formatting

All tests pass:
```bash
go test ./...
# All packages: PASS
```

## Usage Examples

### Normal Completion
```bash
$ ./bin/bandfetch -list urls.txt
[BW] now=150.00 Mbit/s  ewma=145.00 Mbit/s  avg=142.00 Mbit/s  total=1.5 GiB
...
╔══════════════════════════════════════════════════════╗
║              Download Summary Report                 ║
╠══════════════════════════════════════════════════════╣
║  Total Downloaded : 2.45 GiB                         ║
║  Elapsed Time     : 2m 15s                           ║
║  Average Speed    : 145.67 Mbit/s                    ║
║  Peak Speed       : 182.33 Mbit/s                    ║
╚══════════════════════════════════════════════════════╝
```

### Interrupted with Ctrl+C
```bash
$ ./bin/bandfetch -list large-files.txt
[BW] now=150.00 Mbit/s  ewma=145.00 Mbit/s  avg=142.00 Mbit/s  total=1.5 GiB
^C
[INTERRUPT] Received signal interrupt, shutting down gracefully...

╔══════════════════════════════════════════════════════╗
║              Download Summary Report                 ║
╠══════════════════════════════════════════════════════╣
║  Total Downloaded : 1.50 GiB                         ║
║  Elapsed Time     : 45.3s                            ║
║  Average Speed    : 265.15 Mbit/s                    ║
║  Peak Speed       : 312.50 Mbit/s                    ║
╚══════════════════════════════════════════════════════╝

$ echo $?
130
```

## Benefits

1. **Better UX**: Users can stop downloads anytime and see progress
2. **Performance Analysis**: Peak bandwidth shows network capability
3. **Testing**: Quickly test different worker configurations
4. **Debugging**: Understand if average is limited by source or network
5. **Professional**: Clean shutdown without orphaned processes

## Future Enhancements

Possible additions for future versions:
- [ ] Success/failure counts per URL
- [ ] Histogram of bandwidth distribution
- [ ] JSON output for scripting
- [ ] Save summary to file (optional)
- [ ] Multiple peak tracking (top 3 peaks)

## Migration Notes

- **No breaking changes**: Existing usage remains the same
- **New behavior**: Summary always prints (replaces old simple line)
- **Exit codes**: New exit code 130 for Ctrl+C (standard Unix convention)
- **Dependencies**: No new external dependencies added

## Files Changed

```
cmd/bandfetch/main.go            # Signal handling + summary display
internal/metrics/aggregator.go   # Peak tracking + summary generation
internal/metrics/printer.go      # Peak update during monitoring
internal/metrics/aggregator_test.go  # New tests
```

## Build & Test

```bash
# Build
make build

# Run all tests
make test

# Quick verification
bash verify-build.sh

# Test summary report
bash test-summary.sh
```

Enjoy the new feature! 🚀
