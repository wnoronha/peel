# Bottlenecks

Where the system strains. Each links to its owner: [[crawling]],
[[indexing]], [[serving]], [[storage]], [[ranking]], [[ads-targeting]].

## Freshness vs cost

The web changes constantly; every recrawl and reindex burns fetch,
render, CPU and [[storage]] write budget. Frontier scoring decides who
pays — see [[decisions]] D1.

## Tail latency

99th-percentile shard reads dominate user-visible latency in [[serving]]:
deadlines, hedging, caches. Rare queries miss every cache by definition.

## Index size vs recall

Postings for the whole web don't fit in RAM everywhere; compression,
tiering (hot postings resident, cold on SSD) and pruning trade recall
for footprint in [[indexing]].

## Render fleet

JS-heavy pages force [[crawling]] through headless Chromium — orders of
magnitude pricier per byte than raw fetch. Mobile-first makes it
mandatory (see [[decisions]] D4).

## Auction pressure

[[ads-targeting]] must score and price candidates inside the same
[[serving]] deadline; quality models compete with organic [[ranking]]
for the same millisecond budget.
