# Quickstart: Expose Previous Day Price in Latest Intraday Endpoints

**Feature Branch**: `100-intraday-previous-price`  
**Date**: 2026-11-05  
**Spec**: [Feature Specification](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/specs/100-intraday-previous-price/spec.md)  
**Status**: Completed  

---

## Prerequisites

1. Node.js 22 LTS installed.
2. Dependencies installed:
   ```bash
   cd api
   npm install
   ```

---

## 1. Automated Validation (Vitest Integration Suite)

Run the dedicated integration test suite verifying that all four latest intraday endpoints return the newly exposed previous session price attributes:

```bash
cd api
npx vitest run tests/integration/latest-intraday-routes.test.ts
```

**Expected Outcome**:
All tests pass cleanly:
- `previousClose` is asserted on Egyptian and GCC stock quotes.
- `previousClose` is asserted on benchmark index quotes.
- `previousBuyPrice` and `previousSellPrice` are asserted on single-bank and consolidated currency rates.
- `previousBuyPrice` and `previousSellPrice` are asserted on single gold commodities and multi-karat envelopes.

---

## 2. Local Manual Verification (Fastify Server & cURL)

Start the API service in development mode:

```bash
cd api
npm run dev
```

### Scenario 1: Equities Previous Close Verification
```bash
curl -s -H "x-api-key: dev-key" "http://localhost:3000/v1/stocks/COMI/intraday/latest" | jq '{symbol, price, previousClose, change, changePercent}'
```
**Expected Response Shape**:
```json
{
  "symbol": "COMI",
  "price": 84.5,
  "previousClose": 83.0,
  "change": 1.5,
  "changePercent": 1.81
}
```

### Scenario 2: Benchmark Indices Previous Close Verification
```bash
curl -s -H "x-api-key: dev-key" "http://localhost:3000/v1/indices/EGX30/intraday/latest" | jq '{symbol, price, previousClose, change}'
```
**Expected Response Shape**:
```json
{
  "symbol": "EGX30",
  "price": 30500.25,
  "previousClose": 30250.0,
  "change": 250.25
}
```

### Scenario 3: Bank Currency Exchange Rates Previous Close Verification
```bash
curl -s -H "x-api-key: dev-key" "http://localhost:3000/v1/rates/USD/intraday/latest?bank=cbe" | jq '{currency, bank, buy, sell, previousBuyPrice, previousSellPrice}'
```
**Expected Response Shape**:
```json
{
  "currency": "USD",
  "bank": "cbe",
  "buy": 48.65,
  "sell": 48.75,
  "previousBuyPrice": 48.60,
  "previousSellPrice": 48.70
}
```

### Scenario 4: Gold Commodities Previous Close Verification
```bash
# Single Karat
curl -s -H "x-api-key: dev-key" "http://localhost:3000/v1/gold/21/intraday/latest" | jq '{item, buy, sell, previousBuyPrice, previousSellPrice}'

# Multi-Karat Envelope
curl -s -H "x-api-key: dev-key" "http://localhost:3000/v1/gold/intraday/latest" | jq '.karats."21"'
```
**Expected Response Shape (Single Karat)**:
```json
{
  "item": "21",
  "buy": 3650.0,
  "sell": 3670.0,
  "previousBuyPrice": 3630.0,
  "previousSellPrice": 3650.0
}
```

---

## 3. Reference Artifacts

- [Feature Specification](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/specs/100-intraday-previous-price/spec.md)
- [Research & Design Decisions](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/specs/100-intraday-previous-price/research.md)
- [Data Model & Entities](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/specs/100-intraday-previous-price/data-model.md)
- [OpenAPI Contracts](file:///d:/Repo/Bekam_Fe_El_bank_Scrapper/specs/100-intraday-previous-price/contracts/intraday-latest.openapi.json)

