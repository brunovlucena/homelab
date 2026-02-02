# SRE-014: Security Incident Response

**Status**: Backlog
**Priority**: P0
**Story Points**: 8  
**Linear URL**: https://linear.app/bvlucena/issue/BVL-178/sre-014-security-incident-response  
**Created**: 2026-01-19  
**Updated**: 2026-01-19  
**Project**: knative-lambda-operator

---


## 📋 User Story

**As a** SRE Engineer  
**I want to** security incident response  
**So that** I can improve system reliability, security, and performance

---



## 🎯 Acceptance Criteria

- [ ] All requirements implemented and tested
- [ ] Documentation updated
- [ ] Code reviewed and approved
- [ ] Deployed to target environment

---


## Overview

This runbook provides comprehensive procedures for responding to security incidents in the Knative Lambda infrastructure, including detection, investigation, containment, eradication, recovery, and post-incident review.

## Incident Classification

### Severity Levels | Severity | Description | Response Time | Examples | |---------- | ------------- | --------------- | ---------- | | **P0 - Critical** | Active breach, data exfiltration, system compromise | < 15 minutes | Root access compromised, ransomware detected | | **P1 - High** | Potential breach, significant vulnerability | < 1 hour | Suspicious privileged access, CVE exploitation attempt | | **P2 - Medium** | Security policy violation, minor vulnerability | < 4 hours | Failed authentication spike, misconfiguration | | **P3 - Low** | Informational, suspicious activity | < 24 hours | Port scanning, unusual traffic patterns | ## Security Incident Playbooks

### Playbook 1: Unauthorized Access Attempt

**Triggers:**
- Multiple failed authentication attempts
- Login from unusual location/IP
- Privilege escalation detected
- Service account misuse

**Immediate Actions (< 15 min):**

```bash
# Step 1: Identify compromised account
kubectl get events --all-namespaces | grep -i "fail\ | unauthorized"
kubectl logs -n kube-system pod/<auth-pod> | grep "Failed password"

# Step 2: Disable compromised account
kubectl delete serviceaccount <compromised-sa> -n <namespace>
kubectl patch user <username> -p '{"active": false}'

# Step 3: Revoke all active sessions
kubectl delete tokens --all -n <namespace>

# Step 4: Enable enhanced logging
kubectl patch deployment/<app> -n <namespace> \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"<container>","env":[{"name":"LOG_LEVEL","value":"debug"}]}]}}}}'

# Step 5: Isolate affected pods
kubectl label pod <pod-name> -n <namespace> quarantine=true
kubectl patch networkpolicy <policy> -n <namespace> \
  -p '{"spec":{"podSelector":{"matchLabels":{"quarantine":"true"}},"policyTypes":["Ingress","Egress"],"ingress":[],"egress":[]}}'
```

**Containment (< 5 min):**
```bash
# Isolate affected pods
kubectl label pod <pod-name> -n <namespace> quarantine=true
kubectl patch networkpolicy <policy> -n <namespace> \
  -p '{"spec":{"podSelector":{"matchLabels":{"quarantine":"true"}},"policyTypes":["Ingress","Egress"],"ingress":[],"egress":[]}}'
```

**Investigation Steps (< 1 hour):**

```bash
# Collect forensics data
kubectl logs <pod-name> -n <namespace> --all-containers > forensics-logs-$(date +%Y%m%d-%H%M%S).txt
kubectl describe pod <pod-name> -n <namespace> > forensics-pod-$(date +%Y%m%d-%H%M%S).yaml
kubectl get events -n <namespace> --sort-by='.lastTimestamp' > forensics-events-$(date +%Y%m%d-%H%M%S).txt

# Check audit logs
kubectl logs -n kube-system kube-apiserver-* | grep <username> > audit-$(date +%Y%m%d-%H%M%S).txt

# Review access patterns
kubectl get rolebindings,clusterrolebindings --all-namespaces -o json | \
  jq -r '.items[] | select(.subjects[]?.name=="<username>") | .metadata.name'

# Check for privilege escalation
kubectl auth can-i --list --as=<username>

# Identify lateral movement
kubectl get pods --all-namespaces -o json | \
  jq -r '.items[] | select(.spec.serviceAccountName=="<compromised-sa>") | "\(.metadata.namespace)/\(.metadata.name)"'
```

