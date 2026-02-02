# 🏗️ Architecture Decision Record: Why Kaniko?

**Decision**: Use Kaniko for container image builds instead of Docker-in-Docker or BuildKit

---

## 📋 Status

**Status**: ✅ Accepted  
**Date**: 2024-09-15  
**Decision Makers**: Platform Team, SRE Team, Security Team

---

## 🎯 Context

Knative Lambda needs to build container images dynamically from user-provided code. We evaluated several container build solutions to determine the best fit for our serverless platform.

### Requirements

**Must Have**:
- ✅ Secure (no privileged containers)
- ✅ Kubernetes-native (runs in pods)
- ✅ No Docker daemon required
- ✅ Supports Dockerfile syntax
- ✅ Push to remote registries (ECR)

**Nice to Have**:
- ✅ Layer caching
- ✅ Multi-platform builds
- ✅ Active community support
- ✅ Simple configuration

---

## 🔍 Options Considered

### Option 1: Docker-in-Docker (DinD)

**How it works**: Run Docker daemon inside a container.

**Pros**:
- ✅ Full Docker compatibility
- ✅ Well-documented
- ✅ Familiar to developers

**Cons**:
- ❌ **Requires privileged containers** (security risk)
- ❌ Complex setup (daemon management)
- ❌ Higher resource usage (daemon overhead)
- ❌ Security vulnerabilities (container escape)

**Verdict**: ❌ **Rejected** due to security concerns

---

### Option 2: BuildKit

**How it works**: Docker's next-generation build system.

**Pros**:
- ✅ Faster builds (parallel execution)
- ✅ Better caching
- ✅ Active development (Docker official)
- ✅ Multi-platform support

**Cons**:
- ⚠️ Requires privileged mode or rootless mode (complex)
- ⚠️ Rootless mode has limitations
- ⚠️ More complex configuration

**Verdict**: ⚠️ **Possible but complex**

---

### Option 3: Kaniko ✅

**How it works**: Builds container images in Kubernetes without Docker daemon.

**Pros**:
- ✅ **No Docker daemon required**
- ✅ **No privileged containers** (security)
- ✅ Runs as non-root user
- ✅ Kubernetes-native (runs in pods)
- ✅ Supports Dockerfile syntax
- ✅ Layer caching support
- ✅ Push directly to registries
- ✅ Active community (Google OSS)
- ✅ Proven at scale (GCP Cloud Build uses it)

**Cons**:
- ⚠️ Slightly slower than Docker (no daemon optimization)
- ⚠️ Some Dockerfile features not supported (rare edge cases)
- ⚠️ No build context cache (must re-download base images)

**Verdict**: ✅ **Selected**

---

### Option 4: Buildah

**How it works**: RedHat's daemonless container builder.

**Pros**:
- ✅ No daemon required
- ✅ Rootless support
- ✅ OCI-compliant

**Cons**:
- ⚠️ Less Kubernetes-native
- ⚠️ Smaller community vs Kaniko
- ⚠️ More complex scripting (no Dockerfile support out-of-box)

**Verdict**: ❌ **Not selected** (Kaniko better K8s integration)

---

## 🏆 Decision

**We chose Kaniko** because:

1. **Security First**: No Docker daemon = no privileged containers
2. **Kubernetes Native**: Designed to run in Kubernetes pods
3. **Simple Integration**: Works with existing Dockerfiles
4. **Production Proven**: Used by GCP Cloud Build at massive scale
5. **Active Community**: Google-backed, 13k+ GitHub stars

---

## 🔐 Security Analysis

### Kaniko Security Model

```
┌──────────────────────────────────────────────────┐
│  Kaniko Pod (Non-privileged)                      │
│                                                   │
│  User: kaniko (non-root)                         │
│  Capabilities: NONE                              │
│  Privileged: false                               │
│  ReadOnlyRootFilesystem: true                    │
│                                                   │
│  Process:                                        │
│  ├─ Read Dockerfile                              │
│  ├─ Fetch base image (docker.io)                │
│  ├─ Execute Dockerfile commands in userspace    │
│  ├─ Build layers                                 │
│  └─ Push to ECR                                  │
│                                                   │
│  No Docker socket access                         │
│  No host filesystem access                       │
│  No privileged operations                        │
└──────────────────────────────────────────────────┘
```

**vs Docker-in-Docker**:
```
┌──────────────────────────────────────────────────┐
│  Docker-in-Docker Pod (PRIVILEGED)               │
│                                                   │
│  User: root                                      │
│  Capabilities: ALL                               │
│  Privileged: true  ❌ SECURITY RISK              │
│  Volume: /var/run/docker.sock                    │
│                                                   │
│  Risks:                                          │
│  ├─ Container escape possible                   │
│  ├─ Host filesystem access                      │
│  ├─ Kernel exploits                             │
│  └─ Resource exhaustion                         │
└──────────────────────────────────────────────────┘
```

