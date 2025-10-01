package metrics

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"
)

// Aggregator tracks download byte counts for bandwidth calculations.
type Aggregator struct {
	bytesThisSec atomic.Int64
	bytesTotal   atomic.Int64
	peakBps      atomic.Uint64 // stored as uint64 bits representation of float64
	start        time.Time
}

// NewAggregator constructs an Aggregator with current start time.
func NewAggregator() *Aggregator {
	return &Aggregator{start: time.Now()}
}

// AddBytes increases the counters by n bytes.
func (a *Aggregator) AddBytes(n int) {
	if n <= 0 {
		return
	}
	a.bytesThisSec.Add(int64(n))
	a.bytesTotal.Add(int64(n))
}

// SwapBytesThisSecond atomically swaps the per-second counter with zero and returns the previous value.
func (a *Aggregator) SwapBytesThisSecond() int64 {
	return a.bytesThisSec.Swap(0)
}

// TotalBytes returns the total bytes downloaded.
func (a *Aggregator) TotalBytes() int64 {
	return a.bytesTotal.Load()
}

// Elapsed returns the duration since the aggregator was created.
func (a *Aggregator) Elapsed() time.Duration {
	return time.Since(a.start)
}

// HumanBitsPerSecond renders a bit/s value in a readable unit.
func HumanBitsPerSecond(bps float64) string {
	const (
		K = 1000.0
		M = K * 1000
		G = M * 1000
	)
	switch {
	case bps >= G:
		return fmt.Sprintf("%.2f Gbit/s", bps/G)
	case bps >= M:
		return fmt.Sprintf("%.2f Mbit/s", bps/M)
	case bps >= K:
		return fmt.Sprintf("%.2f Kbit/s", bps/K)
	case bps > 0:
		return fmt.Sprintf("%.0f bit/s", bps)
	default:
		return "0 bit/s"
	}
}

// HumanBytes renders a byte value in IEC units.
func HumanBytes(b float64) string {
	const (
		KiB = 1024.0
		MiB = KiB * 1024
		GiB = MiB * 1024
	)
	switch {
	case b >= GiB:
		return fmt.Sprintf("%.2f GiB", b/GiB)
	case b >= MiB:
		return fmt.Sprintf("%.2f MiB", b/MiB)
	case b >= KiB:
		return fmt.Sprintf("%.2f KiB", b/KiB)
	case b > 0:
		return fmt.Sprintf("%.0f B", b)
	default:
		return "0 B"
	}
}

// AverageBps computes the average bit/s throughput.
func (a *Aggregator) AverageBps() float64 {
	elapsed := a.Elapsed().Seconds()
	if elapsed <= 0 {
		return 0
	}
	totalBits := float64(a.TotalBytes()) * 8
	return totalBits / elapsed
}

// UpdatePeakBps updates the peak bandwidth if the current value is higher.
func (a *Aggregator) UpdatePeakBps(bps float64) {
	if bps <= 0 || math.IsNaN(bps) || math.IsInf(bps, 0) {
		return
	}
	newBits := math.Float64bits(bps)
	for {
		oldBits := a.peakBps.Load()
		oldBps := math.Float64frombits(oldBits)
		if bps <= oldBps {
			return
		}
		if a.peakBps.CompareAndSwap(oldBits, newBits) {
			return
		}
	}
}

// PeakBps returns the highest bandwidth observed.
func (a *Aggregator) PeakBps() float64 {
	bits := a.peakBps.Load()
	if bits == 0 {
		return 0
	}
	return math.Float64frombits(bits)
}

// EWMA maintains an exponentially weighted moving average of the supplied series.
type EWMA struct {
	value float64
	alpha float64
}

// NewEWMA creates a new EWMA with provided smoothing factor.
func NewEWMA(alpha float64) *EWMA {
	if alpha <= 0 || alpha >= 1 {
		alpha = 0.25
	}
	return &EWMA{alpha: alpha}
}

// Update adds a new observation and returns the current value.
func (e *EWMA) Update(observation float64) float64 {
	if math.IsNaN(observation) || math.IsInf(observation, 0) {
		return e.value
	}
	if e.value == 0 {
		e.value = observation
	} else {
		e.value = e.alpha*observation + (1-e.alpha)*e.value
	}
	return e.value
}

// Summary contains aggregated statistics for a download session.
type Summary struct {
	TotalBytes   int64
	Elapsed      time.Duration
	AverageBps   float64
	PeakBps      float64
	TotalSizeStr string
	ElapsedStr   string
	AvgBpsStr    string
	PeakBpsStr   string
}

// GetSummary returns a formatted summary of the download statistics.
func (a *Aggregator) GetSummary() Summary {
	totalBytes := a.TotalBytes()
	elapsed := a.Elapsed()
	avgBps := a.AverageBps()
	peakBps := a.PeakBps()

	return Summary{
		TotalBytes:   totalBytes,
		Elapsed:      elapsed,
		AverageBps:   avgBps,
		PeakBps:      peakBps,
		TotalSizeStr: HumanBytes(float64(totalBytes)),
		ElapsedStr:   formatDuration(elapsed),
		AvgBpsStr:    HumanBitsPerSecond(avgBps),
		PeakBpsStr:   HumanBitsPerSecond(peakBps),
	}
}

// FormatSummary returns a multi-line formatted summary report.
func (s Summary) FormatSummary() string {
	return fmt.Sprintf(`
╔══════════════════════════════════════════════════════╗
║              Download Summary Report                 ║
╠══════════════════════════════════════════════════════╣
║  Total Downloaded : %-31s  ║
║  Elapsed Time     : %-31s  ║
║  Average Speed    : %-31s  ║
║  Peak Speed       : %-31s  ║
╚══════════════════════════════════════════════════════╝`,
		s.TotalSizeStr,
		s.ElapsedStr,
		s.AvgBpsStr,
		s.PeakBpsStr,
	)
}

// formatDuration formats a duration in a human-readable format.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%dh %dm %ds", h, m, s)
}
