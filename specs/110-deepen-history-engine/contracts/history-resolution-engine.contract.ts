// Contract definition for HistoryResolutionEngine.
// Authoritative coordinator for single-date historical snapshots and multi-day range queries.

import type { AssetClass, GccMarket, DateSnapshotOptions, RangeOptions, ResolvedDateSnapshot } from '../../src/services/history-resolution-engine.js';

export interface IHistoryResolutionEngine {
  /**
   * Resolves a single calendar date snapshot across any asset domain.
   * Handles Cairo timezone boundaries, future date rejection, tiered caching
   * (RAM -> Negative -> Disk flat-cache -> Weekly bucket -> Legacy daily),
   * and dynamic active-today live RAM + intraday synthesis.
   */
  getDateSnapshot<T = any>(
    assetClass: AssetClass,
    date: string,
    options?: DateSnapshotOptions
  ): Promise<ResolvedDateSnapshot<T>>;

  /**
   * Resolves a continuous multi-day time-series range across any asset domain.
   * Handles future date validation, inverted range rejection, pure historical disk caching,
   * range prefix caching, yesterday defensive synthesis, and dynamic active-today intraday stitching.
   */
  getDateRange(
    assetClass: AssetClass,
    options: RangeOptions
  ): Promise<any[]>;

  /**
   * Clears internal in-memory caches (negative cache and inflight promises).
   * Primarily utilized for testing and administrative cache reset.
   */
  clearCache(): void;
}
