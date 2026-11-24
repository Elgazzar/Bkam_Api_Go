# Research & Architecture Analysis: Unify Fragmented Market Schedules

**Feature Branch**: `111-unify-market-schedules`  
**Date**: 2026-11-15  
**Status**: Completed  

---

## 1. Problem Space & Architectural Flaws

### 1.1 Disjoint Schedule Declarations with Conflicting Trading Hours
Currently, market operating hours are declared in two separate files with conflicting hours:
1. `src/services/market-schedule-service.ts`:
   - Hardcoded solely for EGX in Cairo timezone (`Africa/Cairo`).
   - Declares active trading time as `09:30` to `16:00` Cairo local time (`cairoHourDecimal >= 9.5 && cairoHourDecimal < 16.0`).
   - Hardcoded to calculate `secondsUntilNextMarketOpen` only towards 09:30.
2. `src/config/market-registry.ts` & `src/modules/gcc-stocks/lib/market-schedule.ts`:
   - Declares `EGX` schedule as `09:00` to `15:30` Cairo local time (active trading 10:00 - 14:30 buffered by 1 hour for delayed feeds).
   - Also declares regional schedules for Saudi Arabia (`09:00–16:00`), UAE (`09:00–16:00` Monday–Friday), Qatar (`08:30–14:15`), and Kuwait (`08:00–13:30`).

**Consequence**: Cache TTL calculations in `HistoryResolutionEngine` and `FirestoreListenerService` use `marketScheduleService` (09:30–16:00), while regional market scrapers and catalog definitions use `market-registry.ts` (09:00–15:30). Between 09:00 and 09:30, and between 15:30 and 16:00, different parts of the system report conflicting `open` vs `closed` states.

### 1.2 Iterative Firestore Scanning Loop for Previous Trading Sessions
In `src/services/intraday-coordinator.ts`:
- When an intraday query arrives during off-hours, weekends, or midnight rollover, the system attempts to backtrack to the previous recorded session.
- Lines 356–414 in `intraday-coordinator.ts` call `db.collection(collectionName).listDocuments()` or query descending by document ID, and enter an asynchronous `while` loop that fetches up to 7 documents individually from Firestore:
  ```ts
  while (sortedDates.length > 0) {
    const candidateDate = sortedDates.pop()!;
    ...
    const snap = await db.collection(collectionName).doc(candidateDate).get();
    ...
  }
  ```
- In production, this loop:
  - Consumes up to 7 Firestore read operations per cold off-hours query.
  - Adds 300ms–800ms of avoidable network latency.
  - Causes test flakiness if Firestore mocks do not emulate `listDocuments()`.
  - Is completely unnecessary because the previous trading session date is 100% mathematically deterministic given the market's calendar!

---

## 2. Target Design: Deep MarketScheduleEngine

As outlined in the architectural design, `MarketScheduleEngine` acts as a deep module that hides:
- Multi-Market Registry (`EGX`, `SA`, `AE`, `QA`, `KW`, `RATES`, `GOLD`)
- Timezone Offsets (standard & daylight saving time in `Africa/Cairo`, `Asia/Riyadh`, `Asia/Dubai`, `Asia/Qatar`, `Asia/Kuwait`)
- National Holiday Calendar (Egypt and GCC market closures)
- Instant Session Backtracking (pure memory calendar arithmetic)

### 2.1 The Deep Seam Interface

```ts
export interface IMarketScheduleEngine {
  isMarketOpen(market?: string, referenceDate?: Date): boolean;
  getSecondsUntilNextOpen(market?: string, referenceDate?: Date): number;
  getLatestTradingSessionDate(market?: string, referenceDate?: Date): string;
  getMarketState(market?: string, referenceDate?: Date): MarketScheduleState;
}
```

### 2.2 Instant Session Backtracking Algorithm

Given any `referenceDate` (defaults to current wall clock time):
1. **Timezone Localization**: Format reference date into the target market's local timezone parts (`year`, `month`, `day`, `weekday`, `hour`, `minute`).
2. **Local Decimal Hour**: Compute `hour + minute / 60`.
3. **Session Evaluation**:
   - If `weekday` is in market's `activeDays` AND not in market's `holidays`:
     - If `localDecimalHour >= openDecimalHour`: Today IS an active or completed trading session. Return today's date (`YYYY-MM-DD`).
     - If `localDecimalHour < openDecimalHour`: Market has not opened today yet. Step back to previous calendar day.
   - If `weekday` is not in market's `activeDays` or is in `holidays`:
     - Step back to previous calendar day.
4. **Backtracking Loop**:
   - Decrement calendar day by 1.
   - Check if candidate day's weekday is in `activeDays` and date string is not in `holidays`.
   - The first matching day is returned as `YYYY-MM-DD`.
5. **Performance & Complexity**:
   - Time complexity: $O(1)$ (maximum 4 iterations over weekends/holidays).
   - Database reads: Exactly **0**.
   - Latency: $< 0.1\text{ ms}$.

---

## 3. Consolidation & Deprecation Matrix

| Legacy Module / Method | Target Replacement | Impact / Refactoring |
| :--- | :--- | :--- |
| `MarketScheduleService.isTradingHours()` in `src/services/market-schedule-service.ts` | `MarketScheduleEngine.isMarketOpen(market)` | Canonical EGX hours set to 09:00–15:30; re-exported for backwards compatibility |
| `MarketScheduleService.getSecondsUntilNextMarketOpen()` in `src/services/market-schedule-service.ts` | `MarketScheduleEngine.getSecondsUntilNextOpen(market)` | Multi-market TTL calculation; re-exported for backwards compatibility |
| `isMarketOpenById()` & `isScheduleActive()` in `src/config/market-registry.ts` | `MarketScheduleEngine.isMarketOpen(marketId)` | Delegates directly to `MarketScheduleEngine` |
| `ALL_MARKET_SCHEDULES` in `src/modules/gcc-stocks/lib/market-schedule.ts` | Merged into `MarketScheduleEngine` internal registry | Centralized source of truth |
| 7-day `while (sortedDates.length > 0)` Firestore scanning in `src/services/intraday-coordinator.ts` | `MarketScheduleEngine.getLatestTradingSessionDate(market)` | Completely deleted; single targeted document fetch |

---

## 4. Verification Strategy

1. **Unit Tests**:
   - Multi-market open status checks across active hours, after-hours, weekends, and holidays.
   - Deterministic session backtracking across all days of the week for `EGX` (Sunday–Thursday) and `AE` (Monday–Friday).
   - Midnight rollover tests (00:00–08:59 on trading day returns previous trading date).
   - Seconds-until-open TTL calculations across weekends and overnight closures.
2. **Integration / Coordinator Tests**:
   - Verify `IntradayCoordinator` resolves cold off-hours sessions via single document lookup without calling `listDocuments`.
3. **Differential Parity**:
   - Run `npm run verify:diff` against baseline port 3000 to ensure zero response payload drift.
