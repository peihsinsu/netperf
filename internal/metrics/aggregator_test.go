package metrics

import (
	"testing"
	"time"
)

func TestAggregatorCounts(t *testing.T) {
	agg := NewAggregator()
	agg.AddBytes(512)
	if got := agg.TotalBytes(); got != 512 {
		t.Fatalf("expected 512 bytes, got %d", got)
	}

	swapped := agg.SwapBytesThisSecond()
	if swapped != 512 {
		t.Fatalf("expected swap to return 512, got %d", swapped)
	}
	if agg.SwapBytesThisSecond() != 0 {
		t.Fatalf("counter should have reset to zero")
	}
}

func TestEWMA(t *testing.T) {
	ewma := NewEWMA(0.5)
	values := []float64{100, 200, 300}
	want := []float64{100, 150, 225}
	for i, v := range values {
		got := ewma.Update(v)
		if got != want[i] {
			t.Fatalf("step %d: expected %v, got %v", i, want[i], got)
		}
	}
}

func TestAverageBps(t *testing.T) {
	agg := NewAggregator()
	agg.AddBytes(1024)
	time.Sleep(10 * time.Millisecond)
	if avg := agg.AverageBps(); avg <= 0 {
		t.Fatalf("expected positive average, got %f", avg)
	}
}

func TestHumanFormatting(t *testing.T) {
	if got := HumanBitsPerSecond(1500); got != "1.50 Kbit/s" {
		t.Fatalf("unexpected human bps: %s", got)
	}
	if got := HumanBytes(2048); got != "2.00 KiB" {
		t.Fatalf("unexpected human bytes: %s", got)
	}
}

func TestPeakBps(t *testing.T) {
	agg := NewAggregator()

	// Initially peak should be zero
	if peak := agg.PeakBps(); peak != 0 {
		t.Fatalf("expected initial peak to be 0, got %f", peak)
	}

	// Update with increasing values
	agg.UpdatePeakBps(1000)
	if peak := agg.PeakBps(); peak != 1000 {
		t.Fatalf("expected peak 1000, got %f", peak)
	}

	agg.UpdatePeakBps(5000)
	if peak := agg.PeakBps(); peak != 5000 {
		t.Fatalf("expected peak 5000, got %f", peak)
	}

	// Lower value should not update peak
	agg.UpdatePeakBps(3000)
	if peak := agg.PeakBps(); peak != 5000 {
		t.Fatalf("expected peak to remain 5000, got %f", peak)
	}

	// Zero or negative should not update
	agg.UpdatePeakBps(0)
	agg.UpdatePeakBps(-100)
	if peak := agg.PeakBps(); peak != 5000 {
		t.Fatalf("expected peak to remain 5000 after invalid updates, got %f", peak)
	}
}

func TestGetSummary(t *testing.T) {
	agg := NewAggregator()
	agg.AddBytes(1024 * 1024) // 1 MiB
	agg.UpdatePeakBps(125000000) // 125 Mbit/s

	time.Sleep(100 * time.Millisecond)

	summary := agg.GetSummary()

	if summary.TotalBytes != 1024*1024 {
		t.Fatalf("expected total bytes 1048576, got %d", summary.TotalBytes)
	}

	if summary.PeakBps != 125000000 {
		t.Fatalf("expected peak bps 125000000, got %f", summary.PeakBps)
	}

	if summary.TotalSizeStr != "1.00 MiB" {
		t.Fatalf("unexpected total size string: %s", summary.TotalSizeStr)
	}

	if summary.PeakBpsStr != "125.00 Mbit/s" {
		t.Fatalf("unexpected peak bps string: %s", summary.PeakBpsStr)
	}

	if summary.AverageBps <= 0 {
		t.Fatalf("expected positive average bps, got %f", summary.AverageBps)
	}

	if summary.Elapsed <= 0 {
		t.Fatalf("expected positive elapsed time, got %v", summary.Elapsed)
	}
}

func TestFormatSummary(t *testing.T) {
	summary := Summary{
		TotalBytes:   1024 * 1024 * 100,
		Elapsed:      65 * time.Second,
		AverageBps:   100000000,
		PeakBps:      150000000,
		TotalSizeStr: "100.00 MiB",
		ElapsedStr:   "1m 5s",
		AvgBpsStr:    "100.00 Mbit/s",
		PeakBpsStr:   "150.00 Mbit/s",
	}

	output := summary.FormatSummary()

	// Check that output contains expected strings
	expectedStrings := []string{
		"Download Summary Report",
		"Total Downloaded",
		"100.00 MiB",
		"Elapsed Time",
		"1m 5s",
		"Average Speed",
		"100.00 Mbit/s",
		"Peak Speed",
		"150.00 Mbit/s",
	}

	for _, expected := range expectedStrings {
		if !contains(output, expected) {
			t.Fatalf("expected output to contain '%s', got:\n%s", expected, output)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{5 * time.Second, "5.0s"},
		{45 * time.Second, "45.0s"},
		{90 * time.Second, "1m 30s"},
		{3665 * time.Second, "1h 1m 5s"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.duration)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %s, want %s", tt.duration, got, tt.want)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
