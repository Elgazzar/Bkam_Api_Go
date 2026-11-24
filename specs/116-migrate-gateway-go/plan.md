# Implementation Plan: Migrate Bkam Financial API Gateway to Idiomatic Go

**Branch**: `116-migrate-gateway-go` | **Date**: 2026-11-16 | **Spec**: [specs/116-migrate-gateway-go/spec.md](spec.md)

**Input**: Feature specification from `specs/116-migrate-gateway-go/spec.md`

---

## Summary

Port the Bkam Financial API Gateway from Node.js (TypeScript/Fastify) to idiomatic Go 1.23+ using standard concurrency patterns, official Google Cloud Firestore Go SDK real-time streaming iterators (`Snapshots(ctx)`), and a Deep Module architecture. The service maintains 100% contract, status code, and HTTP header parity across all 49 platform endpoints while delivering sub-millisecond memory lookups and reducing P99 latency below 15ms.

---

## Technical Context

**Language/Version**: Go 1.23+ (Compiled, ARM64/AMD64 target)  
**Primary Dependencies**:
- `cloud.google.com/go/firestore`: Official Google Cloud Firestore client SDK.
- `github.com/go-chi/chi/v5`: Lightweight, idiomatic, zero-allocation HTTP router.
- `golang.org/x/sync/singleflight`: Thundering herd suppression for concurrent queries.
- `github.com/robfig/cron/v3`: Timezone-aware cron scheduler for Cairo midnight and 08:00 AM rollovers.  
**Storage**: Local persistent filesystem directory (`SNAPSHOT_DIR`) storing JSON snapshot envelopes and 30-day sealed historical weeks.  
**Testing**: `go test -v -race -cover ./...` alongside existing verification suite `npm run verify` (49 scenarios) and `npm run verify:report` (80 tests).  
**Target Platform**: Linux ARM64 (OCI Ampere A1 Compute) and local development (Windows/macOS).  
**Project Type**: RESTful API Gateway web service.  
**Performance Goals**: > 5,000 req/sec; P99 response latency < 15ms; cold boot hydration < 500ms.  
**Constraints**: Zero (0) Firestore reads under warmed serving operation; $0 hosting budget; zero data races under `-race`.  
**Scale/Scope**: 49 public endpoints, 6 regional asset classes, millions of requests/month.  

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Requirement | Plan Conformance & Guarantee | Status |
|---|---|---|:---:|
| **I. API-First Contract** | Stable, versioned JSON contracts; zero client response drift | All 49 existing endpoints replicated with exact JSON schema, status codes, and error models | **PASS** |
| **II. Zero-Read Caching** | Warmed operations must incur 0 Firestore reads | Real-time push listeners stream updates into RAM; historical queries resolve from RAM Tier 1 and 30-day disk Tier 2 | **PASS** |
| **III. Freshness & Traceability** | Freshness metadata, timestamps, and `x-trace-id` on every payload | All domain envelopes embed `freshness`, `retrievedAt`, `sourceTraceId`, with fallback self-healing | **PASS** |
| **IV. Performance & SLA** | p95 < 200ms; ETags; compression enabled | Go compiled binaries deliver P99 < 15ms; Chi gzip middleware; FNV/SHA-256 ETag middleware | **PASS** |
| **V. Exhaustive Commenting** | 100% line comment coverage on all source code | Every Go source file, struct, function, and logical line commented per Constitution Principle V | **PASS** |
| **VI. Resource Safeguards** | Must run within OCI Ampere A1 Free Tier budget | Native Go uses < 50MB RAM (vs 250MB Node.js), drastically cutting CPU and memory footprints | **PASS** |
| **VII. Mandatory Verification** | Must pass verification suites before merge | `go test -race` and `npm run verify` (exit code 0) required | **PASS** |
| **VIII. Independent Lifecycle** | Decoupled from scrapers; CI/CD automation | Self-contained Go binary with clean configuration via environment variables | **PASS** |
| **IX. Version Incrementing** | Must increment version upon merge | Version tracked as `2.18.20` in version endpoints and build flags | **PASS** |

---

## Project Structure

### Documentation (this feature)

```text
specs/116-migrate-gateway-go/
├── spec.md              # Feature specification
├── plan.md              # Implementation plan (this file)
├── research.md          # Phase 0 architectural analysis
├── data-model.md        # Phase 1 domain entities & memory models
├── quickstart.md        # Phase 1 setup and verification guide
├── contracts/
│   ├── http-api.md      # 49 HTTP endpoint surface contracts
│   └── firestore-push.md# Firestore stream push contracts
└── checklists/
    └── requirements.md  # Specification quality checklist
```

### Source Code Layout (Target Go Service)

```text
bkam-api/
├── cmd/
│   └── server/
│       └── main.go                 # Process entry point, signal trapping, graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go               # Environment configuration loader
│   ├── domain/                     # Pure domain models (zero external dependencies)
│   │   ├── catalog.go              # Bank catalog with localized Arabic/English names & logos
│   │   ├── rates.go                # Rate models, trends ('up' | 'down'), and freshness
│   │   ├── gold.go                 # Precious metals models
│   │   ├── gas.go                  # Petroleum models
│   │   ├── stocks.go               # Equities models (domestic & regional GCC)
│   │   ├── indices.go              # Benchmark index models
│   │   └── quotes.go               # Batch quote request/response models
│   ├── engine/                     # Core Deep Engines
│   │   ├── live_snapshot.go        # LiveSnapshotCoordinator: RWMutex memory store & singleflight
│   │   ├── history_resolution.go   # HistoryResolutionEngine: Tier 1 RAM + Tier 2 disk + range slicing
│   │   ├── market_schedule.go      # MarketScheduleEngine: Africa/Cairo sessions, holidays, TTLs
│   │   └── intraday.go             # Intraday timeline store & previous session fallback
│   ├── adapter/
│   │   ├── firestore/              # Inbound push stream adapter
│   │   │   └── listener.go         # Goroutine workers with Query.Snapshots & Doc.Snapshots
│   │   ├── storage/                # Local persistent disk persistence adapter
│   │   │   └── disk_adapter.go     # Thread-safe JSON snapshot read/write
│   │   └── http/                   # HTTP handlers, middleware, and Chi router
│   │       ├── router.go           # Chi router assembly and route registration
│   │       ├── middleware.go       # Auth API key validation, ETag generation, tracing
│   │       ├── rates_handler.go
│   │       ├── gold_handler.go
│   │       ├── gas_handler.go
│   │       ├── stock_handler.go
│   │       ├── indices_handler.go
│   │       ├── quotes_handler.go
│   │       ├── health_handler.go
│   │       └── version_handler.go
├── scripts/
│   └── verify/                     # Verification test runners
├── go.mod
└── go.sum
```

**Structure Decision**: Standard Go layout (`cmd/`, `internal/`) enforcing Ports & Adapters architecture. The `internal/domain` layer has 0 external dependencies, `internal/engine` owns substantive domain behaviors (Deep Modules), `internal/adapter` handles transport/IO details, and `cmd/server/main.go` wires dependencies cleanly.

---

## Complexity Tracking

> **Constitution Check has zero violations. No exceptions required.**
