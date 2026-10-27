package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Server struct {
	root       string
	watcher    *fsnotify.Watcher
	mu         sync.Mutex
	feedbackMu sync.Mutex
	clients    map[chan string]struct{}
	debounce   map[string]time.Time
}

func New(root string) *Server {
	return &Server{
		root:     root,
		clients:  make(map[chan string]struct{}),
		debounce: make(map[string]time.Time),
	}
}

var (
	webFS         fs.FS
	staticHandler http.Handler
)

var mediaExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".svg":  true,
	".webp": true,
	".ico":  true,
	".bmp":  true,
	".pdf":  true,
	".mp4":  true,
	".webm": true,
}

// SetWebFS wires the embedded web/ filesystem.
func SetWebFS(f fs.FS) {
	webFS = f
	staticHandler = http.FileServer(http.FS(f))
}

// SetStatic keeps compatibility for custom handlers if needed.
func SetStatic(h http.Handler) {
	staticHandler = h
}

func (s *Server) serveRootAsset(w http.ResponseWriter, r *http.Request) bool {
	clean := filepath.Clean(filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/")))
	if strings.HasPrefix(clean, ".") || strings.Contains(clean, "..") {
		return false
	}
	parts := strings.Split(filepath.ToSlash(clean), "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") || part == "node_modules" {
			return false
		}
	}
	ext := strings.ToLower(filepath.Ext(clean))
	if !mediaExtensions[ext] {
		return false
	}
	abs := filepath.Join(s.root, clean)
	st, err := os.Stat(abs)
	if err != nil || st.IsDir() {
		return false
	}
	http.ServeFile(w, r, abs)
	return true
}

func (s *Server) Run(addr string) error {
	if err := s.startWatcher(); err != nil {
		return fmt.Errorf("watcher: %w", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/files", s.handleFiles)
	mux.HandleFunc("GET /api/doc", s.handleDoc)
	mux.HandleFunc("GET /api/feedback", s.handleFeedbackList)
	mux.HandleFunc("GET /api/feedback/ink", s.handleFeedbackInk)
	mux.HandleFunc("POST /api/feedback", s.handleFeedbackAdd)
	mux.HandleFunc("DELETE /api/feedback", s.handleFeedbackDelete)
	mux.HandleFunc("GET /events", s.handleEvents)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" || path == "index.html" {
			if staticHandler != nil {
				w.Header().Set("Cache-Control", "no-store")
				staticHandler.ServeHTTP(w, r)
				return
			}
		}
		if webFS != nil {
			if f, err := webFS.Open(path); err == nil {
				_ = f.Close()
				w.Header().Set("Cache-Control", "no-store")
				staticHandler.ServeHTTP(w, r)
				return
			}
		}
		if s.serveRootAsset(w, r) {
			return
		}
		http.NotFound(w, r)
	})

	srv := &http.Server{Addr: addr, Handler: mux}
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case sig := <-sigCh:
		fmt.Printf("peel: shutting down (%s)\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (s *Server) broadcast(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (s *Server) startWatcher() error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	s.watcher = w
	root := s.root
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		base := filepath.Base(p)
		if strings.HasPrefix(base, ".") || base == "node_modules" {
			return filepath.SkipDir
		}
		_ = w.Add(p)
		return nil
	})

	go func() {
		for {
			select {
			case ev, ok := <-w.Events:
				if !ok {
					return
				}
				// dynamically watch newly created directories
				if ev.Op&fsnotify.Create != 0 {
					if st, err := os.Stat(ev.Name); err == nil && st.IsDir() {
						base := filepath.Base(ev.Name)
						if !strings.HasPrefix(base, ".") && base != "node_modules" {
							_ = w.Add(ev.Name)
						}
					}
				}
				// watch .md and .c4 changes
				ext := strings.ToLower(filepath.Ext(ev.Name))
				if ext != ".md" && ext != ".c4" {
					continue
				}
				rel, _ := filepath.Rel(root, ev.Name)
				rel = filepath.ToSlash(rel)
				now := time.Now()
				s.mu.Lock()
				last := s.debounce[rel]
				s.mu.Unlock()
				if now.Sub(last) < 300*time.Millisecond {
					continue
				}
				s.mu.Lock()
				s.debounce[rel] = now
				s.mu.Unlock()
				msg, _ := json.Marshal(map[string]string{"type": "reload", "path": rel})
				s.broadcast(string(msg))
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				fmt.Fprintln(os.Stderr, "watch:", err)
			}
		}
	}()
	return nil
}
