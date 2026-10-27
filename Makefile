VERSION ?= v0.1.1
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(DATE)

.PHONY: all build test vet run clean

all: test build

build:
	go build -ldflags "$(LDFLAGS)" -o peel .

test:
	go test -v ./...

vet:
	go vet ./...

run: build
	./peel

clean:
	rm -f peel peel.exe
