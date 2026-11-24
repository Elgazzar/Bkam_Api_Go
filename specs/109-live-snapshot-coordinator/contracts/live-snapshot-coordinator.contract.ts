// Contract Interface: LiveSnapshotCoordinator
// Feature 109 - Deep Live Snapshot In-Memory Storage

export type AssetDomain = "rates" | "gold" | "gas" | "stocks" | "indices" | "gcc-stocks";

export interface LiveSnapshotEnvelope<T = any> {
  // Asset category domain identifier
  domain: AssetDomain;
  // Optional regional market identifier
  market?: string;
  // Currency denomination
  currency: string;
  // Upstream official timestamp
  updatedAt: string;
  // Scraper retrieval timestamp
  scrapedAt: string;
  // Freshness classification
  freshness: "current" | "stale" | "degraded";
  // Trace identifier
  sourceTraceId: string;
  // Domain data payload
  data: T;
}

export interface ILiveSnapshotCoordinator {
  /**
   * Synchronous non-blocking in-memory retrieval of the active market price snapshot.
   * Resolves in <0.01ms purely from RAM without database I/O.
   *
   * @param domain Target asset category (rates, gold, gas, stocks, indices, gcc-stocks)
   * @param market Optional regional market identifier (e.g. "egx", "saudi", "uae", "qatar", "kuwait")
   * @returns Active snapshot payload or null if uninitialized
   */
  getSnapshot<T = any>(domain: AssetDomain, market?: string): T | null;

  /**
   * Centralized push ingestion endpoint for real-time Firestore listeners.
   * Atomically updates in-memory RAM storage and asynchronously writes to disk persistence.
   *
   * @param domain Target asset category
   * @param data Freshly received snapshot data from Firestore onSnapshot listener
   * @param market Optional regional market identifier
   */
  updateSnapshot(domain: AssetDomain, data: any, market?: string): void;

  /**
   * Authoritative on-demand fetch from upstream persistent store.
   * Utilizes in-flight promise coalescing to protect against thundering herds,
   * with automatic fallback to stale/persisted snapshots on network timeouts.
   *
   * @param domain Target asset category
   * @param market Optional regional market identifier
   * @param forceRefresh When true, bypasses in-memory caching and forces upstream query
   * @returns Fresh or degraded snapshot payload, or null on total unrecoverable miss
   */
  fetchFresh<T = any>(domain: AssetDomain, market?: string, forceRefresh?: boolean): Promise<T | null>;

  /**
   * Pre-seeds active in-memory cache from persistent disk storage on process launch.
   * Ensures sub-500ms cold boots before Firestore WebSocket listeners connect.
   */
  warm(): Promise<void>;
}
