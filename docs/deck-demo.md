---
marp: true
title: Demo deck
---

# Demo deck

Minimal slide stack for smoke-testing deck mode.
For the real pitch, use [[deck-peel-tour]] or [[deck-use-cases]].

---

## Diagram slide

```mermaid
sequenceDiagram
  participant R as Reader
  participant P as peel
  participant A as Agent
  R->>P: note
  A->>P: ack
  A->>P: edit sources
  P-->>R: SSE reload
```

---

## Done

Open [[annotate]] and leave a note on this line.
