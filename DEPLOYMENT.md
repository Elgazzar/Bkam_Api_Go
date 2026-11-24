# OCI Deployment Guide: Bkam Fastify API Gateway

This guide explains how to deploy the **Bkam API Gateway** container to an **Oracle Cloud Infrastructure (OCI)** Compute Instance, configured to serve exchange rates, gold prices, fuel prices, and regional stock quotes directly from your production **Google Firestore** database.

> [!NOTE]
> The API container runs as a lightweight Fastify gateway, querying Firestore for pre-populated multi-currency rates, equities, and commodities data. In-memory caching and flat-cache persistence guarantee **100% stability on OCI Free Tier (1GB RAM)** with zero memory leaks!

---

## 📋 Architectural Overview

```mermaid
graph TD
    A[Mobile & Web Clients] -- "HTTP Requests (Port 3001)" --> B[Bkam API Gateway (Fastify / OCI VM)]
    B -- "Reads Snapshots & History" --> C[(Google Firestore)]
    B -- "In-Memory & Flat-Cache" --> D[Local RAM / Disk Cache]
```

---

## 🛠️ Step 1: OCI Security List Configuration

Before deploying, ensure OCI allows public access to the Fastify API port (`3001`):

1. Log in to the **Oracle Cloud Console**.
2. Navigate to **Networking** > **Virtual Cloud Networks**.
3. Select your VCN and click on the **Public Subnet**.
4. Click on your **Default Security List** and select **Add Ingress Rules**:
   - **Source Type**: CIDR
   - **Source CIDR**: `0.0.0.0/0`
   - **IP Protocol**: TCP
   - **Destination Port Range**: `3001`
   - **Description**: Bkam Currency & Equities Fastify API Gateway

---

## 🔑 Step 2: Firestore Service Account Credentials

The OCI API container requires a Google Service Account Key to authenticate with Firestore:

1. In your **Firebase Console**, go to **Project Settings** > **Service accounts**.
2. Click **Generate new private key** and download the JSON file.
3. Rename the file to `firebase-service-account.json`.

---

## 💻 Step 3: Build & Push the API Container

Due to OCI Free Tier's 1GB RAM limits, **always compile and build the Docker image on your local machine** and push it to a registry, rather than building it on the VM.

1. Open a terminal in the `Bkam_Api` repository root:
   ```bash
   # 1. Compile the TypeScript build
   npm run build

   # 2. Build the Docker image
   docker build -t m7moudelgazzar/egp-scraper-api:latest .

   # 3. Push to your Docker Hub registry
   docker push m7moudelgazzar/egp-scraper-api:latest
   ```

---

## 🚀 Step 4: Deployment Options on the Oracle Cloud VM

You can deploy the API container using either the standard Command Line (CLI) or the Portainer Web UI dashboard. Choose the option that fits your workflow.

---

### Option A: Manual Shell Deployment (Command Line)

SSH into your Oracle Cloud VM instance:
```bash
ssh -i <your-private-key> ubuntu@<your-instance-ip>
```

#### 1. Install Docker and Compose (if not already done)
```bash
sudo apt update && sudo apt install -y docker.io docker-compose
sudo systemctl enable --now docker
sudo usermod -aG docker $USER
# Log out and log back in to apply group changes!
```

#### 2. Prepare the Project Directory
On the OCI instance, create a clean directory for your deployment stack:
```bash
mkdir -p ~/egp-api
cd ~/egp-api
```

#### 3. Upload the Service Account Credentials
Choose **one** of the following methods to transfer the credentials file to the VM:

* **Method A: Copy & Paste via Terminal (Easiest & Quickest!)**
  1. Open your local `firebase-service-account.json` file in any text editor and copy its entire text content.
  2. In your SSH session on the VM, run:
     ```bash
     nano ~/egp-api/firebase-service-account.json
     ```
  3. Paste the copied text into the terminal (Right-click or `Shift+Insert` on Windows).
  4. Save and close (`Ctrl+O`, `Enter`, `Ctrl+X`).

* **Method B: Transfer via SCP (Command Line from Local Machine)**
  Open a new Terminal/PowerShell window on your **local machine** (do not SSH first) and run:
  ```bash
  scp -i <path-to-your-private-key> <path-to-local-firebase-service-account.json> ubuntu@<your-instance-ip>:~/egp-api/firebase-service-account.json
  ```
  *(Example: `scp -i C:\Keys\oracle.key D:\firebase-credentials.json ubuntu@129.146.12.34:~/egp-api/firebase-service-account.json`)*

