package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StartPrinter launches a goroutine that periodically prints bandwidth metrics.
func StartPrinter(ctx context.Context, agg *Aggregator, enabled bool, wg *sync.WaitGroup) {
	if !enabled {
		return
	}
	if agg == nil {
		return
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		ewma := NewEWMA(0.25)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				bytes := agg.SwapBytesThisSecond()
				bps := float64(bytes) * 8
				smoothed := ewma.Update(bps)
				avg := agg.AverageBps()
				total := agg.TotalBytes()

				// Track peak bandwidth
				agg.UpdatePeakBps(bps)

				fmt.Printf("[BW] now=%s  ewma=%s  avg=%s  total=%s\n",
					HumanBitsPerSecond(bps),
					HumanBitsPerSecond(smoothed),
					HumanBitsPerSecond(avg),
					HumanBytes(float64(total)),
				)
			}
		}
	}()
}
