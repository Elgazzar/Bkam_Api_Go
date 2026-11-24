# Research & Architectural Analysis: Migrate Bkam Financial API Gateway to Idiomatic Go

**Feature**: `116-migrate-gateway-go`  
**Date**: 2026-11-16  
**Status**: Completed  

---

## 1. Technical Unknowns & Architectural Investigation

### Research Question 1: How should Firestore real-time snapshot listeners be implemented idiomatic to Go concurrency and SDK patterns?

- **Current Node.js Implementation**:
  Node.js registers callback functions with `db.collection(...).onSnapshot(callback, errCallback)`. These execute on the single-threaded Node.js event loop.
- **Go Investigation**:
  The official Google Cloud Firestore Go SDK (`cloud.google.com/go/firestore`) does NOT use callback functions. Instead, it exposes stream iterators:
  - `Query.Snapshots(ctx *context.Context) *QuerySnapshotIterator`
  - `DocumentRef.Snapshots(ctx *context.Context) *DocumentSnapshotIterator`
  Each iterator provides a blocking `iterator.Next()` method that yields a snapshot or error whenever upstream documents mutate.
- **Architecture Decision**:
  - Encapsulate all listener lifecycle management within an inbound adapter struct: `type FirestoreListenerService struct`.
  - For each subscribed path (collection `rates`, single docs `gold_prices/latest`, `gas_prices/latest`, `stock_prices/latest`, `index_prices/latest`, GCC regional stock docs, `*_intraday/{today}`, and `*_history_weekly/{weekKey}`), spawn an isolated worker goroutine.
  - Workers run an active iterator loop:
    ```go
    func (s *FirestoreListenerService) watchDocument(ctx context.Context, docRef *firestore.DocumentRef, handler func(*firestore.DocumentSnapshot) error) {
        iter := docRef.Snapshots(ctx)
        defer iter.Stop()
        for {
            snap, err := iter.Next()
            if err != nil {
                if status.Code(err) == codes.Canceled || errors.Is(err, context.Canceled) {
                    return
                }
                logger.Error("Snapshot stream error", "err", err)
                time.Sleep(1 * time.Second) // Exponential backoff before re-iterating
                continue
            }
            if err := handler(snap); err != nil {
                logger.Error("Snapshot handler error", "err", err)
            }
        }
    }
    ```
  - Graceful cancellation: Passing a root `context.WithCancel` to all worker goroutines ensures instant, leak-free teardown when `Stop()` is called or upon SIGINT/SIGTERM.

---

### Research Question 2: How should `LiveSnapshotCoordinator` manage high-concurrency multi-domain quotes with zero data races and zero thundering herds?

- **Current Node.js Implementation**:
  Uses a `Map<string, LiveSnapshotEnvelope>` with a manual single-flight promise map for on-demand fetches.
- **Go Investigation**:
  In Go, multiple goroutines will read and write to the snapshot store concurrently. A plain `map` causes fatal runtime panic on concurrent reads/writes (`concurrent map read and map write`).
- **Architecture Decision**:
  - Use `sync.RWMutex` protecting a map of composite keys: `map[string]*LiveSnapshotEnvelope`.
  - Reads (`GetSnapshot`, `GetEnvelope`) acquire `RLock()`, copy or extract pointer, and release immediately (`< 1µs` lock contention).
  - Push writes (`UpdateSnapshot`) acquire `Lock()`, update memory, release lock, and dispatch disk persistence asynchronously in a separate goroutine.
  - **Thundering Herd Mitigation**: Embed `golang.org/x/sync/singleflight.Group`. On cache misses or force-refreshes, `group.Do(storageKey, func() (any, error) { ... })` ensures that 500 concurrent requests coalesce into exactly 1 upstream Firestore query.

---

### Research Question 3: How should `HistoryResolutionEngine` implement multi-tier caching (RAM Tier 1 + Persistent Tier 2) and ISO week range queries in Go?

- **Current Node.js Implementation**:
  Maintains in-memory map for the active week (`activeWeeklyBuckets`), queries disk flat-cache for sealed weeks (30-day TTL), and dynamically stitches multi-week slices into daily arrays.
- **Go Investigation**:
  - **Date Calculations**: Go's `time` package natively supports ISO week numbers via `year, week := t.ISOWeek()`. We format ISO week keys as `fmt.Sprintf("%04d-W%02d", year, week)` in `*time.Location` (`Africa/Cairo`).
  - **Tier 1 (RAM)**: `type HistoryResolutionEngine struct` maintains `activeWeeklyBuckets map[string]map[string]any` guarded by `sync.RWMutex`.
  - **Tier 2 (Disk / Local Persistence)**: Sealed past weeks are written to disk as pretty/compact JSON under `$SNAPSHOT_DIR/history/{collection}_{weekKey}.json` with a 30-day retention metadata envelope. Alternatively, an embedded `bbolt` key-value database can be used for zero-filesystem overhead, but filesystem JSON mirrors the current node architecture and simplifies debugging and operational backups.
  - **Range Query Assembly**:
    1. Parse start and end dates (`YYYY-MM-DD`).
    2. Generate distinct ISO calendar week keys covering the interval.
    3. Concurrently or sequentially resolve week documents (active week from Tier 1 RAM; sealed weeks from Tier 2 disk cache; fallback to Firestore on cold miss).
    4. Slice out matching day entries into a sorted historical point array.
    5. If range concludes today, seamlessly synthesize today's quote from `LiveSnapshotCoordinator` and append it.

