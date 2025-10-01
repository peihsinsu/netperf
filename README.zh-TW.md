# netperf - 網路效能測試工具

一個用 Go 開發的網路頻寬測試 CLI 工具，用於壓測下載頻寬與併發 HTTP 傳輸。

## 專案結構

```
netperf/
├── cmd/
│   └── bandfetch/          # CLI 主程式入口
│       └── main.go
├── internal/
│   ├── config/             # 參數解析與驗證
│   │   ├── config.go
│   │   └── config_test.go
│   ├── downloader/         # 下載管理與實作
│   │   ├── download.go     # HTTP 下載核心邏輯（重試、超時）
│   │   ├── download_test.go
│   │   ├── http_client.go  # 優化的 HTTP 客戶端設定
│   │   ├── manager.go      # Worker pool 調度
│   │   ├── manager_test.go
│   │   ├── naming.go       # 檔名處理工具
│   │   ├── sink.go         # 資料寫入介面（檔案/捨棄）
│   │   └── sink_test.go
│   ├── metrics/            # 頻寬統計與輸出
│   │   ├── aggregator.go   # 原子計數器、EWMA 追蹤
│   │   ├── aggregator_test.go
│   │   └── printer.go      # 即時頻寬輸出（ticker）
│   └── urls/
│       └── reader.go       # URL 清單解析
├── prd/
│   └── design.md           # 專案設計文件
├── samples/
│   └── main.go             # 原型範例程式
├── go.mod
├── go.work
├── Makefile
└── README.md
```

## 功能特色

- **併發下載**：Worker pool 架構，可同時處理多個下載任務
- **即時頻寬監控**：每秒更新頻寬統計（瞬時值、EWMA、平均值）
- **靈活的儲存選項**：可選擇儲存檔案或僅測試頻寬（不寫入磁碟）
- **可靠性**：支援自動重試、指數退避、超時控制
- **優化的 HTTP 客戶端**：調整連線池大小、啟用 HTTP/2
- **優雅退出**：支援 Ctrl+C 中斷並顯示完整匯總報告
- **匯總報告**：包含峰值與平均頻寬的詳細統計資訊
- **完整測試覆蓋**：單元測試與整合測試

## 安裝與編譯

### 預編譯執行檔