---

## 📊 Performance Comparison

### Build Time Benchmark (Python function with dependencies)

| Builder | Build Time | Resource Usage | Security |
|---------|-----------|----------------|----------|
| **Docker-in-Docker** | 45s | 1.5 GB RAM | ❌ Privileged |
| **BuildKit** | 40s | 1.2 GB RAM | ⚠️ Rootless complex |
| **Kaniko** | 55s | 1.0 GB RAM | ✅ Non-privileged |

**Verdict**: Kaniko is 20% slower but **significantly more secure**.

---

## 🛠️ Implementation

### Kaniko Job Template

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: kaniko-build-{{parser-id}}
spec:
  template:
    spec:
      restartPolicy: Never
      serviceAccountName: kaniko-builder  # IRSA for ECR
      
      # Fetch code from S3
      initContainers:
        - name: fetch-code
          image: amazon/aws-cli:2.15.0
          command: ["/bin/sh", "-c"]
          args:
            - |
              aws s3 sync s3://{{bucket}}/{{prefix}} /workspace/
          volumeMounts:
            - name: workspace
              mountPath: /workspace
          securityContext:
            runAsNonRoot: true
            runAsUser: 1000
            allowPrivilegeEscalation: false
      
      # Build with Kaniko
      containers:
        - name: kaniko
          image: gcr.io/kaniko-project/executor:v1.19.0
          args:
            - "--dockerfile=/workspace/Dockerfile"
            - "--context=/workspace"
            - "--destination={{ecr-repo}}:{{tag}}"
            - "--cache=true"
            - "--cache-ttl=24h"
            - "--compressed-caching=false"
            - "--snapshot-mode=redo"
            - "--use-new-run"
          volumeMounts:
            - name: workspace
              mountPath: /workspace
          resources:
            requests:
              memory: "1Gi"
              cpu: "500m"
            limits:
              memory: "4Gi"
              cpu: "2000m"
          securityContext:
            runAsNonRoot: true
            runAsUser: 1000
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop:
                - ALL
      
      volumes:
        - name: workspace
          emptyDir: {}
```

---

## ✅ Benefits Realized

### Security

- ✅ **No privileged containers** (eliminated container escape risk)
- ✅ **Non-root execution** (defense in depth)
- ✅ **Read-only filesystem** (immutability)
- ✅ **Dropped all capabilities** (minimal attack surface)

### Operational

- ✅ **Simple deployment** (just a Kubernetes Job)
- ✅ **Auto-cleanup** (TTL controller)
- ✅ **Parallel builds** (10+ concurrent jobs)
- ✅ **Metrics** (Prometheus integration)

### Developer Experience

- ✅ **Dockerfile compatibility** (no retraining)
- ✅ **Multi-language support** (Python, Node, Go)
- ✅ **Fast iteration** (cached layers)

---

## ⚠️ Limitations & Mitigations

### Limitation 1: Slower than Docker daemon

**Impact**: 10-20% slower builds  
**Mitigation**: 
- Enable layer caching (`--cache=true`)
- Use smaller base images (`python:3.9-slim` vs `python:3.9`)
- Pre-warm base images

### Limitation 2: No build context cache

**Impact**: Must re-download base images each build  
**Mitigation**:
- Use kaniko cache (`--cache-repo`)
- Mirror frequently-used images in ECR
- Future: Implement build cache PVC

### Limitation 3: Some Dockerfile features unsupported

**Impact**: Advanced BuildKit features not available  
**Examples**:
- `RUN --mount=type=cache` (BuildKit-specific)
- Multi-stage builds with `--target` (partial support)

**Mitigation**:
- Use standard Dockerfile syntax
- Document unsupported features
- Provide alternative patterns

---

## 🔮 Future Considerations

### Potential Enhancements

1. **Persistent cache** (v1.2.0)
   - Use PVC for Kaniko cache
   - 50% faster builds

2. **Multi-platform builds** (v1.3.0)
   - Build for ARM64 + AMD64
   - Support Apple Silicon

3. **BuildKit integration** (v2.0.0)
   - Evaluate rootless BuildKit
   - Compare performance/security

---

## 📚 References

- [Kaniko GitHub](https://github.com/GoogleContainerTools/kaniko)
- [Kaniko Documentation](https://github.com/GoogleContainerTools/kaniko/blob/main/README.md)
- [GCP Cloud Build Architecture](https://cloud.google.com/build/docs/how-builds-work)
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)

---

## 🔄 Review & Revision

**Last Reviewed**: October 29, 2025  
**Next Review**: January 2026  
**Owned By**: Platform Team

**Revision History**:
- 2024-09-15: Initial decision (ADR-001)
- 2025-10-29: Updated with production learnings

---

**Decision**: ✅ **Kaniko is the right choice for secure, Kubernetes-native container builds**

---

**Last Updated**: October 29, 2025  
**Version**: 1.0.0

