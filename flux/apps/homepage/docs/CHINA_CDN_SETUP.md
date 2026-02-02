# 🇨🇳 China CDN Setup Guide - Alibaba Cloud
## Complete Implementation for Chinese Market

This guide walks you through setting up Alibaba Cloud CDN (OSS + CDN) to serve your homepage assets to Chinese users with optimal performance.

---

## 📋 Overview

### Why Alibaba Cloud CDN?

- ✅ **Excellent China Coverage**: 2800+ edge nodes in China
- ✅ **Affordable**: Pay-as-you-go pricing (~$0.05-0.10/GB)
- ✅ **Fast**: <50ms latency within China
- ✅ **Reliable**: 99.9% uptime SLA
- ✅ **Compatible**: Works alongside Cloudflare and Google CDN

### Architecture

```
Chinese User Request
    ↓
Alibaba Cloud CDN (China Edge)
    ├─ Cache HIT → Serve from Alibaba Edge ✅ (<50ms)
    └─ Cache MISS → Alibaba OSS (China Region) → Origin (Brazil)
```

---

## 🚀 Step 1: Create Alibaba Cloud Account

### 1.1 Sign Up

1. Go to: https://www.alibabacloud.com/
2. Click **"Free Trial"** or **"Sign Up"**
3. Complete registration (requires phone verification)
4. **Note**: You may need to provide business information for some services

### 1.2 Activate OSS Service

1. Log in to Alibaba Cloud Console
2. Go to: **Products** → **Object Storage Service (OSS)**
3. Activate OSS (usually instant)
4. **Free Tier**: 5GB storage, 5GB traffic/month (first 6 months)

---

## 🗂️ Step 2: Create OSS Bucket

### 2.1 Create Bucket via Console

1. Go to: **OSS Console** → **Buckets** → **Create Bucket**
2. Configure:
   - **Bucket Name**: `lucena-cloud-assets-cn` (must be globally unique)
   - **Region**: `China (Hangzhou)` or `China (Beijing)` (closest to users)
   - **Storage Class**: `Standard` (for frequent access)
   - **Access Control**: `Public Read` (for CDN)
   - **Versioning**: `Disabled` (unless needed)
   - **Server-Side Encryption**: `None` (for public assets)

3. Click **"OK"** to create

### 2.2 Create Bucket via CLI (Alternative)

```bash
# Install Aliyun CLI
# macOS
brew install aliyun-cli

# Or download from: https://www.alibabacloud.com/help/en/object-storage-service/latest/install-and-configure-aliyun-cli

# Configure credentials
aliyun configure

# Create bucket
aliyun oss mb oss://lucena-cloud-assets-cn \
  --region cn-hangzhou \
  --acl public-read
```

---

## 🔐 Step 3: Configure Access Control

### 3.1 Set Bucket to Public Read

1. Go to: **OSS Console** → **Buckets** → Select your bucket
2. Go to: **Access Control** → **Bucket Policy**
3. Add policy:
   ```json
   {
     "Version": "1",
     "Statement": [
       {
         "Effect": "Allow",
         "Principal": "*",
         "Action": "oss:GetObject",
         "Resource": "acs:oss:*:*:lucena-cloud-assets-cn/*"
       }
     ]
   }
   ```

### 3.2 Configure CORS

1. Go to: **OSS Console** → **Buckets** → **Data Management** → **Cross-Origin Resource Sharing (CORS)**
2. Click **"Create Rule"**
3. Configure:
   - **Source**: `*` (or your domain: `https://lucena.cloud`)
   - **Allowed Methods**: `GET`, `HEAD`
   - **Allowed Headers**: `*`
   - **Exposed Headers**: `ETag`, `Content-Length`
   - **Max Age**: `3600`

---

## 📤 Step 4: Upload Assets to OSS

### 4.1 Via Console (Small Files)

1. Go to: **OSS Console** → **Buckets** → Select bucket
2. Click **"Upload"**
3. Select files from `public/assets/`
4. Upload

### 4.2 Via CLI (Recommended for Large Files)

```bash
# Install ossutil
# macOS
brew install ossutil

# Or download from: https://www.alibabacloud.com/help/en/object-storage-service/latest/download-and-install-ossutil

# Configure
ossutil config

# Upload assets (parallel upload for speed)
ossutil cp -r ./public/assets/ oss://lucena-cloud-assets-cn/assets/ \
  --parallel 10 \
  --update \
  --checkpoint-dir=/tmp/oss-checkpoint

# Verify upload
ossutil ls oss://lucena-cloud-assets-cn/assets/
```

### 4.3 Via SDK (Automated)

Create script: `scripts/upload-to-aliyun-oss.sh`

