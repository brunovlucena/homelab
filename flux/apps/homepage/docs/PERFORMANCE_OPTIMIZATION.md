# 🚀 Homepage Performance Optimization Guide (FREE Only)

## Current Setup
- **Domain**: `lucena.cloud` via Cloudflare Tunnel
- **Frontend**: React + Vite + nginx
- **Backend**: Go API
- **Current Issues**: Slow page load times

---

## 🎯 FREE Performance Stack

### 1. Cloudflare Page Rules (FREE - Immediate Impact)

Since you're already using Cloudflare Tunnel, configure Page Rules in Cloudflare Dashboard:

**Rule 1: Cache JavaScript Files**
```
URL Pattern: lucena.cloud/*.js
Settings:
  - Cache Level: Cache Everything
  - Edge Cache TTL: 1 month
  - Browser Cache TTL: Respect Existing Headers
```

**Rule 2: Cache Images**
```
URL Pattern: lucena.cloud/storage/*
Settings:
  - Cache Level: Cache Everything
  - Edge Cache TTL: 1 year
  - Browser Cache TTL: 1 year
```

**Rule 3: HTML Caching**
```
URL Pattern: lucena.cloud/
Settings:
  - Cache Level: Cache Everything
  - Edge Cache TTL: 15 minutes (or Bypass if content changes frequently)
  - Browser Cache TTL: Respect Existing Headers
```

### 2. Enable Cloudflare Auto Minify (FREE)

In Cloudflare Dashboard → Speed → Optimization:
- ✅ JavaScript
- ✅ CSS
- ✅ HTML

### 3. Enable Brotli Compression (FREE)

Cloudflare automatically uses Brotli compression. Your nginx is already configured with gzip as fallback.

---

## 🔵 Google Cloud CDN as Fallback (FREE tier available)

Google Cloud CDN serves as a **fallback** for static assets when Cloudflare cache misses occur or for redundancy.

**Free Tier**: 5GB storage, 1GB egress/month

### How It Works

The `getAssetUrl()` function in `src/utils/index.ts` automatically:
1. Uses Google Cloud CDN if `VITE_CDN_BASE_URL` is configured
2. Falls back to relative paths if CDN is not configured
3. Works seamlessly with Cloudflare caching

### Setup Google Cloud CDN

```bash
# 1. Create bucket
gsutil mb -p YOUR_PROJECT -c STANDARD -l us-central1 gs://lucena-cloud-assets

# 2. Make public
gsutil iam ch allUsers:objectViewer gs://lucena-cloud-assets

# 3. Configure CORS
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
gsutil cors set cors.json gs://lucena-cloud-assets

# 4. Upload assets
gsutil -m rsync -r ./public/assets gs://lucena-cloud-assets/assets
```

### Configure in Build

Set environment variable before building:

```bash
export VITE_CDN_BASE_URL=https://storage.googleapis.com/lucena-cloud-assets
npm run build
```

Or add to `.env` file:
```env
VITE_CDN_BASE_URL=https://storage.googleapis.com/lucena-cloud-assets
```

### Usage in Code

The CDN is used automatically via `getAssetUrl()`:

```typescript
import { getAssetUrl } from '../utils'

// Automatically uses CDN if configured, otherwise falls back to relative path
<img src={getAssetUrl('assets/eu.webp')} alt="Profile" />
```

**Result**:
- If CDN configured: `https://storage.googleapis.com/lucena-cloud-assets/assets/eu.webp`
- If not configured: `./assets/eu.webp` (served from your origin)

---

## 🛠️ Application-Level Optimizations

### 1. Improve nginx Caching

Update `nginx.conf`:

```nginx
# Add to http block in nginx-main.conf
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=static_cache:10m max_size=1g 
                inactive=60m use_temp_path=off;

# Update server block
server {
    # ... existing config ...
    
    # Cache static assets aggressively
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg|webp|woff|woff2|ttf|eot)$ {
        root /usr/share/nginx/html;
        expires 1y;
        add_header Cache-Control "public, immutable";
        add_header X-Content-Type-Options "nosniff";
        
        # Enable gzip/brotli
        gzip_static on;
    }
    
    # Cache MinIO images with proxy cache
    location /storage/ {
        proxy_pass http://minio.minio.svc.cluster.local:9000/;
        proxy_cache static_cache;
        proxy_cache_valid 200 1h;
        proxy_cache_use_stale error timeout updating http_500 http_502 http_503 http_504;
        proxy_cache_background_update on;
        proxy_cache_lock on;
        
        add_header X-Cache-Status $upstream_cache_status;
        expires 1h;
        add_header Cache-Control "public";
    }
}
```

### 2. Enable Brotli in nginx

Add to Dockerfile:
```dockerfile
# Install brotli module
RUN apk add --no-cache nginx-mod-http-brotli

# Enable in nginx-main.conf
load_module modules/ngx_http_brotli_filter_module.so;
load_module modules/ngx_http_brotli_static_module.so;

http {
    brotli on;
    brotli_comp_level 6;
    brotli_types text/plain text/css text/xml text/javascript 
                 application/json application/javascript application/xml+rss 
                 image/svg+xml;
}
```

