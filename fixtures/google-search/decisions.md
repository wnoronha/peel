# Decisions

Architecture decision records. Context: [[overview]]; details per area
in [[crawling]], [[indexing]], [[serving]], [[ads-targeting]].

## D1 — Incremental over batch indexing

Caffeine/Percolator replaced periodic full rebuilds. Wins freshness;
costs continuous write load on [[storage]]. Deep dive:
[[2026-10-25-adr-incremental-index-freshness-vs-batch-rebuild-cost-tradeoff]].

## D2 — Deadlines beat completeness

[[serving]] drops slow shards rather than waiting. Rationale: users
prefer fast good answers to slow perfect ones (tail-at-scale).
Counter-pressure: rare-term recall; see [[bottlenecks]].

## D3 — Separate the ads path

[[ads-targeting]] runs as a parallel pipeline merged at blend time, not
interleaved with organic [[ranking]]. Keeps auction iteration independent
and protects result trust.

## D4 — Mobile-first crawl canonical

[[crawling]] renders with the mobile viewport as ground truth, reflecting
where the queries come from. Cost: heavier render fleet (headless
Chromium at web scale).
