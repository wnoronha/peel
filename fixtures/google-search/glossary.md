# Glossary

Shared vocabulary. Linked from [[overview]], [[crawling]],
[[indexing]], [[ranking]], [[serving]], [[storage]], [[ads-targeting]].

- **Frontier** — the scored queue of URLs awaiting a fetch in [[crawling]].
- **Politeness** — per-host rate limits and backoff in [[crawling]].
- **Postings** — per-term sorted docid lists with positions in [[indexing]].
- **Lexicon** — term → term-id dictionary in [[indexing]].
- **Shard** — one partition of the index or serving fleet ([[indexing]], [[serving]]).
- **Fan-out** — one query fanning to all shards in [[serving]].
- **Cascade** — staged rankers from cheap to heavy in [[ranking]].
- **Hedging** — re-issuing slow reads elsewhere in [[serving]].
- **Quality score** — ad relevance estimate in [[ads-targeting]].
- **Webtable** — the crawled-content table in [[storage]] (Bigtable).
