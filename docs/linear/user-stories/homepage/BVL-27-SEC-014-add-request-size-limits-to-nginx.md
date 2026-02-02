# BVL-27: Add Request Size Limits to nginx

**Status**: Backlog  
**Priority**: ⚠️ High  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-27/add-request-size-limits-to-nginx  
**Created**: 2026-01-01T21:42:07.241Z  
**Updated**: 2026-01-01T21:42:07.241Z  
**Project**: homepage  

---

## DoS Protection

Add request size limits to prevent DoS attacks via large uploads.

### Implementation

```nginx
client_max_body_size 10M;
client_body_buffer_size 128k;
client_header_buffer_size 1k;
large_client_header_buffers 4 16k;
```

### Timeouts

```nginx
client_body_timeout 10s;
client_header_timeout 10s;
send_timeout 10s;
```

### Testing

```bash
# Should fail with 413 Request Entity Too Large
curl -X POST -d @large-file.bin https://lucena.cloud/api/upload
```

### Documentation

* See `SECURITY_HARDENING_GUIDE.md` Section 1.1

### Priority

⚠️ **HIGH** - Do this week

---

## 🔐 Security Acceptance Criteria

- [ ] Request size limits prevent DoS attacks
- [ ] Timeout configuration prevents resource exhaustion
- [ ] Buffer overflow protection implemented
- [ ] Proper error handling for oversized requests
- [ ] Security testing validates size limit effectiveness
- [ ] Threat model reviewed for size limit security
- [ ] Security review completed before implementation
