# Portainer Deployment Guide: Bkam Go API Gateway

This guide covers deploying the **Bkam Financial API Gateway (Go)** container using **Portainer** on your server / VM. The API service exposes port **`3001`**, serves market data directly from high-performance in-memory cache (0ms queries), and connects to **PostgreSQL 17** with real-time `LISTEN / NOTIFY` cache invalidation.

---

## 📋 Architectural Overview

```mermaid
graph TD
    Client[HTTP Clients / Mobile App] -- "Port 3001" --> API[Bkam Go API Gateway]
    API -- "0ms In-Memory Read" --> RAM[sync.RWMutex Cache]
    API -- "LISTEN snapshot_update" --> DB[(PostgreSQL 17 / bkam_db)]
    Scraper[Bkam Scraper] -- "NOTIFY snapshot_update" --> DB
    API -- "Tier 2 Disk Snapshot" --> Volume[/app/var Volume Mount]
```

---

## ⚙️ Prerequisites

1. **Docker & Portainer** running on your target host.
2. **PostgreSQL 17 Container** deployed and running on the Docker bridge network `app-net` under the container/service name `bkam_db`.
3. **External Network**: Ensure the shared Docker network `app-net` exists:
   ```bash
   docker network create app-net 2>/dev/null || true
   ```
4. **OCI / Cloud Firewall**: Ensure ingress port **`3001`** (TCP) is open on your firewall / security list.

---

## 🚀 Portainer Stack Deployment

You can deploy the stack in Portainer using either **Method A: Direct Git Repository** (recommended for automated updates) or **Method B: Web Editor (Copy & Paste)**.

---

### Method A: Git Repository Deployment (Recommended)

1. Open your **Portainer Web Dashboard** (e.g. `https://<your-server-ip>:9443`).
2. Select your environment (usually **local**).
3. In the sidebar, navigate to **Stacks** > **Add stack**.
4. Set the **Name**: `bkam-api`
5. Under **Build method**, select **Repository**.
6. Configure the Git repository details:
   - **Repository URL**: `https://github.com/Elgazzar/Bkam_Api_Go.git` (or your repository clone URL)
   - **Repository reference**: `refs/heads/master`
   - **Compose path**: `docker-compose.prod.yml` (for x86_64 / AMD64) or `docker-compose.arm.yml` (for ARM64 / Ampere)
7. **Git Authentication** *(if private repo)*:
   - Toggle **Authentication** to **ON**.
   - **Username**: Your GitHub username.
   - **Personal Access Token**: Your GitHub Personal Access Token (PAT with `repo` read access).
8. **Automatic Updates**:
   - Toggle **Automatic updates** to **ON**.
   - Mechanism: **Polling** (e.g., fetch interval `5m`) or **Webhook** for instant CD triggers.
9. **Environment Variables**:
   Under **Environment variables**, click **Add environment variable** or paste the values:
   ```ini
   PORT=3001
   DB_USER=m7moudelgazzar
   DB_PASSWORD=Premium@2182
   DB_HOST=bkam_db
   DB_PORT=5432
   DB_NAME=bkam_db
   DB_SSLMODE=disable
   API_KEY=local-dev-key
   ```
10. Click **Deploy the stack**.

---

### Method B: Web Editor Deployment (Direct Compose YAML)

1. Open your **Portainer Web Dashboard** and go to **Stacks** > **Add stack**.
2. Set the **Name**: `bkam-api`
3. Under **Build method**, choose **Web editor**.
4. Paste the following production Docker Compose specification:

```yaml
version: '3.8'

services:
  api:
    image: ghcr.io/elgazzar/bkam-api:latest # Use :arm64 for ARM-based instances
    pull_policy: always
    container_name: bkam-api
    restart: always
    ports:
      - "3001:3001"
    volumes:
      # Persistent cache directory for cold-boot snapshots and historical weekly buckets
      - /home/ubuntu/egp-api/var:/app/var
      - /home/ubuntu/egp-api/cache:/app/.cache
    environment:
      - PORT=3001
      - DB_USER=m7moudelgazzar
      - DB_PASSWORD=Premium@2182
      - DB_HOST=bkam_db
      - DB_PORT=5432
      - DB_NAME=bkam_db
      - DB_SSLMODE=disable
      - API_KEY=local-dev-key
      - NODE_ENV=production
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    networks:
      - app-net

networks:
  app-net:
    external: true
    name: app-net
```

5. Click **Deploy the stack**.

---

## 🔄 How to Redeploy Updates

Whenever a new container image is pushed to GitHub Packages (`ghcr.io/elgazzar/bkam-api`):

1. Navigate to **Portainer** > **Stacks** > **bkam-api**.
2. Click on the **Editor** tab.
3. Toggle **Re-pull image** (or **Pull latest image**) to **ON**.
4. Click **Update the stack**.
   Portainer will pull the updated image and restart the container with zero state loss.

---

## 🧪 Verification & Health Checks

Once deployed, verify container health and endpoints on port `3001`:

```bash
# 1. Health Probe
curl -i http://localhost:3001/v1/health

# 2. Version Information
curl -i http://localhost:3001/v1/version

# 3. Foreign Exchange Rates (serves from memory cache in 0ms)
curl -i http://localhost:3001/v1/rates

# 4. Commodities Gold Prices
curl -i http://localhost:3001/v1/gold

# 5. Petroleum Fuel Prices
curl -i http://localhost:3001/v1/gas

# 6. Egyptian Stock Exchange (EGX)
curl -i http://localhost:3001/v1/stocks

# 7. Regional GCC Markets (e.g. Saudi Tadawul)
curl -i http://localhost:3001/v1/stocks/saudi
```

Expected HTTP response for health check:
```json
{
  "status": "healthy",
  "database": "connected",
  "uptime": "..."
}
```

---

## 🛠️ Troubleshooting

- **Container fails to start**: Check logs in Portainer: **Containers** > `bkam-api` > **Logs**.
- **PostgreSQL Connection Refused**: Verify that `bkam_db` is attached to `app-net`:
  ```bash
  docker network inspect app-net
  ```
  Ensure `DB_HOST=bkam_db` can be resolved by containers on `app-net`.
- **Database Notifications**: Confirm that PostgreSQL is receiving `LISTEN snapshot_update` subscriptions. When the scraper issues `NOTIFY snapshot_update`, check `bkam-api` logs for:
  ```text
  [PostgresListener] Real-time invalidation applied: domain=rates, sub_key=
  ```