#### 4. Fetch the Production Docker Compose Configuration
Since the production configuration is version-controlled in Git as `docker-compose.prod.yml`, simply run this command on the VM to fetch the configuration directly:
```bash
curl -sSL https://raw.githubusercontent.com/Elgazzar/Bkam_Api/master/docker-compose.prod.yml -o docker-compose.yml
```

#### 5. Start the API Container
Ensure the shared network `app-net` is created and run the stack:
```bash
docker network create app-net 2>/dev/null || true
docker-compose up -d
```

---

### Option B: Portainer Stack Deployment (Web UI Dashboard)

If you prefer managing your containers visually using Portainer, you can choose either direct Git integration (recommended) or manual copy-paste:

#### 1. Store Credentials on the VM Host File System
Since Portainer stacks run Docker containers on the VM host, you must first upload the credentials file to the VM so Portainer can bind-mount it:
1. SSH into your VM and create the directory:
   ```bash
   mkdir -p /home/ubuntu/egp-api
   ```
2. Upload or paste your `firebase-service-account.json` file to:
   `/home/ubuntu/egp-api/firebase-service-account.json`

   * **Method A: Copy & Paste via Terminal (Easiest & Quickest!)**
     1. Open your local `firebase-service-account.json` file in any text editor and copy its entire text content.
     2. In your SSH session on the VM, run:
        ```bash
        nano ~/egp-api/firebase-service-account.json
        ```
     3. Paste the copied text into the terminal (Right-click or `Shift+Insert` on Windows).
     4. Save and close (`Ctrl+O`, `Enter`, `Ctrl+X`).

#### 2. Deploy the Stack in Portainer

##### Method A: Pull from Git (Recommended — Fully Automated!)
1. Open your **Portainer Dashboard** in your browser and select your environment (usually **local**).
2. Navigate to **Stacks** > **Add stack**.
3. Give your stack a name: `egp-api-stack`.
4. Under **Build method**, select **Repository**.
5. Fill in the repository details:
   * **Repository URL**: `https://github.com/Elgazzar/Bkam_Api.git`
   * **Repository reference**: `refs/heads/master`
   * **Compose path**: `docker-compose.prod.yml`
6. **Git Authentication** *(if private)*: Turn ON and input your GitHub Username and Personal Access Token (PAT) as the password.
7. **🔄 Automatic Updates**: Turn ON and set it to **Polling** (e.g., check every 5 minutes) to automatically pull any updates you push to Git!
8. Scroll down and click **Deploy the stack**.

##### Method B: Web Editor (Alternative Copy-Paste)
1. Open your **Portainer Dashboard** in your browser.
2. Navigate to **Stacks** > **Add stack** and select **Web editor** as the build method.
3. Paste the contents of `docker-compose.prod.yml` directly into the editor:
   ```yaml
   version: '3'

   services:
     api:
       image: m7moudelgazzar/egp-scraper-api:latest
       container_name: egp-scraper-api
       restart: always
       ports:
         - "3001:3001"
       volumes:
         - /home/ubuntu/egp-api/firebase-service-account.json:/app/firebase-service-account.json
       environment:
         - PORT=3001
         - GOOGLE_CLOUD_PROJECT=bekamfeelbank
         - GOOGLE_APPLICATION_CREDENTIALS=/app/firebase-service-account.json
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
4. Scroll down and click **Deploy the stack**.

---

## 🔄 How to Redeploy Updates

Whenever you make updates to the API codebase in the future, first build and push the new container image from your **local machine** (make sure you are in the **Bkam_Api root directory** `D:\Repo\Bkam\Bkam_Api`):
```bash
# Run from repository root directory:
npm run build
docker build -t m7moudelgazzar/egp-scraper-api:latest .
docker push m7moudelgazzar/egp-scraper-api:latest
```

Then, apply the updates using one of your deployment options:

### If using Option A (Command Line)
SSH into your Oracle Cloud VM and run:
```bash
cd ~/egp-api
docker-compose pull api
docker-compose up -d --build api
```

### If using Option B (Portainer Dashboard)
1. Go to your **Portainer Dashboard**.
2. Navigate to **Stacks** > **egp-api-stack**.
3. Click the **Editor** tab at the top.
4. Turn on the **Pull latest image** toggle switch (located directly above the Update button).
5. Click **Update the stack**. Portainer will pull the updated image and restart the container seamlessly!

---

## 🧪 Verifying the API

To verify that the API is up and running, query its health check and rates endpoints:
```bash
# Health check
curl http://localhost:3001/v1/health

# Exchange rates
curl http://localhost:3001/v1/rates
```

The server will return HTTP 200 with active exchange rates across all banks—including full nested multi-currency rates (`USD`, `EUR`, `GBP`, `SAR`, etc.) served directly from in-memory snapshots!
