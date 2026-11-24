# Production Deployment Guide: Serverless Firestore Scraper

This guide outlines the step-by-step process for deploying the serverless scraper architecture to production Google Cloud and Firebase.

## Phase 1: Firebase & Firestore Setup

Before deploying the scraper, you must initialize the production database.

1. **Create a Firebase Project:**
   - Go to the [Firebase Console](https://console.firebase.google.com/).
   - Click **Add Project** and follow the prompts.
   - Note your **Project ID** (e.g., `bekamfeelbank` or whatever specific ID Firebase assigned to it).

2. **Enable Firestore:**
   - In the Firebase Console, navigate to **Build > Firestore Database**.
   - Click **Create Database**.
   - Select **Production Mode** and choose an **Always Free** tier region (`us-central1`, `us-east1`, or `us-west1`). *Note: Choosing one of these specific US regions is required to qualify for the Google Cloud Always Free tier for Firestore and Cloud Run.*

3. **Generate Service Account Credentials (For the API Server):**
   - Go to **Project Settings > Service Accounts**.
   - Click **Generate New Private Key**.
   - Save this JSON file securely (e.g., `firebase-service-account.json`). You will need this for your main API server VM to read from production.

---

## Phase 2: Deploying the Serverless Scraper to Cloud Run

The scraper is packaged as a Docker container. We will push it to Google Artifact Registry and deploy it to Cloud Run.

### Prerequisites
Make sure you have the [Google Cloud CLI (`gcloud`)](https://cloud.google.com/sdk/docs/install) installed and authenticated:
```bash
gcloud auth login
gcloud config set project bekamfeelbank
```

### 1. Build and Push the Docker Image

Enable Artifact Registry and create a repository:
```bash
gcloud services enable artifactregistry.googleapis.com
gcloud artifacts repositories create scraper-repo --repository-format=docker --location=us-east1
gcloud auth configure-docker us-east1-docker.pkg.dev
```

Build the container and push it:
```bash
cd serverless-scraper

# Build the image using our Playwright Dockerfile
docker build -t us-east1-docker.pkg.dev/bekamfeelbank/scraper-repo/scraper:latest .

# Push to Google Cloud
docker push us-east1-docker.pkg.dev/bekamfeelbank/scraper-repo/scraper:latest
```

### 2. Deploy to Cloud Run

Deploy the container as a private service (so it can't be triggered by the public internet, only by your scheduler).

```bash
gcloud run deploy serverless-scraper --image us-east1-docker.pkg.dev/bekamfeelbank/scraper-repo/scraper:latest --region us-east1 --memory 2Gi --cpu 1 --timeout 300s --no-allow-unauthenticated
```
*Note the Service URL that is output after a successful deployment.*
https://serverless-scraper-925311716041.us-east1.run.app

---
## Manual Trigger
```bash
gcloud scheduler jobs run trigger-scraper --location us-east1
```

## Phase 3: Setting Up the Cloud Scheduler Trigger

We need to trigger the scraper exactly every 5 minutes. 

### 1. Create a Service Account for the Scheduler
Create a dedicated identity that has permission to invoke your Cloud Run function.
```bash
gcloud iam service-accounts create scraper-invoker-sa --display-name "Scraper Invoker"

# Grant the invoker role on your Cloud Run service
gcloud run services add-iam-policy-binding serverless-scraper --region us-east1 --member=serviceAccount:scraper-invoker-sa@bekamfeelbank.iam.gserviceaccount.com --role=roles/run.invoker
```

### 2. Create the Cron Job
Set up Cloud Scheduler to ping the Cloud Run URL. 

To stay within the **Google Cloud Always Free tier** (180,000 vCPU-seconds/month), configure the trigger to run every 15 minutes only during Egyptian banking hours (Sunday to Thursday, 9:00 AM to 5:00 PM Cairo Time):

```bash
gcloud scheduler jobs create http trigger-scraper \
  --location=us-east1 \
  --schedule="*/15 9-17 * * 0-4" \
  --time-zone="Africa/Cairo" \
  --uri="https://serverless-scraper-925311716041.us-east1.run.app/scrape" \
  --http-method=POST \
  --headers="Content-Type=application/json" \
  --message-body='{"force": false}' \
  --oidc-service-account-email=scraper-invoker-sa@bekamfeelbank.iam.gserviceaccount.com
```

*(Note: If you run 24/7 or trigger every 5 minutes, the cumulative duration of the 14+ Playwright browser tasks will exceed the monthly compute free-tier limits, resulting in billing charges).*

---

## Phase 4: Deploying the Main API Server

Now that the scraper is populating the production Firestore database every 5 minutes, your main API server simply needs to read from it.

1. **Copy the Service Account Key:**
   Securely transfer the `firebase-service-account.json` (from Phase 1) to your API server VM.

2. **Set Environment Variables:**
   Update the `.env` file on your API server:
   ```env
   GOOGLE_CLOUD_PROJECT=bekamfeelbank
   GOOGLE_APPLICATION_CREDENTIALS=/absolute/path/to/firebase-service-account.json
   ```
   *(Ensure you **delete** `FIRESTORE_EMULATOR_HOST` from the production `.env`!)*

3. **Restart the API:**
   When the API starts up, it will use the service account to seamlessly read the live snapshots from your production Firestore database.

---
**Done!** Your architecture is now fully serverless and highly available.
