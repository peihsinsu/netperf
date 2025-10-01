BINARY ?= bandfetch
LIST ?=
OUT ?=
SAVE ?=
WORKERS ?=
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test fmt run tidy clean build-all build-linux build-windows build-darwin build-arm

build:
	@mkdir -p bin
	GO111MODULE=on go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/bandfetch

# Build for all platforms
build-all: build-linux build-windows build-darwin
	@echo "✓ All platform builds completed"

# Linux builds
build-linux: build-linux-amd64 build-linux-arm64

build-linux-amd64:
	@mkdir -p bin
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-amd64 ./cmd/bandfetch

build-linux-arm64:
	@mkdir -p bin
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 GO111MODULE=on go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-arm64 ./cmd/bandfetch

# Windows builds
build-windows: build-windows-amd64 build-windows-arm64

build-windows-amd64:
	@mkdir -p bin
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 GO111MODULE=on go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-windows-amd64.exe ./cmd/bandfetch

build-windows-arm64:
	@mkdir -p bin
	@echo "Building for Windows (arm64)..."
	GOOS=windows GOARCH=arm64 GO111MODULE=on go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-windows-arm64.exe ./cmd/bandfetch

# macOS builds
build-darwin: build-darwin-amd64 build-darwin-arm64

build-darwin-amd64:
	@mkdir -p bin
	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 GO111MODULE=on go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-darwin-amd64 ./cmd/bandfetch

build-darwin-arm64:
	@mkdir -p bin
	@echo "Building for macOS (arm64/Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 GO111MODULE=on go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-darwin-arm64 ./cmd/bandfetch

test:
	GO111MODULE=on go test ./...

fmt:
	GO111MODULE=on gofmt -w ./

run:
	@if [ -z "$(LIST)" ]; then \
		echo "usage: make run LIST=urls.txt [SAVE=1] [OUT=dir] [WORKERS=16]"; \
		exit 1; \
	fi
	@CMD="GO111MODULE=on go run ./cmd/bandfetch -list $(LIST)"; \
	if [ "$(SAVE)" = "1" ] || [ "$(SAVE)" = "true" ]; then \
		CMD="$$CMD -save"; \
	fi; \
	if [ -n "$(OUT)" ]; then \
		CMD="$$CMD -out $(OUT)"; \
	fi; \
	if [ -n "$(WORKERS)" ]; then \
		CMD="$$CMD -workers $(WORKERS)"; \
	fi; \
	echo "$$CMD"; \
	eval "$$CMD"

tidy:
	GO111MODULE=on go mod tidy

clean:
	rm -rf bin

