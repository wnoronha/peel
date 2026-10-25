---
marp: true
title: Queries become revenue
---

# Queries become revenue

Parallel pipeline, merged at blend.

Scrutiny flows: every money step is inspectable.

---

## Parallel pipeline

```mermaid
flowchart LR
  Q[query] --> CR[candidates]
  CR --> QS[quality x bid]
  QS --> AU[auction]
  AU --> M[merge]
  ORG[organic] --> M
```

Organic [[ranking]] iterates independently of money. See [[ads-targeting]].

---

## Auction scrutiny

```mermaid
sequenceDiagram
  participant Q as query
  participant C as candidates
  participant A as auction
  participant B as blend
  Q->>C: eligible ads
  C->>A: quality x bid
  A-->>B: priced winners
  B-->>Q: merged results
```

Each handoff is logged. Audit any impression. See [[decisions]] D3.

---

## A query's journey

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

Blend scores lowest today. That is the scrutiny backlog. See [[serving]].

---

## Auction state

```mermaid
stateDiagram-v2
  [*] --> Eligible
  Eligible --> Priced : quality x bid
  Priced --> Won : auction clears
  Priced --> Lost : outbid
  Won --> Served : merged at blend
  Lost --> Eligible : next query
```

Losers return to eligibility. No sticky losers. See [[database-reliability]].

---

## Entities behind money

```mermaid
erDiagram
  CAMPAIGN ||--o{ ADGROUP : "has"
  ADGROUP ||--o{ CREATIVE : "has"
  CREATIVE ||--o{ BID : "places"
  BID ||--|| AUCTION : "clears-in"
```

Campaigns own groups own creatives. Bids clear per query. See [[storage]].

---

## Money flow

```mermaid
sankey-beta
query,auction,100
auction,served,80
auction,lost,20
served,revenue,80
```

Four of five queries monetize in this sketch. Real split: [[ads-targeting]].

---

## Quality signals

```mermaid
radar-beta
title Signal weights
axis r["relevance"]
axis b["bid"]
axis f["freshness"]
curve o["organic"]{80, 20, 70}
curve a["ads"]{60, 90, 50}
max 100
```

Organic leans relevance, ads lean bid. Blend weights both. See [[ranking]].

---

## Policy change history

```mermaid
gitGraph
  commit id: "reserve price"
  branch quality-floor
  checkout quality-floor
  commit id: "min relevance"
  checkout main
  merge quality-floor
  commit id: "serve it"
```

Pricing changes ship like code. Review trail: [[decisions]].

---

## Request layout

```mermaid
packet-beta
title Auction request
0-15: "query"
16-31: "budget"
32-47: "floor"
```

Fixed offsets, parsed once. Contract: [[api-design]].

---

## Under the same deadline

Scoring and pricing inside the [[serving]] millisecond budget.

- Auction runs beside ranking, never before it.
- Over budget: serve organic, drop priced ads.

---

## State in Spanner

Transactional auction state in [[storage]].

- Exactly-once impression logging.
- Replays for billing disputes.

---

## Open questions

- Should blend weights vary by query class?
- Reserve price per vertical or global?
- Who owns the scrutiny dashboard?

Parked in [[decisions]] D3. Challenge them.
