VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build dev clean test ui

## build: build UI + compile the binary
build: ui
	CGO_ENABLED=0 go build $(LDFLAGS) -o llmview .

## ui: build the Svelte frontend and copy to embed location
ui:
	cd ui && npm run build
	rm -rf internal/server/ui
	cp -r ui/build internal/server/ui

## dev: run with auto-reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air -c .air.toml || go run .

## test: run all tests
test:
	go test ./... -v -race

## clean: remove build artifacts
clean:
	rm -f llmview
	rm -rf dist/ internal/server/ui
