#!/usr/bin/env python3
"""Peel HTTP helpers — one script, many subcommands.

Agents under strict shell approval: use THIS script only for Peel HTTP/SSE.
Do not invent ad-hoc curls or parallel clients.

  PEEL_URL=http://127.0.0.1:18124 ./skills/peel/peel_api.py health
  ./skills/peel/peel_api.py open
  ./skills/peel/peel_api.py wait-note --timeout 600
  ./skills/peel/peel_api.py ack <note-id> "what changed"
  ./skills/peel/peel_api.py pair --timeout 600
"""
from __future__ import annotations

import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from typing import Any

DEFAULT_URL = "http://127.0.0.1:18124"


def base_url() -> str:
    return os.environ.get("PEEL_URL", DEFAULT_URL).rstrip("/")


def http_json(method: str, path: str, body: dict | None = None, timeout: float = 30) -> Any:
    url = base_url() + path
    data = None
    headers = {"Accept": "application/json"}
    if body is not None:
        data = json.dumps(body).encode()
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            raw = resp.read().decode()
            if not raw.strip():
                return None
            return json.loads(raw)
    except urllib.error.HTTPError as e:
        err = e.read().decode(errors="replace")
        print(f"HTTP {e.code}: {err}", file=sys.stderr)
        raise SystemExit(1) from e
    except urllib.error.URLError as e:
        print(f"connect failed ({base_url()}): {e}", file=sys.stderr)
        raise SystemExit(1) from e


def http_text(path: str, timeout: float = 10) -> str:
    url = base_url() + path
    req = urllib.request.Request(url, headers={"Accept": "text/plain"})
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return resp.read().decode()


def cmd_health(_: argparse.Namespace) -> int:
    print(http_text("/healthz").strip())
    return 0


def cmd_feedback(args: argparse.Namespace) -> int:
    q = args.query or ""
    path = "/api/feedback" + (("?" + q) if q else "")
    print(json.dumps(http_json("GET", path), indent=2 if args.pretty else None))
    return 0


def open_notes() -> list[dict]:
    data = http_json("GET", "/api/feedback") or {}
    return list(data.get("feedback") or [])


def cmd_open(_: argparse.Namespace) -> int:
    notes = open_notes()
    print(f"open: {len(notes)}")
    for n in notes:
        quote = (n.get("quote") or "")[:60]
        note = (n.get("note") or "")[:80]
        print(f"{n.get('id')}\t{n.get('file') or '-'}\t{n.get('action') or '-'}\t{quote}\t{note}")
    return 0


def cmd_ack(args: argparse.Namespace) -> int:
    body = {"type": "ack", "ref": args.ref, "note": args.note or "resolved"}
    print(json.dumps(http_json("POST", "/api/feedback", body)))
    return 0


def cmd_note(args: argparse.Namespace) -> int:
    body: dict[str, Any] = {
        "type": "note",
        "file": args.file,
        "quote": args.quote,
        "note": args.note,
    }
    if args.action:
        body["action"] = args.action
    print(json.dumps(http_json("POST", "/api/feedback", body)))
    return 0


def parse_sse_data_line(line: str) -> dict | None:
    if not line.startswith("data:"):
        return None
    payload = line[5:].lstrip()
    if not payload:
        return None
    try:
        return json.loads(payload)
    except json.JSONDecodeError:
        return None


def is_root_note(ev: dict, since_id: str | None, file_filter: str | None) -> bool:
    if (ev.get("type") or "note") != "note":
        return False
    if ev.get("ref"):
        return False
    if since_id and not (str(ev.get("id") or "") > since_id):
        return False
    if file_filter and ev.get("file") != file_filter:
        return False
    return True


def wait_note_sse(timeout: float, since_id: str | None, file_filter: str | None) -> dict:
    """Block on GET /events until a root note arrives (or timeout)."""
    url = base_url() + "/events"
    req = urllib.request.Request(url, headers={"Accept": "text/event-stream"})
    deadline = time.monotonic() + timeout if timeout > 0 else None
    # urlopen timeout is idle read timeout; use short chunks + overall deadline.
    read_timeout = min(20.0, timeout) if timeout > 0 else 30.0
    try:
        with urllib.request.urlopen(req, timeout=read_timeout) as resp:
            while True:
                if deadline is not None and time.monotonic() >= deadline:
                    raise TimeoutError(f"wait-note: timeout after {int(timeout)}s")
                try:
                    line = resp.readline()
                except TimeoutError:
                    if deadline is not None and time.monotonic() >= deadline:
                        raise TimeoutError(f"wait-note: timeout after {int(timeout)}s") from None
                    continue
                if not line:
                    raise RuntimeError("wait-note: stream ended")
                text = line.decode(errors="replace").rstrip("\r\n")
                msg = parse_sse_data_line(text)
                if not msg or msg.get("type") != "feedback":
                    continue
                ev = msg.get("event") or {}
                if not isinstance(ev, dict):
                    continue
                if is_root_note(ev, since_id, file_filter):
                    return ev
    except urllib.error.URLError as e:
        print(f"SSE failed ({url}): {e}", file=sys.stderr)
        raise SystemExit(1) from e


