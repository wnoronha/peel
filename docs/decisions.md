# Decisions

Short record of choices that still shape the binary. Links: [[architecture]].

## Runtime

- **Go CLI** — one static binary; embed the web UI.
- **Alpine + vanilla fetch** — enough for drawers, SSE reload, and annotate; no SPA framework.
- **Vendored pele** — mermaid-compatible diagrams, including C4, without pulling mermaid.js.
- **Lucide sprites** — icon chips and status tabs without a CDN.

## Feedback

- **Append-only JSONL** under `<root>/.peel/feedback.jsonl` — agents and humans share one log.
- **Folded views** — open = notes minus ack minus tombstone; no PATCH of history.
- **No SSE for feedback** — agents poll or watch the file; SSE is for doc reload only.

## Surfaces

- **Docs / decks / C4** share one reader shell; mode is file-driven (`marp: true`, `.c4`).
- **Dogfood `docs/`** — product story lives here ([[use-cases]], [[deck-peel-tour]], [[c4-peel]]).
