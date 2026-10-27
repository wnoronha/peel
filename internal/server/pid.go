package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// DefaultPIDPath is <root>/.peel/peel.pid. Override with PEEL_PID_FILE.
func DefaultPIDPath(root string) string {
	if v := strings.TrimSpace(os.Getenv("PEEL_PID_FILE")); v != "" {
		return v
	}
	return filepath.Join(root, ".peel", "peel.pid")
}

// WritePID records os.Getpid() at path. Fails if a live process already owns the file.
func WritePID(path string) error {
	if path == "" {
		return fmt.Errorf("empty pid path")
	}
	if b, err := os.ReadFile(path); err == nil {
		if old, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && old > 0 && pidAlive(old) {
			return fmt.Errorf("already running (pid %d); stop it first (kill %d or make stop)", old, old)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())+"\n"), 0o644)
}

// RemovePID deletes the pid file when it still refers to this process.
func RemovePID(path string) {
	if path == "" {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || n != os.Getpid() {
		return
	}
	_ = os.Remove(path)
}

func pidAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal 0 checks existence without killing (Unix).
	return p.Signal(syscall.Signal(0)) == nil
}
