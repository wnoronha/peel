# Ranking

Orders candidates retrieved from [[indexing]]. Runs inside [[serving]]
under a strict deadline; see [[bottlenecks]] for tail-latency pressure.

## Cascade

```mermaid
flowchart TB
  Q[parsed query] --> R1[retrieval: top-k boolean + WAND]
  R1 --> R2[lightweight ranker: hundreds of features]
  R2 --> R3[heavy ML ranker: deep models]
  R3 --> BL[blending: freshness, locale, verticals]
  BL --> OUT[ordered results]
```

Public milestones: PageRank (1998), Hummingbird (2013, conversational
query understanding), RankBrain (2015), neural matching with BERT (2019),
MUM (2021, multimodal). Exact current signals are not public
[assumption: cascade shape inferred from eng talks and papers].

## Query path

```mermaid
sequenceDiagram
  participant U as user
  participant S as serving
  participant I as indexing shards
  participant R as ranker
  U->>S: query string
  S->>S: rewrite, spell-check, expand
  S->>I: postings lookup (fan-out)
  I-->>S: top-k candidates
  S->>R: cascade scoring
  R-->>S: ordered docids
  S-->>U: results + ads
```

Spelling and synonyms expand recall before retrieval; the auction in
[[ads-targeting]] runs in parallel and merges at blend time.
