---
marp: true
title: Loop smoke deck
---

# Loop smoke deck

Minimal slides to exercise deck mode + the agent loop.
Pitch decks: [[deck-peel-tour]] · [[deck-use-cases]]

---

## The happy path

```mermaid
sequenceDiagram
  actor U as User
  participant UI as Web UI
  participant API as /api/feedback
  participant A as peel_api / agent
  U->>UI: select quote + Save
  UI->>API: POST note
  API-->>A: SSE feedback
  A->>A: edit source
  A->>API: POST ack
  API-->>UI: SSE feedback + toast
```

---

## Leave a note on this line

Select the sentence below, chip **Fix this**, Save.
Then pair an agent with `skills/peel/peel_api.py wait-note`.

> Smoke target: this paragraph should become clearer after an agent ack.

Guide: [[annotate]] · [[agent-loop]]
