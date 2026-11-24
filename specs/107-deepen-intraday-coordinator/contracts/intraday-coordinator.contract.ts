/**
 * Contract: IIntradayCoordinator Deep Module Interface
 *
 * Defines the public high-leverage interface for the deepened IntradayCoordinator.
 * Callers delegate all latest session resolution, quote extraction, session backtracking,
 * previous-price baselining, and envelope construction to this contract.
 */

import type {
  AssetCategory,
  IntradayPoint,
  LatestIntradayTimelineResult,
  ResolveLatestTimelineOptions,
  ResolveTimelineOptions,
  RecoverYesterdayOptions
} from '../../../src/services/intraday-types.js';

export type AssetClass = 'rates' | 'gold' | 'stocks' | 'indices';
export type GccMarket = 'saudi' | 'uae' | 'qatar' | 'kuwait';

export interface LatestIntradayOptions {
  bank?: string;
  currency?: string;
  market?: GccMarket;
  liveSnapshot?: any;
}

export interface LatestIntradayResult<T = any> {
  data: T;
  isToday: boolean;
  sessionDate: string;
  cacheTtlSeconds: number;
}

export interface IIntradayCoordinatorContract {
  /**
   * Primary deep entry point: resolves the latest active or backtracked intraday session
   * synthesized with quote rates, points timeline, previous-price baselines, and caching metadata.
   *
   * @param assetClass - 'rates' | 'gold' | 'stocks' | 'indices'
   * @param target - Target symbol, currency, or karat ('USD', 'COMI', '21k', 'EGX30', 'all')
   * @param options - Additional options including bank filter, regional market, or test fixtures
   * @returns Complete standardized intraday envelope
   */
  getLatestIntraday<T = any>(
    assetClass: AssetClass,
    target?: string,
    options?: LatestIntradayOptions
  ): Promise<LatestIntradayResult<T>>;

  /**
   * Existing primitive: coordinates raw symbol-to-points mapping across live/backtracked sessions.
   */
  resolveLatestTimeline(
    options: ResolveLatestTimelineOptions
  ): Promise<LatestIntradayTimelineResult>;

  /**
   * Normalizes heterogeneous intraday point structures into a uniform array of IntradayPoint items.
   */
  normalizePoints(points: any): IntradayPoint[];

  /**
   * Performs case-resilient symbol lookup in a points dictionary.
   */
  lookupSymbolPoints(pointsMap: Record<string, any>, symbol: string): IntradayPoint[];
}

