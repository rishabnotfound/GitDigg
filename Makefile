BINARY := gitdigg
BUILD_DIR := bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X github.com/rishabnotfound/gitdigg/internal/version.Version=$(VERSION) -X github.com/rishabnotfound/gitdigg/internal/version.Commit=$(COMMIT) -X github.com/rishabnotfound/gitdigg/internal/version.Date=$(DATE)"

PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

.PHONY: build clean test fmt tidy cross-build

build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/gitdigg

clean:
	@rm -rf $(BUILD_DIR)

test:
	go test -v -race ./...

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

cross-build:
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-$${platform%/*}-$${platform#*/} ./cmd/gitdigg; \
	done
