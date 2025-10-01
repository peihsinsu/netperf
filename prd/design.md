# Bandwidth Test CLI Design

## Context and Goals
- Provide a Go CLI for stress-testing download bandwidth with concurrent HTTP transfers.
- Default mode streams responses into an in-memory sink so no files are written unless explicitly requested.
- Reuse ideas from `netperf/samples/main.go` (worker pool, atomic counters, live bandwidth reporting) while extracting the logic into a reusable structure for a standalone tool.
- Keep the tool simple to operate, without external dependencies.

## Command Interface
```
bandfetch -list urls.txt [-save] [-out ./downloads] [-workers 16] [-timeout 60s] [-retries 3] [-progress false]
```
- `-list` (required): path to a newline-separated file of URLs; blank lines and lines starting with `#` are ignored.
- `-save` (optional bool): enable file persistence using the default output directory `downloads`.
- `-out` (optional string): custom destination directory for downloads. Supplying this implies `-save` regardless of the flag. Empty string means discard.
- `-workers` (optional int): worker pool size. Default `min(runtime.NumCPU()*2, 64)`.
- `-timeout` (optional duration): per-request timeout parsed with `time.ParseDuration`. Default `60s`.
- `-retries` (optional int): number of retry attempts beyond the initial request. Default `3`.
- `-progress` (optional bool): enable/disable live bandwidth ticker for quiet scripting.
- Output format when saving: `[OK] URL -> path`. When discarding: `[OK] URL (discarded)`.

## High-Level Architecture
```
cmd/bandfetch/main.go      // CLI entrypoint, flag parsing, wiring
internal/config/config.go  // Flag normalization and validation helpers
internal/downloader/       // Download manager and sink implementations
    manager.go             // Worker pool orchestration, job submission
    download.go            // HTTP download routine with retries and sinks
    sink.go                // FileSink (temp file + rename) and DiscardSink
internal/metrics/          // Bandwidth counters and printer
    aggregator.go          // Atomic counters, EWMA tracking, human formatting
    printer.go             // Live ticker goroutine respecting -progress
internal/urls/reader.go    // URL list parsing utilities
```
- `manager` owns the job channel, spawns workers, and handles lifecycle via `context.Context`.
- `download` builds a tuned `http.Client` (MaxIdleConns, HTTP/2) and performs retry with exponential backoff.
- `sink` chooses between `FileSink` (writes to `.part` then rename) and `DiscardSink` (wraps `io.Discard`). Both expose an `io.WriteCloser` that reports byte counts through a `counterWriter`.
- `metrics` keeps track of per-second, EWMA, and average throughput, reusing the strategy from the sample.

## Concurrency and Bandwidth Tracking
- Buffered job queue sized at `workers * 2` to keep goroutines busy without overwhelming memory.
- Each worker reads jobs until the channel closes or context cancels, then calls `Downloader.Download`, passing a sink factory.
- Byte counts captured through `counterWriter` -> `metrics.Aggregator.AddBytes`. Atomic counters enable concurrent updates.
- `metrics.Printer` swaps the "bytes this second" counter every tick, converts to bit/s, updates EWMA (alpha = 0.25), and logs `[BW] now=… ewma=… avg=… total=…`.
- On completion, manager waits for workers, cancels printer, and prints a summary with total bytes.

## Error Handling & Resilience
- Temporary files live at `<target>.part`. On success rename, on error delete partial file.
- Backoff: start at 500ms doubled each retry, jittered slightly (`±20%`) to avoid synchronization.
- Context cancellation respected across retries and HTTP requests.
- HTTP status >= 400 treated as error; message includes status.
- Validation prevents contradictory flags (e.g. negative workers, empty list path).

## Testing Strategy
1. **Unit tests**
   - `sink_test.go`: ensure discard sink never touches filesystem and counts bytes.
   - `aggregator_test.go`: verify EWMA and formatting functions.
   - `config_test.go`: validate flag combinations and defaults.
2. **Integration tests**
   - `download_test.go`: use `httptest.Server` to stream deterministic payloads, assert bytes counted, retries triggered on transient failure, and rename occurs when saving.
   - `manager_test.go`: run a mini job queue with parallel workers to confirm graceful shutdown and summary metrics.
3. **(Optional) Performance sanity**
   - Benchmark aggregator or download path with mocked network to ensure no excessive allocations.

## Open Points & Assumptions
- Only HTTP(S) downloads supported; no auth headers yet.
- No per-host throttling; user must set workers responsibly.
- Checksum verification and progress bars are out of scope for this iteration.

