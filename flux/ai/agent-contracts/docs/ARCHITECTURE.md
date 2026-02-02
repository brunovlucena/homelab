# Agent-Contracts Architecture

## System Context Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                  HOMELAB CLUSTER                                     │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│   ┌──────────────────────────────────────────────────────────────────────────────┐  │
│   │                         KNATIVE SERVING                                       │  │
│   │                                                                               │  │
│   │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │  │
│   │  │  contract   │  │   vuln      │  │  exploit    │  │   notifi    │         │  │
│   │  │  fetcher    │  │  scanner    │  │  generator  │  │   adapter   │         │  │
│   │  │  (0→5)      │  │  (0→10)     │  │  (0→3)      │  │  (1→3)      │         │  │
│   │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘         │  │
│   │         │                │                │                ▲                 │  │
│   └─────────┼────────────────┼────────────────┼────────────────┼─────────────────┘  │
│             │                │                │                │                    │
│             ▼                ▼                ▼                │                    │
│   ┌──────────────────────────────────────────────────────────────────────────────┐  │
│   │                          RABBITMQ CLUSTER                                     │  │
│   │                                                                               │  │
│   │   Exchange: agent-contracts                                                   │  │
│   │   Event Types:                                                                │  │
│   │     • io.homelab.contract.created   (from chain-monitor)                     │  │
│   │     • io.homelab.vuln.found         (from scanner → exploit-gen)             │  │
│   │     • io.homelab.exploit.validated  (from exploit-gen → audit log)           │  │
│   │                                                                               │  │
│   └──────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                      │
│   ┌──────────────────────────────────────────────────────────────────────────────┐  │
│   │                           OBSERVABILITY STACK                                 │  │
│   │                                                                               │  │
│   │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │  │
│   │  │ Prometheus  │──│Alertmanager │  │   Tempo     │  │  Grafana    │         │  │
│   │  │  (metrics)  │  │  (routing)  │  │  (traces)   │  │ (dashboards)│         │  │
│   │  └──────┬──────┘  └──────┬──────┘  └─────────────┘  └─────────────┘         │  │
│   │         │                │                                                    │  │
│   │         │      ┌─────────┘                                                    │  │
│   │         │      │  webhook                                                     │  │
│   │         │      ▼                                                              │  │
│   │         │  ┌─────────────┐                                                    │  │
│   │         │  │   notifi    │◄─────────────────────────────────────────────┐    │  │
│   │         │  │   adapter   │                                              │    │  │
│   │         │  └──────┬──────┘                                              │    │  │
│   │         │         │                                                     │    │  │
│   └─────────┼─────────┼─────────────────────────────────────────────────────┼────┘  │
│             │         │                                                     │       │
│             │         ▼                                                     │       │
│   ┌─────────┼──────────────────────────────────────────────────────────────────┐   │
│   │         │       SUPPORTING SERVICES                                        │   │
│   │         │                                                                   │   │
│   │  ┌──────▼──────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐       │   │
│   │  │   Ollama    │  │   MinIO     │  │   Redis     │  │   Anvil     │       │   │
│   │  │   (LLM)     │  │   (S3)      │  │   (Cache)   │  │   (Fork)    │       │   │
│   │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘       │   │
│   │                                                                            │   │
│   └────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              EXTERNAL SERVICES                                       │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐   │
│  │  Etherscan  │  │  Alchemy    │  │  Anthropic  │  │    notifi-services      │   │
│  │  BSCScan    │  │  Infura     │  │  (fallback) │  │  (Telegram/SMS/Email/   │   │
│  │  (source)   │  │  (RPC)      │  │             │  │   Discord/Webhook)      │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────────┘   │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## Alerting Architecture

