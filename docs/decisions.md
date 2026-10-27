# Decisions

Short record of choices that still shape the binary. Links: [[architecture]], [[c4-peel]], ADR [[2026-10-25-architecture-decision-record-agent-feedback-pipeline-review-workflow]].

## Runtime

- **Go CLI** — one static binary; embed the web UI (`go:embed web/*`).
- **Alpine + vanilla fetch** — drawers, composer, SSE, Goto; no SPA framework.
- **Vendored pele** — mermaid-compatible diagrams (including C4) without shipping mermaid.js.
- **Lucide sprites** — chips and status tabs without a CDN.

## Feedback

- **Append-only JSONL** under `<root>/.peel/feedback.jsonl` — humans and agents share one log; only the server writes.
- **Folded views** — open = notes with no later ack/tombstone for that `id`; no PATCH of history.
- **Typed SSE + file fallback** — `/events` emits `reload` (doc changes) and `feedback` (note/ack/tombstone). Agents prefer `skills/peel/peel_api.py`; they may still `tail -F` the jsonl when SSE is unavailable.
- **Single-thread closes** — one note, then one `ack` or tombstone (no reply chat threads).
- **Activity toasts** — ack / dismiss / remote note / current-file reload (not a persistent footer ticker).

## Surfaces

- **Docs / decks / C4** share one reader shell; mode is file-driven (`marp: true`, `.c4`).
- **C4 click vs drill** — click comments (with `c4Id` anchors); double-click drills a level.
- **No Annotate CTA** — capture is select/click + Feedback drawer (`]`).
- **Compact Save** — composer and Feedback card use a short primary button (not a wide banner CTA).
- **Dogfood `docs/`** — product story lives here ([[use-cases]], [[deck-peel-tour]], [[deck-c4]], [[c4-peel]]).
- **Portable agent skill** — canonical under `skills/peel/`; Cursor `/peel` is a wrapper only.
