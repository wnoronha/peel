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

// Event types for the append-only feedback log.
const (
	FeedbackNote      = "note"
	FeedbackAck       = "ack"
	FeedbackTombstone = "tombstone"
)

type Feedback struct {
	ID        string `json:"id"`
	TS        string `json:"ts"`
	Type      string `json:"type,omitempty"` // note | ack | tombstone; empty on read = note
	Ref       string `json:"ref,omitempty"`  // note id for ack/tombstone
	File      string `json:"file,omitempty"`
	Quote     string `json:"quote,omitempty"`
	StartLine int    `json:"startLine,omitempty"`
	EndLine   int    `json:"endLine,omitempty"`
	Note      string `json:"note,omitempty"`
	Action    string `json:"action,omitempty"` // fix|explain|example|simplify|question
	HeadingID string `json:"headingId,omitempty"`
	Slide     int    `json:"slide,omitempty"`     // 1-based deck slide
	Level     int    `json:"level,omitempty"`     // 1-based C4 level
	LevelKind string `json:"levelKind,omitempty"` // Context|Container|Component|Code|…
}

var allowedActions = map[string]bool{
	"fix": true, "explain": true, "example": true, "simplify": true, "question": true,
}

func feedbackPath(root string) string {
	return filepath.Join(root, ".peel", "feedback.jsonl")
}

func feedbackType(fb Feedback) string {
	t := strings.TrimSpace(fb.Type)
	if t == "" {
		return FeedbackNote
	}
	return t
}

func (s *Server) appendFeedback(fb Feedback) error {
	if err := os.MkdirAll(filepath.Join(s.root, ".peel"), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(feedbackPath(s.root), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(fb); err != nil {
		return err
	}
	return f.Sync()
}

func (s *Server) readFeedbackEvents() ([]Feedback, error) {
	path := feedbackPath(s.root)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
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
		out = append(out, fb)
	}
	return out, sc.Err()
}

// foldOpenNotes returns note events that have no later ack or tombstone for their id.
func foldOpenNotes(events []Feedback) []Feedback {
	closed := map[string]bool{}
	for _, ev := range events {
		switch feedbackType(ev) {
		case FeedbackAck, FeedbackTombstone:
			if ev.Ref != "" {
				closed[ev.Ref] = true
			}
		}
	}
	var out []Feedback
	for _, ev := range events {
		if feedbackType(ev) != FeedbackNote {
			continue
		}
		if closed[ev.ID] {
			continue
		}
		if ev.Type == "" {
			ev.Type = FeedbackNote
		}
		out = append(out, ev)
	}
	return out
}

func noteExists(events []Feedback, id string) bool {
	for _, ev := range events {
		if feedbackType(ev) == FeedbackNote && ev.ID == id {
			return true
		}
	}
	return false
}

func (s *Server) handleFeedbackAdd(w http.ResponseWriter, r *http.Request) {
	var fb Feedback
	if err := json.NewDecoder(r.Body).Decode(&fb); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	t := feedbackType(fb)
	switch t {
	case FeedbackNote:
		if strings.Contains(fb.File, "..") || fb.File == "" || filepath.IsAbs(fb.File) {
			http.Error(w, "bad file", http.StatusBadRequest)
			return
		}
		if fb.Note == "" && fb.Quote == "" {
			http.Error(w, "empty", http.StatusBadRequest)
			return
		}
		if fb.Action != "" && !allowedActions[fb.Action] {
			http.Error(w, "bad action", http.StatusBadRequest)
			return
		}
		fb.Type = FeedbackNote
		fb.Ref = ""
	case FeedbackAck, FeedbackTombstone:
		if strings.TrimSpace(fb.Ref) == "" {
			http.Error(w, "missing ref", http.StatusBadRequest)
			return
		}
		fb.Type = t
		fb.File = ""
		fb.Quote = ""
		fb.StartLine = 0
		fb.EndLine = 0
		fb.Action = ""
		fb.HeadingID = ""
		fb.Slide = 0
		fb.Level = 0
		fb.LevelKind = ""
		if t == FeedbackTombstone {
			fb.Note = ""
		}
	default:
		http.Error(w, "bad type", http.StatusBadRequest)
		return
	}

	s.feedbackMu.Lock()
	defer s.feedbackMu.Unlock()

	fb.ID = time.Now().UTC().Format("20060102T150405.000000000")
	fb.TS = time.Now().UTC().Format(time.RFC3339)
	if err := s.appendFeedback(fb); err != nil {
		http.Error(w, "write", 500)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "id": fb.ID, "type": fb.Type})
}

func (s *Server) handleFeedbackDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	s.feedbackMu.Lock()
	defer s.feedbackMu.Unlock()

	events, err := s.readFeedbackEvents()
	if err != nil {
		http.Error(w, "read", 500)
		return
	}
	if !noteExists(events, id) {
		writeJSON(w, map[string]any{"ok": true, "deleted": false})
		return
	}
	for _, n := range foldOpenNotes(events) {
		if n.ID == id {
			fb := Feedback{Type: FeedbackTombstone, Ref: id}
			fb.ID = time.Now().UTC().Format("20060102T150405.000000000")
			fb.TS = time.Now().UTC().Format(time.RFC3339)
			if err := s.appendFeedback(fb); err != nil {
				http.Error(w, "write", 500)
				return
			}
			writeJSON(w, map[string]any{"ok": true, "deleted": true})
			return
		}
	}
	// Note exists but already closed.
	writeJSON(w, map[string]any{"ok": true, "deleted": false})
}

func (s *Server) handleFeedbackList(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	all := r.URL.Query().Get("all") == "1" || r.URL.Query().Get("all") == "true"

	events, err := s.readFeedbackEvents()
	if err != nil {
		http.Error(w, "read", 500)
		return
	}

	var out []Feedback
	if all {
		// Raw event log (agents). Optional file= filters to note events for that path.
		if file == "" {
			out = events
		} else {
			for _, ev := range events {
				if feedbackType(ev) == FeedbackNote && ev.File == file {
					out = append(out, ev)
				}
			}
		}
	} else {
		out = foldOpenNotes(events)
		if file != "" {
			filtered := out[:0]
			for _, ev := range out {
				if ev.File == file {
					filtered = append(filtered, ev)
				}
			}
			out = filtered
		}
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
