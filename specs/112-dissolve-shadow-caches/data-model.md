# Data Model & Architecture Design: Dissolve Shadow Caches

**Feature Branch**: `112-dissolve-shadow-caches`  
**Date**: 2026-11-15  
**Status**: Completed  

---

## 1. System Architecture: Before vs. After

### 1.1 Before: Fragmented Shadow Caches & Duplicate Fallbacks

```mermaid
flowchart TD
  subgraph QuotesService["QuotesService (6 Dependencies)"]
    Q[QuotesService]
  end

  subgraph DomainServices["Domain Query Services (Private Shadow Caches)"]
    SQS["StockQueryService (cachedPrices)"]
    GQS["GoldQueryService (cachedPrices)"]
    GAS["GasQueryService (cachedPrices)"]
    IQS["IndicesQueryService (cachedPrices)"]
    RQS["RatesSnapshotService"]
    GCCS["GccStockQueryService (cachedPrices)"]
  end

  subgraph FirestoreDB["Direct Firestore Queries (Bypassing Single-Flight)"]
    FS1[("stock_prices/latest")]
    FS2[("gold_prices/latest")]
    FS3[("gas_prices/latest")]
    FS4[("index_prices/latest")]
  end

  subgraph LSC["LiveSnapshotCoordinator"]
    MemStore["Memory Store (Map)"]
    DiskStore["Disk Persistence"]
  end

  Q -->|injected| SQS
  Q -->|injected| GQS
  Q -->|injected| IQS
  Q -->|injected| RQS
  Q -->|injected| GCCS
  Q -->|injected| LSC

  SQS -->|uncoordinated get| FS1
  GQS -->|uncoordinated get| FS2
  GAS -->|uncoordinated get| FS3
  IQS -->|uncoordinated get| FS4

  SQS -.->|also push update| LSC
  GQS -.->|also push update| LSC
  GAS -.->|also push update| LSC
  IQS -.->|also push update| LSC
```

### 1.2 After: Strict Single Authority in LiveSnapshotCoordinator

```mermaid
flowchart TD
  subgraph ClientConsumers["Client / API Layer"]
    BatchRoutes["/v1/quotes"]
    DomainRoutes["/v1/stocks, /v1/gold, /v1/gas, /v1/indices"]
  end

  subgraph ThinPresenters["Stateless Domain Query Presenters"]
    Q[QuotesService]
    SQS[StockQueryService]
    GQS[GoldQueryService]
    GAS[GasQueryService]
    IQS[IndicesQueryService]
  end

  subgraph AuthoritativeCoordinator["LiveSnapshotCoordinator (Single Source of Truth)"]
    SingleFlight["Single-Flight Promise Pooling (inFlightFetches)"]
    MemStore["Volatile In-Memory Cache (RAM)"]
    DiskStore["Persistent Disk Snapshot Fallback"]
    UpstreamFetch["Central Upstream Firestore Reader"]
  end

  BatchRoutes --> Q
  DomainRoutes --> SQS
  DomainRoutes --> GQS
  DomainRoutes --> GAS
  DomainRoutes --> IQS

  Q -->|100% Quote Reads| AuthoritativeCoordinator
  SQS -->|fetchFresh / getSnapshot| AuthoritativeCoordinator
  GQS -->|fetchFresh / getSnapshot| AuthoritativeCoordinator
  GAS -->|fetchFresh / getSnapshot| AuthoritativeCoordinator
  IQS -->|fetchFresh / getSnapshot| AuthoritativeCoordinator

  AuthoritativeCoordinator --> SingleFlight
  SingleFlight --> MemStore
  SingleFlight --> UpstreamFetch
  SingleFlight --> DiskStore
```

---

## 2. Service Contracts & State Definitions

### 2.1 Refactored `QuotesService` Interface

```ts
export interface IQuotesService {
  /**
   * Resolves an array of arbitrary multi-asset identifiers into structured quotes.
   * Reads 100% from LiveSnapshotCoordinator with zero domain service dependencies.
   */
  getQuotes(items: string[]): Promise<BatchQuoteResponse>;
}

export class QuotesService implements IQuotesService {
  /**
   * Constructor accepts ONLY the authoritative LiveSnapshotCoordinator.
   */
  constructor(
    private readonly coordinator: LiveSnapshotCoordinator = liveSnapshotCoordinator
  ) {}
}
```

### 2.2 Refactored Query Services Interface Contracts

