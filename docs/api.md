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

Lists feedback entries in file order. Omit `file` to list every entry across files. The full entry contract is in `feedback-contract.md`.

## `POST /api/feedback`

Appends one entry. The body needs `file` plus a `note`, a `quote`, or both. `startLine` and `endLine` are optional. Returns `{ok: true, id}`. Missing file returns 400, empty note and quote returns 400.

## `DELETE /api/feedback?id=<id>`

Removes the entry with that id by rewriting the file. Always returns `{ok: true, deleted}` with whether a line matched.

## `GET /events`

Server-sent events. Payloads are `{type: "reload", path}` when a file under the root changes, plus `: hb` heartbeats every 20 seconds.

## `GET /healthz`

Returns `ok` with status 200.
