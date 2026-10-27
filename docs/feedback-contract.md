# Feedback contract

Agents and the browser share `<root>/.peel/feedback.jsonl`. One JSON object per line. The file is an **append-only event log**: notes, acks, and tombstones. Never edit prior lines by hand. Only the server writes (`POST` / `DELETE` `/api/feedback`), so `id` and `ts` stay ordered. Prefer SSE wake-ups (`GET /events`, type `feedback`) via `skills/peel/peel_api.py`; fall back to `tail -F` or poll.

Folder layout: the `.peel/` directory is created on first write and should stay gitignored.

## Event types

| `type` | Meaning |
|--------|---------|
| `note` | Human annotation (default when `type` is missing — legacy lines) |
| `ack` | Agent or human finished a note; `ref` is the note `id` |
| `tombstone` | Note dismissed; `ref` is the note `id` |

**Open notes** = `note` events with no later `ack` or `tombstone` for that `id`.

| Read path | Result |
|---|---|
| `GET /api/feedback` | Folded open set |
| `GET /api/feedback?file=` | Open notes for one file |
| `GET /api/feedback?all=1` | Raw log |
| jsonl on disk | Same raw events the API appends |

## Note event

```json
{ "type": "note", "id": "20261025T210527.342062449", "ts": "2026-10-25T21:05:27Z",
  "file": "architecture.md", "quote": "Any write under the docs root broadcasts",
  "startLine": 9, "endLine": 9, "action": "fix",
  "note": "Fix this: also mention feedback SSE",
  "headingId": "server" }
```

C4-shaped note (box or Rel):

```json
{ "type": "note", "id": "…", "ts": "…",
  "file": "c4-peel.c4", "quote": "CLI + browser",
  "action": "question", "note": "Should Rel wording mention SSE?",
  "level": 1, "levelKind": "Context",
  "c4Id": "user-peel", "c4Label": "CLI + browser",
  "c4From": "user", "c4To": "peel" }
```

| Field | Rules |
|---|---|
| `id` | Server-minted; unique per event |
| `ts` | RFC 3339 UTC |
| `file` | Path relative to docs root |
| `quote` | Selected / focused text — prefer when rematching after edits |
| `startLine` / `endLine` | 1-based hints; may be `0` / omitted if unmapped |
| `action` | Optional chip: `fix` \| `explain` \| `example` \| `simplify` \| `question` |
| `note` | Free text |
| `headingId` | Markdown section slug |
| `slide` | Deck slide (1-based) |
| `level` / `levelKind` | C4 level index + kind |
| `c4Id` / `c4Label` | Pele node or edge id / label (`sanitize`: letters, digits, `_`, `-`, max 96) |
| `c4From` / `c4To` | Rel endpoints when commenting an edge |

## Ack event

```json
{ "type": "ack", "id": "20261025T211000.000000001", "ts": "2026-10-25T21:10:00Z",
  "ref": "20261025T210527.342062449", "note": "updated the SSE wording" }
```

Agents POST this after applying a note. UI shows optional `note` as `ack: …` under **Acked**.

## Tombstone event

```json
{ "type": "tombstone", "id": "20261025T211500.000000002", "ts": "2026-10-25T21:15:00Z",
  "ref": "20261025T210527.342062449" }
```

Emitted by `DELETE /api/feedback?id=<note-id>` (Feedback `×`) when the note is still open.

## Live wake-up

Every successful append also broadcasts:

```json
{ "type": "feedback", "event": { "type": "note|ack|tombstone", … } }
```

on `GET /events`. Doc file changes broadcast `{ "type":"reload", "path":"…" }` separately.

## Agent write path

```sh
skills/peel/peel_api.py ack '<note-id>' 'what changed'
```

Do not hand-edit the jsonl. Entries for deleted source files stay until acked or tombstoned.
