package downloader

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// Manager coordinates concurrent downloads.
type Manager struct {
	downloader *Downloader
	workers    int
}

// NewManager constructs a Manager with the specified worker count.
func NewManager(d *Downloader, workers int) *Manager {
	if workers < 1 {
		workers = 1
	}
	return &Manager{downloader: d, workers: workers}
}

// Run processes the provided URLs with the configured worker pool.
func (m *Manager) Run(ctx context.Context, urls []string) error {
	jobs := make(chan string, m.workers*2)
	var wg sync.WaitGroup
	var failed atomic.Int32

	for i := 0; i < m.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case url, ok := <-jobs:
					if !ok {
						return
					}
					res, err := m.downloader.Download(ctx, url)
					if err != nil {
						failed.Add(1)
						fmt.Printf("[FAIL] %s -> %v\n", url, err)
						continue
					}
					if res.Discarded {
						fmt.Printf("[OK]   %s (discarded)\n", url)
					} else {
						fmt.Printf("[OK]   %s -> %s\n", url, res.Destination)
					}
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, u := range urls {
			select {
			case <-ctx.Done():
				return
			case jobs <- u:
			}
		}
	}()

	wg.Wait()

	if n := failed.Load(); n > 0 {
		return fmt.Errorf("%d downloads failed", n)
	}
	return nil
}
