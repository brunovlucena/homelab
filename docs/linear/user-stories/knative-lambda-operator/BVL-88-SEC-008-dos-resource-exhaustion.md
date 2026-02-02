# SEC-008: Denial of Service & Resource Exhaustion Testing

**Priority**: P0 | **Status**: 📋 Backlog K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-252/sec-008-denial-of-service-and-resource-exhaustion-testing

**Priority:** P1 | **Story Points:** 5

## 📋 User Story

**As a** Principal Pentester  
**I want to** validate that the system is protected against denial of service attacks  
**So that** service availability is maintained under attack conditions

## 🎯 Acceptance Criteria

### AC1: Rate Limiting Protection
**Given** APIs are accessible  
**When** sending excessive requests  
**Then** rate limits should prevent service degradation

**Security Tests:**
- ✅ HTTP rate limiting enforced (429 Too Many Requests)
- ✅ CloudEvent rate limiting active
- ✅ Per-IP rate limiting
- ✅ Per-user rate limiting
- ✅ Distributed rate limiting (Redis-based)

**Attack Scenarios:**
```bash
# HTTP flood
for i in {1..1000}; do
  curl http://api/endpoint &
done
# Expected: 429 after rate limit exceeded

# Slowloris attack
slowhttptest -c 1000 -H -g -o slowloris.html \
  -i 10 -r 200 -t GET -u http://api/endpoint
# Expected: Connections limited
```

### AC2: Resource Quota Enforcement
**Given** Kubernetes resources have limits  
**When** attempting to exhaust cluster resources  
**Then** quotas should prevent resource starvation

**Security Tests:**
- ✅ Namespace ResourceQuota enforced
- ✅ LimitRange prevents oversized pods
- ✅ CPU limits enforced
- ✅ Memory limits enforced
- ✅ Storage limits enforced
- ✅ Pod count limits enforced

**Expected Quotas:**
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: compute-quota
spec:
  hard:
    requests.cpu: "10"
    requests.memory: "20Gi"
    limits.cpu: "20"
    limits.memory: "40Gi"
    pods: "50"
    persistentvolumeclaims: "10"
```

### AC3: Pod Disruption Budget Protection
**Given** critical services must remain available  
**When** testing service resilience  
**Then** PodDisruptionBudgets should maintain availability

**Security Tests:**
- ✅ PDB defined for critical services
- ✅ Minimum replicas maintained during disruption
- ✅ Voluntary disruptions controlled
- ✅ Rolling updates don't violate PDB

**Attack Scenarios:**
- ❌ Delete multiple pods simultaneously
- ❌ Drain nodes causing service outage
- ❌ Trigger mass pod evictions

### AC4: Queue Flood Prevention
**Given** events are processed via RabbitMQ  
**When** flooding queues with messages  
**Then** queue limits should prevent overflow

**Security Tests:**
- ✅ Queue length limits enforced
- ✅ Message TTL configured
- ✅ Dead letter queue configured
- ✅ Consumer prefetch limits set
- ✅ Memory limits on RabbitMQ

**Attack Scenarios:**
```bash
# Queue flood
for i in {1..100000}; do
  publish_event "build.created" "{\"data\":\"$i\"}"
