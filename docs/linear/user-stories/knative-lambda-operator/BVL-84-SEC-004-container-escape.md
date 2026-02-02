# SEC-004: Container Escape & Privilege Escalation Testing

**Priority**: P0 | **Status**: 📋 Backlog K  | **Story Points**: 13
**Linear URL**: https://linear.app/bvlucena/issue/BVL-248/sec-004-container-escape-and-privilege-escalation-testing

**Priority:** P0 | **Story Points:** 13

## 📋 User Story

**As a** Principal Pentester  
**I want to** validate that containers cannot escape to the host or escalate privileges  
**So that** the Kubernetes cluster and underlying nodes remain secure

## 🎯 Acceptance Criteria

### AC1: Container Security Context Enforcement
**Given** containers run with security contexts  
**When** attempting to run privileged containers  
**Then** privileged mode should be blocked

**Security Tests:**
- ✅ `privileged: true` blocked by admission controller
- ✅ `allowPrivilegeEscalation: false` enforced
- ✅ Containers run as non-root user
- ✅ Read-only root filesystem enforced
- ✅ Seccomp profile applied
- ✅ AppArmor/SELinux enabled

**Pod Security Standards:**
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 65534
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  seccompProfile:
    type: RuntimeDefault
```

### AC2: Capability Restriction
**Given** Linux capabilities provide granular privileges  
**When** attempting to add dangerous capabilities  
**Then** capabilities should be restricted

**Blocked Capabilities:**
- ❌ `CAP_SYS_ADMIN` (god mode)
- ❌ `CAP_SYS_MODULE` (load kernel modules)
- ❌ `CAP_SYS_RAWIO` (raw I/O)
- ❌ `CAP_SYS_PTRACE` (ptrace any process)
- ❌ `CAP_SYS_BOOT` (reboot)
- ❌ `CAP_NET_ADMIN` (network config)
- ❌ `CAP_DAC_OVERRIDE` (bypass file permissions)

**Allowed Capabilities:**
- ✅ `CAP_NET_BIND_SERVICE` (bind <1024 ports) - if needed
- ✅ `CAP_CHOWN` - if needed
- ✅ Default: ALL capabilities dropped

### AC3: Host Filesystem Protection
**Given** containers may mount volumes  
**When** attempting to mount sensitive host paths  
**Then** dangerous mounts should be blocked

**Blocked Host Paths:**
- ❌ `/` (root)
- ❌ `/proc` (process info)
- ❌ `/sys` (system info)
- ❌ `/dev` (devices)
- ❌ `/etc` (system config)
- ❌ `/var/run/docker.sock` (Docker socket)
- ❌ `/var/run/containerd.sock` (containerd socket)
- ❌ `/run/containerd` (container runtime)

**Attack Scenarios:**
```yaml
# Attempt to mount Docker socket
volumes:
- name: docker-sock
  hostPath:
    path: /var/run/docker.sock
```

### AC4: Host Network/PID/IPC Isolation
**Given** pods can request host namespaces  
**When** attempting to use host namespaces  
**Then** host namespace access should be denied

**Security Tests:**
- ✅ `hostNetwork: true` blocked
- ✅ `hostPID: true` blocked
- ✅ `hostIPC: true` blocked
- ✅ `hostPort` usage restricted
- ✅ Pods cannot see host processes
- ✅ Pods cannot see host network interfaces

### AC5: Kernel Exploitation Prevention
**Given** kernel vulnerabilities may allow escape  
**When** attempting kernel exploits  
**Then** exploit attempts should fail

**Attack Vectors to Test:**
- ❌ Dirty COW (CVE-2016-5195)
- ❌ Dirty Pipe (CVE-2022-0847)
- ❌ runc escape (CVE-2019-5736)
- ❌ Shocker container escape
- ❌ /proc/self/exe symlink manipulation
- ❌ cgroups escape techniques

**Protection Mechanisms:**
- ✅ Kernel version ≥ 5.10 (patched)
- ✅ Container runtime updated (containerd/CRI-O)
- ✅ Seccomp filtering dangerous syscalls
- ✅ AppArmor/SELinux policies active

### AC6: Container Runtime Security
**Given** container runtime manages container lifecycle  
**When** attempting to manipulate the runtime  
**Then** runtime should be protected

**Security Tests:**
- ✅ Runtime socket not accessible from containers
- ✅ Runtime API authentication enabled
- ✅ Container breakout via runtime bugs prevented
- ✅ Image pull restricted to trusted registries
- ✅ No privileged runtime options exposed

**Attack Scenarios:**
- ❌ runc/containerd socket access
- ❌ Runtime API exploitation
- ❌ Container image tampering
- ❌ Malicious image pull

### AC7: Resource Limit Enforcement
**Given** resource limits prevent container abuse  
**When** attempting to consume excessive resources  
**Then** limits should be enforced

**Security Tests:**
- ✅ CPU limits enforced
- ✅ Memory limits enforced
- ✅ PID limits enforced (prevent fork bombs)
- ✅ Ephemeral storage limits enforced
- ✅ Network bandwidth limits enforced (if available)

**Expected Limits:**
```yaml
resources:
  limits:
    cpu: "2000m"
    memory: "2Gi"
    ephemeral-storage: "10Gi"
  requests:
    cpu: "100m"
    memory: "256Mi"
