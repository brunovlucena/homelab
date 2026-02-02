# BVL-30: Block Common Attack Patterns in nginx

**Status**: Backlog  
**Priority**: ⚠️ High  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-30/block-common-attack-patterns-in-nginx  
**Created**: 2026-01-01T21:42:10.221Z  
**Updated**: 2026-01-01T21:42:10.221Z  
**Project**: homepage  

---

## Security Hardening

Block common attack patterns at nginx level before they reach the application.

### Patterns to Block

* SQL injection attempts
* XSS attempts
* Path traversal attempts
* Hidden file access (`.env`, `.git`, etc.)
* Backup file access

### Implementation

```nginx
# Block SQL injection
if ($query_string ~* "union.*select|insert.*into|drop.*table") {
    return 403;
}

# Block XSS
if ($query_string ~* "<script|javascript:|onload=") {
    return 403;
}

# Block path traversal
if ($uri ~* "\.\./|\.\.\\|\.\.%2f") {
    return 403;
}

# Block hidden files
location ~ /\. {
    deny all;
}
```

### Documentation

* See `SECURITY_HARDENING_GUIDE.md` Section 1.1

### Priority

⚠️ **HIGH** - Do this week

---

## 🔐 Security Acceptance Criteria

- [ ] Regex patterns optimized to prevent ReDoS attacks
- [ ] Pattern bypass techniques tested and mitigated
- [ ] All blocked requests logged for analysis
- [ ] Security testing validates pattern blocking
- [ ] Threat model reviewed for attack pattern blocking
- [ ] Security review completed before implementation
- [ ] Regular review and update of attack patterns
