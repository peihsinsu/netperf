package downloader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDiscardSink(t *testing.T) {
	sink, err := newDiscardSink()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sink.discarded {
		t.Fatalf("expected sink to discard")
	}
	if err := sink.closeWriter(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	if err := sink.finalizeSuccess(); err != nil {
		t.Fatalf("finalize success: %v", err)
	}
}

func TestNewFileSinkSuccess(t *testing.T) {
	dir := t.TempDir()
	sink, err := newFileSink(dir, "file.bin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data := []byte("hello world")
	if _, err := sink.writer.Write(data); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := sink.closeWriter(); err != nil {
		t.Fatalf("close error: %v", err)
	}
	if err := sink.finalizeSuccess(); err != nil {
		t.Fatalf("finalize success: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "file.bin")); err != nil {
		t.Fatalf("expected final file to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "file.bin.part")); !os.IsNotExist(err) {
		t.Fatalf("temporary file should be removed, got %v", err)
	}
}

func TestNewFileSinkFailure(t *testing.T) {
	dir := t.TempDir()
	sink, err := newFileSink(dir, "file.bin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := sink.writer.Write([]byte("data")); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := sink.closeWriter(); err != nil {
		t.Fatalf("close error: %v", err)
	}
	sink.finalizeFailure()
	if _, err := os.Stat(filepath.Join(dir, "file.bin.part")); !os.IsNotExist(err) {
		t.Fatalf("temporary file should be cleaned up")
	}
	if _, err := os.Stat(filepath.Join(dir, "file.bin")); !os.IsNotExist(err) {
		t.Fatalf("final file should not exist")
	}
}
