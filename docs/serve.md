# Serve peel

## Start the server

Build once, then point peel at a docs folder:

```sh
go build -o peel .
./peel --dir ./docs --port 18124 --host 127.0.0.1
```

With no `--dir`, peel serves `./docs` when it exists, else the working directory. A bare path works too: `./peel /some/docs --port 18200`. Open the printed address and add `?file=overview.md` to land on a page.

## Restart after a code change

Web assets are embedded in the binary, so rebuild and restart for UI changes. Web assets serve with `Cache-Control: no-store`, so one normal reload shows the new UI.

```sh
go build -o peel . && ./peel --dir fixtures/google-search --host 0.0.0.0 --port 18124
```

Content changes need no restart. The file watcher pushes reloads over SSE, and the browser refreshes by itself. Check `/healthz` for `ok` when something looks dead.

## Flags

`--dir` picks the docs root. `--port` defaults to `18124`. `--host` defaults to `127.0.0.1`; use `0.0.0.0` to expose the server on the local network.
