# 🏪 Agent-POS-Edge: Edge Computing for Retail & Fast-Food

**AI-powered edge agents for monitoring POS systems in gas stations, McDonald's, and retail**

![Agent-POS](https://img.shields.io/badge/System-Agent%20POS%20Edge-orange)
![Knative Lambda](https://img.shields.io/badge/Powered%20by-Knative%20Lambda%20Operator-green)
![CloudEvents](https://img.shields.io/badge/Events-CloudEvents-blue)

```
   ╔══════════════════════════════════════════════════════════════════╗
   ║  🏪  AGENT-POS-EDGE: RETAIL EDGE MONITORING SYSTEM  🏪          ║
   ║                                                                  ║
   ║   ⛽  Gas Station POS & Pump Monitoring                         ║
   ║   🍔  McDonald's Kitchen & Order Queue                          ║
   ║   📊  Real-time Dashboard & Alerting                            ║
   ╚══════════════════════════════════════════════════════════════════╝
```

## 🎯 Concept

**Agent-POS-Edge** deploys intelligent agents at the edge (POS terminals, kitchen displays, fuel pumps) that communicate via CloudEvents to a central command center. Each location runs lightweight agents that:

- 📡 **Monitor in Real-time** - POS transactions, system health, queue status
- 🔄 **Sync When Connected** - Buffer data offline, sync when connectivity restored
- 🤖 **AI-Powered Insights** - Anomaly detection, predictive maintenance
- ⚡ **React Instantly** - Local decision making without cloud latency

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         AGENT COMMAND CENTER                                 │
│                        (Cloud / Central K8s)                                 │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                     🎛️ command-center                                │   │
│  │                       (LambdaAgent)                                  │   │
│  │  - Aggregates all location data                                     │   │
│  │  - Real-time dashboard                                              │   │
│  │  - Alert management                                                 │   │
│  │  - Fleet configuration                                              │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                              ▲                                               │
│                              │ CloudEvents                                   │
│                   ┌──────────┴──────────┐                                   │
│                   │   RabbitMQ Broker   │                                   │
│                   │ (Knative Eventing)  │                                   │
│                   └──────────┬──────────┘                                   │
│                              │                                               │
└──────────────────────────────┼───────────────────────────────────────────────┘
                               │
          ┌────────────────────┼────────────────────┐
          │                    │                    │
          ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   ⛽ GAS STATION │  │   🍔 MCDONALD'S  │  │   🏪 RETAIL     │
│   Location #001  │  │   Location #042  │  │   Location #103  │
├─────────────────┤  ├─────────────────┤  ├─────────────────┤
│ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌─────────────┐ │
│ │ pos-edge    │ │  │ │ pos-edge    │ │  │ │ pos-edge    │ │
│ │ (Agent)     │ │  │ │ (Agent)     │ │  │ │ (Agent)     │ │
│ └──────┬──────┘ │  │ └──────┬──────┘ │  │ └──────┬──────┘ │
│        │        │  │        │        │  │        │        │
│   ┌────┴────┐   │  │   ┌────┴────┐   │  │   ┌────┴────┐   │
│   │         │   │  │   │         │   │  │   │         │   │
│   ▼         ▼   │  │   ▼         ▼   │  │   ▼         ▼   │
│ ┌─────┐ ┌─────┐ │  │ ┌─────┐ ┌─────┐ │  │ ┌─────┐ ┌─────┐ │
│ │⛽   │ │💳   │ │  │ │🍳   │ │🚗   │ │  │ │💳   │ │📦   │ │
│ │Pump │ │ POS │ │  │ │Kitch│ │Drive│ │  │ │ POS │ │Stock│ │
│ │Agent│ │Term │ │  │ │ en  │ │Thru │ │  │ │Term │ │Room │ │
│ └─────┘ └─────┘ │  │ └─────┘ └─────┘ │  │ └─────┘ └─────┘ │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

## 👥 Agents

### 🎛️ Command Center
Central monitoring and control:
```yaml
role: orchestrator
capabilities:
  - Fleet-wide visibility
  - Real-time alerting
  - Configuration management
  - Analytics & reporting
ai_behavior: "Aggregates data, detects patterns, dispatches alerts"
```

### 📱 POS Edge Agent
Runs on each POS terminal:
```yaml
role: edge-monitor
capabilities:
  - Transaction monitoring
  - Health checks (CPU, RAM, disk, network)
  - Receipt printer status
  - Payment terminal status
ai_behavior: "Monitor locally, buffer offline, sync when connected"
```

### 🍳 Kitchen Agent (Fast-Food)
Monitors kitchen operations:
```yaml
role: kitchen-operations
capabilities:
  - Order queue monitoring
  - Kitchen Display System (KDS) status
  - Cook time tracking
  - Inventory alerts
ai_behavior: "Track order flow, predict delays, optimize queue"
```

### ⛽ Pump Agent (Gas Station)
Monitors fuel operations:
```yaml
role: fuel-operations
capabilities:
  - Pump status monitoring
  - Tank level tracking
  - Transaction flow
  - Safety alerts
ai_behavior: "Monitor fuel operations, predict refill needs"
```

## 📡 CloudEvents

### Location Events

| Event Type | Description | Payload |
|------------|-------------|---------|
| `pos.location.heartbeat` | Location alive signal | `{locationId, timestamp, status}` |
| `pos.location.offline` | Location went offline | `{locationId, lastSeen}` |
| `pos.location.config.update` | Config pushed to location | `{locationId, config}` |

### POS Events

| Event Type | Description | Payload |
|------------|-------------|---------|
| `pos.transaction.started` | New transaction | `{posId, transactionId, items}` |
| `pos.transaction.completed` | Transaction finished | `{posId, transactionId, total, paymentType}` |
| `pos.transaction.failed` | Transaction failed | `{posId, transactionId, error}` |
| `pos.health.report` | System health metrics | `{posId, cpu, memory, disk, network}` |
| `pos.alert.raised` | Alert from POS | `{posId, alertType, severity, message}` |

### Kitchen Events (McDonald's)

| Event Type | Description | Payload |
|------------|-------------|---------|
| `pos.kitchen.order.received` | New order in queue | `{orderId, items, priority}` |
| `pos.kitchen.order.started` | Order being prepared | `{orderId, station, estimatedTime}` |
| `pos.kitchen.order.ready` | Order ready for pickup | `{orderId, prepTime}` |
| `pos.kitchen.queue.status` | Queue depth report | `{queueDepth, avgWaitTime}` |
| `pos.kitchen.equipment.alert` | Equipment issue | `{equipmentId, status, alertType}` |

### Pump Events (Gas Station)

| Event Type | Description | Payload |
|------------|-------------|---------|
| `pos.pump.transaction.start` | Pump activated | `{pumpId, fuelType}` |
| `pos.pump.transaction.end` | Pumping complete | `{pumpId, liters, total}` |
| `pos.pump.status` | Pump status change | `{pumpId, status}` |
| `pos.tank.level` | Tank level report | `{tankId, fuelType, level, capacity}` |
| `pos.tank.alert.low` | Low fuel alert | `{tankId, fuelType, currentLevel}` |

### Command Center Events

| Event Type | Description | Payload |
|------------|-------------|---------|
| `pos.command.config.push` | Push config to edge | `{targetLocations, config}` |
| `pos.command.alert.acknowledge` | Acknowledge alert | `{alertId, operator}` |
| `pos.command.maintenance.schedule` | Schedule maintenance | `{locationId, datetime, type}` |

## 🚀 Quick Start

### Deploy to Kubernetes

```bash
# Deploy the POS edge system
kubectl apply -k k8s/kustomize/studio

# Check agent status
kubectl get lambdaagents -n agent-pos-edge

# Watch events flow
kubectl logs -f -l app.kubernetes.io/part-of=agent-pos-edge -n agent-pos-edge
```

### Simulate Edge Location

```bash
# Send heartbeat from location
curl -X POST http://broker.agent-pos-edge.svc/v1/events \
  -H "Content-Type: application/cloudevents+json" \
  -d '{
    "specversion": "1.0",
    "type": "pos.location.heartbeat",
    "source": "/pos-edge/location/gas-station-001",
    "id": "heartbeat-001",
    "data": {
      "locationId": "gas-station-001",
      "locationType": "gas_station",
      "status": "healthy",
      "posCount": 2,
      "pumpCount": 8
    }
  }'
```

### Send Transaction Event

```bash
# Transaction completed
curl -X POST http://broker.agent-pos-edge.svc/v1/events \
  -H "Content-Type: application/cloudevents+json" \
  -d '{
    "specversion": "1.0",
    "type": "pos.transaction.completed",
    "source": "/pos-edge/location/gas-station-001/pos-01",
    "id": "txn-12345",
    "data": {
      "posId": "pos-01",
      "locationId": "gas-station-001",
      "transactionId": "TXN-2024-12345",
      "total": 45.50,
      "paymentType": "card",
      "items": [
        {"name": "Fuel - Regular", "quantity": 30, "unit": "liters"},
        {"name": "Snacks", "quantity": 2, "price": 5.50}
      ]
    }
  }'
```

## 📊 Metrics

Prometheus metrics for operations:

| Metric | Description |
|--------|-------------|
| `pos_transactions_total` | Total transactions by location |
| `pos_transaction_value_sum` | Total transaction value |
| `pos_transaction_duration_seconds` | Transaction processing time |
| `pos_agent_heartbeat_timestamp` | Last heartbeat from agent |
| `pos_kitchen_queue_depth` | Current kitchen queue size |
| `pos_kitchen_avg_wait_seconds` | Average order wait time |
| `pos_pump_utilization_percent` | Pump utilization rate |
| `pos_tank_level_percent` | Fuel tank level |

## 🏢 Use Cases

### Gas Station
- Monitor all pumps in real-time
- Track fuel levels and predict refill needs
- POS transaction monitoring
- Payment terminal health
- C-store inventory alerts

### McDonald's / Fast-Food
- Kitchen order queue monitoring
- Drive-thru timing optimization
- Multi-POS coordination
- Equipment health monitoring
- Peak hour predictions

### Retail Store
- Multi-lane POS monitoring
- Receipt printer health
- Payment terminal status
- Inventory sync
- Staff scheduling insights

## 🔗 Related Projects

- [knative-lambda-operator](../../infrastructure/knative-lambda-operator) - The operator powering agents
- [agent-webinterface](../agent-webinterface) - Web dashboard
- [agent-rpg](../agent-rpg) - Similar multi-agent architecture

## 📜 License

Part of the homelab project. MIT License.

---

**🏪 Deploy intelligent edge agents across your retail fleet! ⛽🍔**
