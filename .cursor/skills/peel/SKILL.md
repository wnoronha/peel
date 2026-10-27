---
name: peel
description: >-
  Peel human↔agent pairing over feedback.jsonl. Invoke with /peel to start
  pairing and keep monitoring notes until the human stops.
disable-model-invocation: true
---

# /peel

Read and follow the **canonical** project skill:

**[skills/peel/SKILL.md](../../../skills/peel/SKILL.md)**

Then enter **pairing mode** immediately:

1. `skills/peel/peel_api.py health` → `open`
2. Block on **`skills/peel/peel_api.py pair --timeout 600`** (or `wait-note`) — one long-lived process so shell approval is not repeated per curl.
3. On each note JSON: edit → `peel_api.py ack <id> "…"` → wait again.

**Do not** use raw `curl` or write a second SSE client — only `peel_api.py`.
