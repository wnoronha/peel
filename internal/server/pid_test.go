package server

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDefaultPIDPathEnv(t *testing.T) {
	t.Setenv("PEEL_PID_FILE", "/tmp/custom-peel.pid")
	if got := DefaultPIDPath("/docs"); got != "/tmp/custom-peel.pid" {
		t.Fatalf("env override: got %s", got)
	}
	t.Setenv("PEEL_PID_FILE", "")
	got := DefaultPIDPath("/docs")
	want := filepath.Join("/docs", ".peel", "peel.pid")
	if got != want {
		t.Fatalf("default path: got %s want %s", got, want)
	}
}

func TestWriteRemovePID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".peel", "peel.pid")
	if err := WritePID(path); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || n != os.Getpid() {
		t.Fatalf("pid file content %q", b)
	}
	RemovePID(path)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected pid removed, err=%v", err)
	}
}

func TestWritePIDRejectsLive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "peel.pid")
	if err := WritePID(path); err != nil {
		t.Fatal(err)
	}
	if err := WritePID(path); err == nil {
		t.Fatal("expected error when pid still live")
	}
	RemovePID(path)
}
