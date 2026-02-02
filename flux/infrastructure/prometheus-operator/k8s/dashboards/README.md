# Grafana Dashboards for Prometheus Operator

This directory contains Grafana dashboards that are automatically deployed via ConfigMaps.

## Dashboard Discovery

Dashboards are automatically discovered by Grafana via the sidecar container. The sidecar watches for ConfigMaps with the label `grafana_dashboard: "1"` in the `prometheus` namespace and automatically imports them into Grafana.

## Available Dashboards

### Knative Lambda Operator Dashboard

**File**: `knative-lambda-operator-dashboard.json`  
**ConfigMap**: `grafana-dashboard-knative-lambda-operator`  
**UID**: `knative-lambda-operator`

#### Metrics Covered

All `knative_lambda_operator_*` metrics are covered:

1. **Reconciliation Metrics**
   - `knative_lambda_operator_reconcile_total` - Total reconciliations by phase and result
   - `knative_lambda_operator_reconcile_duration_seconds` - Reconcile duration (P95/P99)

2. **Lambda Functions**
   - `knative_lambda_operator_lambdafunctions_total` - Lambda count by phase and namespace

3. **Build Metrics**
   - `knative_lambda_operator_build_jobs_active` - Active build jobs
   - `knative_lambda_operator_build_duration_seconds` - Build duration (P95/P99)

4. **Work Queue**
   - `knative_lambda_operator_workqueue_depth` - Current queue depth
   - `knative_lambda_operator_workqueue_latency_seconds` - Queue latency (P95/P99)

5. **Errors**
   - `knative_lambda_operator_errors_total` - Errors by component and type

6. **API Server**
   - `knative_lambda_operator_apiserver_requests_total` - API server request rate

7. **Eventing Resources**
   - `knative_lambda_operator_eventing_resources_total` - Eventing resources by namespace and type

#### Dashboard Sections

- **Overview**: Key metrics at a glance (reconciliations, lambdas, errors, queue depth)
- **Reconciliation Metrics**: Rate and duration of reconciliations
- **Lambda Functions**: Count by phase and namespace
- **Build Metrics**: Active builds and build duration
- **Work Queue & Errors**: Queue depth, latency, and error rates
- **API Server & Eventing Resources**: API server requests and eventing resources

### Knative Lambda Metrics Dashboard

**File**: `knative-lambda-metrics-dashboard.json`  
**ConfigMap**: `grafana-dashboard-knative-lambda-metrics`  
**UID**: `knative-lambda-metrics`

#### Metrics Covered

This dashboard monitors the actual running Knative Lambda function pods and services:

1. **CPU Metrics**
   - CPU usage (5m rate) by pod
   - CPU usage by service (aggregated)
   - Total CPU usage across all lambda services

2. **Memory Metrics**
   - Memory usage (MB) by pod
   - Memory usage by service (aggregated)
   - Total memory usage across all lambda services

3. **Pod Metrics**
   - Active pod count
   - Pod count by service
   - Pod status by service (table view)

#### Dashboard Sections

- **Resource Usage**: CPU and memory usage by pod and service
- **Summary Stats**: Active pods, total CPU, and total memory usage
- **Service Overview**: Aggregated metrics and pod counts by service
- **Pod Status Table**: Detailed pod status breakdown by service

#### Variables

- **Lambda Service**: Multi-select dropdown to filter by specific lambda services (supports regex matching for `.*lambda.*|.*parser.*|.*command.*`)
- **Namespace**: Fixed to `knative-lambda` namespace

## Deployment

Dashboards are deployed automatically when:
1. The ConfigMap is created/updated in the `prometheus` namespace
2. The ConfigMap has the label `grafana_dashboard: "1"`
3. The Grafana sidecar is running and watching for ConfigMaps

The sidecar configuration is in `helmrelease.yaml`:
```yaml
sidecar:
  dashboards:
    enabled: true
    label: grafana_dashboard
    labelValue: "1"
```

## Adding New Dashboards

1. Create a JSON dashboard file in this directory
2. Generate the ConfigMap YAML:
   ```bash
   python3 << 'EOF'
   import json
   import yaml
   
   with open('k8s/dashboards/your-dashboard.json', 'r') as f:
       dashboard_json = f.read()
   
   configmap = {
       'apiVersion': 'v1',
       'kind': 'ConfigMap',
       'metadata': {
           'name': 'grafana-dashboard-your-name',
           'namespace': 'prometheus',
           'labels': {'grafana_dashboard': '1'}
       },
       'data': {'your-dashboard.json': dashboard_json}
   }
   
   with open('k8s/dashboards/your-dashboard-configmap.yaml', 'w') as f:
       f.write('---\n')
       yaml.dump(configmap, f, default_flow_style=False, sort_keys=False)
   EOF
   ```