#### `GoldQueryService`
```ts
export class GoldQueryService {
  // ZERO private cachedPrices field!

  updateCache(data: GoldPriceDocument): void {
    liveSnapshotCoordinator.updateSnapshot("gold", data);
  }

  getCachedPrices(): GoldPriceDocument | null {
    return liveSnapshotCoordinator.getSnapshot<GoldPriceDocument>("gold");
  }

  async fetchLatestGoldPrices(forceRefresh = false): Promise<GoldPriceDocument | null> {
    return liveSnapshotCoordinator.fetchFresh<GoldPriceDocument>("gold", undefined, forceRefresh);
  }
}
```

#### `GasQueryService`
```ts
export class GasQueryService {
  // ZERO private cachedPrices field!

  updateCache(data: GasPriceSnapshot): void {
    liveSnapshotCoordinator.updateSnapshot("gas", data);
  }

  getCachedPrices(): GasPriceSnapshot | null {
    return liveSnapshotCoordinator.getSnapshot<GasPriceSnapshot>("gas");
  }

  async fetchLatestGasPrices(forceRefresh = false): Promise<GasPriceSnapshot | null> {
    return liveSnapshotCoordinator.fetchFresh<GasPriceSnapshot>("gas", undefined, forceRefresh);
  }
}
```

#### `StockQueryService`
```ts
export class StockQueryService {
  // ZERO private cachedPrices field!

  updateCache(data: StockPriceSnapshot): void {
    liveSnapshotCoordinator.updateSnapshot("stocks", data, "egx");
  }

  getCachedPrices(): StockPriceSnapshot | null {
    return liveSnapshotCoordinator.getSnapshot<StockPriceSnapshot>("stocks", "egx");
  }

  async fetchLatestStockPrices(forceRefresh = false): Promise<StockPriceSnapshot | null> {
    return liveSnapshotCoordinator.fetchFresh<StockPriceSnapshot>("stocks", "egx", forceRefresh);
  }

  // Preserved static catalog filtering utility
  filterStocksByIndexOrGroup(stocks: Record<string, any>, filterTarget?: string): Record<string, any> {
    return marketCatalog.filterStocks(stocks, filterTarget);
  }
}
```

#### `IndicesQueryService`
```ts
export class IndicesQueryService {
  // ZERO private cachedPrices or cachedRegionalPrices fields!

  updateCache(data: IndexPriceSnapshot): void {
    liveSnapshotCoordinator.updateSnapshot("indices", data, "egx");
  }

  getCachedPrices(market = "egx"): IndexPriceSnapshot | null {
    return liveSnapshotCoordinator.getSnapshot<IndexPriceSnapshot>("indices", market);
  }

  async fetchLatestIndexPrices(
    marketOrForceRefresh: string | boolean = "egx",
    forceRefresh = false
  ): Promise<IndexPriceSnapshot | null> {
    const isFirstArgBool = typeof marketOrForceRefresh === "boolean";
    const market = isFirstArgBool ? "egx" : marketOrForceRefresh;
    const effectiveForceRefresh = isFirstArgBool ? marketOrForceRefresh : forceRefresh;
    return liveSnapshotCoordinator.fetchFresh<IndexPriceSnapshot>("indices", market, effectiveForceRefresh);
  }

  async fetchConstituentStocks(code: string, forceRefresh = false): Promise<any | null> {
    // Resolves constituent live stocks directly from coordinator
  }
}
```

---

## 3. Concurrency & Single-Flight State Transitions

```mermaid
sequenceDiagram
  autonumber
  participant Client1 as Client 1 (Concurrent)
  participant Client2 as Client 2 (Concurrent)
  participant QueryService as StockQueryService / GoldQueryService
  participant Coordinator as LiveSnapshotCoordinator
  participant Firestore as Firestore ("latest" Doc)

  Client1 ->> QueryService: fetchLatestGoldPrices()
  Client2 ->> QueryService: fetchLatestGoldPrices()
  QueryService ->> Coordinator: fetchFresh("gold")
  Note over Coordinator: Key "gold" not in RAM, inFlight.has("gold") == false
  Coordinator ->> Coordinator: Create fetchPromise & set inFlight("gold")
  QueryService ->> Coordinator: fetchFresh("gold")
  Note over Coordinator: inFlight.has("gold") == true! Coalesce onto existing promise!
  Coordinator ->> Firestore: db.collection("gold_prices").doc("latest").get() (1 read!)
  Firestore -->> Coordinator: Document Snapshot Data
  Coordinator ->> Coordinator: updateSnapshot("gold", data) & delete inFlight("gold")
  Coordinator -->> QueryService: Return Fresh Snapshot Data
  QueryService -->> Client1: 200 OK (Fresh Rates)
  QueryService -->> Client2: 200 OK (Fresh Rates)
```