def wait_note_poll(timeout: float, since_id: str | None, file_filter: str | None) -> dict:
    """Fallback: poll folded open set / all=1 for a newer root note."""
    deadline = time.monotonic() + timeout if timeout > 0 else None
    seen = since_id or ""
    while True:
        data = http_json("GET", "/api/feedback?all=1") or {}
        events = list(data.get("feedback") or [])
        for ev in events:
            if is_root_note(ev, seen or None, file_filter) and str(ev.get("id") or "") > seen:
                return ev
        if deadline is not None and time.monotonic() >= deadline:
            raise TimeoutError(f"wait-note: timeout after {int(timeout)}s")
        time.sleep(1.5)


def _wait_note(timeout: float, since_id: str | None, file_filter: str | None, poll: bool) -> dict:
    if poll:
        return wait_note_poll(timeout, since_id, file_filter)
    try:
        return wait_note_sse(timeout, since_id, file_filter)
    except TimeoutError:
        raise
    except Exception as e:
        print(f"SSE unavailable ({e}); falling back to poll", file=sys.stderr)
        return wait_note_poll(timeout, since_id, file_filter)


def cmd_wait_note(args: argparse.Namespace) -> int:
    timeout = float(args.timeout or 0)
    try:
        note = _wait_note(timeout, args.since_id, args.file, args.poll)
    except TimeoutError as e:
        print(str(e), file=sys.stderr)
        return 124
    print(json.dumps(note))
    return 0


def cmd_pair(args: argparse.Namespace) -> int:
    print("== health ==")
    print(http_text("/healthz").strip())
    print("== open notes (catch-up) ==")
    cmd_open(args)
    print("== waiting for root notes (Ctrl-C to stop) ==")
    last = ""
    timeout = float(args.timeout or 0)
    while True:
        try:
            note = _wait_note(timeout, last or None, None, args.poll)
        except TimeoutError as e:
            print(str(e), file=sys.stderr)
            return 124
        print(json.dumps(note), flush=True)
        last = str(note.get("id") or last)


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(
        prog="peel_api.py",
        description="Peel HTTP/SSE helpers (set PEEL_URL). Prefer this over ad-hoc curl.",
    )
    sub = p.add_subparsers(dest="cmd", required=True)

    sub.add_parser("health", help="GET /healthz").set_defaults(func=cmd_health)

    f = sub.add_parser("feedback", help="GET /api/feedback")
    f.add_argument("query", nargs="?", default="", help="query string e.g. all=1 or file=a.md")
    f.add_argument("--pretty", action="store_true")
    f.set_defaults(func=cmd_feedback)

    sub.add_parser("open", help="summarize open root notes").set_defaults(func=cmd_open)

    a = sub.add_parser("ack", help="POST ack for a note id")
    a.add_argument("ref", help="note id")
    a.add_argument("note", nargs="?", default="resolved", help="what changed")
    a.set_defaults(func=cmd_ack)

    n = sub.add_parser("note", help="POST a note (testing)")
    n.add_argument("file")
    n.add_argument("quote")
    n.add_argument("note")
    n.add_argument("action", nargs="?", default="")
    n.set_defaults(func=cmd_note)

    w = sub.add_parser("wait-note", help="block until next root note (SSE, poll fallback)")
    w.add_argument("--timeout", type=float, default=0, help="seconds; 0 = wait forever")
    w.add_argument("--since-id", default=None, help="only notes with id > this")
    w.add_argument("--file", default=None, help="only this file path")
    w.add_argument("--poll", action="store_true", help="skip SSE; poll API")
    w.set_defaults(func=cmd_wait_note)

    pr = sub.add_parser("pair", help="health + open + loop wait-note (one process)")
    pr.add_argument("--timeout", type=float, default=0, help="per-wait timeout; 0 = forever")
    pr.add_argument("--poll", action="store_true")
    pr.set_defaults(func=cmd_pair)

    return p


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    return int(args.func(args) or 0)


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except KeyboardInterrupt:
        print("interrupted", file=sys.stderr)
        raise SystemExit(130) from None