done
# Expected: Queue limit reached, messages rejected
```

### AC5: Connection Limit Protection
**Given** services accept network connections  
**When** opening excessive connections  
**Then** connection limits should prevent exhaustion

**Security Tests:**
- ✅ TCP connection limits per service
- ✅ HTTP connection pool limits
- ✅ Database connection pool limits
- ✅ Connection timeout enforcement
- ✅ Keep-alive limits

**Expected Limits:**
- HTTP: 1000 concurrent connections
- Database: 100 connections per pool
- RabbitMQ: 500 connections per vhost

### AC6: CPU/Memory Bomb Prevention
**Given** user code executes in containers  
**When** attempting resource exhaustion attacks  
**Then** limits should prevent host impact

**Attack Scenarios:**
- ❌ Fork bomb: `:(){: | :&};:`
- ❌ Memory bomb: `stress --vm 10 --vm-bytes 10G`
- ❌ CPU burn: `yes > /dev/null &`
- ❌ Disk fill: `dd if=/dev/zero of=bigfile bs=1M count=100000`

**Protection Mechanisms:**
- ✅ PID limits (`pids.max` in cgroup)
- ✅ Memory limits (OOMKiller)
- ✅ CPU quotas
- ✅ Ephemeral storage limits

### AC7: Slowloris/Slow POST Protection
**Given** HTTP services may be vulnerable to slow attacks  
**When** sending intentionally slow requests  
**Then** timeouts should terminate slow connections

**Security Tests:**
- ✅ Request timeout enforced (30s)
- ✅ Header timeout enforced (10s)
- ✅ Body read timeout enforced (60s)
- ✅ Idle connection timeout (120s)
- ✅ Slow client detection

### AC8: Amplification Attack Prevention
**Given** services may be used in amplification attacks  
**When** testing for amplification vectors  
**Then** responses should not amplify requests

**Security Tests:**
- ✅ No DNS amplification (recursive queries disabled)
- ✅ No NTP amplification (`monlist` disabled)
- ✅ Response size limited
- ✅ Source IP validation (no spoofing)

## 🔴 Attack Surface Analysis

### DoS Attack Vectors

1. **Application Layer (L7)**
   - HTTP flood
   - Slowloris
   - API abuse

2. **Transport Layer (L4)**
   - SYN flood
   - Connection exhaustion
   - UDP flood

3. **Resource Exhaustion**
   - CPU saturation
   - Memory exhaustion
   - Disk fill
   - Process limits

4. **Queue Flooding**
   - RabbitMQ queue overflow
   - Message storm
   - Dead letter queue abuse

5. **Database**
   - Connection pool exhaustion
   - Expensive queries
   - Table lock contention

## 🛠️ Testing Tools

### Load Testing
```bash
# HTTP load test
wrk -t12 -c400 -d30s http://api/endpoint

# k6 load test
k6 run --vus 100 --duration 30s load-test.js

# Locust distributed load test
locust -f locustfile.py --host=http://api
```

### DoS Attack Simulation
```bash
# SYN flood (requires root)
hping3 -S --flood -p 80 target-ip

# Slowloris
slowhttptest -c 1000 -H -i 10 -r 200 -t GET \
  -u http://target/endpoint

# HTTP flood
ab -n 100000 -c 1000 http://api/endpoint
```

### Resource Exhaustion
```bash
# CPU stress
kubectl exec -it <pod> -- stress --cpu 8 --timeout 60s

# Memory stress
kubectl exec -it <pod> -- stress --vm 4 --vm-bytes 1G --timeout 60s

# Fork bomb (in test environment only!)
kubectl exec -it <pod> -- sh -c ':(){: | :&};:'
# Should be killed by PID limit
```

### Queue Flooding
```bash
# Flood RabbitMQ queue
for i in {1..10000}; do
  curl -X POST http://rabbitmq:15672/api/exchanges/%2f/amq.default/publish \
    -u guest:guest \
    -d '{"properties":{},"routing_key":"test.queue","payload":"test","payload_encoding":"string"}'
done
```

## 📊 Success Metrics

- **Zero** service outages from DoS attacks
- **100%** rate limiting enforced
- **100%** resource quotas respected
- **<5%** legitimate request rejection rate
- **<1s** average response time under attack

## 🚨 Incident Response

If DoS attack is detected:

1. **Immediate** (< 1 min)
   - Enable emergency rate limiting
   - Block attacking IPs
   - Scale up replicas

2. **Short-term** (< 5 min)
   - Activate DDoS protection (CloudFlare, AWS Shield)
   - Review attack patterns
   - Implement temporary blocks

3. **Long-term** (< 1 hour)
   - Analyze attack vectors
   - Update rate limiting rules
   - Implement additional protections

## 📚 Related Stories

- **SEC-003:** API Security & CORS Misconfiguration
- **SEC-007:** Network Segmentation & Data Exfiltration
- **BACKEND-005:** Rate Limiting
- **SRE-002:** Performance Tuning

## 🔗 References

- [OWASP DoS Cheatsheet](https://cheatsheetseries.owasp.org/cheatsheets/Denial_of_Service_Cheat_Sheet.html)
- [Kubernetes Resource Quotas](https://kubernetes.io/docs/concepts/policy/resource-quotas/)
- [Kubernetes Limit Ranges](https://kubernetes.io/docs/concepts/policy/limit-range/)
- [Rate Limiting Patterns](https://cloud.google.com/architecture/rate-limiting-strategies-techniques)

---

**Test File:** `internal/security/security_008_dos_resource_exhaustion_test.go`  
**Owner:** Security Team  
**Last Updated:** October 29, 2025

