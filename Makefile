VERSION ?= $(shell grep -oP 'Version\s*=\s*"\K[^"]+' main.go 2>/dev/null || echo "v0.1.3")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(DATE)

.PHONY: all build test vet run clean

all: vet test build

build:
	go build -ldflags "$(LDFLAGS)" -o peel .

test:
	go test ./...

vet:
	go vet ./...

run: build
	./peel $(ARGS)

clean:
	rm -f peel peel.exe
	rm -rf dist/ peel-*
