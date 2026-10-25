# API Design

Service boundaries that survive breakage. Consumed by [[serving]] and
[[ads-targeting]]; transported over [[grpc]]; see [[outages]] for what
happens when contracts lie.

```mermaid
flowchart LR
  C[client] -->|deadline + idempotency key| E[edge: auth, quota]
  E --> R[router]
  R --> S[service impl]
  S --> D[(storage)]
```

## Rules that prevent pages

- **Every RPC has a deadline** — inherited, never invented at the leaf.
  Missing deadlines turn one slow dependency into a [[outages]] retry storm.
- **Idempotency keys** on all mutations so retries are safe.
- **Backward-compatible fields only**: never rename, never reuse numbers,
  deprecate with sunset headers. A breaking field change is a coordinated
  multi-service outage.
- **Version in the envelope**, not the URL path — one routing fabric.
- **Errors are typed** (rate-limited vs unavailable vs invalid) so callers
  in [[serving]] can decide: shed, degrade, or fail.

Schema evolution and quota policy are capacity questions too:
[[capacity-planning]].
