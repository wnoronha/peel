package server

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// Event types for the append-only feedback log.
const (
	FeedbackNote      = "note"
	FeedbackAck       = "ack"
	FeedbackTombstone = "tombstone"

	maxInkSVGBytes = 200 * 1024
)

type Feedback struct {
	ID        string `json:"id"`
	TS        string `json:"ts"`
	Type      string `json:"type,omitempty"` // note | ack | tombstone; empty on read = note
	Ref       string `json:"ref,omitempty"` // ack/tombstone → note id; note → optional parent note id
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
	C4ID      string `json:"c4Id,omitempty"`      // Pele id: node (reader) or edge (reader-peel)
	C4Label   string `json:"c4Label,omitempty"`   // node name or Rel label e.g. "browses and annotates"
	C4From    string `json:"c4From,omitempty"`    // Rel source id when c4Id is an edge
	C4To      string `json:"c4To,omitempty"`      // Rel target id when c4Id is an edge
	Ink       string `json:"ink,omitempty"`       // relative path e.g. .peel/ink/<id>.svg
	InkSvg    string `json:"inkSvg,omitempty"`    // create-only; never persisted in jsonl
}

var allowedActions = map[string]bool{
	"fix": true, "explain": true, "example": true, "simplify": true, "question": true,
}

func feedbackPath(root string) string {
	return filepath.Join(root, ".peel", "feedback.jsonl")
}

func inkRelPath(id string) string {
	return filepath.ToSlash(filepath.Join(".peel", "ink", id+".svg"))
}

func inkAbsPath(root, id string) string {
	return filepath.Join(root, ".peel", "ink", id+".svg")
}

func validNoteID(id string) bool {
	if id == "" || len(id) > 64 {
		return false
	}
	for _, r := range id {
		if unicode.IsDigit(r) || r == 'T' || r == '.' {
			continue
		}
		return false
	}
	return true
}

// sanitizeC4ID keeps Pele/C4 element or edge ids (letters, digits, _ -).
func sanitizeC4ID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" || len(id) > 96 {
		return ""
	}
	for _, r := range id {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			continue
		}
		return ""
	}
	return id
}

func validateInkSVG(svg string) (string, bool) {
	s := strings.TrimSpace(svg)
	if s == "" {
		return "", true
	}
	if len(s) > maxInkSVGBytes {
		return "", false
	}
	lower := strings.ToLower(s)
	if !strings.HasPrefix(lower, "<svg") || !strings.Contains(lower, "</svg>") {
		return "", false
	}
	if strings.Contains(lower, "<script") || strings.Contains(lower, "javascript:") {
		return "", false
	}
	return s, true
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

// foldOpenNotes returns root note events (no parent ref) with no later ack or tombstone.
// Reply notes (ref set) nest under their parent and are not separate open-queue items.
// Ack/tombstone of a root also closes that root's reply notes.
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
	// Cascade: closing a root closes replies that ref it.
	for _, ev := range events {
		if feedbackType(ev) != FeedbackNote || ev.Ref == "" {
			continue
		}
		if closed[ev.Ref] {
			closed[ev.ID] = true
		}
	}
	var out []Feedback
	for _, ev := range events {
		if feedbackType(ev) != FeedbackNote {
			continue
		}
		if ev.Ref != "" {
			continue // replies are not open-queue roots
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

	inkSvg, inkOK := validateInkSVG(fb.InkSvg)
	if !inkOK {
		http.Error(w, "bad ink", http.StatusBadRequest)
		return
	}
	fb.InkSvg = ""
	fb.Ink = ""

	t := feedbackType(fb)
	switch t {
	case FeedbackNote:
		if strings.Contains(fb.File, "..") || fb.File == "" || filepath.IsAbs(fb.File) {
			http.Error(w, "bad file", http.StatusBadRequest)
			return
		}
		if inkSvg != "" && strings.TrimSpace(fb.Quote) == "" {
			http.Error(w, "empty", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(fb.Note) == "" && strings.TrimSpace(fb.Quote) == "" && inkSvg == "" {
			http.Error(w, "empty", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(fb.Note) == "" && inkSvg != "" {
			fb.Note = "Ink mark"
		}
		if fb.Action != "" && !allowedActions[fb.Action] {
			http.Error(w, "bad action", http.StatusBadRequest)
			return
		}
		parentRef := strings.TrimSpace(fb.Ref)
		fb.Type = FeedbackNote
		fb.Ref = ""
		if parentRef != "" {
			if !validNoteID(parentRef) {
				http.Error(w, "bad ref", http.StatusBadRequest)
				return
			}
			fb.Ref = parentRef // validated against log under lock below
		}
		fb.C4ID = sanitizeC4ID(fb.C4ID)
		fb.C4Label = strings.TrimSpace(fb.C4Label)
		fb.C4From = sanitizeC4ID(fb.C4From)
		fb.C4To = sanitizeC4ID(fb.C4To)
		if len(fb.C4Label) > 200 {
			fb.C4Label = fb.C4Label[:200]
		}
		if fb.C4ID == "" {
			fb.C4Label = ""
			fb.C4From = ""
			fb.C4To = ""
		}
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
		fb.C4ID = ""
		fb.C4Label = ""
		fb.C4From = ""
		fb.C4To = ""
		inkSvg = ""
		if t == FeedbackTombstone {
			fb.Note = ""
		}
	default:
		http.Error(w, "bad type", http.StatusBadRequest)
		return
	}

	s.feedbackMu.Lock()
	defer s.feedbackMu.Unlock()

	if t == FeedbackNote && fb.Ref != "" {
		events, err := s.readFeedbackEvents()
		if err != nil {
			http.Error(w, "read", 500)
			return
		}
		if !noteExists(events, fb.Ref) {
			http.Error(w, "bad ref", http.StatusBadRequest)
			return
		}
	}

	fb.ID = time.Now().UTC().Format("20060102T150405.000000000")
	fb.TS = time.Now().UTC().Format(time.RFC3339)
	if inkSvg != "" && t == FeedbackNote {
		if err := os.MkdirAll(filepath.Join(s.root, ".peel", "ink"), 0o755); err != nil {
			http.Error(w, "write", 500)
			return
		}
		if err := os.WriteFile(inkAbsPath(s.root, fb.ID), []byte(inkSvg), 0o644); err != nil {
			http.Error(w, "write", 500)
			return
		}
		fb.Ink = inkRelPath(fb.ID)
	}
	if err := s.appendFeedback(fb); err != nil {
		http.Error(w, "write", 500)
		return
	}
	writeJSON(w, map[string]any{"ok": true, "id": fb.ID, "type": fb.Type, "ink": fb.Ink})
}

func (s *Server) handleFeedbackInk(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if !validNoteID(id) {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	s.feedbackMu.Lock()
	events, err := s.readFeedbackEvents()
	s.feedbackMu.Unlock()
	if err != nil {
		http.Error(w, "read", 500)
		return
	}

	var ink string
	for _, ev := range events {
		if feedbackType(ev) == FeedbackNote && ev.ID == id {
			ink = ev.Ink
			break
		}
	}
	if ink == "" {
		http.NotFound(w, r)
		return
	}
	// Only serve paths we wrote under .peel/ink/<id>.svg
	want := inkRelPath(id)
	if filepath.ToSlash(ink) != want {
		http.NotFound(w, r)
		return
	}
	data, err := os.ReadFile(inkAbsPath(s.root, id))
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "read", 500)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
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
