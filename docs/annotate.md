# Annotate a doc

Select text in the reader (docs, decks, or C4 labels), optionally draw with a stylus, and save a note agents can finish.

## Capture

1. Click **✎ Annotate** to enter draw mode (toggle / Esc to exit). Feedback stays the sidebar inbox — Annotate does not open it.
2. Drag-select text — a composer appears near the selection.
3. With draw mode on: **pen** strokes mark the page; **mouse** can draw after a selection (finger still scrolls). Stroke color defaults to synth pink (`#ff2ecb`); override via CSS `--color-synth-pink` or `applyInkStroke('#…')` in the app.
4. Pick a chip (**Fix this**, **Explain this**, **Add example**, **Simplify this**, **Question**) or type freely. Ink-only saves as note text `Ink mark`.
5. Press `⌘↵` / `Ctrl↵` or click **Save**. Clear ink from the composer if needed.

Saving needs a quote and either a note or ink. Closing the composer without save drops the draft (including unsaved strokes).

## Review

Open **Feedback** (`]`). Tabs:

| Tab | Meaning |
|---|---|
| **Open** | Active root notes — marks and ink in the doc |
| **Acked** | Agent finished (`type: ack`) |
| **Dismissed** | You discarded (`type: tombstone`) |
| **All** | History for the current scope |

**This file** vs all files filters the list. Click a row to jump to the quote. Click `×` to dismiss (tombstone), not hard-delete. Notes with ink show a small thumbnail.

## Agent side

Each save appends one JSON line under `<root>/.peel/feedback.jsonl`. Optional ink is a separate SVG under `.peel/ink/<note-id>.svg`, referenced by the note’s `ink` field. Reply notes may set `ref` to a parent note id (they are not separate open-queue items). Agents should edit sources, then `POST` an `ack` with `ref` set to the note id. Playbook: [[agent-loop]]. Contract: [[feedback-contract]].

## Anchors

Notes may carry `headingId`, `slide`, `level`, `levelKind`, `c4Id` / `c4Label` / `c4From` / `c4To`, and line ranges so agents can find the quote after edits. On C4 diagrams, select or draw on a **box** (`c4Id: reader`) or a **relationship** (`c4Id: reader-peel`, `c4From`/`c4To`, label text). Prefer matching on `c4Id` / Rel endpoints when lines drift.

## Screenshots

![Annotate panel](assets/annotate-panel.png)
