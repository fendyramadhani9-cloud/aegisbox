.PHONY: all build test vet fmt fmt-check package clean help

BINARY_NAME=aegisbox
BIN_DIR=bin
DIST_DIR=dist
CMD_PATH=./cmd/aegisbox
VERSION=0.1.0
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || echo "unknown")
LDFLAGS=-s -w -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME)

all: fmt vet test build

build:
	@echo "==> Building AegisBox binary..."
	go build -ldflags "$(LDFLAGS)" -v -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)

package:
	@echo "==> Packaging release tarball for Linux amd64..."
	@mkdir -p $(BIN_DIR) $(DIST_DIR)
	@echo "VERSION=$(VERSION)" > RELEASE_METADATA
	@echo "COMMIT_SHA=$(GIT_COMMIT)" >> RELEASE_METADATA
	@echo "BUILD_TIME=$(BUILD_TIME)" >> RELEASE_METADATA
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -v -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)
	tar -czvf $(DIST_DIR)/$(BINARY_NAME)-linux-amd64.tar.gz \
		$(BIN_DIR)/$(BINARY_NAME) \
		configs/config.yaml \
		deploy/aegisbox.service \
		scripts/install.sh \
		scripts/deploy.sh \
		scripts/setup-rootfs.sh \
		RELEASE_METADATA
	@echo "==> Package created: $(DIST_DIR)/$(BINARY_NAME)-linux-amd64.tar.gz"

test:
	@echo "==> Running unit tests..."
	go test -v -race ./...

vet:
	@echo "==> Running go vet..."
	go vet ./...

fmt:
	@echo "==> Formatting code..."
	gofmt -w .

fmt-check:
	@echo "==> Checking code formatting..."
	@test -z "$$(gofmt -l .)" || (echo "Unformatted files found:" && gofmt -l . && exit 1)

clean:
	@echo "==> Cleaning build artifacts..."
	rm -rf $(BIN_DIR) $(DIST_DIR) RELEASE_METADATA

help:
	@echo "AegisBox Build Targets:"
	@echo "  make build      - Build binary to bin/aegisbox with version ldflags"
	@echo "  make package    - Compile and package aegisbox-linux-amd64.tar.gz"
	@echo "  make test       - Run all unit tests with race detector"
	@echo "  make vet        - Run go vet static analysis"
	@echo "  make fmt        - Format all Go source files"
	@echo "  make fmt-check  - Check if files are formatted"
	@echo "  make clean      - Remove build and release artifacts"