從 [Releases](https://github.com/cx009/netperf/releases) 頁面下載適合你平台的最新版本。

**支援平台：**
- Linux (AMD64, ARM64)
- Windows (AMD64, ARM64)
- macOS (AMD64, ARM64/Apple Silicon)

### 從原始碼編譯

```bash
# 取得專案
git clone https://github.com/cx009/netperf.git
cd netperf

# 編譯當前平台
make build

# 編譯所有平台
make build-all

# 或直接執行
go run ./cmd/bandfetch -list urls.txt
```

詳細編譯說明請參考 [BUILD.md](BUILD.md)（英文）。

### 開發指令

```bash
# 執行測試
make test

# 格式化程式碼
make fmt

# 編譯特定平台
make build-linux-amd64
make build-windows-amd64
make build-darwin-arm64

# 編譯所有平台
make build-all

# 建立發布套件
bash build-release.sh

# 清理編譯產物
make clean
```

## 使用方式

### 基本用法

```bash
# 從 URL 清單下載（預設不儲存檔案）
./bin/bandfetch -list urls.txt

# 下載並儲存到預設目錄 (downloads/)
./bin/bandfetch -list urls.txt -save

# 指定輸出目錄（自動啟用儲存）
./bin/bandfetch -list urls.txt -out ./my-downloads

# 調整 worker 數量（預設為 CPU 核心數 * 2，上限 64）
./bin/bandfetch -list urls.txt -workers 16

# 設定請求超時時間
./bin/bandfetch -list urls.txt -timeout 120s

# 設定重試次數
./bin/bandfetch -list urls.txt -retries 5

# 關閉即時頻寬輸出（適合腳本使用）
./bin/bandfetch -list urls.txt -progress=false
```

### 使用 Makefile

```bash
# 執行下載（僅測試頻寬）
make run LIST=urls.txt WORKERS=12

# 下載並儲存
make run LIST=urls.txt SAVE=1 OUT=downloads WORKERS=16
```

### URL 清單格式

建立一個文字檔（如 `urls.txt`），每行一個 URL：

```
https://example.com/file1.bin
https://example.com/file2.bin
# 這是註解，會被忽略

https://example.com/file3.bin
```

- 支援空行
- 以 `#` 開頭的行會被視為註解

## 輸出範例

### 下載過程中

```
[OK] https://example.com/file1.bin (discarded)
[OK] https://example.com/file2.bin -> downloads/file2.bin
[BW] now=125.43 Mbit/s  ewma=118.76 Mbit/s  avg=115.22 Mbit/s  total=1.23 GiB
[BW] now=132.18 Mbit/s  ewma=122.12 Mbit/s  avg=117.45 Mbit/s  total=1.39 GiB
[FAIL] https://example.com/timeout.bin -> Get: context deadline exceeded
```

### 匯總報告（完成或按 Ctrl+C 時）

```
╔══════════════════════════════════════════════════════╗
║              Download Summary Report                 ║
╠══════════════════════════════════════════════════════╣
║  Total Downloaded : 2.45 GiB                         ║
║  Elapsed Time     : 2m 15s                           ║
║  Average Speed    : 145.67 Mbit/s                    ║
║  Peak Speed       : 182.33 Mbit/s                    ║
╚══════════════════════════════════════════════════════╝
```

### 優雅退出

隨時按下 **Ctrl+C** 可優雅地停止下載並查看匯總報告：
```
^C
[INTERRUPT] Received signal interrupt, shutting down gracefully...
（顯示上方的匯總報告）
```

## 技術細節

### 架構設計

- **Manager**：管理 job channel，啟動 workers，處理生命週期
- **Downloader**：使用優化的 `http.Client`（高連線上限、HTTP/2），實作重試與指數退避
- **Sink**：提供兩種實作
  - `FileSink`：寫入 `.part` 暫存檔，成功後 rename
  - `DiscardSink`：包裝 `io.Discard`，僅統計流量
- **Metrics**：原子計數器追蹤每秒/總計流量，計算 EWMA 與平均值

### 併發與頻寬追蹤

- 使用 buffered channel（大小為 `workers * 2`）作為任務佇列
- 每個 worker 從 channel 讀取任務直到關閉或 context 取消
- 位元組計數透過 `counterWriter` → `metrics.Aggregator.AddBytes`
- 使用原子操作確保併發安全
- 每秒 ticker 交換計數器、轉換為 bit/s、更新 EWMA（alpha = 0.25）

### 錯誤處理

- 暫存檔機制：成功時 rename，失敗時刪除 `.part` 檔
- 重試策略：初始 500ms，每次加倍，加入 ±20% 抖動避免同步
- 所有重試與 HTTP 請求都遵守 context 取消
- HTTP 狀態碼 ≥ 400 視為錯誤

## 效能優化建議

### 提升頻寬使用率

1. **增加 worker 數量**：`-workers 24` 或更高（視來源站點與網路環境）
2. **多樣化來源**：從不同主機下載可避免單一來源限速
3. **使用 SSD 或 tmpfs**：避免磁碟 I/O 成為瓶頸
4. **HTTP/2 多路複用**：HTTPS 連線自動啟用，對同一主機更有效率
5. **大 buffer**：程式已使用 1MiB buffer，減少系統呼叫開銷

### 偵測瓶頸

若頻寬長期達不到預期：
- 檢查來源端是否有限速
- 確認 ISP 或路由器是否壅塞
- 驗證本機磁碟寫入速度
- 嘗試 `-save=false` 測試純網路效能

## 開發資訊

- **語言**：Go 1.22+
- **測試框架**：標準 testing 套件
- **外部依賴**：無（僅使用標準函式庫）

## 授權

此專案為個人學習與測試用途。

## 相關文件

- [設計文件](prd/design.md)
- [原型範例](samples/main.go)
