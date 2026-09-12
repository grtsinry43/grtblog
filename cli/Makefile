GO ?= go
# 版本默认跟随项目最新 tag（如 v2.2.0），发布时可用 VERSION=... 覆盖。
VERSION ?= $(shell git describe --tags --abbrev=0 --match 'v[0-9]*' 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build install test vet fmt tidy clean

build:
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o grtblog ./cmd/grtblog

install:
	$(GO) install -trimpath -ldflags "$(LDFLAGS)" ./cmd/grtblog

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w ./cmd ./internal

tidy:
	$(GO) mod tidy

clean:
	rm -f grtblog
