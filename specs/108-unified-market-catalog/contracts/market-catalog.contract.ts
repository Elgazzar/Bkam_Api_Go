// Interface Contract: Deep Market Catalog Module
// Feature: 108-unified-market-catalog
// Strictly adheres to Principle I (API-First Contract) and codebase-design rules

import type { Bank } from "../../../src/models/bank.js";
import type { Currency } from "../../../src/models/currency.js";

/**
 * Normalized Security Profile representing an equity across domestic and GCC markets.
 */
export interface SecurityDefinition {
  /** Canonical uppercase ticker symbol (e.g. "COMI", "RIBL") */
  symbol: string;
  /** Lookup code without exchange suffix */
  code: string;
  /** Optional Tadawul 4-digit numeric code (e.g. "1010") */
  numericCode?: string;
  /** Retail or legacy alias (e.g. "BDC") */
  alias?: string;
  /** TradingView scanner ticker */
  tvSymbol?: string;
  /** Localized Arabic display name */
  nameAr: string;
  /** Localized English display name */
  nameEn: string;
  /** Sector classification */
  sector: string;
  /** Operating exchange */
  exchange: "EGX" | "TADAWUL" | "ADX" | "DFM" | "QSE" | "BK" | string;
  /** ISO 3166-1 alpha-2 country code */
  country: "EG" | "SA" | "AE" | "QA" | "KW";
  /** Trading currency code */
  currency: "EGP" | "SAR" | "AED" | "QAR" | "KWD" | string;
  /** Benchmark indices containing this security */
  indices: string[];
  /** Screening and classification groups */
  groups: string[];
  /** Standard absolute HTTPS URL to official brand logo image */
  logo: string;
}

/**
 * Normalized Benchmark Index definition.
 */
export interface BenchmarkIndexDefinition {
  symbol: string;
  code: string;
  tvSymbol: string;
  exchange: string;
  country: "EG" | "SA" | "AE" | "QA" | "KW";
  currency: string;
  nameAr: string;
  nameEn: string;
  logo: string;
  constituentSymbols: string[];
}

/**
 * Commodity metadata definition (gold karats and fuel grades).
 */
export interface CommodityDefinition {
  code: string;
  category: "gold" | "fuel";
  nameAr: string;
  nameEn: string;
  unit: string;
}

/**
 * Unified parameter validation result.
 */
export interface ValidationResult {
  valid: boolean;
  canonicalValue?: string;
  error?: string;
}

/**
 * Full catalog discovery inventory payload for GET /v1/catalog.
 */
export interface CatalogInventory {
  banks: Bank[];
  currencies: Currency[];
  commodities: CommodityDefinition[];
  indices: BenchmarkIndexDefinition[];
  counts: {
    totalBanks: number;
    totalCurrencies: number;
    totalCommodities: number;
    totalIndices: number;
    totalSecurities: number;
  };
}

/**
 * Deep Market Catalog Interface Contract.
 * Callers interact with platform inventory strictly through this interface.
 */
export interface IMarketCatalog {
  /**
   * Universal security lookup supporting ticker symbols, numeric codes, and aliases.
   * Case-insensitive and whitespace-resilient.
   *
   * @param codeOrSymbol Input ticker, numeric code, or alias
   * @param market Optional exchange market filter ("EGX", "saudi", "uae", "qatar", "kuwait")
   * @returns Canonical SecurityDefinition or undefined if not found
   */
  getSecurity(codeOrSymbol: string, market?: string): SecurityDefinition | undefined;

  /**
   * Retrieve all constituent member securities belonging to a benchmark index.
   *
   * @param indexCode Benchmark index identifier (e.g. "EGX30", "TASI", "DFMGI")
   * @returns Array of constituent SecurityDefinitions (empty if invalid index)
   */
  getConstituents(indexCode: string): SecurityDefinition[];

  /**
   * Retrieve a benchmark index definition by code.
   *
   * @param code Index code (e.g. "EGX30", "EGX70", "TASI")
   */
  getIndex(code: string): BenchmarkIndexDefinition | undefined;

  /**
   * Retrieve a supported bank by unique bank code.
   *
   * @param code Bank code (e.g. "NBE", "CIB", "MISR")
   */
  getBank(code: string): Bank | undefined;

  /**
   * Retrieve a supported currency by currency code.
   *
   * @param code Currency code (e.g. "USD", "EUR", "SAR")
   */
  getCurrency(code: string): Currency | undefined;

  /**
   * Enrich raw price snapshot with authoritative catalog metadata (logos, localized names, indices).
   *
   * @param assetClass "stocks" | "indices" | "rates" | "gold"
   * @param snapshot Raw price snapshot
   * @param market Optional regional market identifier
   * @returns Enriched snapshot
   */
  enrichSnapshot<T>(assetClass: string, snapshot: T, market?: string): T;

  /**
   * Validate parameter against catalog registry.
   *
   * @param type "bank" | "currency" | "index" | "group" | "market" | "symbol"
   * @param value Raw parameter value from request
   * @param market Optional market context
   */
  validate(
    type: "bank" | "currency" | "index" | "group" | "market" | "symbol",
    value: string,
    market?: string
  ): ValidationResult;

  /**
   * Filter a dictionary of stock quotes by index code or group tag.
   *
   * @param stocks Map of symbol to stock quotes
   * @param filterTarget Target index or group name
   * @param exchange Optional exchange identifier
   */
  filterStocks(
    stocks: Record<string, any>,
    filterTarget?: string,
    exchange?: string
  ): Record<string, any>;

  /**
   * Export complete platform discovery inventory for GET /v1/catalog.
   */
  getInventory(): CatalogInventory;
}
