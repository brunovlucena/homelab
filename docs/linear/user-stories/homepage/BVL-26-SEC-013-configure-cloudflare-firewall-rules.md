# BVL-26: Configure Cloudflare Firewall Rules

**Status**: Backlog  
**Priority**: ⚠️ High  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-26/configure-cloudflare-firewall-rules  
**Created**: 2026-01-01T21:42:07.026Z  
**Updated**: 2026-01-01T21:42:07.026Z  
**Project**: homepage  

---

## Edge Security

Configure Cloudflare firewall rules to block common attack patterns at the edge.

### Required Rules

1. **Block SQL Injection**: Block requests with SQL injection patterns
2. **Block XSS**: Block requests with XSS patterns
3. **Rate Limit Chat API**: Challenge after 5 requests/minute for `/api/chat`

### Implementation

Go to: **Security** → **WAF** → **Tools** → **Firewall Rules**

### Rule Examples

```
# SQL Injection
(http.request.uri.query contains "union select") or 
(http.request.uri.query contains "drop table")

# XSS
(http.request.uri.query contains "<script") or 
(http.request.uri.query contains "javascript:")

# Rate Limit Chat
http.request.uri.path contains "/api/chat"
```

### Documentation

* See `SECURITY_HARDENING_GUIDE.md` Phase 2
* See `SECURITY_QUICK_START.md` Step 3

### Priority

⚠️ **HIGH** - Do this week

---

## 🔐 Security Acceptance Criteria

- [ ] Firewall rules tested before deployment
- [ ] False positive monitoring and alerting
- [ ] Blocked requests logged for analysis
- [ ] Firewall rule effectiveness validated
- [ ] Security testing validates firewall rules
- [ ] Threat model reviewed for firewall rules
- [ ] Security review completed before implementation
- [ ] Regular review and update of firewall rules
