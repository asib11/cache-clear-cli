BINARY=medusa-cache-cli
VERSION=1.0.0
GOFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: all build install clean run-status run-clear help

## build: Build the binary for current OS
build:
	go build $(GOFLAGS) -o $(BINARY) .

## build-linux: Build for Linux (amd64)
build-linux:
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o $(BINARY)-linux-amd64 .

## build-mac: Build for macOS (arm64 / Apple Silicon)
build-mac:
	GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -o $(BINARY)-darwin-arm64 .

## build-all: Build for Linux + macOS
build-all: build-linux build-mac

## install: Install the binary to /usr/local/bin
install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed to /usr/local/bin/$(BINARY)"

## clean: Remove built binaries
clean:
	rm -f $(BINARY) $(BINARY)-linux-amd64 $(BINARY)-darwin-arm64

## run-status: Run status check (reads .env)
run-status: build
	./$(BINARY) status

## run-clear: Clear all caches (reads .env, asks for confirmation)
run-clear: build
	./$(BINARY) clear

## run-clear-yes: Clear all caches without confirmation
run-clear-yes: build
	./$(BINARY) clear --yes

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