```

### AC8: Parser Sandbox Isolation
**Given** user-supplied parsers execute in containers  
**When** parser attempts to break sandbox  
**Then** escape attempts should fail

**Sandbox Restrictions:**
- ✅ No network access (except S3)
- ✅ No filesystem write (except /tmp)
- ✅ No process creation
- ✅ Limited syscalls via seccomp
- ✅ Resource limits enforced
- ✅ Timeout enforced (max 5 min)

**Attack Scenarios:**
```python
# Attempt host escape
import os
os.system('curl http://metadata.amazonaws.com/')  # Should fail

# Attempt file system escape
os.system('mount /dev/sda1 /mnt')  # Should fail

# Attempt network pivot
os.system('nc -lvp 1337')  # Should fail
```

## 🔴 Attack Surface Analysis

### Container Escape Vectors

1. **Kernel Vulnerabilities**
   - Dirty COW / Dirty Pipe
   - Use-after-free bugs
   - Race conditions

2. **Container Runtime Bugs**
   - runc vulnerabilities
   - containerd escape
   - CRI-O issues

3. **Misconfigured Security Contexts**
   - Privileged containers
   - Excessive capabilities
   - Host namespace sharing

4. **Mounted Host Paths**
   - Docker socket
   - Procfs/sysfs
   - Device files

5. **User-Supplied Code**
   - Parser execution
   - Build scripts
   - Lambda functions

## 🛠️ Testing Tools

### Container Security Testing
```bash
# Check pod security context
kubectl get pod <pod> -o json | \
  jq '.spec.securityContext'

# Verify no privileged containers
kubectl get pods --all-namespaces -o json | \
  jq -r '.items[] | select(.spec.containers[].securityContext.privileged==true) | "\(.metadata.namespace)/\(.metadata.name)"'

# Check capabilities
kubectl get pod <pod> -o json | \
  jq '.spec.containers[].securityContext.capabilities'

# Test container escape
kubectl exec -it <pod> -- sh
# Try: mount, chroot, nsenter, unshare
```

### Automated Scanning
```bash
# kube-bench (CIS Kubernetes Benchmark)
kubectl apply -f https://raw.githubusercontent.com/aquasecurity/kube-bench/main/job.yaml

# kubesec (Kubernetes manifest security scanner)
kubesec scan pod.yaml

# Falco (runtime security)
kubectl logs -n falco -l app=falco

# Trivy (container vulnerability scanner)
trivy image knative-lambda-builder:latest
```

## 📊 Success Metrics

- **Zero** container escape vulnerabilities
- **Zero** privileged containers in production
- **100%** security contexts enforced
- **Zero** dangerous capabilities granted
- **100%** seccomp/AppArmor coverage

## 🚨 Incident Response

If container escape is detected:

1. **Immediate** (< 2 min)
   - Kill compromised containers
   - Cordon affected nodes
   - Enable enhanced monitoring

2. **Short-term** (< 15 min)
   - Isolate affected workloads
   - Collect forensics
   - Patch vulnerability

3. **Long-term** (< 1 hour)
   - Audit all security contexts
   - Update admission controllers
   - Rotate compromised credentials

## 📚 Related Stories

- **SEC-005:** Cloud Resource Access Control
- **SEC-006:** Secrets Exposure & Credential Leakage
- **SEC-007:** Network Segmentation & Data Exfiltration
- **SRE-014:** Security Incident Response

## 🔗 References

- [Kubernetes Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Container Escape Techniques](https://blog.trailofbits.com/2019/07/19/understanding-docker-container-escapes/)
- [Linux Capabilities](https://man7.org/linux/man-pages/man7/capabilities.7.html)
- [Seccomp](https://kubernetes.io/docs/tutorials/security/seccomp/)
- [AppArmor](https://kubernetes.io/docs/tutorials/security/apparmor/)

---

**Test File:** `internal/security/security_004_container_escape_test.go`  
**Owner:** Security Team  
**Last Updated:** October 29, 2025

