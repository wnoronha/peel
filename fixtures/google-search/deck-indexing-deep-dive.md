---
marp: true
title: Keeping the index fresh
---

# Keeping the index fresh

From batch rebuilds to Caffeine.

Design deep dive: structures, flows, costs.

---

## Old world: batch

Snapshot the web, MapReduce everything, swap the index.

- Simple. Consistent. Stale.
- Every rebuild reprocesses unchanged pages.

---

## New world: incremental

```mermaid
flowchart LR
  F[fresh fetch] --> D[diff vs stored]
  D --> M[merge into shard]
  M --> RT[live in serving]
```

Changed pages flow from [[crawling]] into [[indexing]] within minutes.

See [[2026-10-25-adr-incremental-index-freshness-vs-batch-rebuild-cost-tradeoff]].

---

## Structures that allow it

```mermaid
classDiagram
  class Document {
    +int docid
    +string url
    +tokenize()
  }
  class Posting {
    +int termid
    +int[] docids
  }
  class Shard {
    +int range
    +merge()
  }
  Document "1" --> "*" Posting : indexed-by
  Posting "*" --> "1" Shard : lives-in
```

Lexicon + postings + skip lists, sharded by docid. See [[indexing]].

---

## A document's lifecycle

```mermaid
stateDiagram-v2
  [*] --> Queued
  Queued --> Fetching : politeness slot
  Fetching --> Rendering : bytes ok
  Rendering --> Indexed : extract
  Fetching --> Failed : timeout
  Failed --> Queued : backoff retry
```

Every state is observable. Failures requeue, never drop. See [[crawling]].

---

## Entities involved

```mermaid
erDiagram
  URL ||--o{ FETCH : "has"
  FETCH ||--|| PAGE : "yields"
  PAGE ||--o{ TERM : "contains"
  TERM ||--o{ POSTING : "indexed-by"
```

Crawl facts feed index facts. See [[storage]].

---

## The freshness contract

```mermaid
requirementDiagram
  requirement freshness_req {
    id: 1
    text: changed pages visible within minutes
    risk: high
    verifymethod: test
  }
  element caffeine {
    type: system
  }
  caffeine - satisfies -> freshness_req
```

Assumption: exact SLA not public. Tracked in [[decisions]] D1.

---

## Latency vs scale

```mermaid
xychart-beta
  title "P99 latency by shard count"
  x-axis [4, 8, 16, 32]
  y-axis "ms" 0 --> 400
  bar [120, 180, 260, 350]
  line [120, 180, 260, 350]
```

More shards, more fan-out tax. Numbers illustrative. See [[capacity-planning]].

---

## Crawl budget week

```mermaid
gantt
  title Crawl budget week
  dateFormat YYYY-MM-DD
  section discovery
  sitemaps      :done, 2026-10-19, 2d
  frontier sweep:active, 2026-10-21, 3d
  section fetch
  recrawl hot hosts: 2026-10-22, 2d
```

Hot hosts first, archives when cheap. See [[crawling]].

---

## What to crawl first

```mermaid
quadrantChart
  title What to crawl first
  x-axis cheap --> pricey
  y-axis stale --> fresh
  quadrant-1 reschedule
  quadrant-2 crawl now
  quadrant-3 ignore
  quadrant-4 sample
  sitemaps: [0.2, 0.8]
  forums: [0.7, 0.6]
  archives: [0.2, 0.2]
```

Freshness per cost, not per URL. See [[bottlenecks]].

---

## Cutover tracking

```mermaid
kanban
  Todo
    [Backfill anchors]
  Doing
    [Shard rebalance]
  Done
    [Caffeine cutover]
```

Migration state at a glance. History: [[outages]].

---

## Price paid

- Write amplification on every merge.
- Weaker consistency across shards.
- Harder rollback than a snapshot swap.

Worth it: freshness wins. See [[bottlenecks]] and [[decisions]].
