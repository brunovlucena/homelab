# Google Cloud CDN Setup Guide

This guide explains how to use Google Cloud CDN (free tier) to serve static assets for your homepage.

## Overview

The homepage is now configured to use Google Cloud CDN for serving static assets like images, which provides:
- **Global CDN delivery** for faster load times worldwide
- **Free tier benefits**: 5 GB storage, 1 GB egress per month
- **Automatic fallback** to local assets if CDN is not configured
- **Easy configuration** via environment variables

## Quick Setup

### 1. Automatic Setup (Recommended)

Run the Makefile target to automatically configure Google Cloud Storage and CDN:

```bash
export GCP_PROJECT_ID=your-project-id
export GCP_BUCKET_NAME=lucena-cloud-assets  # Optional
make setup-gcp-cdn
```

The script will:
- Create a Google Cloud Storage bucket
- Configure public read access
- Set up CORS
- Upload assets from `public/assets/`

### 2. Manual Setup

If you prefer to set it up manually:

#### Step 1: Create Google Cloud Storage Bucket

```bash
# Set your project
gcloud config set project YOUR_PROJECT_ID

# Create bucket
gsutil mb -p YOUR_PROJECT_ID -c STANDARD -l us-central1 gs://YOUR_BUCKET_NAME

# Make it publicly readable
gsutil iam ch allUsers:objectViewer gs://YOUR_BUCKET_NAME

# Configure CORS
cat > cors.json << EOF
[
  {
    "origin": ["*"],
    "method": ["GET", "HEAD"],
    "responseHeader": ["Content-Type", "Access-Control-Allow-Origin"],
    "maxAgeSeconds": 3600
  }
]
EOF
gsutil cors set cors.json gs://YOUR_BUCKET_NAME
```

#### Step 2: Upload Assets

```bash
# Upload assets to bucket
gsutil -m rsync -r ./public/assets gs://YOUR_BUCKET_NAME/assets
```

#### Step 3: Configure Environment Variable

Create or update `.env` file:

```env
VITE_CDN_BASE_URL=https://storage.googleapis.com/YOUR_BUCKET_NAME
```

#### Step 4: Rebuild Application

```bash
npm run build
```

## Advanced: Cloud CDN with Load Balancer

For better performance and custom domains, set up Cloud CDN with a load balancer:

1. **Create Backend Bucket**:
   ```bash
   gcloud compute backend-buckets create cdn-backend-bucket \
     --gcs-bucket-name=YOUR_BUCKET_NAME \
     --enable-cdn
   ```

2. **Create URL Map**:
   ```bash
   gcloud compute url-maps create cdn-url-map \
     --default-backend-bucket=cdn-backend-bucket
   ```

3. **Create HTTP(S) Load Balancer**:
   ```bash
   gcloud compute target-http-proxies create cdn-proxy \
     --url-map=cdn-url-map
   
   gcloud compute forwarding-rules create cdn-forwarding-rule \
     --global \
     --target-http-proxy=cdn-proxy \
     --ports=80
   ```

4. **Use Load Balancer IP in `.env`**:
   ```env
   VITE_CDN_BASE_URL=http://LOAD_BALANCER_IP
   ```

## How It Works

### Asset URL Generation

The `getAssetUrl()` utility function in `src/utils/index.ts` handles CDN URL generation:

```typescript
import { getAssetUrl } from '../utils'

// In your components
<img src={getAssetUrl('assets/eu.png')} alt="Profile" />
```

This will return:
- CDN URL if `VITE_CDN_BASE_URL` is set: `https://storage.googleapis.com/bucket/assets/eu.png`
- Relative path if not configured: `./assets/eu.png`

### Environment Variable

The CDN base URL is read from `VITE_CDN_BASE_URL` environment variable at build time. 

**Important**: This is a Vite environment variable, so it must be:
- Prefixed with `VITE_`
- Available at build time (not runtime)
- Set in `.env` file or build environment

## Free Tier Limits

Google Cloud free tier includes:
- **5 GB** of storage per month
- **1 GB** of egress (outbound data transfer) per month
- After free tier: $0.026/GB storage, $0.12/GB egress

Monitor usage in Google Cloud Console to avoid unexpected charges.

## Updating Assets

To update assets on the CDN:

```bash
# Sync local assets to bucket (adds, updates, removes)
gsutil -m rsync -r -d ./public/assets gs://YOUR_BUCKET_NAME/assets

# Or upload specific file
gsutil cp ./public/assets/image.png gs://YOUR_BUCKET_NAME/assets/
```

## Troubleshooting

### Assets not loading from CDN

1. Check that `VITE_CDN_BASE_URL` is set in `.env`
2. Verify assets are uploaded to bucket:
   ```bash
   gsutil ls -r gs://YOUR_BUCKET_NAME/assets
   ```
3. Test CDN URL directly in browser
4. Check CORS configuration if loading from different domain

### CORS Errors

If you see CORS errors, ensure CORS is configured:
```bash
gsutil cors get gs://YOUR_BUCKET_NAME
```

### Build Issues

Ensure environment variable is set before building:
```bash
export VITE_CDN_BASE_URL=https://storage.googleapis.com/YOUR_BUCKET_NAME
npm run build
```

## Cost Optimization

1. **Use Cloud CDN**: Reduces origin egress costs
2. **Enable compression**: Reduces transfer sizes
3. **Set cache headers**: Reduces requests to origin
4. **Monitor usage**: Set up billing alerts in GCP Console

## References

- [Google Cloud Storage Documentation](https://cloud.google.com/storage/docs)
- [Cloud CDN Documentation](https://cloud.google.com/cdn/docs)
- [Pricing Calculator](https://cloud.google.com/products/calculator)
