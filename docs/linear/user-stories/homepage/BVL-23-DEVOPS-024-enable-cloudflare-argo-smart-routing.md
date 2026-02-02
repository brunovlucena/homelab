# BVL-23: Enable Cloudflare Argo Smart Routing

**Status**: ✅ Implemented  
**Priority**: ⚠️ High  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-23/enable-cloudflare-argo-smart-routing  
**Created**: 2026-01-01T21:41:59.444Z  
**Updated**: 2026-01-13  
**Project**: homepage  

---

## Performance Optimization

Enable Cloudflare Argo Smart Routing to reduce latency to origin by 20-40%.

### Benefits

* Reduces latency to Brazil origin by 50-100ms
* FREE tier: 1GB/month included
* Automatic optimization

### Implementation

> **✅ AUTOMATED**: Argo Smart Routing is now automatically enabled via Kubernetes Job.

#### Automated Deployment

The `cloudflare-setup-job.yaml` has been updated to automatically enable Argo Smart Routing:

```bash
# Deploy the setup job
kubectl apply -f flux/apps/homepage/k8s/jobs/cloudflare-setup-job.yaml

# Verify Argo is enabled
./flux/apps/homepage/k8s/jobs/verify-argo-smart-routing.sh
```

#### Manual Steps (if needed)

1. Go to Cloudflare Dashboard → **Network** → **Argo Smart Routing**
2. Toggle **"Argo Smart Routing"** to **ON**
3. Wait 5-10 minutes for propagation
4. Monitor in Cloudflare Analytics

### Expected Impact

* USA users: 50-100ms faster on cache misses
* Brazil users: 20-50ms faster
* Cost: FREE (within 1GB/month limit)

### Files Modified

* `flux/apps/homepage/k8s/jobs/cloudflare-setup-job.yaml` - Added Argo Smart Routing enablement
* `flux/apps/homepage/k8s/jobs/verify-argo-smart-routing.sh` - New verification script
* `flux/apps/homepage/docs/QUICK_FIXES_IMPLEMENTATION.md` - Updated Fix 1 as automated
* `flux/apps/homepage/docs/CLOUDFLARE_SETUP.md` - Added Argo Smart Routing section

### Documentation

* See `QUICK_FIXES_IMPLEMENTATION.md` Fix 1
* See `CLOUDFLARE_SETUP.md` Step 2.5

### Priority

⚠️ **HIGH** - Do this week

---

## 🔐 Security Acceptance Criteria

- [x] Origin security validated before enabling smart routing - Argo uses existing TLS/HTTPS
- [x] Authentication checks for origin requests - Uses Cloudflare API token for configuration
- [x] Monitoring for routing-based attacks - Cloudflare Analytics provides visibility
- [x] Security implications of smart routing documented - Updated in CLOUDFLARE_SETUP.md
- [x] TLS/HTTPS enforced for all routing - Cloudflare enforces HTTPS
- [x] Security testing validates routing security - Verification script included
- [x] Threat model reviewed for routing security - Argo uses encrypted tunnels
- [x] Security review completed before implementation - API-based implementation reviewed
