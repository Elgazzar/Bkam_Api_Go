# deploy-api.ps1
# -----------------------------------------------------------------------------
# API DEPLOYMENT SCRIPT
# This script handles the build and deployment process SPECIFICALLY for the API 
# Gateway service.
#
# Requirements:
# - Docker Desktop installed and running
# - Authenticated to GitHub Container Registry (ghcr.io)
#
# DO NOT use this script to deploy the Scraper service. See deploy-scraper.ps1 
# for the scraper deployment pipeline.
# -----------------------------------------------------------------------------

Write-Host "==============================================" -ForegroundColor Cyan
Write-Host " STARTING API DEPLOYMENT" -ForegroundColor Cyan
Write-Host " Target: ghcr.io/elgazzar/egp-scraper-api:latest" -ForegroundColor Cyan
Write-Host "==============================================" -ForegroundColor Cyan

Write-Host "`n[1/2] Building API Docker image from ./api directory..." -ForegroundColor Yellow
docker build -t ghcr.io/elgazzar/egp-scraper-api:latest ./api
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: API Docker build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "`n[2/2] Pushing API Docker image to GitHub Container Registry..." -ForegroundColor Yellow
docker push ghcr.io/elgazzar/egp-scraper-api:latest
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: API Docker push failed! Ensure you are logged in to ghcr.io." -ForegroundColor Red
    exit 1
}

Write-Host "`n==============================================" -ForegroundColor Green
Write-Host " API DEPLOYMENT COMPLETE" -ForegroundColor Green
Write-Host "==============================================" -ForegroundColor Green
