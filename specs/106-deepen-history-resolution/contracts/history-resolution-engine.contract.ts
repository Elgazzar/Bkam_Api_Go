// Interface Contract: Deepened HistoryResolutionEngine
// Feature 106: Absorb Active Today vs Past Date Synthesis

import type {
  AssetClass,
  DateSnapshotOptions,
  ResolvedDateSnapshot,
  RangeOptions
} from "../data-model.js";

/**
 * Deepened interface for the HistoryResolutionEngine.
 * Provides a single, unified entry point for resolving any calendar date snapshot
 * (active today or sealed past archive) and date ranges across all asset classes.
 */
export interface IHistoryResolutionEngine {
  /**
   * Resolves a daily market snapshot for any calendar date.
   * Transparently handles:
   * - Future date validation rejection (400)
   * - Active today session synthesis from live RAM + intraday coordinator (isToday: true, TTL: 60s)
   * - Past date retrieval from Active Week RAM -> Flat-Cache disk -> Firestore weekly buckets (isToday: false, TTL: 3600s)
   * - Asset-specific filtering (bank, currency, symbol)
   *
   * @param assetClass Target financial category ('rates' | 'gold' | 'stocks' | 'indices')
   * @param date ISO date string (YYYY-MM-DD)
   * @param options Optional filters (market, bank, currency, symbol)
   * @returns ResolvedDateSnapshot containing payload and caching metadata, or null if missing/closed
   */
  getDateSnapshot(
    assetClass: AssetClass,
    date: string,
    options?: DateSnapshotOptions
  ): Promise<ResolvedDateSnapshot | null>;

  /**
   * Resolves a multi-day or multi-week historical price range.
   * Transparently handles:
   * - Slicing date ranges into weekly buckets
   * - Slicing and retrieving disk/Firestore cached daily points
   * - Automatically synthesizing and appending today's live session when endDate >= todayCairo
   * - Applying symbol, bank, or currency filters
   *
   * @param assetClass Target financial category ('rates' | 'gold' | 'stocks' | 'indices')
   * @param options Date range bounds and filters
   * @returns Chronological array of daily snapshot points
   */
  getDateRange(
    assetClass: AssetClass,
    options: RangeOptions
  ): Promise<any[]>;
}

