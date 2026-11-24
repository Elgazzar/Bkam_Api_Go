# Quickstart & Verification Guide: Stateless Batch Quotes API Endpoint

**Feature Branch**: `098-batch-quotes-api`
**Date**: 2026-10-31

## Overview
This guide provides runnable scenarios to validate the stateless batch quotes endpoint (`/v1/quotes`) across both `POST` and `GET` methods.

---

## Prerequisites
- Node.js runtime (v20+ recommended)
- API server running locally on `http://localhost:3000`
- Valid API key in request header `x-api-key: <API_KEY>`

---

## Validation Scenarios

### Scenario 1: Mixed-Domain Batch Request via POST (Happy Path)

Send a POST request containing items across all 4 asset classes:
- Bank currency: `CBE_USD`, `MISR_EUR`
- Gold: `21k`, `pound`
- Index: `EGX30`
- Stocks: `COMI` (EGX), `1120` (Saudi Al Rajhi Bank)

#### Command
```bash
curl -X POST "http://localhost:3000/v1/quotes" \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key-here" \
  -d '{
    "items": [
      "CBE_USD",
      "MISR_EUR",
      "21k",
      "pound",
      "EGX30",
      "COMI",
      "1120"
    ]
  }'
```

#### Expected Outcome
- **HTTP Status**: `200 OK`
- **Payload**:
  - `quotes` contains keys for all 7 items.
  - Each item includes `id`, `category`, `name`, `price`, `currency`, and `updatedAt`.
  - `unresolved` is empty (`[]`).
  - Response time $< 50\text{ ms}$.

---

### Scenario 2: Batch Query via GET with Comma-Separated Items

Retrieve quotes via the GET endpoint:

#### Command
```bash
curl -X GET "http://localhost:3000/v1/quotes?items=CBE_USD,21k,EGX30,COMI" \
  -H "x-api-key: your-api-key-here"
```

#### Expected Outcome
- **HTTP Status**: `200 OK`
- **Headers**:
  - `Cache-Control: public, max-age=15`
  - `ETag: W/"..."`
- **Payload**:
  - Contains quotes for `CBE_USD`, `21k`, `EGX30`, and `COMI`.
  - `unresolved` is `[]`.

---

### Scenario 3: Partial Resolution with Unrecognized Identifiers

Send valid items along with nonexistent or deprecated identifiers:

#### Command
```bash
curl -X POST "http://localhost:3000/v1/quotes" \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key-here" \
  -d '{
    "items": [
      "CBE_USD",
      "21k",
      "UNKNOWN_FOO_BAR",
      "99999999"
    ]
  }'
```

#### Expected Outcome
- **HTTP Status**: `200 OK`
- **Payload**:
  - `quotes` contains entries for `CBE_USD` and `21k`.
  - `unresolved` contains `["UNKNOWN_FOO_BAR", "99999999"]`.
  - Server does NOT throw an error or 500 status.

---

### Scenario 4: Request Validation Safeguards (Empty Body & Exceeding Limit)

#### 4a. Empty Items Array
```bash
curl -X POST "http://localhost:3000/v1/quotes" \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key-here" \
  -d '{"items": []}'
```
- **Expected Outcome**: `400 Bad Request` with `INVALID_FILTER` or `INVALID_REQUEST` error code.

#### 4b. Exceeding Batch Limit (> 100 items)
```bash
# Request with 101 items
curl -X POST "http://localhost:3000/v1/quotes" \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key-here" \
  -d '{"items": [... 101 items ...]}'
```
- **Expected Outcome**: `400 Bad Request` indicating the maximum item threshold (100) was exceeded.

---

## Automated Test Execution

Run the API unit and E2E test suites to verify regression-free operation:

```bash
cd api
npm run test:unit
npm run test:e2e
```
