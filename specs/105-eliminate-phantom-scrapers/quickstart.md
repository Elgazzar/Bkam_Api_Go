# Quickstart Validation Guide: Eliminate Phantom Scraping Machinery

**Feature**: `105-eliminate-phantom-scrapers`

**Date**: 2026-11-15

---

## 1. Prerequisites

- Node.js >= 22 installed.
- Repository root: `D:\Repo\Bkam\Bkam_Api`.
- No Playwright or Chromium installation is required on the host system.

---

## 2. Validation Scenarios

### Scenario A: Clean TypeScript Compilation
Verify that all source code compiles without broken imports or missing types after purging `src/adapters/banks/*`:
```bash
cd D:\Repo\Bkam\Bkam_Api
npm run build
```
*Expected Outcome*: TypeScript compiler exits cleanly with code 0 (`tsc -p tsconfig.build.json`).

---

### Scenario B: Unit Test Execution (Without Playwright)
Run the full unit test suite:
```bash
npm run test:unit
```
*Expected Outcome*: All unit test suites pass (27 test files, 225+ tests) with zero errors regarding missing browser binaries.

---

### Scenario C: Contract Test Verification
Run contract tests asserting public API contract integrity:
```bash
npm run test:contract
```
*Expected Outcome*: All 3 contract test files pass (9/9 tests) without needing to mock `createSupportedAdapters`.

---

### Scenario D: Local Server Startup & Live Health Query
Start the local Fastify development server:
```bash
npm run dev
```
In a secondary terminal, verify the health and rates endpoints:
```bash
# Health check (Port 3001)
curl http://localhost:3001/v1/health

# Exchange rates
curl -H "x-api-key: local-dev-key" http://localhost:3001/v1/rates
```
*Expected Outcome*:
- HTTP 200 returned on `/v1/health`.
- HTTP 200 returned on `/v1/rates` with the complete bank exchange rates payload.
- Zero log entries relating to browser launches, Playwright, or direct bank scraping.
