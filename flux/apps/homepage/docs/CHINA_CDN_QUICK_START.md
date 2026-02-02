# 🇨🇳 China CDN Quick Start (30 Minutes)

Quick setup guide to get Alibaba Cloud CDN working for Chinese users.

---

## ✅ Prerequisites

- Alibaba Cloud account (sign up at https://www.alibabacloud.com/)
- Assets ready in `public/assets/` directory
- Basic knowledge of command line

---

## 🚀 Quick Setup (30 Minutes)

### Step 1: Create OSS Bucket (5 min)

1. **Login** to Alibaba Cloud Console
2. Go to: **Object Storage Service (OSS)** → **Buckets** → **Create Bucket**
3. Configure:
   - **Name**: `lucena-cloud-assets-cn`
   - **Region**: `China (Hangzhou)` or `China (Beijing)`
   - **Storage Class**: `Standard`
   - **Access Control**: `Public Read`
4. Click **OK**

### Step 2: Upload Assets (10 min)

**Option A: Via Console** (for small files)
1. Go to bucket → **Files** → **Upload**
2. Select all files from `public/assets/`
3. Upload

**Option B: Via CLI** (recommended)
```bash
# Install ossutil
brew install ossutil  # macOS
# Or download: https://www.alibabacloud.com/help/en/object-storage-service/latest/download-and-install-ossutil

# Configure (will prompt for credentials)
ossutil config

# Upload assets
ossutil cp -r ./public/assets/ oss://lucena-cloud-assets-cn/assets/ \
  --parallel 10 \
  --update
```

### Step 3: Configure CORS (2 min)

1. Go to: **OSS Console** → **Buckets** → Your bucket → **Data Management** → **CORS**
2. Click **"Create Rule"**
3. Configure:
   - **Source**: `*`
   - **Allowed Methods**: `GET`, `HEAD`
   - **Allowed Headers**: `*`
   - **Max Age**: `3600`
4. Click **OK**

### Step 4: Enable CDN (5 min)

1. Go to: **CDN Console** → **Activate** (if not already activated)
2. Go to: **Domain Names** → **Add Domain**
3. Configure:
   - **Acceleration Domain**: `cdn-cn.lucena.cloud` (or use OSS domain directly)
   - **Origin Domain**: `lucena-cloud-assets-cn.oss-cn-hangzhou.aliyuncs.com`
   - **Origin Type**: `OSS Domain`
4. Click **Submit** (may take 10-15 minutes to activate)

### Step 5: Update Environment Variables (3 min)

Add to your `.env` file:

```env
# Global CDN (Google Cloud)
VITE_CDN_BASE_URL=https://storage.googleapis.com/lucena-cloud-assets

# China CDN (Alibaba Cloud)
VITE_CDN_BASE_URL_CN=https://lucena-cloud-assets-cn.oss-cn-hangzhou.aliyuncs.com
```

### Step 6: Rebuild and Deploy (5 min)

```bash
cd flux/apps/homepage/src/frontend
npm run build

# Deploy as usual
make deploy
```

---

## 🧪 Test

### Test OSS Directly

```bash
curl -I https://lucena-cloud-assets-cn.oss-cn-hangzhou.aliyuncs.com/assets/eu.webp
```

Should return `200 OK` with CORS headers.

### Test from China

1. Use VPN to connect from China
2. Visit `https://lucena.cloud`
3. Open DevTools → Network tab
4. Check assets load from Alibaba CDN (URLs should contain `aliyuncs.com`)

---

## 📊 Verify It's Working

1. **Check CDN Console**:
   - Go to: **CDN Console** → **Statistics**
   - Should show traffic and requests

2. **Check Browser**:
   - Open site from China (or VPN)
   - Assets should load from `aliyuncs.com` domain
   - Latency should be <100ms

---

## 💰 Cost Estimate

**Free Tier** (first 6 months):
- Storage: 5GB free
- CDN Traffic: 10GB free/month

**After Free Tier**:
- Storage: ~$0.06/month (5GB)
- CDN Traffic: ~$0.50/month (10GB)
- **Total**: ~$0.56/month

---

## 🚨 Troubleshooting

### Assets Not Loading

1. **Check bucket is public**:
   ```bash
   ossutil stat oss://lucena-cloud-assets-cn/assets/eu.webp
   ```

2. **Check CORS**:
   - Go to OSS Console → CORS
   - Verify rule is active

3. **Test OSS URL directly**:
   ```bash
   curl https://lucena-cloud-assets-cn.oss-cn-hangzhou.aliyuncs.com/assets/eu.webp
   ```

### Region Detection Not Working

The code uses timezone detection as fallback. For better accuracy:
- Use Cloudflare headers (if available)
- Or implement IP geolocation service

---

## 📝 Next Steps

1. ✅ Monitor performance in Alibaba Cloud Console
2. ✅ Set up billing alerts
3. ✅ Automate asset sync in CI/CD (see `CHINA_CDN_SETUP.md`)
4. ✅ Test from actual China location

---

**Total Time**: 30 minutes  
**Cost**: FREE (first 6 months), then ~$0.50/month  
**Performance**: <100ms latency for China users ✅
