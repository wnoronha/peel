# Agent review loop

How a coding agent should work with peel’s feedback log. Humans use the browser; agents use the API via the skill script (or the file as fallback).

Agent skill (portable): [`skills/peel/SKILL.md`](../skills/peel/SKILL.md). In Cursor run **`/peel`**. Prefer **`skills/peel/peel_api.py`** for all HTTP/SSE so one shell approval covers every subcommand — do not invent ad-hoc curls.

## Setup

```sh
./peel --dir ./docs --port 18124
```

Log path: `docs/.peel/feedback.jsonl` (gitignored). Contract: [[feedback-contract]].

## Wake up

**Primary (default: `pair`):** `skills/peel/peel_api.py pair` or `wait-note` (SSE with poll fallback). On connect, `open` for catch-up.

- **`pair` (default / long-lived process):** Recommended default under shell approval workflows (`skills/peel/peel_api.py pair --timeout 600`). Runs health check, lists open notes for catch-up, and continuously streams new notes to stdout in a single process.
- **`wait-note` (discrete turn-based):** Blocks until **one** new root note arrives, prints its JSON, and exits (code 0, or 124 on timeout). Used when an agent framework requires discrete per-turn tool calls: wait for one note → execute edit → ack → invoke `wait-note --since-id <id>` again. Supports `--since-id <id>` and `--file <path>` filtering.

**Fallback:** watch mtime / `tail -F` the jsonl, or poll `GET /api/feedback?all=1`.

## Handle a note

1. Read new `type: "note"` lines (or open set from `GET /api/feedback`).
2. Prefer matching on `quote`. Use `file`, `startLine`, and anchors (`headingId`, `slide`, `level`, `levelKind`) as hints.
3. Honor `action` when present: `fix` | `explain` | `example` | `simplify` | `question`.
4. Edit the source under the docs root.
5. Ack:

```http
POST /api/feedback
Content-Type: application/json

{"type":"ack","ref":"<note-id>","note":"updated the stale figure"}
```

6. Doc edits trigger SSE `reload` in open browsers. Feedback writes also emit SSE `feedback` so the panel (and `wait-note`) wake without polling. The note moves to **Acked** in the drawer.

## Do not

- Hand-edit `feedback.jsonl` (ids and timestamps are server-owned).
- Rewrite the file to delete — humans dismiss via UI (tombstone); agents ack.
- Assume line numbers survive large edits — re-find the quote.

## Example note

```json
{
  "type": "note",
  "id": "20261026T120000.000000001",
  "file": "architecture.md",
  "quote": "Vanilla JS frontend, no HTMX/Alpine",
  "action": "fix",
  "note": "Fix this: we use Alpine now",
  "headingId": "decisions",
  "startLine": 12,
  "endLine": 12
}
```

## Human UI

- **Open** — work queue (marks in the doc)
- **Acked** — agent finished
- **Dismissed** — human discarded
- **All** — full history for the current file scope

Related: [[annotate]], [[use-cases]], [[api]].
