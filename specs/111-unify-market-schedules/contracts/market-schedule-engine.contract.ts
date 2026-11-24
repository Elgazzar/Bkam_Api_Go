// Contract Specification: MarketScheduleEngine Interface
// Defines the unified 3-method seam and schedule evaluation interface.

import type { MarketId, MarketScheduleState, MarketScheduleConfig } from '../data-model.js';

export interface IMarketScheduleEngine {
  /**
   * Evaluates whether a financial exchange is currently open for trading.
   *
   * @param market - Market identifier or alias (e.g. 'EGX', 'SA', 'AE', 'QA', 'KW', 'stocks', 'indices')
   * @param referenceDate - Optional Date object for testing/deterministic evaluation (defaults to now)
   * @returns Boolean indicating whether market is actively trading
   */
  isMarketOpen(market?: string, referenceDate?: Date): boolean;

  /**
   * Calculates the remaining duration in seconds until the next opening session.
   *
   * @param market - Market identifier or alias (defaults to 'EGX')
   * @param referenceDate - Optional Date object for deterministic evaluation
   * @returns Remaining seconds until next open (minimum 60s)
   */
  getSecondsUntilNextOpen(market?: string, referenceDate?: Date): number;

  /**
   * Deterministically calculates the calendar date (YYYY-MM-DD) of the latest active or completed trading session.
   * Runs in 0ms with 0 database reads, completely replacing 7-day Firestore scanning loops.
   *
   * @param market - Market identifier or alias (defaults to 'EGX')
   * @param referenceDate - Optional Date object for deterministic evaluation
   * @returns ISO date string (YYYY-MM-DD) of the latest trading session
   */
  getLatestTradingSessionDate(market?: string, referenceDate?: Date): string;

  /**
   * Returns the comprehensive schedule state envelope for a market.
   *
   * @param market - Market identifier or alias (defaults to 'EGX')
   * @param referenceDate - Optional Date object
   * @returns Standardized MarketScheduleState
   */
  getMarketState(market?: string, referenceDate?: Date): MarketScheduleState;

  /**
   * Registers custom holiday closures for a market.
   *
   * @param market - Market identifier (e.g. 'EGX', 'SA')
   * @param holidays - Array of YYYY-MM-DD dates representing market closures
   */
  registerHolidays(market: string, holidays: string[]): void;
}
