package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
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
	agent      agentStored
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

// safeJoin resolves a root-relative rel path and rejects traversal and symlinks.
func (s *Server) safeJoin(rel string) (string, bool) {
	if rel == "" || filepath.IsAbs(rel) || filepath.IsAbs(filepath.FromSlash(rel)) {
		return "", false
	}
	cleaned := path.Clean(filepath.ToSlash(rel))
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") {
		return "", false
	}
	for _, part := range strings.Split(cleaned, "/") {
		if part == ".." || part == "." || part == "" {
			return "", false
		}
	}
	rootAbs, err := filepath.Abs(s.root)
	if err != nil {
		return "", false
	}
	abs := filepath.Join(rootAbs, filepath.FromSlash(cleaned))
	relCheck, err := filepath.Rel(rootAbs, abs)
	if err != nil || relCheck == ".." || strings.HasPrefix(relCheck, ".."+string(filepath.Separator)) {
		return "", false
	}
	if fi, err := os.Lstat(abs); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			return "", false
		}
		if resolved, err := filepath.EvalSymlinks(abs); err == nil {
			resAbs, err := filepath.Abs(resolved)
			if err != nil {
				return "", false
			}
			if rc, err := filepath.Rel(rootAbs, resAbs); err != nil || rc == ".." || strings.HasPrefix(rc, ".."+string(filepath.Separator)) {
				return "", false
			}
			return resAbs, true
		}
		return "", false
	}
	// File does not exist: verify parent dir resolves inside root.
	parent := filepath.Dir(abs)
	if presolved, err := filepath.EvalSymlinks(parent); err == nil {
		if rp, err := filepath.Rel(rootAbs, presolved); err != nil || rp == ".." || strings.HasPrefix(rp, ".."+string(filepath.Separator)) {
			return "", false
		}
	}
	return abs, true
}

func (s *Server) serveRootAsset(w http.ResponseWriter, r *http.Request) bool {
	rel := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") || part == "node_modules" {
			return false
		}
	}
	// safeJoin rejects traversal, absolute paths, and symlinks.
	cleaned := path.Clean(filepath.ToSlash(rel))
	ext := strings.ToLower(filepath.Ext(cleaned))
	if !mediaExtensions[ext] {
		return false
	}
	abs, ok := s.safeJoin(cleaned)
	if !ok {
		return false
	}
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
	mux.HandleFunc("POST /api/feedback", s.handleFeedbackAdd)
	mux.HandleFunc("DELETE /api/feedback", s.handleFeedbackDelete)
	mux.HandleFunc("GET /api/agent", s.handleAgentGet)
	mux.HandleFunc("POST /api/agent", s.handleAgentPost)
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
	srv := s.newHTTPServer(addr, mux)
	return srv.ListenAndServe()
}

func (s *Server) newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (s *Server) broadcast(msg string) {
	s.mu.Lock()
	clients := make([]chan string, 0, len(s.clients))
	for ch := range s.clients {
		clients = append(clients, ch)
	}
	s.mu.Unlock()
	for _, ch := range clients {
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
				if now.Sub(last) >= 300*time.Millisecond {
					s.debounce[rel] = now
				}
				s.mu.Unlock()
				if now.Sub(last) < 300*time.Millisecond {
					continue
				}
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