```bash
#!/bin/bash
# Upload assets to Alibaba Cloud OSS

BUCKET_NAME="lucena-cloud-assets-cn"
ASSETS_DIR="./public/assets"
REGION="cn-hangzhou"

echo "📤 Uploading assets to Alibaba Cloud OSS..."

# Check if ossutil is installed
if ! command -v ossutil &> /dev/null; then
    echo "❌ ossutil not found. Install it first:"
    echo "   brew install ossutil"
    exit 1
fi

# Upload with parallel processing
ossutil cp -r "$ASSETS_DIR" "oss://$BUCKET_NAME/assets/" \
  --parallel 10 \
  --update \
  --checkpoint-dir=/tmp/oss-checkpoint

if [ $? -eq 0 ]; then
    echo "✅ Assets uploaded successfully!"
    echo "🌐 CDN URL: https://$BUCKET_NAME.oss-cn-hangzhou.aliyuncs.com"
else
    echo "❌ Upload failed!"
    exit 1
fi
```

---

## 🌐 Step 5: Enable CDN

### 5.1 Activate CDN Service

1. Go to: **Alibaba Cloud Console** → **CDN** → **Activate**
2. Complete activation (may require business verification)
3. **Free Tier**: 10GB traffic/month (first 6 months)

### 5.2 Add CDN Domain

1. Go to: **CDN Console** → **Domain Names** → **Add Domain**
2. Configure:
   - **Domain Type**: `OSS Domain` (if using OSS) or `Accelerate Domain`
   - **Acceleration Domain**: `cdn-cn.lucena.cloud` (subdomain for China CDN)
   - **Origin Domain**: Select your OSS bucket domain
   - **Origin Type**: `OSS Domain`
   - **Origin Domain**: `lucena-cloud-assets-cn.oss-cn-hangzhou.aliyuncs.com`
   - **Back-to-Origin Host**: `lucena-cloud-assets-cn.oss-cn-hangzhou.aliyuncs.com`

3. **Optional**: Use custom domain (requires DNS configuration)

### 5.3 Configure CDN Settings

1. **Cache Rules**:
   - **Static Files** (`.js`, `.css`, `.png`, `.webp`): `30 days`
   - **HTML**: `15 minutes`
   - **Default**: `1 day`

2. **Compression**:
   - Enable **Gzip** and **Brotli** compression
   - Compress: `text/*`, `application/javascript`, `application/json`

3. **HTTPS**:
   - Enable **HTTPS** (free SSL certificate available)
   - Force HTTPS redirect

4. **Cache Key**:
   - Include query string: `No` (for static assets)
   - Ignore case: `Yes`

---

## 🔧 Step 6: Update Application Code

### 6.1 Update Environment Variables

Add to `.env` file:

```env
# Google Cloud CDN (Global - except China)
VITE_CDN_BASE_URL=https://storage.googleapis.com/lucena-cloud-assets

# Alibaba Cloud CDN (China only)
VITE_CDN_BASE_URL_CN=https://lucena-cloud-assets-cn.oss-cn-hangzhou.aliyuncs.com
# Or if using custom CDN domain:
# VITE_CDN_BASE_URL_CN=https://cdn-cn.lucena.cloud
```

### 6.2 Update getAssetUrl() Function

The function has been updated to automatically detect China users and route to Alibaba Cloud CDN. See updated code in `src/utils/index.ts`.

### 6.3 Update Build Configuration

Update your deployment configuration to include the new environment variable:

```yaml
# k8s/kustomize/base/frontend-deployment.yaml
env:
  - name: VITE_CDN_BASE_URL
    value: "https://storage.googleapis.com/lucena-cloud-assets"
  - name: VITE_CDN_BASE_URL_CN
    value: "https://lucena-cloud-assets-cn.oss-cn-hangzhou.aliyuncs.com"
```

---

## 🧪 Step 7: Test from China

### 7.1 Test CDN URL Directly

```bash
# Test asset loading
curl -I https://lucena-cloud-assets-cn.oss-cn-hangzhou.aliyuncs.com/assets/eu.webp

# Should return:
# HTTP/2 200
# Content-Type: image/webp
# Access-Control-Allow-Origin: *
```

### 7.2 Test from China (VPN or Testing Service)

**Option A: Use VPN**
- Connect to China VPN
- Visit `https://lucena.cloud`
- Check browser DevTools → Network tab
- Verify assets load from Alibaba CDN

**Option B: Use Testing Service**
- https://www.17ce.com/ (Chinese site speed test)
- https://gtmetrix.com/ (with China location)
- Enter your domain and test

### 7.3 Verify Region Detection

1. Open browser console on your site
2. Check which CDN is being used:
   ```javascript
   // In browser console
   console.log('CDN Base URL:', import.meta.env.VITE_CDN_BASE_URL)
   console.log('CDN Base URL CN:', import.meta.env.VITE_CDN_BASE_URL_CN)
   ```

---

## 📊 Step 8: Monitor Performance

### 8.1 Alibaba Cloud Console Metrics

1. **CDN Console** → **Statistics**:
   - Bandwidth usage
   - Request count
   - Cache hit ratio (target: >85%)
   - Response time (target: <100ms p95)

2. **OSS Console** → **Statistics**:
   - Storage usage
   - Request count
   - Traffic usage

### 8.2 Set Up Billing Alerts

1. Go to: **Billing** → **Alerts**
2. Create alert:
   - **Alert Type**: `CDN Traffic`
   - **Threshold**: `10GB/month` (adjust based on usage)
   - **Notification**: Email/SMS

