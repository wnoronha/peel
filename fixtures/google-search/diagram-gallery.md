# Diagram Gallery

Every renderer path in one place, themed to [[overview]].
If a block below shows an error instead of a picture, the syntax needs
fixing — that is the point of this file.

## Flowchart

```mermaid
flowchart LR
  Web --> C[crawl]
  C --> I[index]
  Q((query)) --> S[serve]
  I --> S
```

## Sequence

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

## Class

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
  Document "1" --> "*" Posting : indexed-by
```

## State

```mermaid
stateDiagram-v2
  [*] --> Queued
  Queued --> Fetching : politeness slot
  Fetching --> Rendering : bytes ok
  Rendering --> Indexed : extract
  Fetching --> Failed : timeout
  Failed --> Queued : backoff retry
```

## Entity relationship

```mermaid
erDiagram
  URL ||--o{ FETCH : "has"
  FETCH ||--|| PAGE : "yields"
  PAGE ||--o{ TERM : "contains"
```

## Gantt

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

## Git graph

```mermaid
gitGraph
  commit id: "crawl baseline"
  branch freshness
  checkout freshness
  commit id: "percolator merge"
  checkout main
  merge freshness
  commit id: "serve it"
```

## Mindmap

```mermaid
mindmap
  root((Search))
    crawl
      frontier
      fetch
    index
      postings
      shards
    serve
      fan-out
      rank
```

## Pie

```mermaid
pie title Where serving time goes
  "shard reads" : 45
  "ranking" : 30
  "blending" : 15
  "other" : 10
```

## Timeline

```mermaid
timeline
  title Index milestones
  1998 : PageRank paper
       : Google founded
  2004 : MapReduce paper
  2006 : Bigtable paper
  2010 : Caffeine goes live
```

## Journey

```mermaid
journey
  title A query's journey
  section retrieval
    rewrite: 4: serving
    fan-out: 3: serving
  section ranking
    cascade: 5: ranker
    blend ads: 2: mixer
```

## Kanban

```mermaid
kanban
  Todo
    [Backfill anchors]
  Doing
    [Shard rebalance]
  Done
    [Caffeine cutover]
```

## Quadrant

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

## Requirements

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

## XY chart

```mermaid
xychart-beta
  title "P99 latency by shard count"
  x-axis [4, 8, 16, 32]
  y-axis "ms" 0 --> 400
  bar [120, 180, 260, 350]
  line [120, 180, 260, 350]
```

## C4

```mermaid
C4Context
  title Search serving context
  Person(user, "User", "issues queries")
  System(search, "Search", "crawl, index, serve")
  Rel(user, search, "queries")
```
