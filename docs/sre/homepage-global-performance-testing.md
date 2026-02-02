# 🌍 Global Performance Testing Guide: USA & China

**Purpose**: Test homepage (`lucena.cloud`) and agent-bruno chatbot loading performance from USA and China  
**Created**: 2025-01-27  
**Priority**: High (8% Cloudflare cache hit ratio indicates optimization needed)  
**Linear Issue**: [BVL-326](https://linear.app/bvlucena/issue/BVL-326/sre-004-webpagetest-baseline-testing-from-usa-and-china)

---

## 🎯 Testing Objectives

Test from **USA** and **China** to measure:
1. **Page Load Time** - Time to Interactive (TTI), First Contentful Paint (FCP)
2. **Cloudflare Cache Effectiveness** - Cache hit ratio, edge latency
3. **API Response Times** - Origin processing + network latency
4. **Core Web Vitals** - LCP, FID, CLS scores
5. **Geographic Performance** - Compare USA vs China latency

---

## 🔍 Current Performance Issue

**Cloudflare Analytics shows:**
- **Bandwidth Saved**: Only 8% (should be >85%)
- **Cache Hit Ratio**: Very low (most requests hitting origin)
- **Impact**: Users in China/USA experience full origin latency instead of edge cache

**This means:**
- Users in China: ~200-400ms + origin latency (5ms) = **205-405ms total**
- Users in USA: ~50-150ms + origin latency (5ms) = **55-155ms total**
- With proper caching: **5-50ms** (served from Cloudflare edge)

---

## 🧪 Recommended Testing Tools

### Option 1: WebPageTest.org ⭐ **RECOMMENDED (FREE)**

**Best for**: Comprehensive page load testing from multiple locations

#### Setup:
1. Go to: https://www.webpagetest.org
2. Enter URL: `https://lucena.cloud`
3. Select locations:
   - **USA**: `Dulles, VA` (EC2) or `Chicago, IL` (EC2)
   - **China**: `Beijing, China` (if available) or `Hong Kong` (closest)
4. Browser: Chrome (Desktop) or Mobile
5. Connection: Cable (5 Mbps) or 4G
6. Click "Start Test"

#### What to Measure:
- **Load Time**: Time to fully load page
- **First Byte Time (TTFB)**: Server response time
- **First Contentful Paint (FCP)**: First content visible
- **Largest Contentful Paint (LCP)**: Largest element loaded
- **Speed Index**: Visual completeness
- **Total Requests**: Number of assets loaded
- **Cache Hit Ratio**: Check response headers for `CF-Cache-Status: HIT`

#### Advanced Testing:
- **Video**: Record page load video (slow motion)
- **Waterfall Chart**: See which assets load slowest
- **Connection View**: Simulate slow connections
- **Repeat View**: Test cache effectiveness (2nd load should be faster)

#### API for Automation:
```bash
# Test from USA
curl "https://www.webpagetest.org/runtest.php?url=https://lucena.cloud&location=Dulles:Chrome.Cable&runs=3&fvonly=1&k=YOUR_API_KEY"

# Test from China (Hong Kong)
curl "https://www.webpagetest.org/runtest.php?url=https://lucena.cloud&location=ec2-ap-northeast-1:Chrome.Cable&runs=3&fvonly=1&k=YOUR_API_KEY"
```

**Free Tier**: 200 tests/day (more than enough)

---

### Option 2: GTmetrix ⭐ **EASY TO USE (FREE)**

**Best for**: Quick performance scores and Lighthouse reports

#### Setup:
1. Go to: https://gtmetrix.com
2. Enter URL: `https://lucena.cloud`
3. Select location:
   - **USA**: `Vancouver, Canada` (closest to USA)
   - **China**: `Hong Kong, China` or `Tokyo, Japan`
4. Click "Test your site"

#### Metrics Provided:
- **Performance Score**: 0-100 (target: >90)
- **Structure Score**: 0-100 (target: >90)
- **Page Load Time**: Total load time
- **Total Page Size**: All assets combined
- **Requests**: Number of HTTP requests
- **Lighthouse Scores**: Core Web Vitals
- **Waterfall Chart**: Asset loading timeline

**Free Tier**: 3 tests/day per location (enough for basic testing)

---

### Option 3: Cloudflare Speed Test (Built-in) ⭐ **CLOUDFLARE NATIVE**

**Best for**: Testing Cloudflare-specific optimizations

#### Setup:
1. Go to: Cloudflare Dashboard → Speed → Optimization
2. Click "Run Test" or use Cloudflare's built-in analytics
3. Check: Analytics → Performance → Response Times by Country

#### What to Check:
- **Response Times by Country**: Compare USA vs China
- **Cache Hit Ratio by Country**: See if caching works in different regions
- **Bandwidth Saved**: Should increase as cache improves

#### Using Cloudflare Analytics API:
```bash
# Get response times by country (requires API token)
curl -X GET "https://api.cloudflare.com/client/v4/zones/{zone_id}/analytics/dashboard?since=-7d" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "Content-Type: application/json"
```

---

### Option 4: k6 Load Testing (Your Existing Tool) ⭐ **AUTOMATED**

**Best for**: Automated, scriptable testing from your infrastructure

#### Create Test Script:

```javascript
// tests/k6/homepage-global-performance.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('errors');
const pageLoadTime = new Trend('page_load_time');
const apiResponseTime = new Trend('api_response_time');

export const options = {
  scenarios: {
    usa_test: {
      executor: 'constant-vus',
      vus: 5,
      duration: '2m',
      tags: { location: 'usa' },
      // Run from USA-based server/VPS
    },
    china_test: {
      executor: 'constant-vus',
      vus: 5,
      duration: '2m',
      startTime: '0s',
      tags: { location: 'china' },
      // Run from China-based server/VPS or Hong Kong
    },
  },
  thresholds: {
    'http_req_duration{location:usa}': ['p(95)<500', 'p(99)<1000'],
    'http_req_duration{location:china}': ['p(95)<1000', 'p(99)<2000'],
    'errors': ['rate<0.01'],
  },
};

export default function () {
  const baseUrl = 'https://lucena.cloud';
  
  // Test homepage
  const homepageStart = Date.now();
  const homepageRes = http.get(baseUrl);
  const homepageDuration = Date.now() - homepageStart;
  
  check(homepageRes, {
    'homepage status is 200': (r) => r.status === 200,
    'homepage cached by Cloudflare': (r) => 
      r.headers['CF-Cache-Status'] === 'HIT' || 
      r.headers['CF-Cache-Status'] === 'DYNAMIC',
  });
  
  pageLoadTime.add(homepageDuration);
  errorRate.add(homepageRes.status !== 200);
  
  // Test API endpoint
  const apiStart = Date.now();
  const apiRes = http.get(`${baseUrl}/api/projects`);
  const apiDuration = Date.now() - apiStart;
  
  check(apiRes, {
    'api status is 200': (r) => r.status === 200,
    'api response time < 500ms': (r) => apiDuration < 500,
  });
  
  apiResponseTime.add(apiDuration);
  errorRate.add(apiRes.status !== 200);
  
  // Test agent-bruno (if accessible)
  const chatRes = http.post(`${baseUrl}/api/chat`, 
    JSON.stringify({ message: 'Hello' }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  check(chatRes, {
    'chat status is 200': (r) => r.status === 200,
  });
  
  sleep(2);
}
```

#### Running Tests:

```bash
# Test from your local machine (USA perspective)
k6 run tests/k6/homepage-global-performance.js

# Test from China-based server (requires VPN/VPS in China)
# Option A: Use VPN
VPN_CONNECTION=china k6 run tests/k6/homepage-global-performance.js

# Option B: Deploy k6 job to server in China/Hong Kong
kubectl apply -f tests/k6/k6-china-test-job.yaml
```

#### Schedule Regular Tests:

```yaml
# tests/k6/k6-global-performance-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: homepage-global-performance-test
  namespace: homepage
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: k6
            image: grafana/k6:latest
            command:
            - k6
            - run
            - --out=json=/results/usa-results.json
            - tests/k6/homepage-global-performance.js
          restartPolicy: OnFailure
```

---

### Option 5: Browser DevTools with VPN ⭐ **REAL USER SIMULATION**

**Best for**: Manual testing with real browser experience

#### Setup:
1. **Install VPN**: ExpressVPN, NordVPN, or similar
2. **Connect to USA server**: Los Angeles, New York, etc.
3. **Open Chrome DevTools**: F12
4. **Test homepage**:
   ```bash
   # Open https://lucena.cloud
   # Check Network tab for:
   # - Total load time
   # - Response headers (CF-Cache-Status)
   # - Waterfall chart
   ```
5. **Connect to China/Hong Kong server**
6. **Repeat test** and compare

#### What to Check:
- **Network Tab**:
  - Load time (should be in bottom-right)
  - `CF-Cache-Status` header (HIT = cached, MISS = origin)
  - Request waterfall (which assets load slowest)
  - Response sizes

- **Lighthouse Tab**:
  - Performance score
  - Core Web Vitals (LCP, FID, CLS)
  - Opportunities for optimization

- **Performance Tab**:
  - Record page load
  - See main thread blocking
  - Identify JavaScript bottlenecks

---

### Option 6: Cloudflare Workers for Testing ⭐ **EDGE-BASED TESTING**

**Best for**: Testing from Cloudflare's edge locations directly

#### Create Test Worker:

```javascript
// cloudflare-workers/performance-test.js
export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    
    // Test different endpoints
    const testUrls = [
      'https://lucena.cloud',
      'https://lucena.cloud/api/projects',
    ];
    
    const results = await Promise.all(
      testUrls.map(async (testUrl) => {
        const start = Date.now();
        const response = await fetch(testUrl, {
          cf: {
            // Test from different edge locations
            colo: url.searchParams.get('location') || 'US', // US, CN, etc.
          },
          headers: {
            'User-Agent': 'Cloudflare-Worker-Performance-Test',
          },
        });
        const duration = Date.now() - start;
        
        const cacheStatus = response.headers.get('CF-Cache-Status');
        const body = await response.text();
        
        return {
          url: testUrl,
          status: response.status,
          duration: duration,
          cacheStatus: cacheStatus,
          size: body.length,
          headers: Object.fromEntries(response.headers),
        };
      })
    );
    
    return new Response(JSON.stringify({
      timestamp: new Date().toISOString(),
      location: url.searchParams.get('location') || 'US',
      results: results,
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
  },
};
```

#### Deploy and Test:

```bash
# Deploy to Cloudflare Workers
wrangler publish

# Test from USA edge
curl "https://performance-test.YOUR_SUBDOMAIN.workers.dev/?location=US"

# Test from China edge (if available)
curl "https://performance-test.YOUR_SUBDOMAIN.workers.dev/?location=CN"
```

---

### Option 7: Pingdom (Synthetic Monitoring) ⭐ **CONTINUOUS MONITORING**

**Best for**: Automated continuous monitoring from multiple locations

#### Setup:
1. Sign up: https://www.pingdom.com (free tier available)
2. Create new check:
   - Type: HTTP(S) Transaction
   - URL: `https://lucena.cloud`
   - Locations: **USA (Washington)** and **China (Hong Kong)** or **Asia-Pacific**
   - Frequency: Every 5 minutes
   - Alerts: Email/Slack when response time > 2s

#### Metrics Tracked:
- **Uptime**: Availability percentage
- **Response Time**: By location
- **Page Size**: Total bytes
- **Requests**: Number of HTTP requests
- **Performance Grade**: A-F score

**Free Tier**: 1 check, 1 location (limited but useful)

---

## 🔗 Related Linear Issues

- **[BVL-326](https://linear.app/bvlucena/issue/BVL-326)**: WebPageTest Baseline Testing from USA and China (this issue)
- **[BVL-24](https://linear.app/bvlucena/issue/BVL-24)**: Set up Alibaba Cloud CDN for China (related - testing from China)
- **[BVL-25](https://linear.app/bvlucena/issue/BVL-25)**: Fix Google Cloud CDN bucket location (related - global CDN optimization)
- **[BVL-310](https://linear.app/bvlucena/issue/BVL-310)**: k6 Performance Tests for Homepage API (complementary - automated testing)
- **[BVL-311](https://linear.app/bvlucena/issue/BVL-311)**: k6 Browser Performance Tests for Homepage (complementary - automated testing)

---

## 📊 Testing Plan: Step-by-Step

### Phase 1: Baseline Measurement (Do First) - See [BVL-326](https://linear.app/bvlucena/issue/BVL-326)

**1. Test Homepage from USA**
```bash
# Using WebPageTest
https://www.webpagetest.org
URL: https://lucena.cloud
Location: Dulles, VA
```

**Record:**
- [ ] Load Time: _____ seconds
- [ ] First Byte Time: _____ ms
- [ ] Cache Hit Ratio: _____ % (check `CF-Cache-Status` headers)
- [ ] Page Size: _____ MB
- [ ] Requests: _____ count

**2. Test Homepage from China/Hong Kong**
```bash
# Using WebPageTest
Location: Hong Kong or Beijing (if available)
```

**Record:**
- [ ] Load Time: _____ seconds
- [ ] First Byte Time: _____ ms
- [ ] Cache Hit Ratio: _____ %
- [ ] Page Size: _____ MB
- [ ] Requests: _____ count

**3. Test Agent-Bruno API**
```bash
# Test chat endpoint
curl -X POST https://lucena.cloud/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"test"}' \
  -w "\nTime: %{time_total}s\n" \
  -o /dev/null -s
```

**Record:**
- [ ] Response Time (USA perspective): _____ ms
- [ ] Response Time (China perspective): _____ ms

---

### Phase 2: Optimize Cache (Priority Action)

**Your Cloudflare dashboard shows only 8% bandwidth saved - this is the problem!**

#### Immediate Actions:

**1. Check Current Cache Headers**
```bash
curl -I https://lucena.cloud | grep -i cache
# Look for: Cache-Control, CF-Cache-Status, Expires
```

**2. Configure Cloudflare Page Rules** (if not already done)
- Go to: Cloudflare Dashboard → Rules → Page Rules
- Create rules:
  ```
  Rule 1: lucena.cloud/*.js
  - Cache Level: Cache Everything
  - Edge Cache TTL: 1 month
  
  Rule 2: lucena.cloud/*.css
  - Cache Level: Cache Everything
  - Edge Cache TTL: 1 month
  
  Rule 3: lucena.cloud/assets/*
  - Cache Level: Cache Everything
  - Edge Cache TTL: 1 year
  
  Rule 4: lucena.cloud/
  - Cache Level: Standard
  - Browser Cache TTL: 4 hours
  ```

**3. Check Origin Cache Headers**
- Ensure your nginx/go server sends proper `Cache-Control` headers
- Static assets: `Cache-Control: public, max-age=31536000`
- HTML: `Cache-Control: public, max-age=3600`

**4. Enable Cloudflare Caching Settings**
- Go to: Cloudflare Dashboard → Caching → Configuration
- ✅ Enable "Always Online"
- ✅ Enable "Browser Cache TTL: 4 hours"
- ✅ Enable "Caching Level: Standard"

---

### Phase 3: Re-Test After Optimization

**After implementing cache improvements:**

1. **Wait 24 hours** for Cloudflare cache to populate
2. **Re-run WebPageTest** from USA and China
3. **Compare results**:
   - Load time should decrease by 50-80%
   - Cache hit ratio should increase from ~8% to >85%
   - Bandwidth saved should increase from 8% to >85%

---

## 📈 Expected Results

### Before Optimization (Current State):
| Location | Load Time | Cache Hit | TTFB | Status |
|----------|-----------|-----------|------|--------|
| USA | 200-500ms | ~8% | 50-150ms | ⚠️ Slow |
| China | 400-800ms | ~8% | 200-400ms | ❌ Very Slow |

### After Optimization (Target):
| Location | Load Time | Cache Hit | TTFB | Status |
|----------|-----------|-----------|------|--------|
| USA | 50-150ms | >85% | 5-50ms | ✅ Fast |
| China | 100-300ms | >85% | 5-100ms | ✅ Acceptable |

---

## 🔧 Quick Testing Scripts

### Test from Command Line (USA perspective):

```bash
#!/bin/bash
# test-performance.sh

echo "🌍 Testing lucena.cloud Performance"
echo "===================================="

URL="https://lucena.cloud"

echo -e "\n📊 Homepage Test:"
curl -w "\nTime: %{time_total}s\nTTFB: %{time_starttransfer}s\n" \
  -o /dev/null -s -I "$URL" | grep -E "Time:|TTFB:|CF-Cache-Status"

echo -e "\n📊 API Test:"
curl -w "\nTime: %{time_total}s\nTTFB: %{time_starttransfer}s\n" \
  -o /dev/null -s -I "$URL/api/projects" | grep -E "Time:|TTFB:|CF-Cache-Status"

echo -e "\n📊 Agent-Bruno Chat Test:"
curl -X POST "$URL/api/chat" \
  -H "Content-Type: application/json" \
  -d '{"message":"test"}' \
  -w "\nTime: %{time_total}s\n" \
  -o /dev/null -s | grep "Time:"
```

### Test Agent-Bruno Specifically:

```bash
#!/bin/bash
# test-agent-bruno.sh

BASE_URL="https://lucena.cloud/api"

echo "🤖 Testing Agent-Bruno Performance"
echo "=================================="

# Health check
echo -e "\n1. Health Check:"
curl -w "\nTime: %{time_total}s\n" -s "$BASE_URL/chat/health" | jq .

# Chat test
echo -e "\n2. Chat Request:"
START=$(date +%s%N)
RESPONSE=$(curl -s -X POST "$BASE_URL/chat" \
  -H "Content-Type: application/json" \
  -d '{"message":"Hello, what can you help me with?"}')
END=$(date +%s%N)
DURATION=$((($END - $START) / 1000000))

echo "Response: $(echo $RESPONSE | jq -r '.response' | head -c 100)..."
echo "Duration: ${DURATION}ms"
echo "Full Response:"
echo $RESPONSE | jq .
```

---

## 🎯 Recommended Testing Workflow

### Daily Testing (Quick Check):
1. **Cloudflare Dashboard**: Check cache hit ratio (should be >85%)
2. **WebPageTest**: Run 1 test from USA and 1 from China
3. **Review**: Compare load times

### Weekly Testing (Comprehensive):
1. **WebPageTest**: Full test suite (3 runs from each location)
2. **GTmetrix**: Lighthouse report from both locations
3. **k6**: Automated load test from both perspectives
4. **Review**: Document any regressions

### After Each PR/Deployment:
1. **Run k6 smoke test**: Verify no regression
2. **Check Cloudflare Analytics**: Cache hit ratio should not decrease
3. **Quick WebPageTest**: 1 test from USA to verify performance

---

## 📝 Testing Checklist

### Initial Baseline:
- [ ] WebPageTest from USA (Dulles, VA)
- [ ] WebPageTest from China (Hong Kong)
- [ ] GTmetrix report from both locations
- [ ] Record current cache hit ratio (Cloudflare Dashboard)
- [ ] Record current load times
- [ ] Test agent-bruno chat endpoint
- [ ] Document all baseline metrics

### After Cache Optimization:
- [ ] Wait 24 hours for cache population
- [ ] Re-run all tests
- [ ] Compare results (should see 50-80% improvement)
- [ ] Verify cache hit ratio >85%
- [ ] Update baseline metrics

### Ongoing Monitoring:
- [ ] Set up Pingdom or similar (automated monitoring)
- [ ] Schedule weekly k6 tests
- [ ] Review Cloudflare Analytics weekly
- [ ] Track performance trends over time

---

## 🚨 Performance Targets

Based on your current metrics and industry standards:

| Metric | Target (USA) | Target (China) | Current (Origin) |
|--------|-------------|---------------|------------------|
| **Load Time** | < 2s | < 3s | _TBD_ |
| **TTFB** | < 200ms | < 500ms | ~5ms (origin) |
| **Cache Hit Ratio** | >85% | >85% | ~8% ❌ |
| **FCP** | < 1.8s | < 2.5s | _TBD_ |
| **LCP** | < 2.5s | < 4.0s | _TBD_ |
| **API Response** | < 500ms | < 1000ms | ~5ms ✅ |

---

## 🔗 Quick Links

### Testing Tools:
- **WebPageTest**: https://www.webpagetest.org
- **GTmetrix**: https://gtmetrix.com
- **Cloudflare Analytics**: https://dash.cloudflare.com (your dashboard)
- **Pingdom**: https://www.pingdom.com

### Documentation:
- [Homepage RED Metrics Report](./homepage-red-metrics-report.md)
- [Cloudflare Setup Guide](../flux/apps/homepage/docs/CLOUDFLARE_SETUP.md)
- [Performance Optimization](../flux/apps/homepage/docs/PERFORMANCE_OPTIMIZATION.md)

---

## 💡 Immediate Action Items

**Priority 1: Establish Baseline (See [BVL-326](https://linear.app/bvlucena/issue/BVL-326))**
1. ✅ Run WebPageTest from USA (Dulles, VA)
2. ✅ Run WebPageTest from China (Hong Kong)
3. ✅ Record all metrics
4. ✅ Document results

**Priority 2: Fix Caching (Do This After Baseline!)**
1. ✅ Check Cloudflare Page Rules are configured
2. ✅ Verify origin sends proper `Cache-Control` headers
3. ✅ Enable Cloudflare caching settings
4. ✅ Wait 24 hours, then re-test

**Priority 3: Set Up Continuous Monitoring**
1. ✅ Set up Pingdom or similar
2. ✅ Create k6 test scripts
3. ✅ Schedule regular tests
4. ✅ Set up alerts for regressions

---

**Next Steps**: Start with WebPageTest to get baseline metrics, then optimize caching, then re-test.
