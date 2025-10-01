package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/cx009/netperf/internal/metrics"
)

// Options holds parameters for Downloader behaviour.
type Options struct {
	Save    bool
	OutDir  string
	Retries int
}

// Downloader performs download operations with retry policies.
type Downloader struct {
	client *http.Client
	agg    *metrics.Aggregator
	opts   Options
}

// Result describes the outcome of a download attempt.
type Result struct {
	Destination string
	Discarded   bool
}

// New initializes a Downloader instance.
func New(client *http.Client, agg *metrics.Aggregator, opts Options) *Downloader {
	return &Downloader{
		client: client,
		agg:    agg,
		opts:   opts,
	}
}

// Download retrieves a single URL using retry semantics.
func (d *Downloader) Download(ctx context.Context, rawURL string) (Result, error) {
	if d.client == nil {
		return Result{}, errors.New("http client not configured")
	}

	var lastErr error
	baseDelay := 500 * time.Millisecond

	for attempt := 0; attempt <= d.opts.Retries; attempt++ {
		res, err := d.tryOnce(ctx, rawURL)
		if err == nil {
			return res, nil
		}
		lastErr = err

		if ctx.Err() != nil {
			return res, ctx.Err()
		}
		if attempt == d.opts.Retries {
			break
		}

		delay := jitter(baseDelay << attempt)
		select {
		case <-ctx.Done():
			return res, ctx.Err()
		case <-time.After(delay):
		}
	}

	return Result{}, lastErr
}

func (d *Downloader) tryOnce(ctx context.Context, rawURL string) (Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return Result{}, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var handle *sinkHandle
	if d.opts.Save {
		name := FileNameFromURL(rawURL)
		handle, err = newFileSink(d.opts.OutDir, name)
	} else {
		handle, err = newDiscardSink()
	}
	if err != nil {
		return Result{}, err
	}

	writer := &counterWriter{dst: handle.writer, agg: d.agg}
	buf := make([]byte, 1<<20)
	if _, err := io.CopyBuffer(writer, resp.Body, buf); err != nil {
		handle.closeWriter()
		handle.finalizeFailure()
		return Result{}, err
	}

	if err := handle.closeWriter(); err != nil {
		handle.finalizeFailure()
		return Result{}, err
	}

	if err := handle.finalizeSuccess(); err != nil {
		return Result{}, err
	}

	return Result{Destination: handle.destination, Discarded: handle.discarded}, nil
}

// counterWriter records bytes flowing through it.
type counterWriter struct {
	dst io.Writer
	agg *metrics.Aggregator
}

func (cw *counterWriter) Write(p []byte) (int, error) {
	n, err := cw.dst.Write(p)
	if n > 0 && cw.agg != nil {
		cw.agg.AddBytes(n)
	}
	return n, err
}

var (
	jitterOnce sync.Once
	rng        *rand.Rand
	rngMu      sync.Mutex
)

func initRNG() {
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func jitter(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	jitterOnce.Do(initRNG)
	rngMu.Lock()
	defer rngMu.Unlock()
	factor := 0.8 + rng.Float64()*0.4 // 0.8x - 1.2x
	return time.Duration(float64(d) * factor)
}
