# BVL-22: Add Rate Limiting to nginx

**Status**: Backlog  
**Priority**: 🔴 Urgent  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-22/add-rate-limiting-to-nginx  
**Created**: 2026-01-01T21:41:59.241Z  
**Updated**: 2026-01-01T21:59:38.127Z  
**Project**: homepage  

---

## 🎯 Objective

Implement rate limiting at nginx level to protect against DoS attacks and abuse by limiting request rates per IP address.

## 📊 Current State

| Metric | Current Value | Notes |
| -- | -- | -- |
| Rate limiting | Not implemented | Vulnerable to DoS attacks |
| Request rate control | None | No protection against abuse |
| Connection limits | Not configured | No connection throttling |

## 🎯 Target State

| Metric | Target Value | Priority |
| -- | -- | -- |
| General traffic limit | 10 req/s (burst: 20) | P0 |
| API endpoint limit | 5 req/s (burst: 10) | P0 |
| Static asset limit | 50 req/s (burst: 100) | P0 |
| Connection limit | 20 connections per IP | P0 |

## 📋 Requirements

- [ ] Configure rate limit zones for general, API, and static traffic
- [ ] Apply rate limits to appropriate location blocks
- [ ] Set connection limits per IP
- [ ] Configure burst allowances
- [ ] Return 429 status for rate limit violations
- [ ] Log rate limit violations

## 🔧 Implementation Steps

1. **Configure Rate Limit Zones**
   * Add `limit_req_zone` directives for general, API, and static traffic
   * Configure `limit_conn_zone` for connection limits
   * Set appropriate memory sizes and rates
2. **Apply Rate Limits**
   * Apply general limit to server block
   * Apply API limit to `/api/` location
   * Apply static limit to `/assets/` location
   * Configure connection limits
3. **Test and Monitor**
   * Test rate limit enforcement
   * Monitor rate limit violations
   * Adjust limits if needed

## ✅ Acceptance Criteria

- [ ] General traffic limited to 10 req/s
- [ ] API endpoints limited to 5 req/s
- [ ] Static assets limited to 50 req/s
- [ ] Connection limit of 20 per IP enforced
- [ ] 429 status returned when limits exceeded
- [ ] Rate limit violations logged

---

## 🔐 Security Acceptance Criteria

- [ ] Rate limiting prevents DoS attacks effectively
- [ ] Distributed rate limiting implemented (if needed)
- [ ] Rate limit headers included in responses (X-RateLimit-*)
- [ ] Rate limit bypass protection implemented
- [ ] Rate limit violations logged and monitored
- [ ] Security testing validates rate limiting effectiveness
- [ ] Threat model reviewed for rate limiting security
- [ ] Security review completed before implementation

## 🧪 Testing

### Test Commands

```bash
# Should see 429 after 10 requests
for i in {1..15}; do 
  curl -I https://lucena.cloud/api/projects
done

# Should see 429 after 5 API requests
for i in {1..10}; do 
  curl -I https://lucena.cloud/api/chat
done

# Test connection limit
# Open 25 connections simultaneously
# Should see connection refused after 20
```

### Expected Results

* Rate limit violations return 429 status
* Appropriate error message in response
* Rate limit headers in response (if configured)
* Logs show rate limit violations

## 📚 Documentation

* `flux/infrastructure/homepage/SECURITY_HARDENING_GUIDE.md` - Section 1.1
* `flux/infrastructure/homepage/SECURITY_QUICK_START.md` - Step 3
* `flux/apps/homepage/k8s/kustomize/base/frontend-nginx-configmap-secure.yaml` - Template

## 🔗 Related Issues

* Related to: BVL-20 (Security Headers)
* Related to: BVL-21 (CORS Configuration)
* Related to: BVL-27 (Request Size Limits)

## 📅 Timeline

* **Quarter**: Q1 2025
* **Estimated Effort**: 2 hours
* **Target Completion**: 2025-01-02