3. Add the ConfigMap to `kustomization.yaml`
4. Commit and push - Flux will deploy it automatically

### Knative Lambda Eventing & DLQ Dashboard

**File**: `knative-lambda-eventing-dashboard.json`  
**ConfigMap**: `grafana-dashboard-knative-lambda-eventing`  
**UID**: `knative-lambda-eventing`

#### Purpose

Dedicated dashboard for monitoring CloudEvents processing, RabbitMQ queues, and Dead Letter Queues (DLQ).

#### Sections

1. **🎯 Eventing Overview**
   - Total Brokers, Triggers, Queues count
   - Messages in flight
   - Event publish/consume rates

2. **⚠️ Dead Letter Queue (DLQ)**
   - DLQ Messages gauge with threshold
   - DLQ Rate (messages/sec entering DLQ)
   - DLQ Queues count
   - DLQ Messages (1h) - recent additions
   - Oldest DLQ Message age
   - DLQ Messages by Queue timeline
   - DLQ Queue Details table
   - DLQ Ingest Rate timeline

3. **📬 Message Queues**
   - Queue Depth by Queue
   - Message Publish Rate by Queue
   - Message Delivery Rate by Queue
   - Unacknowledged Messages by Queue

4. **🔄 CloudEvents Processing**
   - Events Processed count
   - Event Success Rate
   - Event P95 Latency
   - Retry Rate
   - Max Retries Exceeded count
   - Events in Flight
   - CloudEvents by Type
   - Event Processing Latency (P50/P95/P99)

5. **🏗️ Eventing Resources**
   - Eventing Resources by Type
   - Eventing Resources by Namespace

6. **🐰 RabbitMQ Health** (collapsed)
   - RabbitMQ Nodes, Connections, Channels
   - Memory Used
   - Message Rate (Received/Acknowledged/Delivered)
   - Memory & Disk usage

#### Variables

- **Namespace**: Filter by lambda namespace
- **Queue**: Multi-select filter for specific queues
- **DLQ Threshold**: Hidden constant (default: 100)

---

### Knative Lambda SRE Dashboard

**File**: `knative-lambda-sre-dashboard.json`  
**ConfigMap**: `grafana-dashboard-knative-lambda-sre`  
**UID**: `knative-lambda-sre`

#### Purpose

A unified SRE-focused dashboard covering both `LambdaFunctions` and `LambdaAgents` with:
- SLO tracking (error rate, latency, cold start rate)
- Error budget burn rate monitoring
- Availability calculations

#### Sections

1. **🎯 SLO Overview**
   - Error Rate SLO gauge (vs 1% target)
   - P95 Latency SLO gauge (vs 1s target)
   - Cold Start Rate SLO gauge (vs 5% target)
   - 30-day Availability percentage
   - Total Functions and Agents count

2. **📊 SLO Burn Rate & Error Budget**
   - Multi-window burn rate (5m, 30m, 1h, 6h)
   - Error rate vs SLO target timeline

3. **⚡ LambdaFunctions - RED Metrics**
   - Ready/Failed function counts
   - Request rate, Error rate, P95 latency
   - Cold starts tracking
   - Invocation rate by function
   - Function duration P50/P95/P99

4. **🤖 LambdaAgents - AI Metrics**
   - Ready/Failed agent counts
   - Agent request rate, Token usage
   - AI errors and P95 latency
   - Invocations by provider
   - Token usage by agent

5. **🔧 Operator Health**
   - Reconcile success rate and P95
   - Work queue depth
   - Build success rate
   - Operator errors by type

6. **📬 Eventing & DLQ**
   - Eventing resources count
   - DLQ message monitoring
   - Event rate tracking

7. **🧪 k6 SRE Test Metrics** (collapsed)
   - SLO test error rate
   - SLO test P95 latency
   - SLO compliance status
   - k6 VUs and iterations

#### Variables

- **Namespace**: Filter by lambda namespace
- **Function**: Multi-select filter for specific functions
- **SLO Targets**: Hidden constants for error rate (1%), latency (1s), cold start (5%)

---

### Agent Bruno - AI Chatbot Dashboard