**Resolution (< 2 hours):**
```bash
# Reset credentials
kubectl delete secret <compromised-secret> -n <namespace>
kubectl create secret generic <new-secret> --from-literal=password=$(openssl rand -base64 32)

# Patch vulnerability
kubectl patch deployment <deployment> -n <namespace> --patch-file security-fix.yaml

# Remove quarantine
kubectl label pod <pod-name> -n <namespace> quarantine-
```

### Playbook 2: Malware/Ransomware Detected

**Immediate Actions (< 5 min):**

```bash
# Step 1: Isolate infected nodes
kubectl cordon <node-name>
kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data

# Step 2: Network isolation
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all-quarantine
  namespace: <affected-namespace>
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
EOF
  
# Step 3: Snapshot infected systems for forensics
kubectl get pod <infected-pod> -n <namespace> -o yaml > infected-pod-snapshot.yaml
kubectl exec -n <namespace> <infected-pod> -- tar czf - /var/log /tmp > forensics-files-$(date +%Y%m%d).tar.gz

# Step 4: Terminate infected workloads
kubectl delete pod <infected-pod> -n <namespace> --grace-period=0 --force

# Step 5: Restore from clean backup
velero restore create malware-recovery-$(date +%Y%m%d) \
  --from-backup clean-backup-<timestamp> \
  --include-namespaces <affected-namespace>
```

### Playbook 3: Data Exfiltration Suspected

**Immediate Actions (< 10 min):**

```bash
# Step 1: Block egress traffic
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: block-egress
  namespace: <namespace>
spec:
  podSelector: {}
  policyTypes:
  - Egress
  egress:
  - to:
    - podSelector: {}
    ports:
    - protocol: TCP
      port: 53  # Allow DNS only
EOF

# Step 2: Review recent data access
kubectl logs -n <namespace> <pod> --since=24h | grep -E "GET | POST | PUT | DELETE"

# Step 3: Check for large data transfers
kubectl top pods -n <namespace> --sort-by=memory
kubectl exec -n <namespace> <pod> -- netstat -an | grep ESTABLISHED

# Step 4: Identify exfiltration destination
kubectl exec -n <namespace> <pod> -- tcpdump -w /tmp/capture.pcap -c 1000
kubectl cp <namespace>/<pod>:/tmp/capture.pcap ./capture-$(date +%Y%m%d).pcap

# Step 5: Revoke compromised credentials
aws iam update-access-key --access-key-id <key-id> --status Inactive --user-name <username>
kubectl delete secret <secret-name> -n <namespace>
```

### Playbook 4: Container Escape / Privilege Escalation

**Immediate Actions (< 5 min):**

```bash
# Step 1: Kill compromised container
kubectl delete pod <pod-name> -n <namespace> --grace-period=0 --force

# Step 2: Check for host compromise
kubectl exec -n kube-system <node-pod> -- ps aux | grep -E "malicious | suspicious"
kubectl exec -n kube-system <node-pod> -- find /proc -name "environ" -exec cat {} \; | grep -i "malicious"

# Step 3: Review container capabilities
kubectl get pod <pod-name> -n <namespace> -o json | \
  jq '.spec.containers[].securityContext.capabilities'

# Step 4: Check for privileged containers
kubectl get pods --all-namespaces -o json | \
  jq -r '.items[] | select(.spec.containers[].securityContext.privileged==true) | "\(.metadata.namespace)/\(.metadata.name)"'

# Step 5: Quarantine affected node
kubectl cordon <node-name>
kubectl taint nodes <node-name> quarantine=true:NoSchedule
```

### Playbook 5: Supply Chain Attack

**Immediate Actions (< 30 min):**

