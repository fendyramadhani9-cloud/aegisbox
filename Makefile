.PHONY: all build test vet fmt fmt-check clean help

BINARY_NAME=aegisbox
BIN_DIR=bin
CMD_PATH=./cmd/aegisbox

all: fmt vet test build

build:
	@echo "==> Building AegisBox binary..."
	go build -v -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)

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
	rm -rf $(BIN_DIR)

help:
	@echo "AegisBox Build Targets:"
	@echo "  make build      - Build binary to bin/aegisbox"
	@echo "  make test       - Run all unit tests with race detector"
	@echo "  make vet        - Run go vet static analysis"
	@echo "  make fmt        - Format all Go source files"
	@echo "  make fmt-check  - Check if files are formatted"
	@echo "  make clean      - Remove build artifacts"
