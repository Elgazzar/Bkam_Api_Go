# Quickstart & Validation Guide: Tiered Weekly Caching & API Performance

## Prerequisites
- Node.js >= 22
- Working directory: `api/`
- Dependencies: `@fastify/compress`, `@fastify/etag`, `flat-cache`

---

## 1. Setup & Dependency Verification

```bash
cd api
npm install @fastify/compress @fastify/etag
```

---

## 2. Automated Test Suite Execution

### Unit Tests
Execute unit tests covering session-aware caching, range result caching, and prefix stitching:
```bash
npm run test:unit tests/unit/firestore-service.test.ts
npm run test:unit tests/unit/stock-query-service.test.ts
```

### Integration Tests
Execute integration tests validating HTTP compression, ETag 304 revalidation, and dynamic Cache-Control headers:
```bash
npm test tests/integration/stock-history.test.ts
```

---

## 3. Manual / cURL Verification Scenarios

### Scenario A: Verify Gzip Compression & ETag Header
```bash
curl -i -H "x-api-key: secret" -H "Accept-Encoding: gzip" \
  "http://localhost:3000/v1/stocks/COMI/history/range?start=2026-05-01&end=2026-10-20"
```
**Expected Outcome**:
- Response header `content-encoding: gzip`
- Response header `etag: W/"..."`
- Response header `cache-control: public, max-age=86400, immutable`
- Content length is ~85% smaller than uncompressed JSON.

### Scenario B: Verify Conditional HTTP 304 Revalidation
Take the `etag` value from Scenario A and repeat the request:
```bash
curl -i -H "x-api-key: secret" -H "If-None-Match: W/\"...\"" \
  "http://localhost:3000/v1/stocks/COMI/history/range?start=2026-05-01&end=2026-10-20"
```
**Expected Outcome**:
- HTTP status `304 Not Modified`
- Zero payload bytes transferred
- Response time `<10ms`

### Scenario C: Verify Dynamic Today Stitching & 60s Cache-Control
```bash
curl -i -H "x-api-key: secret" \
  "http://localhost:3000/v1/stocks/COMI/history/range?start=2026-05-01&end=2026-10-24"
```
**Expected Outcome**:
- Response header `cache-control: public, max-age=60`
- Historical dates contain only closing prices without intraday arrays
- Today's record (`2026-10-24`) contains the latest live quote and real-time `intraday` tick array.