The alerting system uses **Prometheus + Alertmanager + notifi-services** instead of a custom alert-dispatcher:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         ALERTING FLOW                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. METRICS EMISSION                                                         │
│     ┌─────────────┐                                                          │
│     │vuln-scanner │──► agent_contracts_critical_vulns_total{chain, type}    │
│     │             │──► agent_contracts_high_vulns_total{chain, type}        │
│     │             │──► agent_contracts_vulnerabilities_total{severity}      │
│     └─────────────┘                                                          │
│     ┌─────────────┐                                                          │
│     │exploit-gen  │──► agent_contracts_exploits_validated_total{profit}     │
│     └─────────────┘                                                          │
│                                                                              │
│  2. PROMETHEUS SCRAPING                                                      │
│     ┌─────────────┐     ┌──────────────────────────────────────┐            │
│     │ Prometheus  │◄────│ ServiceMonitor (agent-contracts ns)  │            │
│     │             │     └──────────────────────────────────────┘            │
│     └──────┬──────┘                                                          │
│            │                                                                 │
│  3. ALERT EVALUATION (PrometheusRule)                                        │
│            │                                                                 │
│            │  increase(agent_contracts_critical_vulns_total[5m]) > 0        │
│            │  increase(agent_contracts_exploits_validated_total[5m]) > 0    │
│            │                                                                 │
│            ▼                                                                 │
│  4. ALERTMANAGER ROUTING (AlertmanagerConfig)                                │
│     ┌─────────────┐                                                          │
│     │Alertmanager │                                                          │
│     │             │  severity=critical → notifi-critical (all channels)     │
│     │             │  severity=high     → notifi-high (telegram, discord)    │
│     │             │  severity=warning  → notifi-warning (discord only)      │
│     │             │  severity=info     → grafana-only (annotations)         │
│     └──────┬──────┘                                                          │
│            │ webhook                                                         │
│            ▼                                                                 │
│  5. NOTIFI ADAPTER                                                           │
│     ┌─────────────┐                                                          │
│     │notifi-adapter│  Transforms Alertmanager payload → notifi format       │
│     │             │  Routes to channels based on severity                   │
│     └──────┬──────┘                                                          │
│            │ HTTP POST                                                       │
│            ▼                                                                 │
│  6. NOTIFI-SERVICES                                                          │
│     ┌─────────────┐                                                          │
│     │  notifi-    │──► Telegram Bot                                         │
│     │  services   │──► Discord Webhook                                      │
│     │             │──► SMS (Twilio)                                         │
│     │             │──► Email (SES)                                          │
│     │             │──► Webhooks                                             │
│     └─────────────┘                                                          │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Severity Routing

| Severity | Channels | Group Wait | Repeat Interval |
|----------|----------|------------|-----------------|
| CRITICAL | Telegram, Discord, Email, Webhook | 0s | 1h |
| HIGH | Telegram, Discord, Webhook | 30s | 2h |
| WARNING | Discord, Webhook | 1m | 4h |
| INFO | Grafana Annotations | 5m | 24h |

## Function Details

### 1. Contract Fetcher (`contract-fetcher`)

**Purpose**: Fetch smart contract source code and metadata from block explorers or RPC.

**Triggers**:
- CloudEvent `io.homelab.scan.request` (manual request)
- CloudEvent `io.homelab.block.new` (from chain monitor)

**Flow**:
```
Input: {chain, address} OR {chain, block_number}
  │
  ▼
┌─────────────────────────────────────────┐
│ 1. Check Redis cache for contract       │
│ 2. If miss: fetch from Etherscan API    │
│ 3. If not verified: fetch bytecode      │
│ 4. Store in MinIO (S3)                  │
│ 5. Emit contract.created CloudEvent     │
└─────────────────────────────────────────┘
  │
  ▼
Output: CloudEvent(io.homelab.contract.created)
```

**Dependencies**:
- Redis (caching)
- MinIO (storage)
- Etherscan API

---

### 2. Vulnerability Scanner (`vuln-scanner`)

**Purpose**: Analyze contracts using static analysis tools + LLM reasoning.

**Triggers**:
- CloudEvent `io.homelab.contract.created`

**Flow**:
```
Input: CloudEvent(contract.created) with source code
  │
  ▼
┌─────────────────────────────────────────┐
│ 1. Run Slither (fast, low FP)           │
│ 2. Run Mythril (deep, symbolic exec)    │
│ 3. Aggregate findings                   │
│ 4. If confidence < threshold:           │
│    └─ Send to LLM for reasoning         │
│ 5. Score vulnerability severity         │
│ 6. Emit Prometheus metrics (alertable)  │
│ 7. Emit vuln.found CloudEvent           │
└─────────────────────────────────────────┘
  │
  ▼
Output: CloudEvent(io.homelab.vuln.found)
        + Prometheus metrics for Alertmanager
```

**Alertable Metrics Emitted**:
```python
# Critical vulns → immediate alert
agent_contracts_critical_vulns_total{chain, vuln_type, contract_address}

# High vulns → high-priority alert
agent_contracts_high_vulns_total{chain, vuln_type, contract_address}

# All vulns for tracking
agent_contracts_vulnerabilities_total{chain, severity, vuln_type}
```

---

### 3. Exploit Generator (`exploit-generator`)

**Purpose**: Use LLM to generate exploit PoC for validation.

**Triggers**:
- CloudEvent `io.homelab.vuln.found` (severity >= HIGH)

