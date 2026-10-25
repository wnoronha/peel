---
marp: true
title: Search in five minutes
---

# Search in five minutes

Crawl the web, index it, serve it. Money via ads.

Audience: anyone new. Time: five minutes.

---

## Map of this talk

```mermaid
mindmap
  root((Search))
    crawl
      frontier
      politeness
      fetch
    index
      postings
      shards
    serve
      fan-out
      rank
    money
      auction
      blend
```

Four stages, one loop. Details: [[overview]].

---

## System context

```mermaid
C4Context
  title Search serving context
  Person(user, "User", "issues queries")
  System(search, "Search", "crawl, index, serve")
  System_Ext(ads, "Ads", "priced results")
  Rel(user, search, "queries")
  Rel(search, ads, "blends")
```

Two systems, one results page. Details: [[overview]].

---

## The loop

```mermaid
flowchart LR
  Web --> C[crawl]
  C --> I[index]
  Q[user query] --> S[serve]
  I --> S
  S --> R[results]
```

- Crawl discovers pages.
- Index makes them retrievable.
- Serve answers queries.

---

## A query's path

```mermaid
sequenceDiagram
  participant U as user
  participant S as serving
  participant I as shards
  U->>S: query
  S->>I: fan-out
  I-->>S: top-k
  S-->>U: results
```

Fan-out, gather, rank, blend. Details: [[serving]].

---

## Where serving time goes

```mermaid
pie title Where serving time goes
  "shard reads" : 45
  "ranking" : 30
  "blending" : 15
  "other" : 10
```

Reads dominate. Cache them. Details: [[caching]].

---

## Serving containers

```mermaid
C4Container
  title Search serving containers
  Container(blend, "Blender", "merges organic and ads")
  ContainerDb(shards, "Shard pool", "postings, fan-out reads")
  Rel(blend, shards, "top-k reads")
```

Blend first, shards second. Details: [[serving]], [[api-design]].

---

## Crawl: discover the web

Frontier → politeness → fetch → render → extract.

- Politeness caps per-host rate.
- Render executes JS before extract.

See [[crawling]].

---

## Index: build once, read billions

Tokens → postings → shards by docid.

- Incremental updates, not rebuilds.
- Skip lists for fast intersections.

See [[indexing]] and [[deck-indexing-deep-dive]].

---

## Serve: decide in milliseconds

Rewrite → fan-out → cascade rank → blend.

- gRPC fan-out, tight deadlines.
- Organic and ads scored apart, merged at blend.

See [[serving]] and [[grpc]].

---

## Money beside organic

```mermaid
flowchart LR
  Q[query] --> CR[candidates]
  CR --> AU[auction]
  AU --> M[merge]
  ORG[organic] --> M
  M --> R[results]
```

Auction beside ranking, merged at blend. See [[ads-targeting]] and [[deck-ads-monetization]].

---

## Index milestones

```mermaid
timeline
  title Index milestones
  1998 : PageRank paper
       : Google founded
  2004 : MapReduce paper
  2006 : Bigtable paper
  2010 : Caffeine goes live
```

Batch, then incremental. See [[2026-10-25-adr-incremental-index-freshness-vs-batch-rebuild-cost-tradeoff]].

---

## Where to go next

| If you want | Read |
|---|---|
| Crawl details | [[crawling]] |
| Index freshness | [[deck-indexing-deep-dive]] |
| Money pipeline | [[deck-ads-monetization]] |
| Vocabulary | [[glossary]] |
