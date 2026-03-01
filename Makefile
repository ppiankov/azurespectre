BINARY_NAME := azurespectre
MAIN_PATH   := ./cmd/azurespectre
BUILD_DIR   := bin

VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
VERSION_NUM  = $(patsubst v%,%,$(VERSION))
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE        ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS     := -s -w -X main.version=$(VERSION_NUM) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build test lint coverage clean install

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

test:
	go test -race ./...

lint:
	golangci-lint run --timeout=5m

coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

clean:
	rm -rf $(BUILD_DIR) coverage.out

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
