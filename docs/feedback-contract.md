# Feedback contract spec

Agents read `<root>/.peel/feedback.jsonl`. One JSON object per line, appended on every save, never edited in place. Delete rewrites the file without the matching line.

## Entry

```json
{ "id": "20261025T210527.342062449", "ts": "2026-10-25T21:05:27Z",
  "file": "overview.md", "quote": "How web search works",
  "startLine": 3, "endLine": 3, "note": "Fix this: stale number" }
```

`id` is the server timestamp at write time and is unique per entry. `ts` is RFC 3339. `file` is the path relative to the docs root. `quote` is the selected text verbatim, up to 2000 characters. `startLine` is the 1-based source line where the quote starts, `endLine` where it ends; both are `0` when the quote cannot be mapped. `note` is the human text, usually prefixed with a quick action such as `Fix this: `.

## Reading

Watch the file for appends or poll `GET /api/feedback` with no `file` param for every entry across files. Entries for a deleted file stay in the log. Treat `file` plus `startLine` as a hint, and match on `quote` text, because edits shift lines.

## Writing

Only the server writes this file, through `POST /api/feedback`. Agents consume it and must not append by hand, so ids and timestamps stay server ordered.
