# Quickstart Guide: MarketScheduleEngine

**Feature Branch**: `111-unify-market-schedules`  
**Date**: 2026-11-15  
**Status**: Completed  

---

## 1. Quick Usage Examples

### 1.1 Checking if a Market is Open

```ts
import { marketScheduleEngine } from '../services/market-schedule-engine.js';

// Check domestic Egyptian Exchange (defaults to EGX)
const isEgxOpen = marketScheduleEngine.isMarketOpen('EGX');

// Check regional markets
const isSaudiOpen = marketScheduleEngine.isMarketOpen('SA');
const isUaeOpen = marketScheduleEngine.isMarketOpen('AE');
```

### 1.2 Computing Off-Hours Cache TTL (Seconds Until Next Open)

```ts
import { marketScheduleEngine } from '../services/market-schedule-engine.js';

// Returns exact seconds until Sunday 09:00 AM Cairo when called on weekend
const ttlSeconds = marketScheduleEngine.getSecondsUntilNextOpen('EGX');

// Use as flat-cache TTL for off-hours reboots
cacheService.set('stock_prices_history_weekly/2026-W46', weekData, ttlSeconds);
```

### 1.3 Instant Session Backtracking (0 Database Reads)

```ts
import { marketScheduleEngine } from '../services/market-schedule-engine.js';

// Resolve latest trading session date (e.g. today during trading hours, or Thursday on Friday/Saturday)
const sessionDate = marketScheduleEngine.getLatestTradingSessionDate('stocks');

// Probe exactly that single date without scanning across a 7-day Firestore loop:
const docSnap = await db.collection('stock_prices_intraday').doc(sessionDate).get();
```

---

## 2. Backward-Compatible Service Facade

Existing callers importing `marketScheduleService` from `src/services/market-schedule-service.ts` remain 100% compatible:

```ts
import { marketScheduleService } from './market-schedule-service.js';

// Delegates directly to marketScheduleEngine.isMarketOpen('EGX')
const isOpen = marketScheduleService.isTradingHours();

// Delegates directly to marketScheduleEngine.getSecondsUntilNextOpen('EGX')
const ttl = marketScheduleService.getSecondsUntilNextMarketOpen();
```

---

## 3. Running Verification Tests

```bash
# Run unit tests for market schedule engine
npx vitest run tests/unit/market-schedule-engine.test.ts

# Run intraday coordinator tests to verify zero-read backtracking
npx vitest run tests/unit/intraday-coordinator.test.ts

# Run full unit suite
npm run test:unit

# Run full contract suite
npm run test:contract

# Differential parity verification against baseline server
npm run verify:diff
```
