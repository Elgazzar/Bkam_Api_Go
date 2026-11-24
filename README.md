# Bkam API Gateway (Go)

High-performance, statically compiled Go API Gateway serving Egyptian foreign exchange rates, domestic and international gold prices, fuel prices, and regional Middle Eastern equity quotes (EGX, Tadawul, ADX, DFM, QSE, Boursa Kuwait).

Designed for sub-10ms latency, zero-read cloud database caching, and high availability.
Powered by a consolidated **LiveSnapshotCoordinator** in-memory storage engine providing sub-microsecond synchronous quote lookups, atomic lock management, and single-flight thundering herd protection.

---

## 🚀 Quickstart

### Prerequisites
- Go >= 1.26
- Docker (optional, for containerized execution)

### Development
```bash
# Run locally with hot reload / direct execution
go run ./cmd/server
```
Server starts by default on `http://localhost:3001`.

### Running Tests
```bash
# Run unit and integration tests across all packages
go test -v ./...
```

### Compiling Static Binary
```bash
# Build optimized Linux statically linked executable
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server
```

### Docker Multi-Stage Build
```bash
# Build native container (~20MB total size)
docker build -t bkam-api-go .

# Run container
docker run -p 3001:3001 -e PORT=3001 bkam-api-go
```

---

## 🏛️ Architecture & Principles

- **Framework**: `chi/v5` lightweight router with idiomatic HTTP middleware stack.
- **Principle V**: 100% line comment coverage on every line of Go source code.
- **Cache Architecture**:
  - **Tier 1 (RAM)**: In-memory live snapshots and active weekly buckets via `sync.RWMutex`.
  - **Tier 2 (Disk)**: Flatted JSON weekly bucket caching on persistent storage.
  - **Tier 3 (Cloud)**: Google Cloud Firestore real-time snapshot listeners.

---

## 📖 API Documentation

Comprehensive endpoint specifications, request parameters, and response schemas are detailed in:
- [API Endpoints Guide](./API_ENDPOINTS_GUIDE.md)
- [API Usage & Integration Guide](./API_USAGE.md)
- [Deployment Guide](./DEPLOYMENT.md)
- [Oracle Cloud OCI Deployment](./ORACLE_CLOUD_API_DEPLOYMENT.md)
