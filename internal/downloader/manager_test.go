package downloader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cx009/netperf/internal/metrics"
)

func TestManagerRunSuccess(t *testing.T) {
	payload := []byte("payload")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	t.Cleanup(srv.Close)

	agg := metrics.NewAggregator()
	dl := New(NewHTTPClient(5*time.Second), agg, Options{Save: false, Retries: 0})
	mgr := NewManager(dl, 4)

	urls := []string{srv.URL, srv.URL, srv.URL}
	if err := mgr.Run(context.Background(), urls); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := int64(len(payload) * len(urls))
	if got := agg.TotalBytes(); got != want {
		t.Fatalf("expected %d bytes, got %d", want, got)
	}
}

func TestManagerRunFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	agg := metrics.NewAggregator()
	dl := New(NewHTTPClient(time.Second), agg, Options{Save: false, Retries: 0})
	mgr := NewManager(dl, 2)

	if err := mgr.Run(context.Background(), []string{srv.URL}); err == nil {
		t.Fatalf("expected failure error")
	}
}
