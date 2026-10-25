# Outages

How big systems break. Each pattern links to the layer that owns it:
[[serving]], [[crawling]], [[indexing]], [[storage]], [[grpc]],
[[caching]], [[api-design]], [[database-reliability]].

```mermaid
flowchart LR
  T[traffic spike] --> O[overload]
  O --> R[retry storm]
  R --> C[cascading failure]
  C --> G[global outage]
  G --- M[mitigations: shedding, backoff, bulkheads]
```

## Classic shapes

1. **Retry storm** — clients retry timeouts; retries *are* the outage.
   Fixed by budgets, backoff + jitter, circuit breakers ([[grpc]]).
2. **Cache stampede** — a hot key expires; every host recomputes at once
   ([[caching]]). Request coalescing or probabilistic early refresh.
3. **Hotspotting** — one [[storage]] tablet / shard takes all traffic;
   range-partitioned keys are the usual culprit ([[database-reliability]]).
4. **Slow drain** — [[crawling]] fetchers hang on dead hosts without
   deadlines; threads pile up until the pool is gone.
5. **Config push** — a bad flag rolls out faster than the rollback path.

Rule of thumb: the failure you plan for is never the one you get —
so every layer needs independent shedding. See [[capacity-planning]].
