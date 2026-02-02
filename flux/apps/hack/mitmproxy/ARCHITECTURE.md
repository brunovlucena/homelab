# YouTube SSAI Bypass - Architecture Plan

## Executive Summary

This document outlines a multi-layered approach to bypass YouTube's Server-Side Ad Insertion (SSAI). Since SSAI embeds ads directly into the video stream, we need multiple detection and filtering mechanisms working together.

## The Problem

```
Traditional Ad Blocking:
┌─────────┐     ┌─────────┐     ┌─────────┐
│ Client  │────▶│ Ad URL  │  X  │ Blocked │
└─────────┘     └─────────┘     └─────────┘

SSAI (Server-Side Ad Insertion):
┌─────────┐     ┌─────────────────────────────┐     ┌─────────┐
│ YouTube │────▶│ Video + Ads (same stream)   │────▶│ Client  │
│ Server  │     │ googlevideo.com             │     │         │
└─────────┘     └─────────────────────────────┘     └─────────┘
                        ↑
                  Can't block without
                  blocking video too
```

## Solution Architecture

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                           ENHANCED MITMPROXY                                  │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐      │
│  │   Layer 1   │   │   Layer 2   │   │   Layer 3   │   │   Layer 4   │      │
│  │ URL Pattern │   │  Manifest   │   │ SponsorBlock│   │   Audio     │      │
│  │  Blocking   │   │  Filtering  │   │ Integration │   │Fingerprint  │      │
│  └──────┬──────┘   └──────┬──────┘   └──────┬──────┘   └──────┬──────┘      │
│         │                 │                 │                 │              │
│         ▼                 ▼                 ▼                 ▼              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                      Decision Engine                                 │    │
│  │  - Combines signals from all layers                                  │    │
│  │  - Maintains video state and timing                                  │    │
│  │  - Generates skip/block decisions                                    │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                    │                                         │
│                                    ▼                                         │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                         Redis Cache                                  │    │
│  │  - Video ID → Ad segments mapping                                    │    │
│  │  - Audio fingerprint cache                                           │    │
│  │  - SponsorBlock data cache                                           │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                        SUPPORTING SERVICES                                    │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐              │
│  │ Audio Analyzer  │  │  ML Detector    │  │ Segment Tracker │              │
│  │    Service      │  │    Service      │  │    Service      │              │
│  │                 │  │                 │  │                 │              │
│  │ - Chromaprint   │  │ - SponsorBlock  │  │ - Track video   │              │
│  │ - Audio hash DB │  │   ML model      │  │   playback      │              │
│  │ - Ad signature  │  │ - Transcript    │  │ - Detect skips  │              │
│  │   matching      │  │   analysis      │  │ - Log patterns  │              │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘              │
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Layer Details

### Layer 1: Enhanced URL Pattern Blocking

**Purpose**: Block known ad-related requests before they reach YouTube.

**Improvements over basic blocking**:
- Dynamic pattern updates from community blocklists
- Machine learning-based URL classification
- Request header analysis (ad-specific headers)

**Patterns to block**:
```
# Direct ad endpoints
/pagead/, /ptracking, /api/stats/ads, /get_midroll_info

# Tracking pixels
/api/stats/atr, /api/stats/qoe.*ad

# Ad manifest requests
ctier=L (ad tier marker in googlevideo URLs)

# Doubleclick/Google Ads
googlesyndication.com, doubleclick.net, googleadservices.com
```

### Layer 2: Manifest Filtering (DASH/HLS)

**Purpose**: Parse and modify video manifests to remove ad segments.

**How YouTube manifests work**:
```xml
<!-- DASH MPD Structure -->
<MPD>
  <Period id="content-1" duration="PT120S">
    <!-- Main video content -->
  </Period>
  <Period id="ad-midroll-1" duration="PT30S">
    <!-- INJECTED AD - REMOVE THIS -->
  </Period>
  <Period id="content-2" duration="PT180S">
    <!-- More video content -->
  </Period>
</MPD>
```

**Detection signals**:
- Period ID containing "ad", "midroll", "preroll"
- SCTE-35 markers (industry standard ad insertion markers)
- EventStream elements with ad events
- Unusual segment durations (15s, 30s, 60s - common ad lengths)
- Missing or different `<ContentProtection>` for ad periods

### Layer 3: SponsorBlock Integration

**Purpose**: Leverage crowdsourced segment data for known videos.

**API Integration**:
```
GET https://sponsor.ajay.app/api/skipSegments?videoID={id}

Response:
[
  {
    "segment": [0, 30.5],
    "category": "sponsor",
    "UUID": "..."
  }
]
```