### 8.3 Monitor Costs

**Estimated Monthly Cost** (after free tier):
- **Storage**: 5GB × $0.012/GB = $0.06/month
- **CDN Traffic**: 10GB × $0.05/GB = $0.50/month
- **OSS Requests**: 10k requests × $0.01/10k = $0.01/month
- **Total**: ~$0.57/month (very affordable!)

---

## 🔄 Step 9: Automate Asset Sync

### 9.1 Create Sync Script

Create `scripts/sync-assets-to-aliyun.sh`:

```bash
#!/bin/bash
# Sync assets to Alibaba Cloud OSS (for CI/CD)

BUCKET_NAME="lucena-cloud-assets-cn"
ASSETS_DIR="./public/assets"
REGION="cn-hangzhou"

echo "🔄 Syncing assets to Alibaba Cloud OSS..."

# Check if ossutil is configured
if ! ossutil ls "oss://$BUCKET_NAME" &> /dev/null; then
    echo "❌ OSS not configured. Run: ossutil config"
    exit 1
fi

# Sync (only uploads changed files)
ossutil sync "$ASSETS_DIR" "oss://$BUCKET_NAME/assets/" \
  --delete \
  --update \
  --parallel 10

if [ $? -eq 0 ]; then
    echo "✅ Assets synced successfully!"
    
    # Purge CDN cache (optional)
    echo "🔄 Purging CDN cache..."
    # Note: Requires Alibaba Cloud API or console
    echo "   Go to CDN Console → Refresh Cache → Enter URLs"
else
    echo "❌ Sync failed!"
    exit 1
fi
```

### 9.2 Add to CI/CD Pipeline

```yaml
# .github/workflows/deploy.yml (example)
- name: Sync assets to Alibaba Cloud OSS
  run: |
    chmod +x scripts/sync-assets-to-aliyun.sh
    ./scripts/sync-assets-to-aliyun.sh
  env:
    ALIBABA_CLOUD_ACCESS_KEY_ID: ${{ secrets.ALIBABA_CLOUD_ACCESS_KEY_ID }}
    ALIBABA_CLOUD_ACCESS_KEY_SECRET: ${{ secrets.ALIBABA_CLOUD_ACCESS_KEY_SECRET }}
```

---

## 🚨 Troubleshooting

### Issue: Assets Not Loading from China

**Check**:
1. Verify OSS bucket is public read
2. Check CORS configuration
3. Verify CDN domain is configured correctly
4. Test OSS URL directly (bypass CDN)

**Solution**:
```bash
# Test OSS directly
curl -I https://lucena-cloud-assets-cn.oss-cn-hangzhou.aliyuncs.com/assets/eu.webp

# If 403, check bucket permissions
# If 404, verify file exists
```

### Issue: CORS Errors

**Check**:
1. CORS rules in OSS console
2. Allowed origins include your domain
3. Allowed methods include GET, HEAD

**Solution**:
- Update CORS rules in OSS console
- Ensure `Access-Control-Allow-Origin: *` is set

### Issue: High Costs

**Check**:
1. CDN traffic usage
2. OSS storage usage
3. Request count

**Solution**:
- Enable aggressive caching
- Optimize image sizes (WebP)
- Use Cloudflare for non-China traffic
- Set up billing alerts

### Issue: Region Detection Not Working

**Check**:
1. `getAssetUrl()` function is updated
2. Environment variables are set
3. Browser timezone/location detection

**Solution**:
- Use Cloudflare headers for better detection
- Fallback to user agent or IP geolocation
- Test with VPN from China

---

## 📈 Performance Expectations

### Before (No China CDN)
- **China Users**: 300-500ms latency
- **Cache Hit**: Unreliable (Cloudflare limited in China)
- **User Experience**: Poor

### After (With Alibaba Cloud CDN)
- **China Users**: <100ms latency ✅
- **Cache Hit**: >85% hit ratio ✅
- **User Experience**: Excellent ✅

### Improvement
- **Latency**: 200-400ms faster
- **Reliability**: 99.9% uptime
- **Cost**: ~$0.50-1.00/month

---

## 📝 Next Steps

1. ✅ Set up Alibaba Cloud account
2. ✅ Create OSS bucket and upload assets
3. ✅ Enable CDN and configure
4. ✅ Update application code
5. ✅ Test from China
6. ✅ Monitor performance and costs
7. ✅ Automate asset sync in CI/CD

---

## 🔗 Resources

- [Alibaba Cloud OSS Documentation](https://www.alibabacloud.com/help/en/object-storage-service)
- [Alibaba Cloud CDN Documentation](https://www.alibabacloud.com/help/en/cdn)
- [OSS Pricing](https://www.alibabacloud.com/product/oss/pricing)
- [CDN Pricing](https://www.alibabacloud.com/product/cdn/pricing)
- [OSS CLI (ossutil) Guide](https://www.alibabacloud.com/help/en/object-storage-service/latest/ossutil)

---

**Total Setup Time**: ~2 hours  
**Monthly Cost**: ~$0.50-1.00 (after free tier)  
**Performance Improvement**: 200-400ms faster for China users ✅
