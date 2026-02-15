VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test test-integration test-all lint install clean

build:
	go build $(LDFLAGS) -o bin/wpgo ./cmd/wpgo

test:
	go test -race -count=1 ./internal/...

test-integration:
	@if [ -n "$(shell go list -tags=integration ./tests/integration/... 2>/dev/null)" ]; then \
		go test -race -tags=integration ./tests/integration/...; \
	else \
		echo "No integration test packages found under ./tests/integration/..."; \
	fi

test-all:
	go test -race -tags=integration ./...

lint:
	go vet ./...
	golangci-lint run

install:
	go install $(LDFLAGS) ./cmd/wpgo

clean:
	rm -rf bin/ dist/
