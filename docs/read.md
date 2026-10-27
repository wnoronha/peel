# Read docs with peel

Serve a folder, open a page, leave a first note. You need Go and a markdown tree.

## Serve this dogfood folder

```sh
go build -o peel .
./peel --dir ./docs --port 18124
```

Open `http://127.0.0.1:18124/`. You get Files, the hub [[index]], and empty side panels until you annotate.

Fixture alternative (search-domain sample):

```sh
./peel --dir fixtures/google-search --port 18124
```

## Read a page

Click a file or press `Ctrl-P` / `Cmd-P` (Goto Anything). Use **Files**, **Decks**, or **C4** in the top bar for the left drawer tabs.

Worth opening once: [[use-cases]], [[deck-peel-tour]], [[c4-peel]].

## Save a first note

1. Drag-select a sentence.
2. Chip **Fix this** (or type your own note).
3. `⌘↵` / **Save**.
4. Open **Feedback** — the note is **Open**.

It was appended to `docs/.peel/feedback.jsonl` (or the fixture’s `.peel/`):

```sh
tail -n 1 docs/.peel/feedback.jsonl
```

Next: [[annotate]] for the full UI, [[agent-loop]] to wire an agent.