**Flow**:
```
Input: CloudEvent(vuln.found) with vulnerability details
  │
  ▼
┌─────────────────────────────────────────┐
│ 1. Build prompt with:                   │
│    - Contract source                    │
│    - Vulnerability type                 │
│    - Attack vector description          │
│ 2. Query Ollama (DeepSeek-Coder)        │
│ 3. If confidence < 0.8:                 │
│    └─ Fallback to Claude API            │
│ 4. Parse exploit script from response   │
│ 5. Validate on Anvil fork               │
│ 6. Emit metrics (triggers CRITICAL)     │
│ 7. Emit exploit.validated CloudEvent    │
└─────────────────────────────────────────┘
  │
  ▼
Output: CloudEvent(io.homelab.exploit.validated)
        + agent_contracts_exploits_validated_total metric
```

**Alertable Metrics Emitted**:
```python
# Validated exploit → CRITICAL alert
agent_contracts_exploits_validated_total{chain, vuln_type, profit_potential, contract_address}
```

---

### 4. Notifi Adapter (`notifi-adapter`)

**Purpose**: Bridge between Alertmanager and notifi-services for multi-channel delivery.

**Triggers**:
- HTTP webhook from Alertmanager

**Flow**:
```
Input: Alertmanager webhook payload
  │
  ▼
┌─────────────────────────────────────────┐
│ 1. Parse Alertmanager payload           │
│ 2. Transform to notifi-services format  │
│ 3. Route to channels based on severity  │
│ 4. POST to notifi-services webhook      │
│ 5. Log delivery status                  │
└─────────────────────────────────────────┘
  │
  ▼
Output: Multi-channel notifications via notifi-services
```

**Why not a custom alert-dispatcher?**

| Custom Dispatcher | Alertmanager + notifi-services |
|-------------------|--------------------------------|
| ❌ Duplicates notification logic | ✅ Reuses notifi-services |
| ❌ No grouping/silencing | ✅ Full Alertmanager features |
| ❌ Manual channel routing | ✅ Declarative routing rules |
| ❌ No repeat intervals | ✅ Configurable repeat/inhibit |
| ❌ Maintenance overhead | ✅ Standard observability stack |

---

## Deployment Model

### Knative Service Configuration

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: vuln-scanner
  namespace: agent-contracts
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/min-scale: "0"
        autoscaling.knative.dev/max-scale: "10"
        autoscaling.knative.dev/target: "5"
    spec:
      containerConcurrency: 1
      timeoutSeconds: 300
      containers:
        - image: ${ECR_REGISTRY}/agent-contracts/vuln-scanner:latest
          resources:
            requests:
              memory: "512Mi"
              cpu: "500m"
```

### Alertmanager Integration

```yaml
apiVersion: monitoring.coreos.com/v1alpha1
kind: AlertmanagerConfig
metadata:
  name: agent-contracts-routing
spec:
  route:
    receiver: 'notifi-critical'
    routes:
      - matchers: [{name: severity, value: critical}]
        receiver: 'notifi-critical'
        groupWait: 0s
  receivers:
    - name: 'notifi-critical'
      webhookConfigs:
        - url: 'http://notifi-adapter.agent-contracts/webhook/alertmanager'
```

---

## Monitoring & Observability

### Prometheus Metrics

```python
# Counters
agent_contracts_scanned_total{chain, status}
agent_contracts_vulnerabilities_total{chain, severity, vuln_type}
agent_contracts_critical_vulns_total{chain, vuln_type, contract_address}  # Alertable
agent_contracts_high_vulns_total{chain, vuln_type, contract_address}     # Alertable
agent_contracts_exploits_validated_total{chain, vuln_type, profit_potential}  # Alertable

# Histograms
agent_contracts_scan_duration_seconds{chain, analyzer}
agent_contracts_llm_inference_seconds{model, operation}
agent_contracts_exploit_validation_seconds{chain}

