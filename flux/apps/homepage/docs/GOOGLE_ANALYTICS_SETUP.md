# Google Analytics Setup Guide

This guide explains how to set up Google Analytics for your homepage and integrate it with Grafana.

## Part 1: Get Your Google Analytics Measurement ID

### Step 1: Create or Access Your Google Analytics Property

1. Go to [Google Analytics](https://analytics.google.com)
2. Sign in with your Google account
3. If you don't have a property yet:
   - Click "Admin" (bottom left)
   - Click "Create Property"
   - Fill in the property details
   - Click "Create"

### Step 2: Create a Data Stream (if not already created)

1. In Admin, under the Property column, click "Data Streams"
2. Click "Add stream" → "Web"
3. Enter your website details:
   - **Website URL**: `https://lucena.cloud`
   - **Stream name**: `Homepage` (or any name you prefer)
4. Click "Create stream"

### Step 3: Copy Your Measurement ID

1. After creating the stream (or clicking on an existing stream), you'll see the stream details
2. Find the **"Measurement ID"** - it looks like `G-XXXXXXXXXX`
3. Copy this ID

**Example:**
```
Measurement ID: G-ABC123XYZ
```

## Part 2: Configure Your Homepage

### Option A: Using Environment Variable (Recommended)

1. Create a `.env` file in the `src/frontend` directory:
   ```bash
   cd flux/infrastructure/homepage/src/frontend
   touch .env
   ```

2. Add your Measurement ID:
   ```env
   VITE_GA_MEASUREMENT_ID=G-ABC123XYZ
   ```
   *(Replace `G-ABC123XYZ` with your actual Measurement ID)*

3. Rebuild the application:
   ```bash
   npm run build
   ```

### Option B: Direct Replacement in index.html

If you prefer not to use environment variables, you can directly replace the placeholder in `index.html`:

1. Open `src/frontend/index.html`
2. Find the lines with `__VITE_GA_MEASUREMENT_ID__`
3. Replace both occurrences with your actual Measurement ID

**Before:**
```html
<script async src="https://www.googletagmanager.com/gtag/js?id=__VITE_GA_MEASUREMENT_ID__"></script>
<script>
  gtag('config', '__VITE_GA_MEASUREMENT_ID__', {
```

**After:**
```html
<script async src="https://www.googletagmanager.com/gtag/js?id=G-ABC123XYZ"></script>
<script>
  gtag('config', 'G-ABC123XYZ', {
```

## Part 3: Set Up Grafana Google Analytics Datasource

The datasource ConfigMap has already been created with your service account credentials. Now you need to grant permissions:

### Step 1: Grant Service Account Access to Google Analytics

1. Go to [Google Analytics](https://analytics.google.com)
2. Click "Admin" (bottom left)
3. Under the **Account** column, click "Account Access Management"
4. Click the "+" button (top right) → "Add users"
5. Enter your service account email:
   ```
   homelab@homelab-481500.iam.gserviceaccount.com
   ```
6. Select the following roles:
   - **Viewer** (minimum required)
   - Or **Analyst** (recommended for full read access)
7. Uncheck "Notify new users by email" (service accounts don't need notifications)
8. Click "Add"

### Step 2: Enable Required APIs (if not already enabled)

The Google Analytics API is already enabled in your GCP project, but ensure these are also enabled:

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Select project: `homelab-481500`
3. Navigate to "APIs & Services" → "Library"
4. Search for and enable:
   - ✅ **Google Analytics API** (already enabled)
   - ✅ **Google Analytics Data API** - Enable this if not already enabled
   - ✅ **Google Analytics Admin API** - Enable this if not already enabled

### Step 3: Deploy the Datasource

The datasource ConfigMap is already configured at:
```
flux/infrastructure/prometheus-operator/k8s/datasources/google-analytics-datasource.yaml
```

To deploy:
1. Commit and push your changes
2. Flux will automatically deploy the ConfigMap
3. The Grafana sidecar will detect it and provision the datasource

### Step 4: Verify in Grafana

1. Go to [Grafana](https://grafana.lucena.cloud)
2. Navigate to "Connections" → "Data sources"
3. You should see "Google Analytics" in the list
4. Click on it to test the connection

## Troubleshooting

### Issue: "Permission denied" in Grafana

**Solution:** Make sure you added the service account email to your Google Analytics property with at least "Viewer" role.

### Issue: Measurement ID not replaced in build

**Solution:** 
- Ensure your `.env` file is in the correct location (`src/frontend/.env`)
- Verify the environment variable is set: `echo $VITE_GA_MEASUREMENT_ID`
- Make sure you ran `npm run build` after setting the environment variable

### Issue: Datasource not appearing in Grafana

**Solution:**
- Check if the ConfigMap was created: `kubectl get configmap grafana-datasource-google-analytics -n prometheus`
- Check Grafana sidecar logs: `kubectl logs -n prometheus deployment/kube-prometheus-stack-grafana -c grafana-sc-datasources`
- Wait a few minutes - the sidecar checks for new ConfigMaps periodically

## Verification

After setup, you should see:
1. ✅ Google Analytics tracking on your homepage (check browser Network tab for gtag requests)
2. ✅ Google Analytics datasource in Grafana
3. ✅ Ability to query Google Analytics data in Grafana dashboards

## Next Steps

- Create Grafana dashboards using the Google Analytics datasource
- Monitor homepage traffic, user behavior, and conversions
- Set up alerts based on analytics metrics
