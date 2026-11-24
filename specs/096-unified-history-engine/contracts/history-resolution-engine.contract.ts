/**
 * Contract specification for the HistoryResolutionEngine
 * Defined in specs/096-unified-history-engine/contracts/history-resolution-engine.contract.ts
 */

export type AssetClass = 'rates' | 'gold' | 'stocks' | 'indices';

export type GccMarket = 'saudi' | 'uae' | 'qatar';

export interface RangeOptions {
  startDate: string;         // YYYY-MM-DD
  endDate: string;           // YYYY-MM-DD
  symbol?: string;           // Ticker symbol (required for stocks and indices)
  bank?: string;             // Optional bank filter (for rates)
  currency?: string;         // Optional currency filter (for rates)
  market?: GccMarket;        // Regional market (for GCC equities)
}

export interface IHistoryResolutionEngine {
  /**
   * Resolves a historical closing price snapshot for a single calendar date.
   *
   * Hierarchy:
   * 1. If date is in the Active Calendar Week, resolve from RAM (ActiveWeeklyBucketStore) in <0.1ms.
   * 2. If date is recorded in negative cache (holiday/weekend), resolve immediately as null.
   * 3. If week is sealed, check CacheService (flat-cache on disk, 30-day TTL).
   * 4. If cache miss, fetch weekly bucket from Firestore, persist to CacheService, and extract daily record.
   * 5. If weekly bucket document does not exist (pre-migration archive), fallback to legacy daily collection.
   *
   * @param assetClass The target asset class ('rates', 'gold', 'stocks', 'indices')
   * @param date Calendar date string formatted as YYYY-MM-DD
   * @param market Optional GCC market identifier ('saudi', 'uae', 'qatar')
   * @returns Daily closing snapshot payload or null if market closed/missing
   */
  getDateSnapshot(
    assetClass: AssetClass,
    date: string,
    market?: GccMarket
  ): Promise<any | null>;

  /**
   * Resolves historical market quotes across a multi-day or multi-week date range.
   *
   * Slices distinct ISO calendar weeks covering [startDate, endDate], batch resolves
   * each weekly bucket from RAM (Active Week) or CacheService (Sealed Weeks), extracts
   * daily closing snapshots within the range, and applies optional domain filters.
   *
   * @param assetClass The target asset class ('rates', 'gold', 'stocks', 'indices')
   * @param options Query range and domain-filtering options
   * @returns Array of daily historical records sorted chronologically
   */
  getDateRange(
    assetClass: AssetClass,
    options: RangeOptions
  ): Promise<any[]>;
}
