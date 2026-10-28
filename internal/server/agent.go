package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

const (
	EventAgent      = "agent" // ephemeral presence — not written to feedback.jsonl
	agentWorkingTTL = 5 * time.Minute
)

// AgentStatus is the public JSON for GET /api/agent and SSE type=agent.
type AgentStatus struct {
	State string `json:"state"` // working | idle
	Ref   string `json:"ref,omitempty"`
	Note  string `json:"note,omitempty"`
	Since string `json:"since,omitempty"` // RFC3339
}

type agentStored struct {
	State string
	Ref   string
	Note  string
	Since string
	until time.Time
}

func (s *Server) getAgentStatus() AgentStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.agent.State == "working" && !s.agent.until.IsZero() && s.agent.until.Before(time.Now()) {
		s.agent = agentStored{State: "idle"}
	}
	out := AgentStatus{State: s.agent.State, Ref: s.agent.Ref, Note: s.agent.Note, Since: s.agent.Since}
	if out.State == "" {
		out.State = "idle"
	}
	return out
}

func (s *Server) setAgentStatus(state, ref, note string) AgentStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if state != "working" {
		s.agent = agentStored{State: "idle"}
	} else {
		s.agent = agentStored{
			State: "working",
			Ref:   ref,
			Note:  note,
			Since: now.Format(time.RFC3339),
			until: now.Add(agentWorkingTTL),
		}
	}
	return AgentStatus{State: s.agent.State, Ref: s.agent.Ref, Note: s.agent.Note, Since: s.agent.Since}
}

func (s *Server) clearAgentIfRef(ref string) {
	if ref == "" {
		return
	}
	s.mu.Lock()
	clear := s.agent.State == "working" && s.agent.Ref != "" && s.agent.Ref == ref
	if clear {
		s.agent = agentStored{State: "idle"}
	}
	s.mu.Unlock()
	if clear {
		s.emitAgent(AgentStatus{State: "idle"})
	}
}

func (s *Server) emitAgent(st AgentStatus) {
	payload := map[string]any{"type": EventAgent, "state": st.State}
	if st.Ref != "" {
		payload["ref"] = st.Ref
	}
	if st.Note != "" {
		payload["note"] = st.Note
	}
	if st.Since != "" {
		payload["since"] = st.Since
	}
	msg, err := json.Marshal(payload)
	if err != nil {
		return
	}
	s.broadcast(string(msg))
}

func (s *Server) handleAgentGet(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.getAgentStatus())
}

func (s *Server) handleAgentPost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 256<<10)
	var body struct {
		State string `json:"state"`
		Ref   string `json:"ref"`
		Note  string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	state := strings.ToLower(strings.TrimSpace(body.State))
	if state != "working" && state != "idle" {
		http.Error(w, "state must be working or idle", http.StatusBadRequest)
		return
	}
	ref := strings.TrimSpace(body.Ref)
	if len(ref) > 128 {
		http.Error(w, "ref too large", http.StatusBadRequest)
		return
	}
	note := truncateRunes(strings.TrimSpace(body.Note), 200)
	st := s.setAgentStatus(state, ref, note)
	s.emitAgent(st)
	writeJSON(w, map[string]any{"ok": true, "state": st.State, "ref": st.Ref, "note": st.Note, "since": st.Since})
}
