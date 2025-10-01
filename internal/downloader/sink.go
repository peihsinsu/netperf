package downloader

import (
	"io"
	"os"
	"path/filepath"
)

// sinkHandle represents a target for a download.
type sinkHandle struct {
	writer      io.WriteCloser
	finalize    func(success bool) error
	destination string
	discarded   bool
}

func newDiscardSink() (*sinkHandle, error) {
	return &sinkHandle{
		writer:    nopCloser{Writer: io.Discard},
		finalize:  func(bool) error { return nil },
		discarded: true,
	}, nil
}

func newFileSink(outDir, fileName string) (*sinkHandle, error) {
	finalPath := filepath.Join(outDir, fileName)
	tmpPath := finalPath + ".part"

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}

	f, err := os.Create(tmpPath)
	if err != nil {
		return nil, err
	}

	finalize := func(success bool) error {
		if success {
			return os.Rename(tmpPath, finalPath)
		}
		if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	return &sinkHandle{
		writer:      f,
		finalize:    finalize,
		destination: finalPath,
		discarded:   false,
	}, nil
}

func (s *sinkHandle) closeWriter() error {
	if s == nil || s.writer == nil {
		return nil
	}
	return s.writer.Close()
}

// finalizeFailure ensures temp files are removed when the caller handles failure paths.
func (s *sinkHandle) finalizeFailure() {
	if s == nil || s.finalize == nil {
		return
	}
	_ = s.finalize(false)
}

// finalizeSuccess promotes the temp file on successful downloads.
func (s *sinkHandle) finalizeSuccess() error {
	if s == nil || s.finalize == nil {
		return nil
	}
	return s.finalize(true)
}

// nopCloser adapts an io.Writer to io.WriteCloser.
type nopCloser struct {
	io.Writer
}

func (n nopCloser) Close() error { return nil }
