# peel

CLI to browse markdown docs, slide decks, and C4 architecture models in the browser. One Go binary, embedded UI, live reload over SSE, and an append-only feedback log for agent review loops.

## Usage

```sh
# default: ./docs if it exists, else cwd
./peel

# explicit docs directory and port
./peel --dir ./docs --port 18124 --host 127.0.0.1
./peel /path/to/docs --port 18200

# or
make run

./peel --version
```

Open `http://127.0.0.1:18124/?file=index.md`. Dogfood docs live under `docs/` (tour: `deck-peel-tour.md`, model: `c4-peel.c4`).

## Features

- **Docs / decks / C4** in one shell — plain `*.md`, Marp (`marp: true`), multi-level `*.c4`
- **C4 viewer** — levels split on `---`; click box/Rel to comment (`c4Id` / `c4From`/`c4To`); double-click box to drill; zoom + kind icons
- **Goto Anything** — `Ctrl/Cmd-P` or `K`; wikilinks + local markdown links; backlinks
- **Feedback loop** — select text (or C4 click) → chips → Save → `<root>/.peel/feedback.jsonl` (note / ack / tombstone)
- **Live events** — `GET /events` emits `reload` (doc changes) and `feedback` (note/ack/tombstone); browser toasts + Feedback drawer
- **Agent skill** — portable `skills/peel/` (`peel_api.py` for health / open / wait-note / ack / pair). Cursor `/peel` is a thin wrapper

## Keyboard shortcuts

| Key | Action |
|---|---|
| `Ctrl/Cmd-P` or `K` | Goto Anything |
| `[` | Toggle files sidebar |
| `]` | Toggle feedback drawer |
| `←` / `→` | Previous / next slide or C4 level |
| `Home` / `End` | First / last slide or C4 level |
| `⌘↵` / `Ctrl↵` | Save composer |
| `Esc` | Close Goto / composer |

## Feedback (agent review loop)

1. Select text or **click** a C4 shape/Rel → chip or free text → **Save**.
2. Server appends a `note` to `<root>/.peel/feedback.jsonl` and broadcasts SSE `feedback`.
3. Agent: `skills/peel/peel_api.py pair --timeout 600` (or `wait-note`), edits sources, then `ack <id> "…"`.
4. Doc edits push SSE `reload`; the ack moves the note to **Acked** and shows a toast.

Contract: `docs/feedback-contract.md`. Playbook: `docs/agent-loop.md`.

## API

| Endpoint | Role |
|---|---|
| `GET /api/files` | File list (`type`: doc \| deck \| c4, counts, titles) |
| `GET /api/doc?path=` | Rendered doc / slides / C4 levels |
| `GET /api/feedback` | Folded open notes; `?all=1` raw log |
| `POST /api/feedback` | Append note, ack, or tombstone |
| `DELETE /api/feedback?id=` | Tombstone an open note |
| `GET /events` | SSE: `{type:reload\|feedback,…}` |
| `GET /healthz` | `ok` |

Full reference: `docs/api.md`.

## Build

```sh
make build   # → ./peel
make test
```

## Test fixture

`fixtures/google-search/` — sample docs + decks (search architecture).

```sh
./peel --dir fixtures/google-search --port 18125
```

## Credits

- Design tokens inspired by [opencode.ai](https://opencode.ai)
- Diagrams by [Pele](https://github.com/obsidianmd/pele)
- UI by [Alpine.js](https://alpinejs.dev/) · icons [Lucide](https://lucide.dev/)
- Markdown [Goldmark](https://github.com/yuin/goldmark) · watch [fsnotify](https://github.com/fsnotify/fsnotify)

## License

[MIT](LICENSE)
