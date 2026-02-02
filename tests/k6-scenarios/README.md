# 🧪 Homelab Performance Test Scenarios

**Comprehensive k6 load testing suite for all homelab AI agents**

## 📋 Overview

This directory contains performance test scenarios that simulate real-world interactions between all homelab AI agents. Each scenario tests specific business workflows and cross-agent communication patterns.

## 🤖 Agent Summary

| Agent | Purpose | Key Functions |
|-------|---------|---------------|
| **agent-bruno** | AI Chatbot Assistant | Chat, Q&A, cross-agent queries |
| **agent-contracts** | Smart Contract Security | Fetch, scan, exploit generation, alerts |
| **agent-restaurant** | Fine Dining Experience | Host, Waiter, Sommelier, Chef coordination |
| **agent-pos-edge** | Retail/Fast-Food POS | Transactions, kitchen orders, pump monitoring |
| **agent-store-multibrands** | E-commerce WhatsApp | AI sellers, orders, product catalog |
| **agent-medical** | HIPAA Medical Records | Patient records, RBAC, audit logging |
| **agent-redteam** | Security Testing | Exploit execution, vulnerability testing |
| **agent-blueteam** | Security Defense | Threat detection, MAG7 battle defense |
| **agent-chat** | Multi-modal Chat | Voice, media, location, messaging |
| **agent-devsecops** | Container Security | Image scanning, compliance |
| **agent-tools** | K8s Operations | Cluster management tools |
| **agent-rpg** | AI-Driven RPG Game | Characters, combat, story progression |

---

## 🎯 Test Scenarios

### 1. 🔴🛡️ Security Battle Arena (Redteam vs Blueteam)

**File:** `k6-security-battle-arena.yaml`

Simulates a full security testing cycle where redteam launches attacks and blueteam defends.

**Workflow:**
```
┌─────────────────┐     Attack Events      ┌─────────────────┐
│  AGENT-REDTEAM  │ ─────────────────────► │  AGENT-BLUETEAM │
│                 │                        │                 │
│  • Launch SSRF  │     Defense Events     │  • Block SSRF   │
│  • Cmd Injection│ ◄───────────────────── │  • Log Threat   │
│  • Path Traversal│                       │  • MAG7 Damage  │
└─────────────────┘                        └─────────────────┘
```

**Scenarios Tested:**
- Sequential exploit execution
- Parallel attack waves
- Defense response timing
- MAG7 boss battle mechanics
- Cross-agent event routing

---

### 2. 🍽️ Restaurant Full Service (All Restaurant Agents)

**File:** `k6-restaurant-full-service.yaml`

Complete fine dining simulation from reservation to departure.

**Workflow:**
```
Customer Arrives
       │
       ▼
┌──────────────────┐
│ 🎩 Host Maximilian│ ─── Greets, seats guest
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ 👔 Waiter Pierre │ ─── Presents menu, takes order
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ 🍷 Sommelier     │ ─── Recommends wine pairing
│    Isabella      │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ 👨‍🍳 Chef Marco   │ ─── Prepares dishes
└────────┬─────────┘
         │
         ▼
    Dish Served
```

**Scenarios Tested:**
- VIP anniversary dinner (premium flow)
- Busy Friday night (concurrent tables)
- Large party coordination (8+ guests)
- Special dietary requirements
- Wine cellar integration
- Kitchen timing optimization

---

### 3. 🏪 Store MultiBrands Customer Journey

**File:** `k6-store-customer-journey.yaml`

E-commerce workflow with WhatsApp integration and AI sellers.

**Workflow:**
```
WhatsApp Message
       │
       ▼
┌──────────────────┐
│ WhatsApp Gateway │ ─── Routes to brand
└────────┬─────────┘
         │
    ┌────┴────┬────────┬────────┬────────┐
    ▼         ▼        ▼        ▼        ▼
┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│👗Fashion│ │📱Tech  │ │🏠Home  │ │💄Beauty│ │🎮Gaming│
│ Seller │ │ Seller │ │ Seller │ │ Seller │ │ Seller │
└────┬───┘ └────┬───┘ └────┬───┘ └────┬───┘ └────┬───┘
     │          │          │          │          │
     └──────────┴──────────┴──────────┴──────────┘
                           │
                           ▼
                  ┌──────────────────┐
                  │  Order Processor │
                  └────────┬─────────┘
                           │
                           ▼
                  ┌──────────────────┐
                  │  Sales Assistant │ (Human escalation)
                  └──────────────────┘
```

