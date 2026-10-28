package server

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path"
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
	C4ID      string `json:"c4Id,omitempty"`      // Pele id: node (reader) or edge (reader-peel)
	C4Label   string `json:"c4Label,omitempty"`   // node name or Rel label
	C4From    string `json:"c4From,omitempty"`    // Rel source id when c4Id is an edge
	C4To      string `json:"c4To,omitempty"`      // Rel target id when c4Id is an edge
}

var allowedActions = map[string]bool{
	"fix": true, "explain": true, "example": true, "simplify": true, "question": true,
}

func feedbackPath(root string) string {
	return filepath.Join(root, ".peel", "feedback.jsonl")
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

func feedbackType(fb Feedback) string {
	t := strings.TrimSpace(fb.Type)
	if t == "" {
		return FeedbackNote
	}
	return t
}

func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	count := 0
	for i := range s {
		if count == n {
			return s[:i]
		}
		count++
	}
	return s
}

func pathCleanSlash(p string) string {
	return path.Clean(filepath.ToSlash(p))
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

func newFeedbackID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UTC().Format("20060102T150405.000000000")
	}
	return time.Now().UTC().Format("20060102T150405.000000000") + "-" + hex.EncodeToString(b[:])
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
	br := bufio.NewReader(f)
	for {
		line, err := br.ReadString('\n')
		if line != "" {
			// Skip overlong lines instead of failing the whole listing.
			if len(line) <= 2<<20 {
				var fb Feedback
				if jerr := json.Unmarshal([]byte(line), &fb); jerr == nil {
					out = append(out, fb)
				}
			}
		}
		if err != nil {
			break
		}
	}
	return out, nil
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
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var fb Feedback
	if err := json.NewDecoder(r.Body).Decode(&fb); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	t := feedbackType(fb)
	switch t {
	case FeedbackNote:
		if _, ok := s.safeJoin(fb.File); !ok {
			http.Error(w, "bad file", http.StatusBadRequest)
			return
		}
		fb.File = pathCleanSlash(fb.File)
		if len(fb.Note) > 20000 || len(fb.Quote) > 20000 {
			http.Error(w, "note too large", http.StatusBadRequest)
			return
		}
		if len(fb.File) > 512 || len(fb.HeadingID) > 512 || len(fb.LevelKind) > 64 {
			http.Error(w, "field too large", http.StatusBadRequest)
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
		fb.C4ID = sanitizeC4ID(fb.C4ID)
		fb.C4Label = truncateRunes(strings.TrimSpace(fb.C4Label), 200)
		fb.C4From = sanitizeC4ID(fb.C4From)
		fb.C4To = sanitizeC4ID(fb.C4To)
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
		if len(fb.Ref) > 128 || len(fb.Note) > 20000 {
			http.Error(w, "field too large", http.StatusBadRequest)
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
		if t == FeedbackTombstone {
			fb.Note = ""
		}
	default:
		http.Error(w, "bad type", http.StatusBadRequest)
		return
	}

	s.feedbackMu.Lock()
	defer s.feedbackMu.Unlock()

	fb.ID = newFeedbackID()
	fb.TS = time.Now().UTC().Format(time.RFC3339)
	if err := s.appendFeedback(fb); err != nil {
		http.Error(w, "write", 500)
		return
	}
	s.emitFeedback(fb)
	if t == FeedbackAck {
		s.clearAgentIfRef(fb.Ref)
	}
	writeJSON(w, map[string]any{"ok": true, "id": fb.ID, "type": fb.Type})
}

// emitFeedback wakes SSE clients (agents + browser Feedback panel).
func (s *Server) emitFeedback(fb Feedback) {
	msg, err := json.Marshal(map[string]any{"type": "feedback", "event": fb})
	if err != nil {
		return
	}
	s.broadcast(string(msg))
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
			fb.ID = newFeedbackID()
			fb.TS = time.Now().UTC().Format(time.RFC3339)
			if err := s.appendFeedback(fb); err != nil {
				http.Error(w, "write", 500)
				return
			}
			s.emitFeedback(fb)
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
