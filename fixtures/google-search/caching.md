# Caching

The layer that hides [[storage]] and [[serving]] latency — until it
doesn't. Outage shapes: [[outages]].

```mermaid
flowchart LR
  Q[query] --> L1[L1: in-process]
  L1 -->|miss| L2[L2: distributed cache]
  L2 -->|miss| DB[(storage)]
  DB --> L2
  L2 --> L1
```

## How caches cause outages

- **Stampede** — hot key expires; N hosts recompute simultaneously.
  Mitigations: request coalescing (single-flight), probabilistic early
  refresh, stale-while-revalidate.
- **Poisoning** — one bad write cached everywhere; version keys or
  short TTLs on derived data.
- **Inconsistency windows** — index updates in [[indexing]] race cached
  postings; acceptable for search, fatal for [[ads-targeting]] pricing.
- **Capacity cliff** — cache hit ratio drops 2% and [[storage]] sees
  10x; size for the trough, alarm on the ratio, shed on the queue
  ([[capacity-planning]]).

Rule: caches are a performance feature with availability consequences.
Every cache needs an explicit miss-path budget.
