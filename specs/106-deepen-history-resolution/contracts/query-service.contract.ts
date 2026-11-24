// Interface Contract: Query Service Integration
// Feature 106: Simplified caller query services delegating to the deepened engine

/**
 * Illustrates the drastically simplified signature of caller query services.
 * Callers no longer perform manual isToday checks, timezone conversions, or intraday stitching.
 */
export interface IAssetQueryService {
  /**
   * Fetch daily snapshot for date (delegates directly to HistoryResolutionEngine)
   */
  fetchDailyHistory(date: string, filters?: any): Promise<any | null>;

  /**
   * Fetch historical range (delegates directly to HistoryResolutionEngine)
   */
  fetchHistoryRange(start: string, end: string, filters?: any): Promise<any[]>;
}

