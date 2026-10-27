# API reference

Base URL is the serving address (default `http://127.0.0.1:18124/`). REST returns JSON; live updates use SSE (`GET /events`). Error bodies are plain text with an HTTP status.

Agents should prefer [`skills/peel/peel_api.py`](../skills/peel/peel_api.py) over raw curls. Specs: [[feedback-contract]], [[agent-loop]], [[c4-format]].

## Endpoints

| Method & path | Description | Format |
|---|---|---|
| `GET /api/files` | List markdown and C4 files | JSON |
| `GET /api/doc?path=` | Render document, slides, or C4 levels | JSON |
| `GET /api/feedback` | Folded open notes or raw log (`all=1`) | JSON |
| `POST /api/feedback` | Append note, ack, or tombstone | JSON |
| `DELETE /api/feedback?id=` | Tombstone a note | JSON |
| `GET /api/agent` | Ephemeral agent presence (`working` \| `idle`) | JSON |
| `POST /api/agent` | Set presence — pulses logo when `working` | JSON |
| `GET /events` | Live stream: reloads, feedback, agent | SSE |
| `GET /healthz` | Health probe (`ok`) | Text |

## `GET /api/files`

Sorted by path. Skips `.git`, `.peel`, `node_modules`. Only `.md`, `.markdown`, `.c4`.

```json
{ "root": "docs", "files": [
  { "path": "architecture.md", "name": "architecture.md",
    "type": "doc", "isDeck": false, "isC4": false,
    "title": "Architecture",
    "size": 2400, "modTime": "2026-10-27T20:00:00Z",
    "words": 380, "links": 12, "backlinks": 4 }
]}
```

| Field | Notes |
|---|---|
| `type` | `doc` \| `deck` \| `c4` (UI tabs / Goto icons) |
| `isDeck` / `isC4` | Convenience booleans |
| `links` | `[[wikilinks]]` + local Markdown doc/C4 links |
| `backlinks` | Other files that reference this path |

## `GET /api/doc?path=architecture.md`

Accepts `path` or `file`.

```json
{ "path": "architecture.md", "title": "Architecture",
  "type": "doc", "isDeck": false, "isC4": false,
  "raw": "...", "body": "...", "html": "<h1>…</h1>",
  "slides": [], "levels": [],
  "outline": [{"level":2,"text":"Server","line":7,"id":"server"}],
  "backlinks": ["index.md"] }
```

| Mode | How | Payload |
|---|---|---|
| Doc | plain `.md` | `html`, heading `outline` |
| Deck | `marp: true` | `slides[]` HTML per `---` section; `type`/`isDeck` |
| C4 | `.c4` | `levels[]` of `{name, kind, source}`; `html` empty |

Title order: frontmatter `title` → first heading (md) / first `title` line (c4) → basename.

C4 `kind` values: `Context` \| `Container` \| `Component` \| `Code` \| `Deployment`. Directives `C4Code` and `C4Dynamic` both map to kind **Code**. Unknown path → 404; path escape → 400.

## `GET /api/feedback`

| Query | Result |
|---|---|
| (none) | All **open** notes (folded) |
| `?file=path` | Open notes for that file |
| `?all=1` | Raw event log |
| `?all=1&file=path` | Raw events filtered to notes for that path |

Response shape: `{ "feedback": [ … ] }`. Folding rule: open = `note` with no later `ack` / `tombstone` for that `id`.

## `POST /api/feedback`

Appends one event and `fsync`s. Also broadcasts SSE `{ "type":"feedback", "event": … }`.

### note (default)

Requires `file` plus `note` and/or `quote`. Optional:

| Field | Meaning |
|---|---|
| `action` | `fix` \| `explain` \| `example` \| `simplify` \| `question` |
| `startLine` / `endLine` | 1-based hints |
| `headingId` | Markdown section |
| `slide` | Deck slide (1-based) |
| `level` / `levelKind` | C4 level |
| `c4Id` / `c4Label` | Pele node or edge id / label |
| `c4From` / `c4To` | Rel endpoints when `c4Id` is an edge |

### ack

```json
{ "type": "ack", "ref": "<note-id>", "note": "patched the figure" }
```

### tombstone

```json
{ "type": "tombstone", "ref": "<note-id>" }
```

Returns `{ "ok": true, "id": "…", "type": "note|ack|tombstone" }`. Bad file/type/action/ref → 400.

## `DELETE /api/feedback?id=<note-id>`

Appends a tombstone for an **open** note (no rewrite). `{ "ok": true, "deleted": true|false }`.

## `GET /events`

`Content-Type: text/event-stream`. Comment heartbeats (`: hb`). Each `data:` line is JSON:

| `type` | Payload | When |
|---|---|---|
| `reload` | `{ "type":"reload", "path":"…" }` | `.md` / `.c4` under root changes |
| `feedback` | `{ "type":"feedback", "event": {…Feedback} }` | After note / ack / tombstone |
| `agent` | `{ "type":"agent", "state":"working\|idle", … }` | Presence change (logo pulse) |

`peel_api.py wait-note` / `pair` subscribe here (poll fallback with `--poll`). Use `busy` / `idle` for the pulsing logo.

## `GET|POST /api/agent`

Ephemeral presence (not written to `feedback.jsonl`). TTL ~5 minutes; matching `ack` clears `working`.

```json
// POST { "state":"working", "note":"editing architecture", "ref":"<note-id>" }
// GET  { "state":"working", "note":"…", "ref":"…", "since":"…" }
```

## `GET /healthz`

Returns body `ok`, status 200.
