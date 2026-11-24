# Bkam API Gateway - Domain Glossary

| Term | Definition |
|------|------------|
| **BankRate** | Buy and sell exchange rate record for a specific currency pair at an Egyptian bank. |
| **BenchmarkIndexDefinition** | Registered benchmark index entity (e.g. EGX30, TASI, FADX15) with market affiliation, currency, calculation basis, and pre-indexed constituent member sets. |
| **Catalog** | Discoverable inventory of supported banks, currencies, commodities, and regional stock markets. |
| **ConstituentReverseIndex** | In-memory inverted hash map mapping benchmark index codes directly to their constituent member equities in O(1) time, eliminating live runtime scanning loops. |
| **FlatCache** | Disk-persisted key-value store (`flat-cache`) holding immutable sealed weekly ranges for 30 days. |
| **IntradayTimeline** | Minute-by-minute or event-based tick series recorded during an active daily market session. |
| **LiveSnapshotCoordinator** | Deep centralized in-memory storage, warming, and push ingestion coordinator encapsulating all 6 asset domains (`rates`, `gold`, `gas`, `stocks`, `indices`, `gcc-stocks`) behind an expressive 3-method interface (`getSnapshot`, `updateSnapshot`, `fetchFresh`). |
| **LiveSnapshotEnvelope** | Standardized metadata envelope wrapping cached live market quotes with domain, market, currency, timestamps, and freshness classifications (`current`, `stale`, `degraded`). |
| **MarketCatalog** | Deep in-memory domain inventory encapsulating 1,180+ multi-asset platform securities, benchmark indices, banks, currencies, and commodities behind O(1) multi-key lookup, alias resolution, constituent indexing, quote enrichment, and parameter validation. |
| **PushListener** | Real-time WebSocket connection to Firestore (`onSnapshot`) updating RAM caches with 0 query cost. |
| **SecurityDefinition** | Normalized domain metadata model for tradable securities (symbol, ISIN/code, numericCode, TradingView ticker, sector, country, exchange, currency, benchmark indices, groups, logo). |
| **SnapshotCache** | Primary L1 in-memory data store providing sub-10ms response delivery to incoming API calls. |
| **ZeroReadDelivery** | Architectural guarantee that client requests are served from memory/disk without Firestore billing charges. |