# Gauges
agent_contracts_active_scans{chain}
agent_contracts_pending_queue{chain}
agent_contracts_llm_queue_depth
```

### Grafana Dashboard Panels

1. **Overview**
   - Contracts scanned (24h)
   - Vulnerabilities found (by severity)
   - Alert delivery status

2. **Performance**
   - Scan latency (p50, p95, p99)
   - LLM inference time
   - Cold start frequency

3. **Cost**
   - API calls (Etherscan, RPC)
   - LLM tokens used
   - Estimated cost per contract

4. **Security**
   - Validated exploits
   - False positive rate
   - Time-to-detection

---

## Flux CD Integration (CDEvents)

Agent-contracts CloudEvents can trigger Flux CD GitOps reconciliation using the Notification Controller's CDEvents Receiver. This enables **automated security responses** via GitOps.

### Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    AGENT-CONTRACTS → FLUX CD INTEGRATION                         │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │                    AGENT-CONTRACTS PIPELINE                                 │  │
│  │                                                                             │  │
│  │   vuln-scanner                      exploit-generator                       │  │
│  │   ├─ io.homelab.vuln.found          ├─ io.homelab.exploit.validated        │  │
│  │   └─ Prometheus metrics             └─ CRITICAL alert trigger              │  │
│  └────────────────────┬──────────────────────────┬─────────────────────────────┘  │
│                       │                          │                                │
│                       │ CloudEvents              │                                │
│                       ▼                          ▼                                │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │                    FLUX NOTIFICATION CONTROLLER                             │  │
│  │                                                                             │  │
│  │  Receiver: security-alert-receiver                                          │  │
│  │  ├─ type: cdevents                                                          │  │
│  │  ├─ events:                                                                 │  │
│  │  │   - io.homelab.exploit.validated                                        │  │
│  │  │   - io.homelab.vuln.found                                               │  │
│  │  ├─ resourceFilter: severity == "critical"                                  │  │
│  │  └─ Triggers Kustomization reconciliation                                   │  │
│  └────────────────────┬───────────────────────────────────────────────────────┘  │
│                       │                                                          │
│                       │ Reconcile                                                │
│                       ▼                                                          │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │                         GITOPS SECURITY RESPONSE                            │  │
│  │                                                                             │  │
│  │  Kustomization: security-policies                                           │  │
│  │  ├─ NetworkPolicy (quarantine compromised contracts)                        │  │
│  │  ├─ WAF rules (block exploit patterns)                                      │  │
│  │  └─ Security patches (from Git repository)                                  │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### Flux Receiver Configuration

```yaml
apiVersion: notification.toolkit.fluxcd.io/v1
kind: Receiver
metadata:
  name: security-alert-receiver
  namespace: flux-system
spec:
  # CDEvents receiver validates CloudEvents payloads
  type: cdevents
  
  # Filter agent-contracts security events
  events:
    - "io.homelab.exploit.validated"
    - "io.homelab.vuln.found"
  
  # Authentication
  secretRef:
    name: agent-contracts-webhook-token
  
  # Flux resources to reconcile on security events
  resources:
    - kind: Kustomization
      name: security-policies
      namespace: flux-system
    - kind: Kustomization
      name: network-policies
      namespace: flux-system
  
  # CEL filter: only trigger on critical severity
  resourceFilter: |
    request.body.data.severity == "critical" || 
    request.body.type == "io.homelab.exploit.validated"
```

### Use Cases

| Event | Flux Action | Automated Response |
|-------|-------------|-------------------|
| `io.homelab.exploit.validated` | Deploy security patches | NetworkPolicy quarantine, WAF rules |
| `io.homelab.vuln.found` (critical) | Update configs | Alert escalation, monitoring rules |
| `io.homelab.contract.created` | Reconcile scan configs | Update scan schedules |

### Publishing Events to Flux

Add a Knative Trigger to forward security events to Flux:

```yaml
apiVersion: eventing.knative.dev/v1
kind: Trigger
metadata:
  name: security-to-flux
  namespace: agent-contracts
spec:
  broker: agent-contracts-broker
  filter:
    attributes:
      type: io.homelab.exploit.validated
  subscriber:
    # Flux Receiver webhook URL
    uri: http://notification-controller.flux-system.svc.cluster.local/hook/<webhook-path>
```

### Webhook Secret Setup

```bash
# Generate secure token
TOKEN=$(head -c 12 /dev/urandom | shasum | head -c 32)

# Create secret
kubectl create secret generic agent-contracts-webhook-token \
  --from-literal=token=$TOKEN \
  -n flux-system

# Get webhook path after Receiver creation
kubectl get receiver security-alert-receiver -n flux-system \
  -o jsonpath='{.status.webhookPath}'
```

### Security Response Flow

1. **Detection**: `vuln-scanner` detects critical vulnerability
2. **Validation**: `exploit-generator` validates exploit on Anvil fork
3. **CloudEvent**: `io.homelab.exploit.validated` emitted
4. **Flux Receiver**: Receives event, validates, triggers reconciliation
5. **GitOps**: `security-policies` Kustomization applies remediation
6. **Response**: NetworkPolicies quarantine affected resources

### References

- [Flux Notification API v1](https://fluxcd.io/flux/components/notification/api/v1/)
- [Flux Receivers Documentation](https://fluxcd.io/flux/components/notification/receivers/)
- [CDEvents Specification](https://cdevents.dev/)
- [Knative Lambda CloudEvents Specification](../../infrastructure/knative-lambda-operator/docs/04-architecture/CLOUDEVENTS_SPECIFICATION.md)
