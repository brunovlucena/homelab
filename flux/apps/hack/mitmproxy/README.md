# YouTube SSAI Bypass Proxy

A multi-layered mitmproxy deployment for your homelab that attempts to bypass YouTube's Server-Side Ad Insertion (SSAI) using multiple detection and filtering mechanisms.

## 🎯 Overview

This project implements a comprehensive approach to blocking YouTube ads at the network level, including the challenging Server-Side Ad Insertion that YouTube uses to embed ads directly into video streams.

```
┌─────────────────────────────────────────────────────────────────────┐
│                    ENHANCED MITMPROXY STACK                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐        │
│  │  Layer 1  │  │  Layer 2  │  │  Layer 3  │  │  Layer 4  │        │
│  │    URL    │  │ Manifest  │  │  Sponsor  │  │   Audio   │        │
│  │  Pattern  │  │ Filtering │  │   Block   │  │Fingerprint│        │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘        │
│        │              │              │              │               │
│        └──────────────┴──────────────┴──────────────┘               │
│                              │                                       │
│                    ┌─────────▼─────────┐                            │
│                    │  Decision Engine  │                            │
│                    └─────────┬─────────┘                            │
│                              │                                       │
│                    ┌─────────▼─────────┐                            │
│                    │   Redis Cache     │                            │
│                    └───────────────────┘                            │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

## 🏗️ Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed technical documentation.

### Components

| Component | Description | Port |
|-----------|-------------|------|
| **mitmproxy** | Main HTTPS proxy with SSAI bypass addons | 30080 (proxy), 30081 (web UI) |
| **Redis** | Caching layer for SponsorBlock data and fingerprints | 6379 (internal) |
| **Audio Analyzer** | Audio fingerprinting service for ad detection | 30083 (API) |

### Detection Layers

1. **URL Pattern Blocking** - Blocks known ad-related URLs, tracking endpoints
2. **Manifest Filtering** - Parses DASH/HLS manifests to remove ad periods
3. **SponsorBlock Integration** - Leverages crowdsourced segment data
4. **Audio Fingerprinting** - Detects ads by audio signature (advanced)

## 📦 Deployment

### Prerequisites

- Kubernetes cluster (Kind, k3s, etc.)
- kubectl configured
- Optional: Flux for GitOps

### Quick Start

```bash
# Deploy all components
kubectl apply -k flux/infrastructure/mitmproxy/

# Check deployment status
kubectl get pods -n mitmproxy

# View logs
kubectl logs -n mitmproxy -l app=mitmproxy -f
```

### Build Audio Analyzer (Optional)

If you want to use the audio fingerprinting feature:

```bash
# Build the image
cd flux/infrastructure/mitmproxy/audio-analyzer
docker build -t ghcr.io/homelab/audio-analyzer:latest .

# Push to your registry
docker push ghcr.io/homelab/audio-analyzer:latest
```

## 🔧 Client Setup

### 1. Get CA Certificate

```bash
# Download from the pod
kubectl cp mitmproxy/$(kubectl get pod -n mitmproxy -l app=mitmproxy -o jsonpath='{.items[0].metadata.name}'):/home/mitmproxy/.mitmproxy/mitmproxy-ca-cert.pem ./mitmproxy-ca-cert.pem
```

### 2. Install Certificate

**macOS:**
```bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain mitmproxy-ca-cert.pem
```

**Linux:**
```bash
sudo cp mitmproxy-ca-cert.pem /usr/local/share/ca-certificates/mitmproxy.crt
sudo update-ca-certificates
```

**iOS:**
1. Transfer cert to device (AirDrop, email, etc.)
2. Install profile in Settings
3. Settings → General → About → Certificate Trust Settings → Enable

**Android:**
1. Copy cert to device
2. Settings → Security → Install from storage

### 3. Configure Proxy

**Environment Variables:**
```bash
export HTTP_PROXY=http://<NODE_IP>:30080
export HTTPS_PROXY=http://<NODE_IP>:30080
export NO_PROXY=localhost,127.0.0.1,.local
```

**macOS System Proxy:**
```bash
networksetup -setwebproxy "Wi-Fi" <NODE_IP> 30080
networksetup -setsecurewebproxy "Wi-Fi" <NODE_IP> 30080
```

**Browser (Firefox):**
1. Settings → Network Settings → Manual proxy configuration
2. HTTP Proxy: `<NODE_IP>`, Port: `30080`
3. Check "Also use this proxy for HTTPS"

## 🖥️ Web Interfaces

| Interface | URL | Description |
|-----------|-----|-------------|
| mitmproxy Web | `http://<NODE_IP>:30081` | Real-time traffic monitoring |
| Audio Analyzer API | `http://<NODE_IP>:30083/docs` | Swagger UI for fingerprint management |

