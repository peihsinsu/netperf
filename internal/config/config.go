package config

import (
	"errors"
	"flag"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Config holds parsed CLI settings.
type Config struct {
	ListPath string
	Save     bool
	OutDir   string
	Workers  int
	Timeout  time.Duration
	Retries  int
	Progress bool
}

// DefaultWorkers returns the default worker count based on CPU cores.
func DefaultWorkers() int {
	w := runtime.NumCPU() * 2
	if w < 1 {
		w = 1
	}
	if w > 64 {
		return 64
	}
	return w
}

// Normalize validates and fills derived values.
func (c *Config) Normalize() error {
	if c.ListPath == "" {
		return errors.New("-list is required")
	}
	if c.Workers <= 0 {
		c.Workers = DefaultWorkers()
	}
	if c.Timeout <= 0 {
		return errors.New("-timeout must be greater than 0")
	}
	if c.Retries < 0 {
		return errors.New("-retries cannot be negative")
	}

	if strings.TrimSpace(c.OutDir) != "" {
		c.OutDir = filepath.Clean(c.OutDir)
		c.Save = true
	} else if c.Save {
		c.OutDir = "downloads"
	}

	if c.Save && c.OutDir == "" {
		c.OutDir = "downloads"
	}

	return nil
}

// Parse uses the provided FlagSet to load configuration from flags.
func Parse(fs *flag.FlagSet, args []string) (*Config, error) {
	list := fs.String("list", "", "path to URL list (required)")
	save := fs.Bool("save", false, "persist downloads to disk (default discards)")
	out := fs.String("out", "", "directory for downloads; implies -save when set")
	workers := fs.Int("workers", 0, "number of concurrent download workers")
	timeout := fs.Duration("timeout", 60*time.Second, "per-request timeout (e.g. 45s, 2m)")
	retries := fs.Int("retries", 3, "retry attempts beyond the first request")
	progress := fs.Bool("progress", true, "enable live bandwidth output")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg := &Config{
		ListPath: *list,
		Save:     *save,
		OutDir:   *out,
		Workers:  *workers,
		Timeout:  *timeout,
		Retries:  *retries,
		Progress: *progress,
	}

	if err := cfg.Normalize(); err != nil {
		return nil, err
	}
	return cfg, nil
}