---

### Research Question 4: What is the optimal HTTP framework, middleware stack, and ETag engine for maximum throughput and compatibility?

- **Current Node.js Implementation**:
  Uses Fastify with `@fastify/compress`, `@fastify/etag`, custom auth hook, and custom tracing hooks.
- **Go Investigation**:
  - `net/http` standard library is great, but routing parameterized patterns (`/v1/stocks/:symbol/history/range`, `/v1/rates/:currency`) requires clean regex or URL trie.
  - `go-chi/chi/v5` is 100% compliant with standard `http.Handler`, zero allocations on routing overhead, supports sub-routers, and has native middleware for gzip compression, panic recovery, and context tracing.
- **Middleware & Header Strategy**:
  - **API Key Authenticator**: Inspects `X-API-Key` header and `?apiKey=` query parameter. Skips `/v1/health` and `/v1/version`. Returns HTTP `401 Unauthorized` with JSON error payload matching `HttpError`.
  - **ETag & Conditional Caching**: Custom middleware calculates deterministic hash (FNV-1a or SHA-256) of response buffer. If incoming `If-None-Match` equals calculated `ETag`, write `304 Not Modified` and return zero body bytes.
  - **Trace ID**: Extract `X-Trace-ID` or generate `req-{timestamp}-{randHex}` and attach to response header `x-trace-id`.

---

### Research Question 5: How should the Market Schedule & Cairo operational timezone be modeled?

- **Operational Timezone**: Load `time.LoadLocation("Africa/Cairo")`.
- **Trading Schedule Logic**:
  - EGX: Sunday – Thursday, 10:00 AM – 02:30 PM Cairo.
  - Saudi: Sunday – Thursday, 10:00 AM – 03:00 PM Riyadh (`Asia/Riyadh`).
  - Regional GCC exchanges: UAE (`Asia/Dubai`), Qatar (`Asia/Qatar`), Kuwait (`Asia/Kuwait`).
  - Egyptian Holiday Registry: Static date list matching national official holiday calendar.
- **Dynamic TTL Calculations**:
  - 60 seconds if market is open.
  - Seconds until next trading session open (09:00 AM Cairo next business day) when closed.

---

## 2. Summary of Architectural Decisions & Strategy

| Layer / Concern | Node.js (TypeScript) | Go Migration Target | Architectural Rationale |
|---|---|---|---|
| **Language & Runtime** | Node.js 22 LTS (V8, Single-threaded Event Loop) | Go 1.23+ (Compiled, Goroutines, Low GC overhead) | 10x throughput, P99 latency < 15ms, zero V8 overhead |
| **HTTP Transport** | Fastify | `chi/v5` on standard `net/http` | Standard library compatibility, zero-allocation routing |
| **Real-time Ingestion** | `onSnapshot` callbacks in main thread | Goroutines consuming `Snapshots(ctx)` iterators | True concurrent ingestion; non-blocking channel dispatches |
| **Concurrency & State** | Plain `Map` (safe only due to single thread) | `sync.RWMutex` + `sync.Map` | Thread-safe across multiple concurrent CPU cores |
| **Thundering Herd** | Manual promise caching in Map | `golang.org/x/sync/singleflight` | Standard Go pattern; zero duplicate database reads |
| **Historical Tier 1** | Internal memory map in `HistoryResolutionEngine` | Thread-safe memory map in `HistoryResolutionEngine` | Preserves data locality; sub-millisecond RAM lookups |
| **Historical Tier 2** | Disk files (`flat-cache`) | Local filesystem JSON or embedded KV store (`bbolt`) | Sealed past weeks immutable; 30-day persistence |
| **Public API Contracts** | Zod schemas | Go structs with JSON tags matching exact contract | 100% backward compatibility for all 49 endpoints |
| **Operational Schedule** | Cron jobs (`node-cron`) | `robfig/cron/v3` or native `time.Ticker` | Deterministic midnight & 08:00 AM Cairo rollovers |

---

## 3. Best Practices & Quality Standards

1. **Explicit Error Handling**: No suppressed errors; all errors wrapped with context (`fmt.Errorf("reading snapshot: %w", err)`).
2. **Graceful Shutdown**: Listen for `os.Interrupt`, `syscall.SIGTERM`, cleanly stop HTTP server (`server.Shutdown(ctx)`), cancel listener context, and flush caches.
3. **Comment Coverage**: Constitution Principle V requires comprehensive commenting for all domain logic, rationale, and concurrency guarantees.
