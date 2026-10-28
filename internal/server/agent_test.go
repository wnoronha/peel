package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAgentPresenceWorkingIdle(t *testing.T) {
	s := New(t.TempDir())

	req := httptest.NewRequest(http.MethodPost, "/api/agent", bytes.NewBufferString(`{"state":"working","note":"editing","ref":"n1"}`))
	rr := httptest.NewRecorder()
	s.handleAgentPost(rr, req)
	if rr.Code != 200 {
		t.Fatalf("POST working: %d %s", rr.Code, rr.Body.String())
	}
	st := s.getAgentStatus()
	if st.State != "working" || st.Note != "editing" || st.Ref != "n1" {
		t.Fatalf("unexpected status: %+v", st)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/agent", nil)
	rr = httptest.NewRecorder()
	s.handleAgentGet(rr, req)
	var got AgentStatus
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.State != "working" {
		t.Fatalf("GET: %+v", got)
	}

	req = httptest.NewRequest(http.MethodPost, "/api/agent", bytes.NewBufferString(`{"state":"idle"}`))
	rr = httptest.NewRecorder()
	s.handleAgentPost(rr, req)
	if s.getAgentStatus().State != "idle" {
		t.Fatalf("expected idle after POST")
	}
}

func TestAckClearsMatchingAgent(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)

	s.setAgentStatus("working", "note-1", "patching")
	s.clearAgentIfRef("note-1")
	if s.getAgentStatus().State != "idle" {
		t.Fatal("expected clear on matching ref")
	}

	s.setAgentStatus("working", "note-1", "patching")
	s.clearAgentIfRef("other")
	if s.getAgentStatus().State != "working" {
		t.Fatal("expected keep when ref mismatches")
	}

	s.setAgentStatus("working", "", "any")
	s.clearAgentIfRef("anything")
	if s.getAgentStatus().State != "working" {
		t.Fatal("expected keep when working has no ref")
	}
}
