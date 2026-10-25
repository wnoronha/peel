# Annotate a doc

## Capture a quote

Drag-select any text in the center column, including labels inside rendered diagrams. A composer appears by the selection. If you click elsewhere first, the selection is kept and the sidebar form holds the quote instead.

## Pick a quick action

The form offers five presets: **Explain this**, **Fix this**, **Add example**, **Simplify this**, **Question**. Only icons show until you hover a chip. Clicking one prefixes your note, for example `Add example: `. Add detail after the colon, then click **Save**. Saving needs both a quote and a note.

## Review the list

Click **Feedback** in the top bar to open the right panel across all files. The **This file** tab filters to the open doc. Each entry shows its age, file, note, quote, and source lines. Click an entry to jump to its quoted text. Click `×` to delete it.

## Feed the agent loop

Every save appends one JSON object per line to `<root>/.peel/feedback.jsonl` with `id`, `ts`, `file`, `quote`, `startLine`, `endLine`, and `note`. Your agent or IDE watches that file. The exact contract is in `feedback-contract.md`.

![Annotate panel](assets/annotate-panel.png)
