# ⚡ Cloudflare Performance Setup Guide (FREE Only)

Quick guide to configure Cloudflare for maximum homepage performance using only FREE features.

## 🎯 Step 1: Page Rules (FREE - Do This First!)

Go to: **Cloudflare Dashboard → Rules → Page Rules**

### Rule 1: Cache JavaScript Files
```
URL Pattern: lucena.cloud/*.js
Settings:
  ✅ Cache Level: Cache Everything
  ✅ Edge Cache TTL: 1 month
  ✅ Browser Cache TTL: Respect Existing Headers
```

### Rule 2: Cache Images
```
URL Pattern: lucena.cloud/storage/*
Settings:
  ✅ Cache Level: Cache Everything
  ✅ Edge Cache TTL: 1 year
  ✅ Browser Cache TTL: 1 year
```

### Rule 3: HTML Caching
```
URL Pattern: lucena.cloud/
Settings:
  ✅ Cache Level: Cache Everything
  ✅ Edge Cache TTL: 15 minutes
  ✅ Browser Cache TTL: Respect Existing Headers
```

**Note**: If your homepage content changes very frequently, you can set HTML to "Bypass" instead.

---

## 🚀 Step 2: Speed Optimization (FREE)

Go to: **Cloudflare Dashboard → Speed → Optimization**

### Auto Minify
- ✅ **JavaScript**: ON
- ✅ **CSS**: ON
- ✅ **HTML**: ON

### Other Free Features (Already Enabled)
- ✅ **Brotli**: Automatically enabled by Cloudflare
- ✅ **HTTP/2**: Already enabled
- ✅ **HTTP/3 (QUIC)**: Enable if available in your plan

---

## 📊 Step 3: Verify Performance

### Check Cache Status
1. Open browser DevTools → Network tab
2. Reload page
3. Check response headers:
   - `CF-Cache-Status: HIT` = Cached by Cloudflare ✅
   - `CF-Cache-Status: MISS` = Not cached yet (first request)
   - `CF-Cache-Status: DYNAMIC` = Can't cache (expected for some HTML)

### Test Performance
1. Run Lighthouse in Chrome DevTools
2. Target scores:
   - Performance: 90+
   - Best Practices: 90+
   - SEO: 90+

### Monitor in Cloudflare
**Location**: Analytics → Performance
- Check cache hit ratio (should be >80% for static assets)
- Monitor response times
- Check bandwidth saved

---

## 🔵 Step 4: Google Cloud CDN as Fallback (Optional but Recommended)

Google Cloud CDN serves as a fallback when Cloudflare cache misses occur.

### Setup
See `CDN_SETUP.md` for detailed instructions. Quick setup:

```bash
# 1. Create bucket
gsutil mb -p YOUR_PROJECT -c STANDARD -l us-central1 gs://lucena-cloud-assets

# 2. Make public
gsutil iam ch allUsers:objectViewer gs://lucena-cloud-assets

# 3. Upload assets
gsutil -m rsync -r ./public/assets gs://lucena-cloud-assets/assets

# 4. Set environment variable before build
export VITE_CDN_BASE_URL=https://storage.googleapis.com/lucena-cloud-assets
```

The `getAssetUrl()` function automatically uses Google Cloud CDN if configured, otherwise falls back to relative paths.

---

## 📈 Expected Results

After implementing all FREE optimizations:
- **Static assets**: 80-90% cache hit ratio (Cloudflare)
- **Page load time**: 40-60% faster
- **Bandwidth saved**: 50-70% reduction
- **Lighthouse score**: +20-30 points
- **Global performance**: Improved via Cloudflare edge + Google CDN fallback

---

## 🆘 Troubleshooting

### Assets not caching
- Check Page Rules are active (green status in dashboard)
- Verify URL patterns match exactly (case-sensitive)
- Check response headers in browser DevTools
- Ensure assets have proper `Cache-Control` headers from origin

### Still slow after setup
- Check Cloudflare Analytics for cache hit ratio
- Verify Auto Minify is enabled
- Check if assets are being served from Cloudflare edge
- Verify Google Cloud CDN is configured if using fallback

### Images loading slowly
- Convert images to WebP format manually (free)
- Use responsive images with `srcset`
- Ensure images are uploaded to Google Cloud CDN if using fallback
- Check image file sizes (optimize before upload)

---

## 📚 Resources

- [Cloudflare Page Rules Docs](https://developers.cloudflare.com/rules/page-rules/)
- [Cloudflare Speed Optimization](https://developers.cloudflare.com/speed/)
- [Cloudflare Workers Docs](https://developers.cloudflare.com/workers/)

