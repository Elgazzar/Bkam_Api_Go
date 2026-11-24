// Contract Specification: LiveSnapshotCoordinator and Thin Presenter Seams
// Governed by Constitution Principle I (API-First Contracts) and Principle II (Zero-Read Caching)

// Supported asset domains managed by the snapshot coordinator
export type AssetDomain = 'rates' | 'gold' | 'gas' | 'stocks' | 'indices' | 'gcc-stocks';

// Standardized live snapshot envelope
export interface LiveSnapshotEnvelope<T = any> {
  domain: AssetDomain;
  market?: string;
  currency: string;
  updatedAt: string;
  scrapedAt: string;
  freshness: 'current' | 'partial' | 'stale' | 'degraded';
  sourceTraceId: string;
  data: T;
}

// Contract interface for the centralized LiveSnapshotCoordinator
export interface ILiveSnapshotCoordinatorContract {
  // Synchronous zero-read retrieval of a domain snapshot from volatile RAM
  getSnapshot<T = any>(domain: AssetDomain, market?: string): T | null;

  // Retrieve raw envelope metadata
  getEnvelope<T = any>(domain: AssetDomain, market?: string): LiveSnapshotEnvelope<T> | null;

  // Push-based real-time listener update ingestion
  updateSnapshot(domain: AssetDomain, data: any, market?: string): void;

  // Authoritative on-demand fetch with single-flight promise pooling and stale fallbacks
  fetchFresh<T = any>(domain: AssetDomain, market?: string, forceRefresh?: boolean): Promise<T | null>;

  // Pre-seed active in-memory cache from persistent disk storage on server boot
  warm(): Promise<void>;
}

// Contract interface for the decoupled QuotesService
export interface IQuotesServiceContract {
  // Resolves an array of multi-asset identifiers directly from LiveSnapshotCoordinator
  getQuotes(items: string[]): Promise<{
    quotes: Record<string, any>;
    unresolved: string[];
    timestamp: string;
  }>;
}
