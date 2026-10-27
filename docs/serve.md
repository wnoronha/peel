# Serve peel

## Start the server

Build once, then point peel at a docs folder:

```sh
go build -o peel .
./peel --dir ./docs --port 18124 --host 127.0.0.1
```

With no `--dir`, peel serves `./docs` when it exists, else the working directory. A bare path works too: `./peel /some/docs --port 18200`. Open the printed address; add `?file=index.md` (or [[deck-peel-tour]]) to land on a page.

## Restart after a code change

Web assets are embedded in the binary, so rebuild and restart for UI changes. Web assets serve with `Cache-Control: no-store`, so one normal reload shows the new UI.

```sh
make restart          # stop via pid file, then rebuild+run --dir docs
# or
make stop && make run
```

**Do not** `kill $(lsof -t -i :18124)` — that also hits browser tabs connected to peel (e.g. Firefox). Prefer the pid file.

Content changes need no restart. The file watcher pushes reloads over SSE, and the browser refreshes by itself. Check `/healthz` for `ok` when something looks dead.

## Pid file

On start, peel writes `<docs-root>/.peel/peel.pid` (gitignored with `.peel/`). Override with `PEEL_PID_FILE=/path/to/peel.pid`. A second start fails if that pid is still alive. Ctrl-C / SIGTERM removes the file after shutdown.

```sh
make stop             # kill only the pid in docs/.peel/peel.pid
PEEL_PID_FILE=./x.pid make stop
```

## Flags

`--dir` picks the docs root. `--port` defaults to `18124`. `--host` defaults to `127.0.0.1`; use `0.0.0.0` to expose the server on the local network.