```bash
# Step 1: Identify affected images
kubectl get pods --all-namespaces -o json | \
  jq -r '.items[] | "\(.metadata.namespace)/\(.metadata.name): \(.spec.containers[].image)"' | \
  grep <compromised-image>

# Step 2: Stop pulling compromised images
kubectl patch deployment <deployment> -n <namespace> \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"<container>","imagePullPolicy":"Never"}]}}}}'

# Step 3: Roll back to known good version
kubectl rollout undo deployment/<deployment> -n <namespace>
kubectl rollout status deployment/<deployment> -n <namespace>

# Step 4: Scan all images in cluster
for image in $(kubectl get pods --all-namespaces -o jsonpath='{.items[*].spec.containers[*].image}' | tr ' ' '\n' | sort -u); do
  echo "Scanning: $image"
  trivy image "$image" --severity HIGH,CRITICAL
done

# Step 5: Update image pull secrets and policies
kubectl create secret docker-registry secure-registry \
  --docker-server=<new-registry> \
  --docker-username=<username> \
  --docker-password=<password>
```

### Playbook 6: Suspicious Runtime Behavior

**Triggers:**
- Unusual process execution
- Unexpected network connections
- File system modifications in read-only areas
- Crypto-mining activity detected

**Immediate Actions (< 10 min):**

```bash
# Step 1: Capture process snapshot
kubectl exec -n <namespace> <pod> -- ps auxf > /tmp/process-snapshot.txt

# Step 2: Check network connections
kubectl exec -n <namespace> <pod> -- netstat -tunapl > /tmp/network-snapshot.txt

# Step 3: Isolate the pod
kubectl label pod <pod> -n <namespace> security-isolated=true

# Step 4: Enable verbose logging
kubectl patch deployment <deployment> -n <namespace> \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"<container>","env":[{"name":"LOG_LEVEL","value":"trace"}]}]}}}}'

# Step 5: Analyze with runtime security
kubectl logs -n falco -l app=falco | grep <pod-name>
```

**Containment:**
- Isolate affected pod with network policy
- Prevent horizontal spread
- Capture forensics data

**Investigation:**
- Analyze process tree
- Review file modifications
- Check network traffic patterns
- Compare with baseline behavior

**Resolution:**
- Terminate suspicious processes
- Remove malicious files
- Update security policies
- Deploy patched version

### Playbook 7: Critical CVE in Container Image

**Immediate Actions (< 30 min):**

```bash
# Step 1: Identify affected images
kubectl get pods --all-namespaces -o json | \
  jq -r '.items[] | "\(.metadata.namespace)/\(.metadata.name): \(.spec.containers[].image)"' | \
  grep <compromised-image>

# Step 2: Stop pulling compromised images
kubectl patch deployment <deployment> -n <namespace> \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"<container>","imagePullPolicy":"Never"}]}}}}'

# Step 3: Roll back to known good version
kubectl rollout undo deployment/<deployment> -n <namespace>
kubectl rollout status deployment/<deployment> -n <namespace>

# Step 4: Scan all images in cluster
for image in $(kubectl get pods --all-namespaces -o jsonpath='{.items[*].spec.containers[*].image}' | tr ' ' '\n' | sort -u); do
  echo "Scanning: $image"
  trivy image "$image" --severity HIGH,CRITICAL
done

# Step 5: Update image pull secrets and policies
kubectl create secret docker-registry secure-registry \
  --docker-server=<new-registry> \
  --docker-username=<username> \
  --docker-password=<password>
```

## Forensics Data Collection

### Essential Data to Collect (Processes, Logs, Network Traffic)

