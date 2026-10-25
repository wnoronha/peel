---
marp: true
title: Peel in five minutes
---

# Peel in five minutes

One binary. Docs, decks, and C4 models. Notes flow to agents.

---

## What it serves

- `*.md` files become docs.
- `marp: true` files become slide decks.
- `*.c4` files become C4 models.

See [[peel-overview]].

---

## The loop

```mermaid
flowchart LR
  S[select text] --> N[write note]
  N --> J[feedback.jsonl]
  J --> A[agent acts]
  A --> R[browser reloads]
```

---

## Modes

Docs for reading, decks for presenting, C4 for architecture.

- Model: [[c4-peel]]
- Hub: [[index]]

---

## Notes carry the work

```json
{ "file": "overview.md", "quote": "...", "note": "Fix this: ..." }
```

One object per line in `.peel/feedback.jsonl`. Agents watch the file.

---

## Keys

Arrows move slides and levels. `Ctrl-K` filters files. `[` and `]` toggle panels. `⌘↵` saves the composer.
