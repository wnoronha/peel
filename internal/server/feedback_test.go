package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

	// 2. Add feedback
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
		OK bool   `json:"ok"`
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &addResp)
	if !addResp.OK || addResp.ID == "" {
		t.Fatalf("unexpected add response: %+v", addResp)
	}

	// 3. List feedback for doc-a.md
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/feedback?file=doc-a.md", nil)
	s.handleFeedbackList(rec, req)
	_ = json.Unmarshal(rec.Body.Bytes(), &listResp)
	if len(listResp.Feedback) != 1 || listResp.Feedback[0].ID != addResp.ID {
		t.Fatalf("expected 1 item with ID %s, got %+v", addResp.ID, listResp.Feedback)
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

	// 5. Delete actual ID
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/api/feedback?id="+addResp.ID, nil)
	s.handleFeedbackDelete(rec, req)
	_ = json.Unmarshal(rec.Body.Bytes(), &delResp)
	if !delResp.OK || !delResp.Deleted {
		t.Fatalf("expected ok=true, deleted=true, got %+v", delResp)
	}

	// 6. Verify empty again
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/feedback", nil)
	s.handleFeedbackList(rec, req)
	_ = json.Unmarshal(rec.Body.Bytes(), &listResp)
	if len(listResp.Feedback) != 0 {
		t.Fatalf("expected 0 feedback after deletion, got %d", len(listResp.Feedback))
	}
}

func TestFeedbackValidation(t *testing.T) {
	s := New(t.TempDir())

	// Empty note and quote
	badBody := []byte(`{"file": "a.md", "note": "", "quote": ""}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/feedback", bytes.NewReader(badBody))
	s.handleFeedbackAdd(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty note and quote, got %d", rec.Code)
	}

	// Path traversal in file
	badPath := []byte(`{"file": "../secret.md", "note": "bad", "quote": "q"}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/feedback", bytes.NewReader(badPath))
	s.handleFeedbackAdd(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for traversal file, got %d", rec.Code)
	}
}
