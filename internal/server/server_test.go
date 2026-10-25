package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestRootAssetServing(t *testing.T) {
	tmp := t.TempDir()
	assetsDir := filepath.Join(tmp, "assets")
	_ = os.MkdirAll(assetsDir, 0o755)

	imagePath := filepath.Join(assetsDir, "diagram.png")
	_ = os.WriteFile(imagePath, []byte("fake png content"), 0o644)

	secretPath := filepath.Join(tmp, "main.go")
	_ = os.WriteFile(secretPath, []byte("package main"), 0o644)

	s := New(tmp)

	// 1. Serving allowed media file
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/assets/diagram.png", nil)
	served := s.serveRootAsset(rec, req)
	if !served {
		t.Fatalf("expected serveRootAsset to return true for allowed image")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// 2. Rejecting non-media file
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/main.go", nil)
	served = s.serveRootAsset(rec, req)
	if served {
		t.Fatalf("expected serveRootAsset to return false for .go file")
	}

	// 3. Rejecting path traversal
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/../secret.png", nil)
	served = s.serveRootAsset(rec, req)
	if served {
		t.Fatalf("expected serveRootAsset to return false for traversal")
	}
}

func TestWebFSAndFallback(t *testing.T) {
	mockWeb := fstest.MapFS{
		"index.html": {Data: []byte("<html>Peel</html>")},
		"style.css":  {Data: []byte("body { color: white; }")},
	}
	SetWebFS(mockWeb)

	tmp := t.TempDir()
	imgDir := filepath.Join(tmp, "img")
	_ = os.MkdirAll(imgDir, 0o755)
	_ = os.WriteFile(filepath.Join(imgDir, "test.png"), []byte("png"), 0o644)

	s := New(tmp)

	// Web asset
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/style.css", nil)
	s.serveRootAsset(rec, req) // won't serve style.css as root asset
	// Test the full fallback routing logic
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[1:]
		if path == "" || path == "index.html" {
			if staticHandler != nil {
				staticHandler.ServeHTTP(w, r)
				return
			}
		}
		if webFS != nil {
			if f, err := webFS.Open(path); err == nil {
				_ = f.Close()
				staticHandler.ServeHTTP(w, r)
				return
			}
		}
		if s.serveRootAsset(w, r) {
			return
		}
		http.NotFound(w, r)
	})

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/style.css", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for style.css from webFS, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/img/test.png", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for img/test.png from s.root, got %d", rec.Code)
	}
}
