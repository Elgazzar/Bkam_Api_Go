// Contract definition for Firestore Storage Adapter.
// Pure stateless document reader mediating low-level Firestore reads without caching or schedule dependencies.

export interface IFirestoreHistoryAdapter {
  /**
   * Reads a raw weekly consolidated bucket document from Firestore.
   * Returns a map of date -> daily snapshot, or null if the document does not exist.
   * Pure stateless read: does not perform negative caching, disk writes, or memory caching.
   */
  fetchWeeklyBucketDocument(
    collectionName: string,
    weekKey: string
  ): Promise<Record<string, any> | null>;

  /**
   * Reads a legacy daily document from Firestore by collection name and date key.
   * Used strictly as a fallback for pre-migration archives prior to weekly bucket aggregation.
   * Pure stateless read: returns the raw document data or null.
   */
  fetchDailyDocument(
    collectionName: string,
    date: string
  ): Promise<any | null>;
}
