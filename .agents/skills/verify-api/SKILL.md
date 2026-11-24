---
name: verify-api
description: Mandatory verification workflow to run after any code modification, bug fix, or feature implementation in the API repository.
---

# API Verification Workflow

Whenever you modify any file in this repository, execute this exact verification sequence before finishing:

## Prerequisites
- Ensure required services/env vars are active (e.g., check `.env` exists; do not modify production configs).

## Verification & Reporting Steps

### Step 1: Comprehensive Verification & HTML Report Generation
Execute the unified verification suite to run all 37 schema/SLA scenarios (Gate 1) and all 31 differential parity checks (Gate 2), generating a self-contained HTML audit report:
```bash
npm run verify:report
```
- **Output Artifact:** `reports/verification-report.html`
- Validates Zod response schemas, HTTP status codes, and RAM-cached latency SLAs (< 150ms).
- Compares baseline server (port 3000) against candidate deployment (port 3001) for 100% structural parity.

### Alternative Quick Terminal Check
If you only need a quick terminal smoke check without HTML generation:
```bash
npm run verify
npm run verify:diff
```
