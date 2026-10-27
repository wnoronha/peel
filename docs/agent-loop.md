# Agent review loop

How a coding agent pairs with peel. Humans annotate in the browser; agents wake on **SSE** (or files), edit sources, and **ack** over HTTP.

Portable skill: [`skills/peel/SKILL.md`](../skills/peel/SKILL.md). In Cursor, **`/peel`** is a thin wrapper. Prefer **`skills/peel/peel_api.py`** for all Peel HTTP/SSE so one shell approval covers every subcommand — do not invent ad-hoc curls.

## Setup

```sh
./peel --dir ./docs --port 18124
# optional
export PEEL_URL=http://127.0.0.1:18124
```

Log path: `<root>/.peel/feedback.jsonl` (gitignored). Contract: [[feedback-contract]]. API: [[api]].

## Wake up

**Primary — `peel_api.py`:**

| Command | Role |
|---|---|
| `pair --timeout 600` | Health + open catch-up + loop (long-lived; best under shell approval) |
| `wait-note --timeout 600` | Block until **one** root note; exit 0 or **124** on timeout |
| `wait-note --since-id <id>` | Skip notes already handled |
| `open` / `feedback` | Catch-up without waiting |
| `ack <id> "…"` | Close a note |

SSE payload for notes: `{ "type":"feedback", "event": { "type":"note", … } }`. The script falls back to poll if the stream fails (`--poll` to force).

**Fallback — files:** `tail -F <root>/.peel/feedback.jsonl` or poll `GET /api/feedback?all=1`.

## Handle a note

1. Take a root `note` (empty `ref`, not yet acked/tombstoned).
2. Prefer matching on `quote`. Use `file`, `startLine`, and anchors:
   - docs: `headingId`
   - decks: `slide` (1-based)
   - C4: `level`, `levelKind`, `c4Id`, `c4Label`, `c4From`, `c4To`
3. Honor `action` when present: `fix` | `explain` | `example` | `simplify` | `question`.
4. Edit sources under the docs root (not `.peel/`).
5. Ack:

```sh
skills/peel/peel_api.py ack '<note-id>' 'updated the stale figure'
```

Equivalent HTTP:

```http
POST /api/feedback
Content-Type: application/json

{"type":"ack","ref":"<note-id>","note":"updated the stale figure"}
```

6. Doc edits emit SSE `reload` (browser refreshes). The ack emits SSE `feedback` — the note moves to **Acked** and a toast appears.

## Do not

- Hand-edit `feedback.jsonl` (ids and timestamps are server-owned).
- Rewrite the file to delete — humans dismiss via UI (tombstone); agents ack.
- Assume line numbers survive large edits — re-find the quote or use `c4Id`.
- Mass-ack the open backlog unless the user points at specific notes.
- Call raw `curl` for Peel while pairing — extend `peel_api.py` instead.

## Example note

```json
{
  "type": "note",
  "id": "20261026T120000.000000001",
  "file": "architecture.md",
  "quote": "Any write under the docs root broadcasts",
  "action": "fix",
  "note": "Fix this: also mention feedback SSE, not only reload",
  "headingId": "server",
  "startLine": 9,
  "endLine": 9
}
```

C4 example anchors: `"c4Id":"reader-peel","c4From":"reader","c4To":"peel","level":1,"levelKind":"Context"`.

## What the human sees

| Surface | Signal |
|---|---|
| Feedback → **Open** | Work queue; peach marks on quotes |
| Feedback → **Acked** | Your `ack` text as `ack: …` |
| Toast (bottom-right) | Agent finished / note dismissed / doc updated |
| Logo pulse | `busy` → amber peel mark + status chip |
| Reader | Live reload after source edits |

![Agent working](assets/agent-working.png)

![Toast](assets/toast.png)

No chat thread — one note, one close.

Related: [[annotate]], [[use-cases]], [[api]], [[c4-peel]].
