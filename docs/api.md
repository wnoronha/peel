# API reference

Base URL is the serving address, for example `http://127.0.0.1:18124/`. Every response is JSON. Error bodies are plain text with an HTTP status.

## `GET /api/files`

Lists every served file, sorted by path.

```json
{ "root": "google-search", "files": [
  { "path": "overview.md", "name": "overview.md",
    "isDeck": false, "isC4": false, "title": "Google Search Architecture",
    "size": 1183, "modTime": "2026-10-25T18:18:00Z",
    "words": 120, "links": 9, "backlinks": 7 }
]}
```

`links` counts `[[wikilinks]]` and local Markdown doc links (`[label](./doc.md)`) in the file. `backlinks` counts other files that reference it via either form. Only `.md`, `.markdown`, and `.c4` files are listed. `.git`, `.peel`, and `node_modules` are skipped.

## `GET /api/doc?path=overview.md`

Renders one file. Accepts `path` or `file`.

```json
{ "path": "overview.md", "title": "Google Search Architecture",
  "isDeck": false, "isC4": false, "raw": "...", "body": "...",
  "html": "<h1>...</h1>", "slides": [],
  "levels": [], "outline": [], "backlinks": ["crawling.md"] }
```

Titles resolve as frontmatter `title`, then first markdown heading, then filename. Decks fill `slides` with rendered HTML per `---` section. C4 files fill `levels` with `{name, kind, source}` per section (`kind` is Context, Container, Component, Code, or Deployment) and leave `html` empty. `outline` holds `{level, text, line, id}` per heading, or per C4 level. Unknown paths return 404, unsafe paths return 400.

## `GET /api/feedback?file=overview.md`

Returns the **folded open-note set** (notes without a later ack or tombstone). Omit `file` for every open note. Pass `all=1` for the raw event log. Full contract: `feedback-contract.md`.

## `POST /api/feedback`

Appends one event (`fsync`).

- **note** (default): `file` plus `note` and/or `quote`. Optional `action`, `startLine`/`endLine`, `headingId`, `slide`, `level`, `levelKind`, `c4Id`, `c4Label`, `ref` (parent note id), `inkSvg` (SVG string → writes `.peel/ink/<id>.svg` and sets `ink`). Ink requires `quote`; ink-only notes get `note: "Ink mark"`. On C4, prefer `c4Id` (element id) over mashed SVG text quotes.
- **ack**: `{ "type": "ack", "ref": "<note-id>", "note": "…" }`
- **tombstone**: `{ "type": "tombstone", "ref": "<note-id>" }`

Returns `{ok: true, id, type, ink?}`. Bad file/type/action/ref/ink → 400.

## `GET /api/feedback/ink?id=<note-id>`

Returns the SVG for a note that has `ink`, `Content-Type: image/svg+xml`. 404 if missing.

## `DELETE /api/feedback?id=<note-id>`

Appends a **tombstone** for an open note (does not rewrite the log). Returns `{ok: true, deleted}` (`deleted` false if missing or already closed).

## `GET /events`

Server-sent events. Payloads are `{type: "reload", path}` when a file under the root changes, plus `: hb` heartbeats every 20 seconds.

## `GET /healthz`

Returns `ok` with status 200.