```bash
# Create forensics directory
mkdir -p /tmp/forensics/$(date +%Y%m%d-%H%M%S)
cd /tmp/forensics/$(date +%Y%m%d-%H%M%S)

# 1. Pod logs (all containers)
for pod in $(kubectl get pods -n <namespace> -o name); do
  kubectl logs $pod -n <namespace> --all-containers > logs-$(basename $pod).txt
done

# 2. Process information
kubectl exec -n <namespace> <pod> -- ps aux > processes.txt

# 3. Pod descriptions
kubectl get pods -n <namespace> -o yaml > pods.yaml

# 3. Events
kubectl get events -n <namespace> --sort-by='.lastTimestamp' > events.txt

# 4. Network policies
kubectl get networkpolicies -n <namespace> -o yaml > networkpolicies.yaml

# 5. Service accounts and RBAC
kubectl get sa,roles,rolebindings -n <namespace> -o yaml > rbac.yaml

# 6. ConfigMaps and Secrets (sanitized)
kubectl get configmaps -n <namespace> -o yaml > configmaps.yaml
kubectl get secrets -n <namespace> -o yaml | grep -v "data:" > secrets-metadata.yaml

# 7. Persistent storage
kubectl get pvc,pv -n <namespace> -o yaml > storage.yaml

# 8. Network traffic capture
kubectl exec -n <namespace> <pod> -- tcpdump -w /tmp/traffic.pcap -c 10000
kubectl cp <namespace>/<pod>:/tmp/traffic.pcap ./traffic-$(date +%Y%m%d).pcap

# 9. Audit logs
kubectl logs -n kube-system kube-apiserver-* --since=24h > audit-logs.txt

# 10. Node information
kubectl describe nodes > nodes.txt

# Package forensics data
tar czf forensics-$(date +%Y%m%d-%H%M%S).tar.gz .
aws s3 cp forensics-*.tar.gz s3://security-forensics-bucket/$(date +%Y-%m-%d)/
```

## Escalation Paths

### Internal Escalation

**Level 1: On-Call SRE**
- **Contact:** PagerDuty escalation
- **Response Time:** < 15 minutes
- **Scope:** P2-P3 incidents

**Level 2: Security Team Lead**
- **Contact:** security-lead@company.com + Slack #security-incidents
- **Response Time:** < 30 minutes
- **Scope:** P1 incidents, any data breach suspicion

**Level 3: CISO / Executive Team**
- **Contact:** Direct phone + Email
- **Response Time:** < 1 hour
- **Scope:** P0 incidents, confirmed breaches, regulatory implications

### External Escalation

**When to Escalate Externally:**
- Confirmed data breach affecting customer data
- Ransomware with payment demand
- Suspected nation-state actor
- Regulatory reporting requirements (GDPR, HIPAA, etc.)

**External Contacts:**
- **Law Enforcement:** FBI Cyber Division (IC3.gov)
- **Incident Response Partner:** <Partner Company> - +1-XXX-XXX-XXXX
- **Legal Team:** legal@company.com
- **PR/Communications:** pr@company.com

## Post-Incident Review (PIR)

### PIR Template

```markdown
# Post-Incident Review: [Incident Title]

**Date:** YYYY-MM-DD  
**Incident ID:** INC-XXXXXX  
**Severity:** P0/P1/P2/P3  
**Duration:** XX hours  
**Status:** Resolved/Ongoing

## Executive Summary
[Brief description of what happened, impact, and resolution]

## Timeline | Time (UTC) | Event | Action Taken | |------------ | ------- | -------------- | | HH:MM | Initial detection | ... | | HH:MM | Containment started | ... | | HH:MM | Root cause identified | ... | | HH:MM | Incident resolved | ... | ## Root Cause Analysis
**What Happened:**
[Detailed explanation of the incident]

**Why It Happened:**
[Root cause and contributing factors]

**Impact:**
- **Users Affected:** X users
- **Data Compromised:** Yes/No - [details]
- **Services Down:** [list of services]
- **Financial Impact:** $XXX

## Response Evaluation
**What Went Well:**
- [List positive aspects]

**What Went Wrong:**
- [List root causes and failures]

**What Could Be Improved:**
- [List areas for improvement]

## Action Items | Action | Owner | Due Date | Priority | |-------- | ------- | ---------- | ---------- | | [Action 1] | [Name] | YYYY-MM-DD | P0/P1/P2 | | [Action 2] | [Name] | YYYY-MM-DD | P0/P1/P2 | ## Lessons Learned
[Key takeaways and preventive measures]
```

### PIR Schedule

