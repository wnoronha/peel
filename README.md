# peel

CLI to browse markdown docs, slide decks, and C4 architecture models in the browser. Single Go binary, embedded renderer, live reload, and block annotations for agent review loops.

## Usage

```sh
# default: ./docs if it exists, else cwd
./peel

# explicit docs directory and port
./peel --dir ./docs --port 18124 --host 127.0.0.1
./peel /path/to/docs --port 18200

# print version
./peel --version
```

Open `http://127.0.0.1:18124/?file=architecture.md`.

## Features

- **Format support**: Serves `*.md` docs, Marp slide stacks (`marp: true`), and multi-level `*.c4` diagrams.
- **C4 Architecture viewer**: `*.c4` files separate levels with `---`. Supports `C4Context`, `C4Container`, `C4Component`, and `C4Dynamic`. Includes level tabs, breadcrumbs, zoom controls, and click-a-box drill-down.
- **Embedded engine**: Single Go binary embedding all frontend assets. Markdown is parsed server-side via Goldmark; diagrams are rendered client-side via vendored Pele.
- **Graph & Links**: Resolves `[[wikilinks]]` by basename (anchors and `|labels` supported) and standard Markdown links to local `.md`/`.c4` files (relative to the current doc). External `http(s)` links open in a new tab. Computes backlinks and link counts across both forms.
- **Real-time live reload**: File watcher pushes disk modifications over Server-Sent Events (`/events`) to reload docs instantly.
- **Media asset fallback**: Automatically serves local diagram images and media files (`.png`, `.svg`, `.webp`, `.pdf`) referenced in markdown docs.

## Keyboard shortcuts

| Key | Action |
|---|---|
| `Ctrl-P` / `Cmd-P` | Open Goto Anything (jump to file) |
| `Ctrl-K` / `Cmd-K` | Open Goto Anything (same as `P`) |
| `[` | Toggle files sidebar |
| `]` | Toggle feedback drawer |
| `←` / `→` | Previous / next slide or C4 level |
| `Home` / `End` | First / last slide or C4 level |
| `Esc` | Close Goto palette, blur filter, or close composer |

## Feedback (agent review loop)

1. Drag-select text in any document or diagram label.
2. Select an action chip or write a note, then click **Save** (`Cmd+Enter`).
3. Appends structured JSONL to `<root>/.peel/feedback.jsonl`:
   `{"id","ts","file","quote","startLine","endLine","note"}`
4. Your coding agent or IDE watches `feedback.jsonl` to apply changes.
5. Disk writes push via SSE (`/events`), refreshing the document in the browser.

## API

- `GET /api/files` → `{root, files: [{path,name,type,title,size,modTime,words,links,backlinks}]}`
- `GET /api/doc?path=a.md` → `{path,title,type,isDeck,isC4,raw,body,html,slides[],levels[],outline[],backlinks[]}`
- `GET /api/feedback?file=a.md` → `{feedback: []}`
- `POST /api/feedback` → appends entry
- `DELETE /api/feedback?id=<id>` → removes entry by id
- `GET /events` → Server-Sent Events (`{type:"reload", path}`)
- `GET /healthz` → `ok` (status 200)

## Build

```sh
# build binary
make build

# run test suite
make test
```

## Test fixture

`fixtures/google-search/` — 11 docs + 3 decks reverse-engineering Google Search architecture from public sources. Rich wikilinks, Mermaid diagrams, and backlinks.

```sh
./peel --dir fixtures/google-search --port 18125
```

## Credits

- Design tokens, typography, and paper layout inspired by [opencode.ai](https://opencode.ai).
- Diagram rendering by [Pele](https://github.com/obsidianmd/pele) (Obsidian zero-dependency Mermaid/C4 SVG engine).
- Client-side interactivity by [Alpine.js](https://alpinejs.dev/).
- Server-side markdown parsing by [Goldmark](https://github.com/yuin/goldmark).
- File system event monitoring by [fsnotify](https://github.com/fsnotify/fsnotify).

## License

[MIT](LICENSE)