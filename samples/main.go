package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// -------- 參數 --------
var (
	listFile   = flag.String("list", "", "含多個下載網址的文字檔 (每行一個 URL)")
	outDir     = flag.String("out", "downloads", "下載輸出目錄")
	workers    = flag.Int("workers", 8, "同時下載的 worker 數量（併發）")
	timeoutSec = flag.Int("timeout", 60, "單次請求逾時（秒）")
	retries    = flag.Int("retries", 3, "失敗重試次數")
)

// -------- 全域頻寬計數 --------
var (
	bytesThisSec int64 // 每秒累計，用於即時速率
	bytesTotal   int64 // 全程累計
)

// counterWriter 會把寫入的位元組數累加到原子計數器
type counterWriter struct {
	dst io.Writer
}

func (cw *counterWriter) Write(p []byte) (int, error) {
	n, err := cw.dst.Write(p)
	atomic.AddInt64(&bytesThisSec, int64(n))
	atomic.AddInt64(&bytesTotal, int64(n))
	return n, err
}

// -------- 工具 --------
func must(err error) {
	if err != nil {
		panic(err)
	}
}

func ensureDir(d string) {
	must(os.MkdirAll(d, 0o755))
}

func fileNameFromURL(u string) string {
	// 簡單擷取檔名；若最後沒有檔名，給一個時間戳
	parts := strings.Split(u, "/")
	fn := parts[len(parts)-1]
	if fn == "" || strings.HasSuffix(u, "/") {
		return fmt.Sprintf("download_%d.bin", time.Now().UnixNano())
	}
	// 去 querystring
	if i := strings.Index(fn, "?"); i >= 0 {
		fn = fn[:i]
	}
	return fn
}

func humanBitsPerSec(bps float64) string {
	// 轉換為可讀的 Mbit/s、Gbit/s
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
	default:
		return fmt.Sprintf("%.0f bit/s", bps)
	}
}

func humanBytes(b float64) string {
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
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}

// 自訂 HTTP Client：加大連線上限，盡量吃滿頻寬
func makeHTTPClient(timeout time.Duration) *http.Client {
	transport := &http.Transport{
		// 更積極的連線配置
		MaxIdleConns:        1024,
		MaxConnsPerHost:     0,    // 0 表示不限制（Go 1.20+）
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false, // 讓伺服器可視需要壓縮；大檔多半是二進位，不太有差
		ForceAttemptHTTP2:   true,  // HTTPS 預設就啟 HTTP/2，有助多路複用
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   15 * time.Second,
		}).DialContext,
		// 對付某些慢伺服器，增加 connection pool 容量
		MaxIdleConnsPerHost: 256,
		// ExpectContinueTimeout: 0，
	}
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// 單一檔案下載（含重試）
func downloadOne(ctx context.Context, client *http.Client, url string, outPath string, retries int) error {
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		func() {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				lastErr = err
				return
			}
			resp, err := client.Do(req)
			if err != nil {
				lastErr = err
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
				return
			}

			tmp := outPath + ".part"
			f, err := os.Create(tmp)
			if err != nil {
				lastErr = err
				return
			}
			defer f.Close()

			// 用 counterWriter 統計即時位元組數
			cw := &counterWriter{dst: f}
			// 自行搬運（比 io.Copy 還要能在需要時做細緻控制）
			buf := make([]byte, 1<<20) // 1 MiB buffer：大 buffer 可減少 syscalls
			_, err = io.CopyBuffer(cw, resp.Body, buf)
			if err != nil {
				lastErr = err
				return
			}
			// 成功才 rename
			if err := os.Rename(tmp, outPath); err != nil {
				lastErr = err
				return
			}
			lastErr = nil
		}()
		if lastErr == nil {
			return nil
		}
		// 指數退避後重試
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(500*(1<<attempt)) * time.Millisecond):
		}
	}
	return lastErr
}

// worker：從 job chan 拿 URL，執行下載
func worker(ctx context.Context, wg *sync.WaitGroup, client *http.Client, jobs <-chan string, outDir string, retries int) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case url, ok := <-jobs:
			if !ok {
				return
			}
			fn := fileNameFromURL(url)
			outPath := filepath.Join(outDir, fn)
			if err := downloadOne(ctx, client, url, outPath, retries); err != nil {
				fmt.Printf("[FAIL] %s -> %v\n", url, err)
			} else {
				fmt.Printf("[OK]   %s -> %s\n", url, outPath)
			}
		}
	}
}

// 每秒輸出即時頻寬
func startLiveBandwidthPrinter(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		var ewma float64
		alpha := 0.25 // 平滑係數
		start := time.Now()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 取得並清零「這一秒」的 byte 計數
				b := atomic.SwapInt64(&bytesThisSec, 0)
				// 轉成 bit/s
				bps := float64(b) * 8
				if ewma == 0 {
					ewma = bps
				} else {
					ewma = alpha*bps + (1-alpha)*ewma
				}
				total := atomic.LoadInt64(&bytesTotal)
				elapsed := time.Since(start).Seconds()
				avgBps := (float64(total) * 8) / elapsed

				fmt.Printf("[BW] now=%s  ewma=%s  avg=%s  total=%s\n",
					humanBitsPerSec(bps),
					humanBitsPerSec(ewma),
					humanBitsPerSec(avgBps),
					humanBytes(float64(total)),
				)
			}
		}
	}()
}

func readURLs(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var urls []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}
	return urls, sc.Err()
}

func main() {
	flag.Parse()
	if *listFile == "" {
		fmt.Println("用法：go run . -list urls.txt [-workers 8] [-out downloads]")
		os.Exit(1)
	}

	urls, err := readURLs(*listFile)
	must(err)
	if len(urls) == 0 {
		fmt.Println("URL 清單是空的。")
		os.Exit(1)
	}

	ensureDir(*outDir)
	client := makeHTTPClient(time.Duration(*timeoutSec) * time.Second)

	// 控制整體生命週期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	startLiveBandwidthPrinter(ctx, &wg)

	jobs := make(chan string, *workers*2)
	// 啟動 workers
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go worker(ctx, &wg, client, jobs, *outDir, *retries)
	}

	// 丟任務
	go func() {
		for _, u := range urls {
			jobs <- u
		}
		close(jobs)
	}()

	// 等 workers 完成
	wgWorkers := sync.WaitGroup{}
	wgWorkers.Add(1)
	go func() {
		defer wgWorkers.Done()
		for {
			// 粗暴的等待：偵測 jobs 已關閉且 channel 清空，並確認沒有 goroutine 在工作
			// 這裡交由主 wg（live printer + workers）在 cancel 後結束
			time.Sleep(250 * time.Millisecond)
			if len(jobs) == 0 {
				// 讓 worker 稍等跑完當前任務
				time.Sleep(2 * time.Second)
				cancel()
				return
			}
		}
	}()
	wgWorkers.Wait()
	wg.Wait()

	// 最後印出總結
	total := atomic.LoadInt64(&bytesTotal)
	fmt.Printf("\n[SUMMARY] 下載完成：總量=%s\n", humanBytes(float64(total)))
}

