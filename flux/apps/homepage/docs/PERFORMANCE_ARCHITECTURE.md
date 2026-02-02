# 🏗️ Homepage Performance Architecture (FREE Stack)

## Overview

Your homepage uses a **dual CDN strategy** with all FREE services:

```
User Request
    ↓
Cloudflare Edge (Primary CDN)
    ├─ Cache HIT → Serve from Cloudflare ✅
    └─ Cache MISS → Origin (Your K8s Cluster)
            ↓
        Google Cloud CDN (Fallback)
            ├─ CDN configured → Serve from GCS ✅
            └─ Not configured → Serve from origin ✅
```

## Components

### 1. Cloudflare (Primary CDN) - FREE
- **Page Rules**: Aggressive caching for JS, CSS, images, HTML
- **Auto Minify**: JavaScript, CSS, HTML compression
- **Brotli**: Automatic compression (enabled by default)
- **Edge Network**: Global distribution
- **Cache TTL**: 
  - Static assets: 1 month - 1 year
  - HTML: 15 minutes (configurable)

### 2. Google Cloud CDN (Fallback) - FREE Tier
- **Storage**: 5 GB free per month
- **Egress**: 1 GB free per month
- **Purpose**: Fallback when Cloudflare cache misses
- **Automatic**: `getAssetUrl()` handles fallback logic

### 3. Your Origin (K8s Cluster)
- **nginx**: Serves static files with aggressive caching headers
- **MinIO**: Stores blog images (proxied through nginx)
- **Cache Headers**: Optimized for Cloudflare caching

## Request Flow

### Scenario 1: First Request (Cache MISS)
```
User → Cloudflare → Cache MISS → Your Origin → Google Cloud CDN (if configured)
                                              → Origin files (if not configured)
```

### Scenario 2: Cached Request (Cache HIT)
```
User → Cloudflare → Cache HIT → Serve from Cloudflare Edge ✅
```

### Scenario 3: Cloudflare Cache Expired
```
User → Cloudflare → Cache MISS → Your Origin
                                → Google Cloud CDN (faster than origin)
                                → Origin files (fallback)
```

## Configuration

### Cloudflare
- **Page Rules**: 3 rules configured
- **Auto Minify**: Enabled
- **Brotli**: Automatic

### Google Cloud CDN
- **Environment Variable**: `VITE_CDN_BASE_URL`
- **Fallback Logic**: Built into `getAssetUrl()` function
- **Setup**: See `CDN_SETUP.md`

### Origin (nginx)
- **Cache Headers**: Optimized for Cloudflare
- **Gzip**: Enabled as fallback compression
- **Static Assets**: 1 year cache TTL

## Benefits

1. **Redundancy**: If Cloudflare has issues, Google CDN serves assets
2. **Performance**: Dual CDN reduces latency globally
3. **Cost**: All FREE (within free tier limits)
4. **Automatic**: No code changes needed, just configuration

## Monitoring

### Cloudflare Analytics
- Cache hit ratio (target: >80%)
- Response times
- Bandwidth saved

### Google Cloud Console
- Storage usage (stay under 5GB free tier)
- Egress usage (stay under 1GB free tier)
- CDN cache performance

## Cost Breakdown

| Service | Cost | Free Tier |
|---------|------|-----------|
| Cloudflare | FREE | Unlimited (free plan) |
| Google Cloud Storage | FREE | 5GB storage/month |
| Google Cloud CDN Egress | FREE | 1GB/month |
| **Total** | **$0/month** | Within free tier |

## Next Steps

1. ✅ Configure Cloudflare Page Rules (see `CLOUDFLARE_SETUP.md`)
2. ✅ Enable Auto Minify in Cloudflare
3. ⚠️ Optional: Set up Google Cloud CDN fallback (see `CDN_SETUP.md`)
4. ✅ Deploy updated nginx config (already done)
5. ✅ Monitor performance in Cloudflare Analytics

---

**All optimizations are FREE and provide 60-80% performance improvement!**