## 📊 Monitoring

### View Blocked Requests

```bash
# Stream mitmproxy logs
kubectl logs -n mitmproxy -l app=mitmproxy -f | grep SSAI

# Example output:
# [SSAI] BLOCKED: https://youtube.com/api/stats/ads... | blocked path: /api/stats/ads
# [SSAI] Filtered DASH manifest: removed 2 ad periods
# [SSAI] SponsorBlock: Found 3 segments for dQw4w9WgXcQ
```

### Metrics

The SSAI bypass addon tracks these metrics:
- `ads_blocked_url` - Requests blocked by URL pattern
- `ads_blocked_manifest` - Ad periods removed from manifests
- `sponsorblock_hits` - Segments found via SponsorBlock
- `manifests_modified` - Total manifests filtered

## 🎵 Audio Fingerprinting

### Register Known Ads

```bash
# Upload an ad audio file to the database
curl -X POST "http://<NODE_IP>:30083/ads/register?ad_name=YouTube%20Premium%20Ad" \
  -F "file=@youtube_premium_ad.mp3"

# List registered ads
curl "http://<NODE_IP>:30083/ads"

# Match audio against database
curl -X POST "http://<NODE_IP>:30083/match/audio" \
  -F "file=@unknown_segment.mp3"
```

### Building the Ad Database

1. Record/download known YouTube ads
2. Register them using the API
3. The service will match future audio against these signatures

## ⚠️ Limitations

### What Works
- ✅ Blocking ad tracking/analytics endpoints
- ✅ Filtering ad periods from DASH/HLS manifests
- ✅ SponsorBlock integration for known videos
- ✅ Audio fingerprint matching (when database populated)

### What's Challenging
- ⚠️ **SSAI Encrypted Streams** - When ads are encrypted with same DRM as content
- ⚠️ **Dynamic Ad Injection** - Real-time ad insertion varies by user
- ⚠️ **YouTube Updates** - Patterns change frequently

### Certificate Pinning
Some apps (YouTube mobile app) use certificate pinning and won't work with mitmproxy. Use these alternatives:
- **ReVanced** - Modified YouTube app
- **NewPipe** - Open-source YouTube client
- **FreeTube** - Desktop YouTube client

## 🔗 Integration with Pi-hole

For maximum ad blocking, combine with Pi-hole:

```yaml
# Pi-hole blocks DNS for ad domains
# mitmproxy handles same-domain ad URLs

# Configure mitmproxy to use Pi-hole as DNS:
# Add to deployment environment:
env:
  - name: DNS_SERVER
    value: "pihole-dns.pihole.svc.cluster.local"
```

## 📁 File Structure

```
mitmproxy/
├── ARCHITECTURE.md        # Detailed technical documentation
├── README.md              # This file
├── namespace.yaml         # Kubernetes namespace
├── configmap.yaml         # Configuration and Python addons
├── deployment.yaml        # Deployment specs (mitmproxy, Redis, Audio Analyzer)
├── service.yaml           # Service definitions
├── pvc.yaml               # Persistent storage for certs
├── kustomization.yaml     # Kustomize configuration
└── audio-analyzer/        # Audio fingerprinting service
    ├── Dockerfile
    ├── requirements.txt
    └── main.py
```

## 🚀 Future Enhancements

- [ ] ML-based ad detection using SponsorBlock-ML
- [ ] Visual frame analysis for ad detection
- [ ] Browser extension companion for skip injection
- [ ] Community ad signature sharing
- [ ] Prometheus metrics endpoint

## 📚 Resources

- [mitmproxy Documentation](https://docs.mitmproxy.org/)
- [SponsorBlock API](https://wiki.sponsor.ajay.app/w/API_Docs)
- [SponsorBlock-ML](https://github.com/xenova/sponsorblock-ml)
- [Chromaprint](https://acoustid.org/chromaprint)
- [YouTube SSAI Analysis](https://adguard.com/en/blog/youtube-server-side-ad-insertion.html)

## ⚖️ Disclaimer

This project is for educational purposes. Using ad-blocking tools may violate YouTube's Terms of Service. Consider supporting content creators through YouTube Premium or other means.
