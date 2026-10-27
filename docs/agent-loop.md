# Agent review loop

How a coding agent should work with peel’s feedback log. Humans use the browser; agents use the file (or the API).

## Setup

```sh
./peel --dir ./docs --port 18124
```

Log path: `docs/.peel/feedback.jsonl` (gitignored). Contract: [[feedback-contract]].

## Wake up

Watch **mtime** or **tail** the jsonl. Polling `GET /api/feedback?all=1` also works. Prefer the file for local agents — no SSE required for feedback.

## Handle a note

1. Read new `type: "note"` lines (or open set from `GET /api/feedback`).
2. Prefer matching on `quote`. Use `file`, `startLine`, and anchors (`headingId`, `slide`, `level`, `levelKind`, `c4Id` / `c4Label`) as hints. On C4, `c4Id` is the source element id (`Person(reader, …)` → `reader`).
3. Honor `action` when present: `fix` | `explain` | `example` | `simplify` | `question`.
4. Edit the source under the docs root.
5. Ack:

```http
POST /api/feedback
Content-Type: application/json

{"type":"ack","ref":"<note-id>","note":"updated the stale figure"}
```

6. Doc edits trigger SSE reload in open browsers. The note moves to **Acked** in the drawer.

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

## Ink

Notes may include `"ink": ".peel/ink/<id>.svg"`. Fetch via `GET /api/feedback/ink?id=<id>`. Prefer `quote` + `note` for edits; treat ink as visual context. Reply notes may set `ref` to a parent note id (not separate open-queue items).

## Human UI

- **Feedback** — sidebar inbox (`]`)
- **Annotate** — draw mode (select text, then pen)
- **Open** — work queue (root notes; marks + ink in the doc)
- **Acked** — agent finished
- **Dismissed** — human discarded
- **All** — full history for the current file scope

Related: [[annotate]], [[use-cases]], [[api]].
