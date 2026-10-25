# Serving

Answers queries in ~hundreds of milliseconds. Reads [[indexing]] shards,
invokes [[ranking]], blends [[ads-targeting]]. Parent: [[overview]].

```mermaid
flowchart LR
  U[user] --> LB[front-end + cache]
  LB --> QR[query understanding]
  QR --> FO[fan-out to index shards]
  FO --> S1[shard 1]
  FO --> S2[shard 2]
  FO --> SN[shard N]
  S1 & S2 & SN --> AG[aggregator]
  AG --> RK[ranking cascade]
  RK --> AD[ads merge]
  AD --> OUT[response]
```

## Latency discipline

- **Deadlines everywhere**: every fan-out has a budget; slow shards get
  dropped (slightly worse results beat slow results).
- **Hedged requests** (Dean–Barroso "tail at scale"): re-issue slow
  shard reads elsewhere after a percentile wait.
- **Layered caches**: results cache, postings cache, doc-store cache
  in front of [[storage]].

Failure handling: replica shards, graceful degradation (fewer shards,
cheaper ranker) before errors. More in [[bottlenecks]].
