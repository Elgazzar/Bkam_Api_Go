# Quickstart & Verification Guide: Migrate Bkam Financial API Gateway to Idiomatic Go

**Feature**: `116-migrate-gateway-go`  
**Date**: 2026-11-16  

---

## 1. Prerequisites

- **Go SDK**: Go 1.23 or newer (`go version`).
- **Google Cloud Credentials**: `GOOGLE_APPLICATION_CREDENTIALS` pointing to valid Firebase/Firestore service account JSON, or local Firestore emulator configured via `FIRESTORE_EMULATOR_HOST`.
- **Node.js Environment**: Node.js 22+ (to run the 49-scenario differential verification suite against the Go service).

---

## 2. Configuration (`.env`)

The Go gateway reads environment variables matching the Node.js service:

```bash
PORT=3001
HOST=0.0.0.0
NODE_ENV=development
SNAPSHOT_DIR=./var/snapshots
API_KEY=local-dev-key
SERVER_ID=server-go-01
```

---

## 3. Building and Running

### Development Mode (with hot-reload)
```bash
# Using air (recommended for Go dev)
air

# Or standard go run
go run cmd/server/main.go
```

### Production Build
```bash
go build -ldflags="-s -w" -o dist/bkam-api cmd/server/main.go
./dist/bkam-api
```

---

## 4. Testing & Verification Sequence

### 4.1 Unit Tests with Race Detector
Verify thread safety, zero data races, and unit logic across all engines:
```bash
go test -v -race -cover ./...
```

### 4.2 Local Smoke Test
```bash
# Probe liveness
curl -i http://localhost:3001/v1/health

# Probe runtime version
curl -i http://localhost:3001/v1/version

# Query rates with API key
curl -i -H "x-api-key: local-dev-key" http://localhost:3001/v1/rates

# Test ETag revalidation (should return HTTP 304)
curl -i -H "x-api-key: local-dev-key" -H 'If-None-Match: W/"..."' http://localhost:3001/v1/rates
```

### 4.3 Running Mandatory Platform Verification Gate
Point the verification suite to the running Go instance on port 3001:
```bash
# Run 49-scenario verification suite
npm run verify

# Run 80-test differential parity comparison against base server
npm run verify:report
```

All 49 scenarios must pass with exit code `0` and P99 latency strictly under **15ms**.