**Scenarios Tested:**
- Single product inquiry
- Cross-brand shopping (fashion + beauty)
- Order placement and confirmation
- Human seller escalation
- Product recommendations
- Concurrent customer sessions

---

### 4. ⛽🍔 POS Edge Multi-Location Operations

**File:** `k6-pos-edge-fleet.yaml`

Fleet management for gas stations, McDonald's, and retail locations.

**Workflow:**
```
┌──────────────────────────────────────────────────────────────┐
│                     COMMAND CENTER                            │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                  Dashboard & Alerts                      │ │
│  └─────────────────────────────────────────────────────────┘ │
└──────────────────────────┬───────────────────────────────────┘
                           │
          ┌────────────────┼────────────────┐
          ▼                ▼                ▼
   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
   │ ⛽ Gas Stn  │  │ 🍔 McDonald's│  │ 🏪 Retail   │
   │  #001       │  │   #042       │  │   #103      │
   ├─────────────┤  ├─────────────┤  ├─────────────┤
   │ POS Agent   │  │ POS Agent   │  │ POS Agent   │
   │ Pump Agent  │  │ Kitchen Agt │  │             │
   └─────────────┘  └─────────────┘  └─────────────┘
```

**Scenarios Tested:**
- Multi-location heartbeat monitoring
- High-volume transaction processing
- Gas pump operations and tank levels
- McDonald's kitchen queue management
- Drive-thru timing optimization
- Location offline/recovery handling
- Fleet-wide configuration push

---

### 5. 💬🔗 Agent-Bruno Cross-Agent Communication

**File:** `k6-bruno-cross-agent.yaml`

Tests Bruno's ability to communicate with other agents.

**Workflow:**
```
User Chat
    │
    ▼
┌──────────────────┐
│  AGENT-BRUNO     │
│  (Chatbot)       │
└────────┬─────────┘
         │
    ┌────┴────────────────────────────┐
    ▼                                 ▼
┌──────────────────┐        ┌──────────────────┐
│ AGENT-CONTRACTS  │        │  ALERTMANAGER    │
│ (Security Query) │        │  (Alert Query)   │
└────────┬─────────┘        └────────┬─────────┘
         │                           │
         └───────────────────────────┘
                      │
                      ▼
              Response to User
```

**Scenarios Tested:**
- Basic Q&A conversations
- Security status queries (triggers contracts scan)
- Alert status queries (retrieves active alerts)
- Multi-turn conversations
- Concurrent chat sessions
- LLM response latency under load

---

### 6. 🏥 Medical Records HIPAA Compliance

**File:** `k6-medical-hipaa.yaml`

HIPAA-compliant medical records access with RBAC.

**Workflow:**
```
Request with Token
       │
       ▼
┌──────────────────┐
│   RBAC Check     │ ─── Doctor/Nurse/Patient/Admin
└────────┬─────────┘
         │
    ┌────┴────┐
    ▼         ▼
 ALLOWED   DENIED
    │         │
    ▼         ▼
┌────────┐  ┌────────────┐
│Query DB│  │Access Denied│
└────┬───┘  │   Event     │
     │      └────────────┘
     ▼
┌────────────┐
│ Audit Log  │
└────────────┘
```

**Scenarios Tested:**
- Doctor accessing patient records (allowed)
- Nurse accessing assigned patients (allowed)
- Patient accessing own records (allowed)
- Unauthorized access attempts (denied + logged)
- Cross-patient access prevention
- Audit log generation
- HIPAA compliance verification

---

### 7. 🎮 MAG7 Dragon Battle (Gamification)

**File:** `k6-mag7-battle.yaml`

