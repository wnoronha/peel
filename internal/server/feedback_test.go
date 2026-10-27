package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestFeedbackOperations(t *testing.T) {
	tmp := t.TempDir()
	s := New(tmp)

	// 1. Initially empty
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/feedback", nil)
	s.handleFeedbackList(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var listResp struct {
		Feedback []Feedback `json:"feedback"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &listResp)
	if len(listResp.Feedback) != 0 {
		t.Fatalf("expected empty feedback, got %d", len(listResp.Feedback))
	}

	// 2. Add feedback note
	fbInput := Feedback{
		File:  "doc-a.md",
		Quote: "quoted text",
		Note:  "review this",
	}
	body, _ := json.Marshal(fbInput)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/feedback", bytes.NewReader(body))
	s.handleFeedbackAdd(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on add, got %d", rec.Code)
	}
	var addResp struct {
		OK   bool   `json:"ok"`
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &addResp)
	if !addResp.OK || addResp.ID == "" || addResp.Type != FeedbackNote {
		t.Fatalf("unexpected add response: %+v", addResp)
	}

	// 3. List open notes for doc-a.md
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/feedback?file=doc-a.md", nil)
	s.handleFeedbackList(rec, req)
	_ = json.Unmarshal(rec.Body.Bytes(), &listResp)
	if len(listResp.Feedback) != 1 || listResp.Feedback[0].ID != addResp.ID {
		t.Fatalf("expected 1 item with ID %s, got %+v", addResp.ID, listResp.Feedback)
	}
	if listResp.Feedback[0].Type != FeedbackNote {
		t.Fatalf("expected type note, got %q", listResp.Feedback[0].Type)
	}

	// 4. Delete nonexistent ID
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/api/feedback?id=nonexistent", nil)
	s.handleFeedbackDelete(rec, req)
	var delResp struct {
		OK      bool `json:"ok"`
		Deleted bool `json:"deleted"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &delResp)
	if !delResp.OK || delResp.Deleted {
		t.Fatalf("expected ok=true, deleted=false, got %+v", delResp)
	}

	// 5. Delete actual ID → tombstone append (not rewrite)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/api/feedback?id="+addResp.ID, nil)
	s.handleFeedbackDelete(rec, req)
	_ = json.Unmarshal(rec.Body.Bytes(), &delResp)
	if !delResp.OK || !delResp.Deleted {
		t.Fatalf("expected ok=true, deleted=true, got %+v", delResp)
	}

	// 6. Open list empty; raw log still has note + tombstone
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/feedback", nil)
	s.handleFeedbackList(rec, req)
	_ = json.Unmarshal(rec.Body.Bytes(), &listResp)
	if len(listResp.Feedback) != 0 {
		t.Fatalf("expected 0 open feedback after tombstone, got %d", len(listResp.Feedback))
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/feedback?all=1", nil)
	s.handleFeedbackList(rec, req)
	_ = json.Unmarshal(rec.Body.Bytes(), &listResp)
	if len(listResp.Feedback) != 2 {
		t.Fatalf("expected 2 raw events, got %d (%+v)", len(listResp.Feedback), listResp.Feedback)
	}
	if feedbackType(listResp.Feedback[0]) != FeedbackNote || feedbackType(listResp.Feedback[1]) != FeedbackTombstone {
		t.Fatalf("expected note then tombstone, got %+v", listResp.Feedback)
	}
	if listResp.Feedback[1].Ref != addResp.ID {
		t.Fatalf("tombstone ref want %s got %s", addResp.ID, listResp.Feedback[1].Ref)
	}

	// File must still contain both lines (append-only)
	f, err := os.Open(feedbackPath(tmp))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	lines := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines++
	}
	if lines != 2 {
		t.Fatalf("expected 2 jsonl lines on disk, got %d", lines)
	}
}

func TestFeedbackAckFoldsOpen(t *testing.T) {
	s := New(t.TempDir())

	body, _ := json.Marshal(Feedback{File: "a.md", Quote: "q", Note: "n"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/feedback", bytes.NewReader(body))
	s.handleFeedbackAdd(rec, req)
	var addResp struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &addResp)

	ackBody, _ := json.Marshal(Feedback{Type: FeedbackAck, Ref: addResp.ID, Note: "fixed"})
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/feedback", bytes.NewReader(ackBody))
	s.handleFeedbackAdd(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("ack expected 200, got %d body %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/feedback", nil)
	s.handleFeedbackList(rec, req)
	var listResp struct {
		Feedback []Feedback `json:"feedback"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &listResp)
	if len(listResp.Feedback) != 0 {
		t.Fatalf("acked note should not be open, got %+v", listResp.Feedback)
	}
}

func TestFeedbackValidation(t *testing.T) {
	s := New(t.TempDir())

	badBody := []byte(`{"file": "a.md", "note": "", "quote": ""}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/feedback", bytes.NewReader(badBody))
	s.handleFeedbackAdd(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty note and quote, got %d", rec.Code)
	}

	badPath := []byte(`{"file": "../secret.md", "note": "bad", "quote": "q"}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/feedback", bytes.NewReader(badPath))
	s.handleFeedbackAdd(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for traversal file, got %d", rec.Code)
	}

	badAck := []byte(`{"type":"ack","note":"x"}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/feedback", bytes.NewReader(badAck))
	s.handleFeedbackAdd(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for ack without ref, got %d", rec.Code)
	}
}

func TestFoldOpenNotes(t *testing.T) {
	events := []Feedback{
		{ID: "1", Type: FeedbackNote, File: "a.md", Note: "n"},
		{ID: "2", Type: FeedbackNote, File: "a.md", Note: "n2"},
		{ID: "3", Type: FeedbackAck, Ref: "1"},
		{ID: "4"}, // legacy note (empty type)
	}
	open := foldOpenNotes(events)
	if len(open) != 2 {
		t.Fatalf("want 2 open, got %d %+v", len(open), open)
	}
	if open[0].ID != "2" || open[1].ID != "4" {
		t.Fatalf("unexpected open set: %+v", open)
	}
	if open[1].Type != FeedbackNote {
		t.Fatalf("legacy should normalize type to note, got %q", open[1].Type)
	}
}