**File**: `agent-bruno-dashboard.json`  
**ConfigMap**: `grafana-dashboard-agent-bruno`  
**UID**: `agent-bruno`

#### Purpose

Monitor the Agent Bruno AI chatbot for the homepage. Tracks LLM performance, message processing, and token usage.

#### Metrics Covered

1. **Message Processing**
   - `agent_bruno_messages_total` - Total messages by status (success/error)
   - `agent_bruno_active_conversations` - Active conversation count

2. **LLM Performance**
   - `agent_bruno_response_duration_seconds` - Response time distribution
   - `agent_bruno_llm_inference_seconds` - LLM inference time by model

3. **Token Usage & Costs**
   - `agent_bruno_tokens_total` - Token usage by model and type (input/output)
   - `agent_bruno_api_calls_total` - API calls to LLM services

4. **Build Info**
   - `agent_bruno_build_info` - Version and commit information

#### Sections

- **🤖 Overview**: Messages processed, success rate, active conversations, P95 response time
- **💬 Message Processing**: Message rate by status, active conversations timeline
- **⚡ LLM Performance**: Response duration P50/P95/P99, inference time by model
- **🪙 Token Usage & Costs**: Token usage by type, API calls by service
- **🔧 Build Info**: Build version and commit information

#### Variables

- **Model**: Filter by LLM model (ollama, etc.)

---

### Agent Contracts - Smart Contract Security Dashboard

**File**: `agent-contracts-dashboard.json`  
**ConfigMap**: `grafana-dashboard-agent-contracts`  
**UID**: `agent-contracts`

#### Purpose

Monitor the Agent Contracts security pipeline for smart contract vulnerability detection and exploit generation.

#### Metrics Covered

1. **Contract Processing**
   - `agent_contracts_scanned_total` - Contracts scanned by chain
   - `agent_contracts_fetched_total` - Contracts fetched from explorers

2. **Vulnerability Detection**
   - `agent_contracts_vulnerabilities_total` - Vulnerabilities by chain, severity, type
   - `agent_contracts_critical_vulns_total` - Critical vulnerabilities (triggers alerts)
   - `agent_contracts_high_vulns_total` - High severity vulnerabilities

3. **Exploit Generation**
   - `agent_contracts_exploits_generated_total` - Exploits generated by LLM
   - `agent_contracts_exploits_validated_total` - Exploits validated on Anvil fork
   - `agent_contracts_exploits_failed_validation_total` - Failed validations

4. **Performance**
   - `agent_contracts_scan_duration_seconds` - Scan duration by analyzer
   - `agent_contracts_llm_inference_seconds` - LLM inference time
   - `agent_contracts_exploit_validation_seconds` - Exploit validation time

5. **Resources**
   - `agent_contracts_active_scans` - Running scans
   - `agent_contracts_pending_queue` - Contracts waiting
   - `agent_contracts_llm_queue_depth` - LLM queue depth

6. **Costs**
   - `agent_contracts_api_calls_total` - External API calls
   - `agent_contracts_llm_tokens_total` - LLM token consumption

#### Sections

- **🔒 Security Pipeline Overview**: Scanned contracts, vulnerabilities, critical alerts, validated exploits
- **🔍 Vulnerability Detection**: Pie chart by severity, vulnerabilities over time
- **💥 Exploit Generation & Validation**: Generated vs validated exploits, validation rate
- **⚡ Performance Metrics**: Scan duration by analyzer, validation duration
- **📊 Contract Fetching & API Usage**: Fetch rate by source, API calls by service
- **🔧 Build Info**: Build version and commit information

#### Variables

- **Chain**: Filter by blockchain (ethereum, polygon, arbitrum)
- **Severity**: Filter by vulnerability severity (critical, high, medium, low)

---

---

### 🎮 Agent Battle Arena - Game Dashboard

**File**: `agent-battle-arena-dashboard.json`  
**ConfigMap**: `grafana-dashboard-agent-battle-arena`  
**UID**: `agent-battle-arena`

#### Purpose

An epic game-style dashboard that visualizes the battle between Red Team (attackers) and Blue Team (defenders)! Uses game design principles to make observability fun and engaging.

#### Game Design Principles Applied

1. **🎯 Clear Goals & Progress** - Boss health bar, wave counter, scores
2. **⚡ Immediate Feedback** - Real-time attack/defense counters with visual gauges
3. **🏆 Social Competition** - Leaderboard comparing Red vs Blue team effectiveness
4. **📖 Narrative** - Epic battle against the MAG7 dragon boss
5. **✨ Visual Appeal** - Colorful gauges, bars, and exciting visualizations

