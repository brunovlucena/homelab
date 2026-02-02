# BVL-24: Set Up Alibaba Cloud CDN for China

**Status**: Backlog  
**Priority**: ⚠️ High  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-24/set-up-alibaba-cloud-cdn-for-china  
**Created**: 2026-01-01T21:42:03.718Z  
**Updated**: 2026-01-01T21:42:03.718Z  
**Project**: homepage  

---

## China Market CDN

Set up Alibaba Cloud CDN to serve Chinese users with optimal performance (<100ms latency).

### Implementation Steps

1. Create Alibaba Cloud account
2. Create OSS bucket in China region (Hangzhou or Beijing)
3. Upload assets to OSS
4. Enable CDN
5. Configure environment variable: `VITE_CDN_BASE_URL_CN`
6. Test from China (VPN or testing service)

### Expected Results

* China users: <100ms latency (vs 300-500ms before)
* 200-300ms improvement
* Cost: ~$0.50-1.00/month after free tier

### Documentation

* See `CHINA_CDN_SETUP.md` (complete guide)
* See `CHINA_CDN_QUICK_START.md` (30-minute setup)

### Priority

⚠️ **HIGH** - Do this week

---

## 🔐 Security Acceptance Criteria

- [ ] All assets encrypted at rest in China CDN
- [ ] All assets encrypted in transit (TLS)
- [ ] Access control lists (ACL) configured for CDN assets
- [ ] Data sovereignty compliance documented
- [ ] Monitoring for unauthorized CDN access
- [ ] Secrets management for CDN credentials
- [ ] Security testing validates CDN security
- [ ] Threat model reviewed for China CDN security
- [ ] Security review completed before implementation
