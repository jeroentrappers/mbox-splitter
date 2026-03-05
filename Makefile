VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build build-cli clean test dev

build:
	CGO_ENABLED=1 go build -tags "production webkit2_41" -ldflags "$(LDFLAGS)" -o mbox-splitter .

build-cli:
	CGO_ENABLED=0 go build -tags cli -ldflags "$(LDFLAGS)" -o mbox-splitter-cli .

test:
	go test -v -tags cli ./...

dev:
	~/go/bin/wails dev -tags webkit2_41

clean:
	rm -f mbox-splitter mbox-splitter-cli
	rm -rf dist/
