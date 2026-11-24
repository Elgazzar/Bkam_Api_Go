/**
 * Contract: Intraday Response Envelopes
 *
 * Enforces byte-for-byte backward compatibility for client-facing REST endpoints:
 * - GET /v1/rates/intraday/latest
 * - GET /v1/gold/intraday/latest
 * - GET /v1/stocks/:symbol/intraday/latest
 * - GET /v1/indices/:code/intraday/latest
 */

import type { IntradayPoint } from '../../../src/services/intraday-types.js';

export interface RatesIntradayResponseContract {
  currency: string;
  bank?: string;
  buy: number;
  sell: number;
  previousBuyPrice: number;
  previousSellPrice: number;
  intraday: IntradayPoint[];
  updatedAt: string;
}

export interface GoldKaratIntradayContract {
  buy: number;
  sell: number;
  previousBuyPrice: number;
  previousSellPrice: number;
  intraday: IntradayPoint[];
}

export interface GoldIntradayResponseContract {
  date: string;
  karats: {
    "24": GoldKaratIntradayContract;
    "21": GoldKaratIntradayContract;
    "18": GoldKaratIntradayContract;
    pound: GoldKaratIntradayContract;
  };
  international?: GoldKaratIntradayContract;
  updatedAt: string;
}

export interface StockIntradayResponseContract {
  symbol: string;
  name_en: string;
  name_ar: string;
  price: number;
  change: number;
  changePercent: number;
  previousPrice: number;
  intraday: IntradayPoint[];
  updatedAt: string;
  country?: string;
  exchange?: string;
}

export interface IndexIntradayResponseContract {
  code: string;
  name_en: string;
  name_ar: string;
  value: number;
  change: number;
  changePercent: number;
  previousValue: number;
  intraday: IntradayPoint[];
  updatedAt: string;
}

