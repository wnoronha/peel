VERSION ?= v0.1.2
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(DATE)

# Docs root for run/stop (pid lives at $(DIR)/.peel/peel.pid unless PEEL_PID_FILE is set)
DIR ?= docs
PID_FILE ?= $(DIR)/.peel/peel.pid

.PHONY: all build test vet run stop restart clean

all: test build

build:
	go build -ldflags "$(LDFLAGS)" -o peel .

test:
	go test -v ./...

vet:
	go vet ./...

run: build
	./peel --dir $(DIR)

# Kill only the peel recorded in the pid file — never port-wide lsof (that hits Firefox).
stop:
	@if [ -n "$$PEEL_PID_FILE" ]; then pidfile="$$PEEL_PID_FILE"; else pidfile="$(PID_FILE)"; fi; \
	if [ ! -f "$$pidfile" ]; then echo "no pid file ($$pidfile)"; exit 0; fi; \
	pid=$$(tr -d '[:space:]' < "$$pidfile"); \
	if [ -z "$$pid" ]; then echo "empty pid file"; rm -f "$$pidfile"; exit 0; fi; \
	if kill -0 "$$pid" 2>/dev/null; then echo "stopping peel pid $$pid"; kill "$$pid"; \
	else echo "stale pid $$pid"; fi; \
	rm -f "$$pidfile"

restart: stop run

clean:
	rm -f peel peel.exe
