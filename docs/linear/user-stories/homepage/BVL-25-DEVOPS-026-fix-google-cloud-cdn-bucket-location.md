# BVL-25: Fix Google Cloud CDN Bucket Location

**Status**: Backlog  
**Priority**: ⚠️ High  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-25/fix-google-cloud-cdn-bucket-location  
**Created**: 2026-01-01T21:42:03.913Z  
**Updated**: 2026-01-01T21:42:03.913Z  
**Project**: homepage  

---

## CDN Optimization

Current bucket is in `us-central1` (Iowa), which is suboptimal for global distribution.

### Problem

* Far from Brazil origin
* Far from China (blocked anyway)
* Only optimal for USA users

### Solution

Use multi-region bucket or regional buckets:

* **Option A**: Multi-region (`-l US`)
* **Option B**: Regional buckets (us-central1, sa-east1, asia-east1)

### Implementation

```bash
# Option A: Multi-region
gsutil mb -p YOUR_PROJECT -c STANDARD -l US gs://lucena-cloud-assets

# Option B: Regional (requires migration)
# Create buckets in multiple regions and sync
```

### Expected Impact

* 30-50ms faster for USA users
* 20-40ms faster for Brazil users

### Documentation

* See `QUICK_FIXES_IMPLEMENTATION.md` Fix 2

### Priority

⚠️ **HIGH** - Do this week

---

## 🔐 Security Acceptance Criteria

- [ ] Bucket access controls configured properly
- [ ] Encryption at rest enabled for all buckets
- [ ] Encryption in transit enforced (TLS)
- [ ] CORS configuration for CDN buckets
- [ ] Bucket security monitoring implemented
- [ ] Secrets management for bucket credentials
- [ ] Security testing validates bucket security
- [ ] Threat model reviewed for CDN bucket security
- [ ] Security review completed before implementation
