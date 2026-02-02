# BVL-21: Fix CORS Configuration - Remove Wildcard

**Status**: Backlog  
**Priority**: 🔴 Urgent  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-21/fix-cors-configuration-remove-wildcard  
**Created**: 2026-01-01T21:41:56.655Z  
**Updated**: 2026-01-01T21:59:38.168Z  
**Project**: homepage  

---

## 🎯 Objective

Fix insecure CORS configuration in nginx that allows any origin to access resources, restricting access to only `https://lucena.cloud`.

## 📊 Current State

| Metric | Current Value | Notes |
| -- | -- | -- |
| CORS policy | Wildcard (`*`) | Allows any origin (security risk) |
| Origin validation | None | No origin checking |
| Preflight handling | Basic | May allow unauthorized origins |

## 🎯 Target State

| Metric | Target Value | Priority |
| -- | -- | -- |
| CORS policy | Restricted to `https://lucena.cloud` | P0 |
| Origin validation | Strict validation | P0 |
| Preflight handling | Secure OPTIONS handling | P0 |

## 📋 Requirements

- [ ] Remove wildcard (`*`) from CORS header
- [ ] Restrict to `https://lucena.cloud` only
- [ ] Update OPTIONS preflight handling
- [ ] Add proper CORS headers for allowed origin
- [ ] Test with unauthorized origins (should fail)
- [ ] Test with authorized origin (should succeed)

## 🔧 Implementation Steps

1. **Update nginx ConfigMap**
   * Locate `/storage/` location block
   * Replace `Access-Control-Allow-Origin *` with `Access-Control-Allow-Origin "https://lucena.cloud"`
   * Update OPTIONS method handling
   * Add proper CORS headers
2. **Deploy and Test**
   * Apply updated ConfigMap
   * Restart nginx pods
   * Test with unauthorized origin
   * Test with authorized origin
3. **Documentation**
   * Update security documentation
   * Add to security hardening guide

## ✅ Acceptance Criteria

- [ ] CORS header restricted to `https://lucena.cloud`
- [ ] Unauthorized origins receive 403 or CORS error
- [ ] Authorized origin can access resources
- [ ] OPTIONS preflight requests handled correctly
- [ ] No wildcard (`*`) in CORS configuration

---

## 🔐 Security Acceptance Criteria

- [ ] CORS configuration restricts to authorized origin only
- [ ] Credential handling configured properly (`Access-Control-Allow-Credentials`)
- [ ] Preflight caching configured appropriately
- [ ] CORS error logging implemented
- [ ] CORS bypass attempts logged and monitored
- [ ] Security testing validates CORS restrictions
- [ ] Threat model reviewed for CORS security implications
- [ ] Security review completed before implementation

## 🧪 Testing

### Test Commands

```bash
# Should fail (unauthorized origin)
curl -H "Origin: https://evil.com" -I https://lucena.cloud/storage/test.jpg

# Should succeed (authorized origin)
curl -H "Origin: https://lucena.cloud" -I https://lucena.cloud/storage/test.jpg

# Test preflight
curl -X OPTIONS -H "Origin: https://lucena.cloud" \
  -H "Access-Control-Request-Method: GET" \
  -I https://lucena.cloud/storage/test.jpg
```

### Expected Results

* Unauthorized origins receive CORS error or 403
* Authorized origin receives 200 OK with CORS headers
* Preflight requests return proper CORS headers
* No wildcard in response headers

## 📚 Documentation

* `flux/infrastructure/homepage/SECURITY_HARDENING_GUIDE.md` - Section 1.2
* `flux/infrastructure/homepage/SECURITY_QUICK_START.md` - Step 2
* `flux/apps/homepage/k8s/kustomize/base/frontend-nginx-configmap-secure.yaml` - Template

## 🔗 Related Issues

* Related to: BVL-20 (Security Headers)
* Related to: BVL-22 (Rate Limiting)

## 📅 Timeline

* **Quarter**: Q1 2025
* **Estimated Effort**: 1 hour
* **Target Completion**: 2025-01-02
