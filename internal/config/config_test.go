package config

import (
	"flag"
	"testing"
	"time"
)

func TestNormalizeDefaults(t *testing.T) {
	cfg := &Config{
		ListPath: "urls.txt",
		Save:     true,
		Timeout:  time.Second,
		Retries:  1,
	}
	if err := cfg.Normalize(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutDir == "" {
		t.Fatalf("expected default out dir, got empty")
	}
	if cfg.Workers <= 0 {
		t.Fatalf("expected workers to be set, got %d", cfg.Workers)
	}
}

func TestNormalizeOutDirImpliesSave(t *testing.T) {
	cfg := &Config{
		ListPath: "urls.txt",
		OutDir:   "./downloads/test",
		Timeout:  time.Second,
		Retries:  0,
	}
	if err := cfg.Normalize(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Save {
		t.Fatalf("expected save to be true when out dir provided")
	}
	if got := cfg.OutDir; got != "downloads/test" {
		t.Fatalf("expected cleaned path downloads/test, got %s", got)
	}
}

func TestParseWithFlagSet(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg, err := Parse(fs, []string{"-list", "urls.txt", "-workers", "4", "-timeout", "30s"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Workers != 4 {
		t.Fatalf("expected 4 workers, got %d", cfg.Workers)
	}
	if cfg.Timeout != 30*time.Second {
		t.Fatalf("unexpected timeout: %v", cfg.Timeout)
	}
}

func TestParseMissingList(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	if _, err := Parse(fs, []string{}); err == nil {
		t.Fatalf("expected error for missing list flag")
	}
}
