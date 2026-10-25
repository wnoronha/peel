# Storage

Durable substrate under [[crawling]], [[indexing]] and [[serving]].
Public lineage: GFS (2003) → Colossus; MapReduce (2004);
Bigtable (2006); Spanner (2012). Compute runs on Borg; RPC via
Stubby/gRPC [sane guess: names current, internals evolved].

```mermaid
flowchart TB
  subgraph Compute
    B[Borg jobs]
  end
  subgraph Files
    C[Colossus]
  end
  subgraph Tables
    BT[Bigtable: webtable, anchors]
    SP[Spanner: metadata, ads state]
  end
  B --> C
  B --> BT
  B --> SP
  BT --> IX[indexing]
  SP --> AD[ads-targeting]
```

## Who stores what

| Data | Home | Access pattern |
| --- | --- | --- |
| Raw fetches | Colossus | Write-once, scan |
| Webtable + anchors | Bigtable | Row reads, bulk scans |
| Index shards | Colossus + local SSD | Bulk load, mmap reads |
| Ads/auction state | Spanner | Transactional |

Locality matters: index shards are pushed near [[serving]] hosts;
rebuilds stream from Colossus at datacenter-network rates.
