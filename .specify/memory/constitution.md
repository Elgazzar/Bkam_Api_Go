<!--
Sync Impact Report:
- Version Change: 1.0.0 → 1.1.0
- Bump Type: MINOR (Added Principle IX: Mandatory Version Incrementing and Governance section)
- Modified Principles: None
- Added Sections:
  - Core Principles: ### IX. Mandatory Version Incrementing
  - ## Governance
- Removed Sections: None
- Deferred Items / TODOs: None
-->

# Bkam API Gateway Constitution

## Core Principles

### I. API-First Contract
The system MUST be designed around stable, versioned JSON contracts for Android, web, and external clients. Every client-facing capability MUST define request shape, response shape, error schema, and freshness metadata before implementation begins. Breaking API changes MUST be versioned (e.g. `/v2/`) or paired with an explicit client migration plan.
Rationale: Downstream mobile and client apps depend on predictable payloads and cannot safely absorb silent response drift.

### II. Zero-Read Memory & Disk Caching
The API service MUST maximize resource efficiency and protect Cloud database free tier boundaries (enforcing <50,000 reads/day).
- All real-time asset quotes (Bank exchange rates, Gold karats, Fuel gas prices, EGX/GCC Stocks, Benchmark Indices) MUST be served directly from in-memory RAM caches populated via push listeners (`onSnapshot`) or background refresh cycles.
- Sealed past historical week ranges MUST be resolved from persistent disk cache (`flat-cache` with 30-day TTL) or local memory buckets.
- Under active serving conditions, standard HTTP queries MUST incur zero (0) direct Firestore read operations.

### III. Freshness, Traceability, and Fallbacks
Every returned rate and asset payload MUST include data source identification, retrieval timestamps, freshness indicators, and trace IDs (`x-trace-id`). When an upstream data provider experiences an outage, the API MUST serve cached snapshots with clear degradation markers rather than returning 500 errors.

### IV. Performance and SLA Guarantees
- The API service MUST maintain a p95 response latency of under 200ms for all cached endpoint requests.
- Transport Brotli and Gzip compression MUST be enabled for large collection responses.
- HTTP ETags and conditional `304 Not Modified` headers MUST be evaluated on historical and intraday timeline routes.

### V. Exhaustive Code Commenting
Every single line of source code in the repository MUST be explicitly and clearly commented detailing its logic, rationale, and runtime expectations. Comments MUST be updated in tandem with any code modification. No uncommented lines of code are permitted under any circumstances.

### VI. Cloud Resource and Budget Safeguards
The API service runs on Oracle Cloud Infrastructure (OCI) Ampere A1 Compute (4 ARM cores, 24GB RAM) within the Always Free tier, maintaining a strict $0 direct hosting budget.

### VII. Mandatory Testing and Contract Verification
Every code modification MUST be accompanied by passing unit tests, contract schema tests (`npm run test:contract`), and integration test suites. Merges without passing tests are strictly blocked.

### VIII. Independent Deployment Lifecycle
The API service MUST remain entirely decoupled from external scraping daemons. It is deployed automatically to the production OCI VM on push to `master` via GitHub Actions (`.github/workflows/deploy-api.yml`).

### IX. Mandatory Version Incrementing
With every change, modification, feature addition, bug fix, or dependency update, the application version in `package.json` MUST be incremented prior to merge into `master`.
- Version increments MUST follow Semantic Versioning (SemVer: MAJOR.MINOR.PATCH):
  - **PATCH** (`x.y.Z`): Backward-compatible bug fixes, refactoring, performance improvements, internal optimizations, and code documentation updates.
  - **MINOR** (`x.Y.0`): New client-facing features, new API endpoints, or backward-compatible schema/capability additions.
  - **MAJOR** (`X.0.0`): Breaking contract changes, modified existing response shapes, or incompatible architectural overhauls.
- The version running in production MUST be observable and verifiable in real-time via `GET /v1/version` and `GET /version`, as well as logged during process startup.
- Pull requests, commits, and merges that alter codebase behavior or schemas without incrementing the application version MUST be rejected during review and CI validation.
Rationale: Deterministic, automated version incrementing guarantees end-to-end traceability for clients, monitoring probes, and CI/CD pipelines, removing all ambiguity as to whether an environment is serving the latest build.

## Governance

This Constitution represents the supreme architectural law of the Bkam API Gateway repository and supersedes all ad-hoc conventions or team preferences.

1. **Compliance Verification**: All pull requests, code reviews, and automated verification suites MUST explicitly verify compliance with Principles I through IX. Any violation of a MUST normative statement is automatically a CRITICAL blocker.
2. **Amendment Procedure**: Amendments to this Constitution require documented rationale, impact analysis on existing clients, and formal approval. Proposed amendments MUST be submitted via `/speckit-constitution`.
3. **Constitution Versioning**:
   - **MAJOR**: Incompatible governance shifts, principle removals, or fundamental policy redefinitions.
   - **MINOR**: Addition of new principles or sections, or materially expanded architectural guidance.
   - **PATCH**: Clarifications, non-semantic refinements, or typo fixes.

**Version**: 1.1.0 | **Ratified**: 2026-11-14 | **Last Amended**: 2026-11-15
