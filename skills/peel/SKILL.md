---
name: peel
description: >-
  Peel human↔agent pairing over feedback.jsonl (SSE or poll). Invoke with
  /peel to catch up on open notes and keep monitoring until the human stops.
---

# Peel agent skill

Portable skill for [peel](https://github.com/wnoronha/peel). Humans annotate in the browser (Feedback panel); agents pair, edit docs, then **ack** via API. Never hand-edit `.peel/feedback.jsonl`.

Default base URL: `http://127.0.0.1:18124` (or the host/port peel printed). Docs root = peel `--dir`.

| Path | Role |
|------|------|
| `.peel/feedback.jsonl` | Append-only notes / acks / tombstones |

## Shell API (mandatory)

**All Peel HTTP/SSE goes through** [`peel_api.py`](peel_api.py) (stdlib Python 3).

```sh
chmod +x skills/peel/peel_api.py   # once
export PEEL_URL=http://127.0.0.1:18124   # optional
```

**Do not** write ad-hoc `curl`, extra Python SSE clients, or one-off poll loops. If a verb is missing, **extend `peel_api.py`** — do not invent a parallel client.

### One shell approval

Under per-command approval, always invoke **this same executable path** so the human can allow it once (e.g. “Always allow” for `skills/peel/peel_api.py`). Prefer a **single long-lived** process:

```sh
skills/peel/peel_api.py pair --timeout 600
```

That does health + open catch-up, then blocks and prints each new root note as JSON. For discrete steps in the same session, still call only `peel_api.py …` (never raw curl).

```sh
skills/peel/peel_api.py health
skills/peel/peel_api.py open
skills/peel/peel_api.py wait-note --timeout 600
skills/peel/peel_api.py wait-note --since-id <id> --timeout 600
skills/peel/peel_api.py busy "editing architecture" <note-id>   # pulse logo
skills/peel/peel_api.py ack <note-id> "what changed"           # clears pulse
skills/peel/peel_api.py idle                                   # clear pulse without ack
```

Source edits use normal file tools; only Peel HTTP uses this script. Call **`busy`** when you start working a note so the browser logo pulses.

## Pairing mode

When asked to pair / monitor / run this skill:

1. `peel_api.py health` → `ok`. If not, tell them to run `./peel --dir <root>` (or `make run`).
2. Catch-up: `peel_api.py open`. List open notes briefly. Do **not** mass-ack old smoke/test notes unless the user points at them.
3. Tell the user you are paired, then **block** on `pair` or `wait-note` (await the process — do not background and walk away).
4. When stdout prints a note JSON: `busy "…" <id>`, edit → `ack`, tell the user what you did, then wait again (`wait-note --since-id <id>` or continue `pair`).
5. Exit code **124** = timeout — say you are still paired and wait again (or ask).
6. If SSE fails, the script falls back to poll automatically (`--poll` to force).

Do **not** push or force-push unless the user asks.

## Handle a note

1. Open root = `note` with empty `ref` and no later ack/tombstone for that `id`.
2. Match with `quote` first; then `file` + anchors (`headingId`, `slide`, `level`/`levelKind`, `c4Id`/`c4From`/`c4To`) + `startLine`.
3. Honor `action`: `fix` | `explain` | `example` | `simplify` | `question`.
4. Edit sources under the docs root (not `.peel/`).
5. `peel_api.py busy "short status" <note-id>` (logo pulses).
6. Ack: `peel_api.py ack <note-id> "what changed"` (clears pulse).
7. Resume waiting.

## UX (what the human sees)

- Note appears under Feedback → **Open** (peach highlight on the quote).
- Your ack moves it to **Acked**; optional ack text shows as `ack: …` on the card.
- No chat thread — one note, one close.

## Do not

- Hand-edit or truncate `feedback.jsonl`.
- Mass-clear the open queue without being asked.
- Assume `startLine` survives large edits.
- Commit `.peel/` or push unless asked.
- Call `curl` / custom SSE for Peel — use `peel_api.py` only.

## API map

`GET|POST /api/feedback` · `DELETE /api/feedback?id=` · `GET|POST /api/agent` · `GET /events` · `GET /healthz`

Project docs: `docs/agent-loop.md`, `docs/feedback-contract.md`, `docs/api.md`.
