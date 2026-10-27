# Feedback contract spec

Agents read `<root>/.peel/feedback.jsonl`. One JSON object per line. The file is an **append-only event log**: notes, acks, and tombstones. Never edit prior lines by hand. Only the server writes, through `POST` / `DELETE` on `/api/feedback`, so `id` and `ts` stay server-ordered. Watch the file (tail or mtime) or poll the API.

## Event types

| `type` | Meaning |
|--------|---------|
| `note` | Annotation (default when `type` is missing — legacy lines). Optional `ink` path; optional `ref` to a parent note |
| `ack` | Agent or human handled a note; `ref` is the note `id` |
| `tombstone` | Note dismissed; `ref` is the note `id` (replaces rewrite-delete) |

**Open notes** = root `note` events (`ref` empty) with no later `ack` or `tombstone` for that `id`. Ack/tombstone of a root also folds its reply notes. `GET /api/feedback` returns the folded open set. `GET /api/feedback?all=1` returns the raw log.

## Note event

```json
{ "type": "note", "id": "20261025T210527.342062449", "ts": "2026-10-25T21:05:27Z",
  "file": "overview.md", "quote": "How web search works",
  "startLine": 3, "endLine": 3, "action": "fix",
  "note": "Fix this: stale number",
  "headingId": "how-web-search-works",
  "ink": ".peel/ink/20261025T210527.342062449.svg" }
```

- `id` — server timestamp at write; unique per event  
- `ts` — RFC 3339 UTC  
- `file` — path relative to the docs root  
- `quote` — selected text (prefer this when matching after edits); required when attaching ink  
- `startLine` / `endLine` — 1-based source hints; `0` if unmapped  
- `action` — optional chip intent: `fix` | `explain` | `example` | `simplify` | `question`  
- `note` — free text (defaults to `Ink mark` when only ink is sent)  
- `ink` — optional relative path to an SVG under `.peel/ink/<id>.svg` (never embed SVG in the jsonl line)  
- `ref` — optional parent note `id` (reply / follow-up ink or text). Reply notes are not open-queue roots  
- Anchors (optional): `headingId` (markdown), `slide` (1-based deck), `level` + `levelKind` (C4), `c4Id` + `c4Label` (C4 element — Pele `data-id` and source display name, e.g. `peel` / `"Peel"`)

Create with optional `inkSvg` in the POST body (SVG string, max 200KB). The server writes the file, sets `ink`, and does not persist `inkSvg` in the log. Fetch SVG via `GET /api/feedback/ink?id=<note-id>`.

## Ack event

```json
{ "type": "ack", "id": "20261025T211000.000000001", "ts": "2026-10-25T21:10:00Z",
  "ref": "20261025T210527.342062449", "note": "updated the number" }
```

Agents `POST` this after applying a note. Do not hand-edit the jsonl file.

## Tombstone event

```json
{ "type": "tombstone", "id": "20261025T211500.000000002", "ts": "2026-10-25T21:15:00Z",
  "ref": "20261025T210527.342062449" }
```

Emitted by `DELETE /api/feedback?id=<note-id>` when the note is still open.

## Reading

- Open backlog: `GET /api/feedback` or `?file=path`  
- Full history: `GET /api/feedback?all=1` or tail the jsonl file  
- Prefer `quote` (and anchors) over `startLine` after edits. Entries for deleted files stay in the log until tombstoned or acked.

## Writing

Only the server appends (and `fsync`s). Clients and agents use the HTTP API.