#### Metrics Translation (Game Language!)

| Prometheus Metric | Game Element |
|-------------------|--------------|
| `agent_redteam_exploits_executed_total` | ⚡ Laser Attacks |
| `agent_redteam_attacks_by_severity_total` | 🔥 Weapon Power Levels |
| `blueteam_threats_blocked_total` | 🛡️ Shield Blocks |
| `blueteam_defense_activations_total` | ⚔️ Counter Attacks |
| `blueteam_mag7_health` | 🐉 Boss HP |
| `blueteam_game_score` | 🎯 Battle Score |
| `blueteam_game_wave` | 🌊 Wave Number |
| `*_exploit_duration_seconds` | ⏱️ Weapon Charge Time |
| `*_cloudevents_received_total` | 📡 Battle Communications |

#### Dashboard Sections

1. **🏟️ BATTLE ARENA**
   - MAG7 Boss Health gauge (color-coded by HP level)
   - Red Team Attack Power gauge
   - Blue Team Shield Power gauge
   - Battle Score, Current Wave, Active Weapons

2. **⚡ LIVE BATTLE - Attack vs Defense Streams**
   - Red Team Laser Attacks (stacked by severity/category)
   - Blue Team Shield Blocks (stacked by defense type)

3. **🏆 LEADERBOARD & BATTLE STATS**
   - Red Team Attack Arsenal (bar gauge)
   - Blue Team Defense Arsenal (bar gauge)
   - Battle Outcome pie chart (attacks landed vs blocked)
   - Key battle statistics

4. **💥 WEAPON POWER - Attack Effectiveness**
   - Laser Intensity by severity (critical/high/medium/low)
   - Target distribution showing which systems are attacked

5. **⏱️ WEAPON CHARGE TIME**
   - Attack charge time P50/P95/P99
   - Shield activation time P50/P95/P99

6. **🐉 MAG7 DRAGON BOSS BATTLE**
   - Boss health over time (with threshold zones)
   - Damage dealt to MAG7 by attack type

7. **📡 CLOUDEVENTS - Battle Communications**
   - Red Team communications (attack orders)
   - Blue Team communications (defense orders)

8. **🔧 KNATIVE LAMBDA - Supporting Infrastructure** (collapsed)
   - Lambda function invocations
   - Operator reconcile rate

9. **📜 AGENT CONTRACTS - Vulnerability Scanner** (collapsed)
   - Vulnerabilities found
   - Exploits generated
   - Contract scanning activity

#### Variables

- **Battle Speed**: Interval for rate calculations (5s, 10s, 30s, 1m, 5m)

#### Annotations

- 🔥 **Critical Attacks**: Markers when critical severity exploits are executed
- 🛡️ **Defense Activated**: Markers when Blue Team blocks threats

#### Related Resources

- [MAG7 Battle Demo](../../../ai/demo-mag7-battle/) - HTML5 game version
- [Agent RedTeam](../../../ai/agent-redteam/) - Offensive security agent
- [Agent BlueTeam](../../../ai/agent-blueteam/) - Defensive security agent
- [Agent Contracts](../../../ai/agent-contracts/) - Smart contract scanner

---

## Dashboard Organization

Dashboards are organized into folders in Grafana for better navigation:

### Folders

- **Agents** (`af9kgc8lh4g74d`) - Contains all agent dashboards
- **Homepage** (`bf9kgc8ljmcjkf`) - Contains homepage-related dashboards

### Organizing Dashboards

Since dashboards are provisioned via ConfigMaps, they need to be moved to folders after they're loaded. Use the `organize-dashboards.sh` script:

```bash
export GRAFANA_URL="http://your-grafana-url:3000"
export GRAFANA_API_KEY="your-api-key"
./organize-dashboards.sh
```

The script will:
1. Move all agent dashboards to the "Agents" folder
2. Move homepage dashboards to the "Homepage" folder
3. Update dashboard titles as needed

### Dashboard Titles

- `homepage-metrics` → "homepage" (in Homepage folder)
- `google-analytics-homepage` → "Google Analytics" (in Homepage folder)
- All agent dashboards remain in the Agents folder

## Migration from notifi

The dashboards were originally in `notifi/repos/infra/20-platform/services/dashboards/deploy/dashboards/knative/` but have been moved here for centralized management in the homelab prometheus-operator.
