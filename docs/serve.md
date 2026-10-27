# Serve peel

## Start the server

Build once, then point peel at a docs folder:

```sh
make build          # → ./peel
./peel --dir ./docs --port 18124 --host 127.0.0.1

# shortcuts
make run            # build + ./peel (defaults to ./docs when present)
./peel /some/docs --port 18200
```

With no `--dir`, peel serves `./docs` when it exists, else the working directory. Open the printed address; add `?file=index.md` (or [[deck-peel-tour]]) to land on a page.

## Pair an agent

In another terminal (or the agent shell):

```sh
export PEEL_URL=http://127.0.0.1:18124   # optional
skills/peel/peel_api.py health
skills/peel/peel_api.py pair --timeout 600
```

Skill docs: [[agent-loop]], [`skills/peel/SKILL.md`](../skills/peel/SKILL.md).

## Restart after a code change

Web assets are embedded (`go:embed web/*`), so rebuild and restart for UI/server changes. Assets serve with `Cache-Control: no-store`, so one normal reload shows the new UI.

```sh
go build -o peel . && ./peel --dir ./docs --port 18124
```

Content changes need no restart. The file watcher pushes `reload` over SSE; feedback appends push `feedback`. Check `/healthz` for `ok` when something looks dead.

## Flags

| Flag | Default | Role |
|---|---|---|
| `--dir` | `./docs` or cwd | Docs root |
| `--port` | `18124` | Listen port |
| `--host` | `127.0.0.1` | Bind address (`0.0.0.0` for LAN) |
| `--version` / `-v` | — | Print version and exit |

Feedback log path: `<root>/.peel/feedback.jsonl` (created on first note).