- **Draft PIR:** Within 24 hours of resolution
- **PIR Review Meeting:** Within 3 business days
- **Final PIR Published:** Within 5 business days
- **Follow-up Review:** 30 days after incident

## Security Tool Integration

### Integrated Security Tools

#### 1. **Falco** - Runtime Security Monitoring

```bash
# Install Falco
helm repo add falcosecurity https://falcosecurity.github.io/charts
helm install falco falcosecurity/falco \
  --namespace falco \
  --set ebpf.enabled=true

# View Falco alerts
kubectl logs -n falco -l app=falco -f | grep -i "warning\ | error"

# Custom Falco rules
kubectl edit configmap falco-rules -n falco
```

#### 2. **Trivy** - Vulnerability Scanning

```bash
# Scan all images in cluster
kubectl get pods --all-namespaces -o jsonpath='{.items[*].spec.containers[*].image}' | \
  tr ' ' '\n' | sort -u | while read image; do
    echo "Scanning: $image"
    trivy image --severity HIGH,CRITICAL "$image"
  done

# Automated scanning (CronJob)
kubectl apply -f trivy-scanner-cronjob.yaml
```

#### 3. **OPA/Gatekeeper** - Policy Enforcement

```bash
# Install Gatekeeper
kubectl apply -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/master/deploy/gatekeeper.yaml

# View policy violations
kubectl get constrainttemplates
kubectl get <constraint-kind> --all-namespaces
```

#### 4. **Prometheus/Grafana** - Security Metrics

```prometheus
# Security-related metrics queries

# Failed authentication attempts
rate(authentication_attempts{result="failure"}[5m])

# Privileged container count
count(kube_pod_container_status_running{container_security_context_privileged="true"})

# Network policy violations
rate(network_policy_violations_total[5m])

# High-risk API calls
rate(apiserver_audit_event_total{verb=~"create | update | delete",user!="system:*"}[5m])
```

#### 5. **SIEM Integration** - Log Aggregation

```bash
# Forward logs to SIEM
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: logging
data:
  output.conf: | [OUTPUT]
        Name   syslog
        Match  *
        Host   siem.company.com
        Port   514
        Mode   tcp
        Syslog_Format rfc5424
EOF
```

## Compliance and Reporting

### Regulatory Requirements

- **GDPR:** Report data breaches to supervisory authority within 72 hours
- **HIPAA:** Report breaches affecting 500+ individuals to HHS within 60 days
- **PCI-DSS:** Report compromised card data to card brands immediately
- **SOC 2:** Document and report security incidents to customers within 24 hours

### Incident Reporting Template

```markdown
# Security Incident Notification

**To:** [Customer/Regulator]  
**From:** Security Team  
**Date:** YYYY-MM-DD  
**Subject:** Security Incident Notification - [Brief Description]

Dear [Recipient],

We are writing to inform you of a security incident that occurred on [DATE] affecting [SCOPE].

**Incident Summary:**
[Brief description]

**Impact:**
- Data affected: [Yes/No - details]
- Services affected: [list]
- User accounts affected: [number]

**Actions Taken:**
[List of response actions]

**Next Steps:**
[Recommendations for affected parties]

**Contact:**
For questions, contact security@company.com or call +1-XXX-XXX-XXXX

Sincerely,  
[Name]  
Chief Information Security Officer
```

## Training and Drills

### Quarterly Security Drill Schedule

**Q1:** Unauthorized Access Simulation  
**Q2:** Ransomware Response Exercise  
**Q3:** Data Exfiltration Detection Drill  
**Q4:** Full Incident Response Tabletop Exercise

### Drill Procedure

```bash
# Schedule drill
# Notify: Security Team + On-call SREs (NOT entire organization)

# Execute drill
# Follow incident playbook as if real incident

# Debrief within 24 hours
# Document lessons learned and update playbooks
```

## Related Documentation

- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [SRE Runbook Index](../README.md)

## Revision History | Version | Date | Author | Changes | |--------- | ------ | -------- | --------- | | 1.0.0 | 2024-01-15 | Security Team | Initial runbook creation |
