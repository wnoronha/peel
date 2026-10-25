package server

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Feedback struct {
	ID        string `json:"id"`
	TS        string `json:"ts"`
	File      string `json:"file"`
	Quote     string `json:"quote"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	Note      string `json:"note"`
}

func feedbackPath(root string) string {
	return filepath.Join(root, ".peel", "feedback.jsonl")
}

func (s *Server) handleFeedbackAdd(w http.ResponseWriter, r *http.Request) {
	var fb Feedback
	if err := json.NewDecoder(r.Body).Decode(&fb); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if strings.Contains(fb.File, "..") || fb.File == "" || filepath.IsAbs(fb.File) {
		http.Error(w, "bad file", http.StatusBadRequest)
		return
	}
	if fb.Note == "" && fb.Quote == "" {
		http.Error(w, "empty", http.StatusBadRequest)
		return
	}

	s.feedbackMu.Lock()
	defer s.feedbackMu.Unlock()

	fb.ID = time.Now().UTC().Format("20060102T150405.000000000")
	fb.TS = time.Now().UTC().Format(time.RFC3339)
	if err := os.MkdirAll(filepath.Join(s.root, ".peel"), 0o755); err != nil {
		http.Error(w, "mkdir", 500)
		return
	}
	f, err := os.OpenFile(feedbackPath(s.root), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		http.Error(w, "open", 500)
		return
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(fb); err != nil {
		http.Error(w, "write", 500)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "id": fb.ID})
}

func (s *Server) handleFeedbackDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	s.feedbackMu.Lock()
	defer s.feedbackMu.Unlock()

	path := feedbackPath(s.root)
	f, err := os.Open(path)
	if err != nil {
		writeJSON(w, map[string]any{"ok": true, "deleted": false})
		return
	}

	var kept []string
	deleted := false
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		var fb Feedback
		if err := json.Unmarshal(line, &fb); err == nil && fb.ID == id {
			deleted = true
			continue
		}
		kept = append(kept, string(line))
	}
	f.Close()

	if !deleted {
		writeJSON(w, map[string]any{"ok": true, "deleted": false})
		return
	}

	tmp := path + ".tmp"
	body := strings.Join(kept, "\n")
	if len(kept) > 0 {
		body += "\n"
	}
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil {
		http.Error(w, "write", 500)
		return
	}
	if err := os.Rename(tmp, path); err != nil {
		http.Error(w, "rename", 500)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "deleted": true})
}

func (s *Server) handleFeedbackList(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	path := feedbackPath(s.root)
	f, err := os.Open(path)
	if err != nil {
		writeJSON(w, map[string]any{"feedback": []Feedback{}})
		return
	}
	defer f.Close()
	var out []Feedback
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		var fb Feedback
		if err := json.Unmarshal(sc.Bytes(), &fb); err != nil {
			continue
		}
		if file != "" && fb.File != file {
			continue
		}
		out = append(out, fb)
	}
	if out == nil {
		out = []Feedback{}
	}
	writeJSON(w, map[string]any{"feedback": out})
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "sse unsupported", 500)
		return
	}
	ch := make(chan string, 16)
	s.mu.Lock()
	s.clients[ch] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		s.mu.Unlock()
	}()
	ctx := r.Context()
	w.Write([]byte(": connected\n\n"))
	fl.Flush()
	tick := time.NewTicker(20 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			w.Write([]byte("data: " + msg + "\n\n"))
			fl.Flush()
		case <-tick.C:
			w.Write([]byte(": hb\n\n"))
			fl.Flush()
		}
	}
}
