VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build dev clean test

## build: compile the binary
build:
	CGO_ENABLED=1 go build $(LDFLAGS) -o llmview .

## dev: run with auto-reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air -c .air.toml || go run .

## test: run all tests
test:
	go test ./... -v -race

## clean: remove build artifacts
clean:
	rm -f llmview
	rm -rf dist/

## release: cross-compile for major platforms
release:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build $(LDFLAGS) -o dist/llmview-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o dist/llmview-darwin-amd64 .
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build $(LDFLAGS) -o dist/llmview-linux-amd64 .
