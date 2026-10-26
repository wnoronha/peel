# Read docs with peel

You will serve a docs folder, read a page, and save a first note. You need Go and a folder of markdown files.

## Serve the fixture docs

First, build the binary:

```sh
go build -o peel .
```

Next, serve the bundled fixture:

```sh
./peel --dir fixtures/google-search --port 18124
```

Open `http://127.0.0.1:18124/`. You see the file list, the first doc, and empty side panels.

## Read a page

Click `overview.md` in the file list. The doc renders in the center column. Press `Ctrl-P` (or `Cmd-P`) to open Goto Anything, type a few letters, and press Enter to open another file.

Click **Files**, **Decks**, or **C4** in the top bar to open the left panel on that tab.

## Save a first note

Now, do the following:

1. Drag-select a sentence in the doc. A small composer pops up near the selection.
2. Pick a quick action, for example **Fix this**. The box fills with `Fix this: `.
3. Type the problem after the colon.
4. Press `⌘↵` (or `Ctrl↵`) or click **Save**.

Open the **Feedback** link in the top bar. Your note is listed there. The same entry was appended to `fixtures/google-search/.peel/feedback.jsonl`:

```sh
tail -n 1 fixtures/google-search/.peel/feedback.jsonl
```

You now have a running loop: select text, save a note, and the JSONL file carries it to your agent. `annotate.md` covers the whole flow.
