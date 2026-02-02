# 📊 K6 Performance Test Scenario Dashboards

**Grafana dashboards for monitoring k6 performance test scenarios and Lambda function metrics**

## 🎯 Dashboard Overview

| Dashboard | UID | Description |
|-----------|-----|-------------|
| **K6 Scenarios Overview** | `k6-scenarios-overview` | Master dashboard showing all test scenarios at a glance |
| **Security Battle Arena** | `k6-security-battle` | Redteam vs Blueteam attack/defense metrics |
| **Restaurant Full Service** | `k6-restaurant-service` | Fine dining flow with all restaurant agents |
| **Store Customer Journey** | `k6-store-customer` | E-commerce with WhatsApp integration |
| **POS Edge Fleet** | `k6-pos-edge-fleet` | Multi-location POS operations |
| **Cross-Agent Communication** | `k6-cross-agent` | Agent Bruno's interactions with other agents |
| **Medical HIPAA** | `k6-medical-hipaa` | HIPAA compliance testing metrics |

---

## 📁 Directory Structure

```
dashboards/
├── README.md                              # This file
├── k6-dashboards-configmaps.yaml          # Kubernetes ConfigMaps for deployment
├── k6-scenarios-overview-dashboard.json   # Master overview dashboard
├── k6-security-battle-dashboard.json      # Security battle arena
├── k6-restaurant-service-dashboard.json   # Restaurant full service
├── k6-store-customer-dashboard.json       # Store customer journey
├── k6-pos-edge-fleet-dashboard.json       # POS edge fleet
├── k6-cross-agent-dashboard.json          # Cross-agent communication
└── k6-medical-hipaa-dashboard.json        # Medical HIPAA compliance
```

---

## 🚀 Deployment

### Option 1: Using Grafana Sidecar (Recommended)

Deploy the ConfigMaps with the `grafana_dashboard: "1"` label for automatic import:

```bash
kubectl apply -f k6-dashboards-configmaps.yaml
```

### Option 2: Manual Import

Import the JSON files directly through Grafana UI:

1. Go to Grafana → Dashboards → Import
2. Upload the `.json` file or paste the JSON content
3. Select the Prometheus datasource
4. Click Import

### Option 3: Copy to Dashboard Directory

Copy the JSON files to the existing dashboards directory:

```bash
cp *.json /path/to/homelab/flux/infrastructure/prometheus-operator/k8s/dashboards/
```

---

## 📈 Metrics Reference

### K6 Standard Metrics

| Metric | Description |
|--------|-------------|
| `k6_http_reqs_total` | Total HTTP requests made |
| `k6_http_req_duration_bucket` | HTTP request duration histogram |
| `k6_http_req_failed_total` | Failed HTTP requests |
| `k6_vus` | Current number of virtual users |
| `k6_iterations_total` | Total test iterations completed |
| `k6_data_received_total` | Total data received |
| `k6_data_sent_total` | Total data sent |

### Security Battle Metrics

| Metric | Description |
|--------|-------------|
| `battle_attacks_launched` | Total attacks launched by redteam |
| `battle_attacks_blocked` | Attacks blocked by blueteam |
| `battle_attacks_succeeded` | Successful attacks |
| `battle_mag7_damage_dealt` | Damage dealt to MAG7 dragon |
| `battle_success_rate` | Overall battle success rate |

### Restaurant Metrics

| Metric | Description |
|--------|-------------|
| `restaurant_guests_served` | Total guests served |
| `restaurant_orders_placed` | Orders placed |
| `restaurant_dishes_prepared` | Dishes prepared by chef |
| `restaurant_wine_pairings` | Wine pairings served |
| `restaurant_dining_success_rate` | Dining success rate |

### Store Metrics

| Metric | Description |
|--------|-------------|
| `store_whatsapp_messages` | WhatsApp messages processed |
| `store_product_inquiries` | Product inquiries |
| `store_orders_placed` | Orders placed |
| `store_human_escalations` | Human escalations |
| `store_conversion_rate` | Conversion rate |

### POS Edge Metrics

| Metric | Description |
|--------|-------------|
| `pos_locations_online` | Online locations |
| `pos_transactions_total` | Total transactions |
| `pos_kitchen_orders` | Kitchen orders |
| `pos_pump_operations` | Gas pump operations |
| `pos_fleet_availability` | Fleet availability |

### Cross-Agent Metrics

| Metric | Description |
|--------|-------------|
| `bruno_chat_messages` | Chat messages processed |
| `bruno_cross_agent_calls` | Cross-agent calls |
| `bruno_llm_response_time_bucket` | LLM response time |
| `bruno_chat_success_rate` | Chat success rate |

### Medical HIPAA Metrics

| Metric | Description |
|--------|-------------|
| `medical_hipaa_compliance_rate` | HIPAA compliance rate |
| `medical_records_accessed` | Records accessed |
| `medical_authorized_access` | Authorized access attempts |
| `medical_unauthorized_attempts` | Unauthorized access attempts |
| `medical_audit_events` | Audit events generated |

---

## 🔧 Lambda Function Metrics

All dashboards include Lambda function metrics from the Knative Lambda Operator:

| Metric | Description |
|--------|-------------|
| `knative_lambda_operator_invocations_total` | Total function invocations |
| `knative_lambda_operator_cold_starts_total` | Cold start events |
| `knative_lambda_operator_warm_starts_total` | Warm start events |
| `knative_lambda_operator_function_duration_seconds_bucket` | Function execution duration |
| `knative_lambda_operator_reconcile_total` | Operator reconciliation events |
| `knative_lambda_operator_workqueue_depth` | Operator workqueue depth |

---

## 🔗 Dashboard Links

Each dashboard includes links to:
- **K6 Scenarios Overview** - Master dashboard
- **K6 Knative Lambda** - General k6/Lambda metrics
- **Agent-specific dashboards** - Related agent dashboards

---

## 📊 Variables

All dashboards support these variables:

| Variable | Description |
|----------|-------------|
| `datasource` | Prometheus datasource selector |
| `testid` | Filter by specific test ID (for k6 metrics) |

---

## 🎨 Dashboard Features

### Common Features
- **Auto-refresh**: 10 seconds default
- **Time range**: 30 minutes to 1 hour default
- **Annotations**: Alerts overlay on panels
- **Cross-linking**: Navigation between related dashboards

### Panel Types Used
- **Stat panels**: Key metrics at a glance
- **Time series**: Trends over time
- **Bar charts**: Comparisons
- **Gauges**: Performance indicators

### Color Coding
- 🟢 **Green**: Good/Success/Healthy
- 🟡 **Yellow**: Warning/Caution
- 🔴 **Red**: Error/Critical/Alert
- 🔵 **Blue**: Information/Active
- 🟣 **Purple**: Special/Premium

---

## 📜 License

Part of the homelab project. MIT License.

---

**📊 Monitor your k6 tests and Lambda functions in real-time! 🚀**