### 3. Optimize Vite Build

Update `vite.config.ts`:

```typescript
export default defineConfig({
  // ... existing config ...
  build: {
    outDir: 'dist',
    sourcemap: false,
    minify: 'esbuild',
    cssMinify: true,
    rollupOptions: {
      output: {
        // Better code splitting
        manualChunks: {
          vendor: ['react', 'react-dom'],
          router: ['react-router-dom'],
          query: ['@tanstack/react-query'],
          ui: ['framer-motion', 'lucide-react'],
        },
        // Optimize chunk names
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
      },
    },
    // Enable chunk size warnings
    chunkSizeWarningLimit: 1000,
  },
});
```

### 4. Add Resource Hints

Update `index.html`:

```html
<head>
  <!-- Preconnect to CDN -->
  <link rel="preconnect" href="https://storage.googleapis.com">
  <link rel="dns-prefetch" href="https://storage.googleapis.com">
  
  <!-- Preload critical resources -->
  <link rel="preload" href="/assets/fonts/main.woff2" as="font" type="font/woff2" crossorigin>
  
  <!-- Prefetch likely next pages -->
  <link rel="prefetch" href="/blog">
</head>
```

### 5. Image Optimization

**Convert images to WebP**:
```bash
# Install cwebp
brew install webp  # macOS
# or
apt-get install webp  # Linux

# Convert images
find public/assets -name "*.png" -o -name "*.jpg" | while read img; do
  cwebp -q 80 "$img" -o "${img%.*}.webp"
done
```

**Use responsive images**:
```tsx
<picture>
  <source srcSet={getAssetUrl('assets/eu.webp')} type="image/webp" />
  <img src={getAssetUrl('assets/eu.png')} alt="Profile" loading="lazy" />
</picture>
```

---

## 📊 Infrastructure Optimizations

### 1. Increase Frontend Replicas

Update `frontend-deployment.yaml`:

```yaml
spec:
  replicas: 3  # Increase from 2
  # ... rest of config ...
```

### 2. Optimize Resource Limits

```yaml
resources:
  requests:
    cpu: 100m  # Increase from 50m
    memory: 256Mi  # Increase from 128Mi
  limits:
    cpu: 500m  # Increase from 300m
    memory: 1Gi  # Increase from 512Mi
```

### 3. Enable HPA (Horizontal Pod Autoscaler)

Already configured in `hpa.yaml`, but verify:

```yaml
spec:
  minReplicas: 2
  maxReplicas: 5
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

---

## 🎯 Recommended Implementation Order (All FREE)

### Phase 1: Immediate (Do Today - 15 minutes)
1. ✅ Configure Cloudflare Page Rules (3 rules)
2. ✅ Enable Auto Minify in Cloudflare
3. ✅ Deploy updated nginx config (already done)
4. ✅ Deploy updated Vite build config (already done)

### Phase 2: This Week (Optional but Recommended)
1. ✅ Set up Google Cloud CDN as fallback
2. ✅ Upload assets to Google Cloud Storage
3. ✅ Configure `VITE_CDN_BASE_URL` environment variable
4. ✅ Rebuild and deploy

### Phase 3: Advanced (Optional)
1. ✅ Convert images to WebP format manually
2. ✅ Increase frontend replicas if needed
3. ✅ Monitor cache hit ratios in Cloudflare Analytics

---

## 📈 Expected Performance Improvements

| Optimization | Impact | Cost |
|-------------|--------|------|
| Cloudflare Page Rules | 40-60% faster static assets | FREE |
| Auto Minify | 10-20% smaller JS/CSS | FREE |
| Brotli Compression | 15-25% smaller responses | FREE |
| Google Cloud CDN (Fallback) | 30-50% faster global loads | FREE tier (5GB/1GB) |
| WebP Images | 50-70% smaller images | FREE (manual conversion) |

**Combined Expected Improvement**: 60-80% faster page loads globally (all FREE)

---

## 🔍 Monitoring Performance

### Cloudflare Analytics
- Dashboard → Analytics → Performance
- Monitor: Cache hit ratio, response times, bandwidth saved

### Google Cloud CDN Metrics
```bash
# Check CDN cache hit ratio
gcloud compute backend-services get-health cdn-backend \
  --global
```

### Browser DevTools
- Lighthouse score (target: 90+)
- Network tab: Check cache headers
- Performance tab: Identify bottlenecks

---

## 🚨 Common Issues

### Images not caching
- Check `Cache-Control` headers
- Verify Cloudflare Page Rules are active
- Check nginx `expires` directive

### Slow API responses
- Consider API caching with Redis
- Add API response compression
- Optimize database queries

### High bandwidth costs
- Enable Cloudflare caching aggressively
- Use WebP images
- Enable compression everywhere

---

## 📚 References

- [Cloudflare Page Rules](https://developers.cloudflare.com/rules/page-rules/)
- [Cloudflare Speed Optimization](https://developers.cloudflare.com/speed/)
- [Google Cloud CDN](https://cloud.google.com/cdn/docs)
- [Web.dev Performance](https://web.dev/performance/)