**Enhancement**: Use segment data to:
- Pre-mark known ad timestamps
- Adjust for SSAI timestamp shifts
- Report new ad segments back to community

### Layer 4: Audio Fingerprinting

**Purpose**: Detect ads by their audio signature, regardless of where they appear in the stream.

**How it works**:
```
1. Extract audio from video chunks
2. Generate fingerprint (Chromaprint/AcoustID)
3. Compare against known ad audio database
4. If match > threshold → mark as ad segment
```

**Database sources**:
- Known YouTube ad campaigns
- Common ad music/jingles
- Silence patterns (ads often have different audio profiles)

## Implementation Phases

### Phase 1: Enhanced URL Blocking (Current)
- [x] Basic URL pattern matching
- [x] Domain blocking
- [ ] Dynamic blocklist updates
- [ ] Request header analysis

### Phase 2: Manifest Filtering
- [ ] DASH MPD parser
- [ ] HLS M3U8 parser
- [ ] Ad period detection
- [ ] Manifest rewriting

### Phase 3: SponsorBlock Integration
- [ ] API client
- [ ] Caching layer
- [ ] Timestamp adjustment for SSAI
- [ ] Segment injection into player

### Phase 4: Audio Analysis (Advanced)
- [ ] Audio extraction service
- [ ] Fingerprint generation
- [ ] Ad signature database
- [ ] Real-time matching

### Phase 5: ML-Based Detection
- [ ] Deploy SponsorBlock-ML model
- [ ] Transcript analysis
- [ ] Visual frame analysis (future)

## Data Flow

```
Request Flow:
─────────────

Client                mitmproxy              YouTube
  │                       │                      │
  │ ──── GET video ────▶ │                      │
  │                       │ ──── GET video ────▶│
  │                       │ ◀──── Response ─────│
  │                       │                      │
  │                       ▼                      │
  │               ┌───────────────┐              │
  │               │ Layer 1: URL  │              │
  │               │ Is it an ad?  │              │
  │               └───────┬───────┘              │
  │                       │ No                   │
  │                       ▼                      │
  │               ┌───────────────┐              │
  │               │ Layer 2:      │              │
  │               │ Parse manifest│              │
  │               │ Remove ad     │              │
  │               │ periods       │              │
  │               └───────┬───────┘              │
  │                       │                      │
  │                       ▼                      │
  │               ┌───────────────┐              │
  │               │ Layer 3:      │              │
  │               │ Check         │              │
  │               │ SponsorBlock  │              │
  │               │ Add skip data │              │
  │               └───────┬───────┘              │
  │                       │                      │
  │ ◀─── Modified ────────│                      │
  │      Response                                │
```

## Technical Challenges

### 1. Encrypted Streams
- YouTube uses DRM (Widevine) for some content
- Manifest is readable, but video chunks are encrypted
- **Solution**: Focus on manifest-level filtering

### 2. Dynamic Ad Insertion Timing
- Ads are inserted at different times for different users
- Timestamps shift based on ad length
- **Solution**: Use relative timing + audio fingerprinting

### 3. Anti-Detection
- YouTube may detect modified manifests
- Could serve different content or block access
- **Solution**: Minimal modifications, preserve structure

### 4. Performance
- Manifest parsing adds latency
- Audio analysis is CPU-intensive
- **Solution**: Aggressive caching, async processing

## Metrics to Track

| Metric | Description |
|--------|-------------|
| `ads_blocked_url` | Requests blocked by URL pattern |
| `ads_blocked_manifest` | Ad periods removed from manifests |
| `sponsorblock_hits` | Segments found via SponsorBlock |
| `audio_fingerprint_matches` | Ads detected by audio |
| `false_positives` | Content incorrectly blocked |
| `latency_added_ms` | Processing time overhead |

## Future Enhancements

1. **Browser Extension Companion**
   - Inject skip buttons based on proxy data
   - Report playback position for timing analysis

2. **Community Contribution**
   - Submit new ad patterns
   - Share audio fingerprints
   - Crowdsource detection improvements

3. **Smart TV / Chromecast Support**
   - Integrate with CastBlock
   - Network-wide ad blocking for all devices

## Resources

- [mitmproxy Addons Documentation](https://docs.mitmproxy.org/stable/addons-overview/)
- [DASH-IF Guidelines](https://dashif.org/guidelines/)
- [SponsorBlock API](https://wiki.sponsor.ajay.app/w/API_Docs)
- [Chromaprint](https://acoustid.org/chromaprint)
- [SponsorBlock-ML](https://github.com/xenova/sponsorblock-ml)

