package downloader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cx009/netperf/internal/metrics"
)

func TestDownloadDiscard(t *testing.T) {
	payload := []byte("abcdefghijklmnopqrstuvwxyz")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	t.Cleanup(srv.Close)

	agg := metrics.NewAggregator()
	dl := New(NewHTTPClient(5*time.Second), agg, Options{Save: false, Retries: 0})

	res, err := dl.Download(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Discarded {
		t.Fatalf("expected discarded result")
	}
	if agg.TotalBytes() != int64(len(payload)) {
		t.Fatalf("expected %d bytes, got %d", len(payload), agg.TotalBytes())
	}
}

func TestDownloadSave(t *testing.T) {
	payload := []byte("hello world")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	agg := metrics.NewAggregator()
	dl := New(NewHTTPClient(5*time.Second), agg, Options{Save: true, OutDir: dir, Retries: 0})

	res, err := dl.Download(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Discarded {
		t.Fatalf("expected file to be saved")
	}
	data, err := os.ReadFile(res.Destination)
	if err != nil {
		t.Fatalf("reading destination: %v", err)
	}
	if string(data) != string(payload) {
		t.Fatalf("payload mismatch")
	}
	if agg.TotalBytes() != int64(len(payload)) {
		t.Fatalf("expected %d bytes, got %d", len(payload), agg.TotalBytes())
	}
	if _, err := os.Stat(filepath.Join(dir, filepath.Base(res.Destination)+".part")); !os.IsNotExist(err) {
		t.Fatalf("expected temp file to be removed")
	}
}

func TestDownloadRetries(t *testing.T) {
	var hits int32
	payload := []byte("retry-success")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&hits, 1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write(payload)
	}))
	t.Cleanup(srv.Close)

	agg := metrics.NewAggregator()
	dl := New(NewHTTPClient(5*time.Second), agg, Options{Save: false, Retries: 2})

	res, err := dl.Download(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("expected retry to succeed, got error: %v", err)
	}
	if !res.Discarded {
		t.Fatalf("expected discard mode")
	}
	if hits < 2 {
		t.Fatalf("expected at least 2 hits, got %d", hits)
	}
}

func TestDownloadHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	agg := metrics.NewAggregator()
	dl := New(NewHTTPClient(5*time.Second), agg, Options{Save: false, Retries: 0})

	if _, err := dl.Download(context.Background(), srv.URL); err == nil {
		t.Fatalf("expected error on 500 response")
	}
}