Boss battle simulation with redteam attacks and blueteam defense.

**MAG7 Heads:**
| Head | Company | Attack | Damage |
|------|---------|--------|--------|
| 🍎 Apple | Walled Garden | 30 |
| 🪟 Microsoft | Blue Screen | 25 |
| 🔍 Google | Data Harvest | 35 |
| 📦 Amazon | Cloud Lock | 20 |
| 👓 Meta | Privacy Void | 40 |
| ⚡ Tesla | Self-Drive | 45 |
| 🎮 Nvidia | GPU Meltdown | 50 |

**Scenarios Tested:**
- Game start and initialization
- Attack/defense rounds
- Damage calculation
- Victory/defeat conditions
- Score tracking
- Real-time event streaming

---

### 8. ⚔️ RPG Multi-Character Interaction

**File:** `k6-rpg-adventure.yaml`

AI-driven RPG with character interactions and combat.

**Characters:**
- 🗡️ Crono (Warrior) - Lightning attacks
- ⚡ Lucca (Mage) - Fire magic
- 💫 Marle (Healer) - Ice/Healing
- 🐸 Frog (Paladin) - Water attacks
- 🤖 Robo (Tank) - Shadow/Support
- 🦖 Ayla (Berserker) - Physical power

**Scenarios Tested:**
- Character action selection
- Combat turn order (ATB system)
- Combo attacks between characters
- Story progression events
- Save/load game states
- AI decision making timing

---

## 🚀 Running Tests

### Prerequisites

```bash
# Install k6
brew install k6

# Or run in Kubernetes with k6-operator
kubectl apply -f https://github.com/grafana/k6-operator/releases/latest/download/bundle.yaml
```

### Run Individual Scenarios

```bash
# Security Battle Arena
kubectl apply -f k6-security-battle-arena.yaml

# Restaurant Full Service
kubectl apply -f k6-restaurant-full-service.yaml

# Store Customer Journey
kubectl apply -f k6-store-customer-journey.yaml

# POS Edge Fleet
kubectl apply -f k6-pos-edge-fleet.yaml

# Bruno Cross-Agent
kubectl apply -f k6-bruno-cross-agent.yaml
```

### Run All Scenarios

```bash
# Apply all tests
kubectl apply -f .

# Watch results
kubectl get testruns -A --watch
```

---

## 📊 Metrics & Observability

All tests export metrics to Prometheus via remote write:

| Metric | Description |
|--------|-------------|
| `scenario_success_rate` | Overall scenario success |
| `stage_latency_ms` | Per-stage latency |
| `agent_response_time_ms` | Individual agent response time |
| `cross_agent_events_total` | Cross-agent CloudEvents |
| `business_metric_*` | Domain-specific metrics |

### Grafana Dashboards

Pre-built dashboards available:
- Agent Performance Overview
- Cross-Agent Communication
- Scenario Success Rates
- Business Metrics by Domain

---

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NAMESPACE` | Target namespace | `agent-*` |
| `*_URL` | Agent service URLs | Auto-discovered |
| `K6_PROMETHEUS_RW_SERVER_URL` | Prometheus write endpoint | Required |
| `K6_VUS` | Virtual users override | Scenario default |
| `K6_DURATION` | Duration override | Scenario default |

---

## 📁 Directory Structure

```
tests/k6-scenarios/
├── README.md                        # This file
├── k6-security-battle-arena.yaml    # Redteam vs Blueteam
├── k6-restaurant-full-service.yaml  # Restaurant ecosystem
├── k6-store-customer-journey.yaml   # E-commerce flow
├── k6-pos-edge-fleet.yaml           # POS multi-location
├── k6-bruno-cross-agent.yaml        # Chatbot interactions
├── k6-medical-hipaa.yaml            # HIPAA compliance
├── k6-mag7-battle.yaml              # Boss battle game
└── k6-rpg-adventure.yaml            # RPG game scenarios
```

---

## 📜 License

Part of the homelab project. MIT License.

---

**🧪 Test your agents, measure performance, ensure reliability! 🚀**
