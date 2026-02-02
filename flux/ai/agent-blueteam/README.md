# 🛡️ Agent Blueteam - Defense Runner

**Defensive security agent for threat mitigation and MAG7 battle!**

Agent Blueteam is a defensive security agent that monitors for threats, blocks exploits, and battles the MAG7 dragon boss. It works in conjunction with [agent-redteam](../agent-redteam) to provide a complete security testing and defense demonstration.

## 🎮 MAG7 Battle Demo

The MAG7 (Magnificent 7) dragon is a seven-headed boss representing the tech giants. Each head has unique powers:

| Head | Company | Special Attack | Damage |
|------|---------|---------------|--------|
| 🍎 | Apple | Walled Garden | 30 |
| 🪟 | Microsoft | Blue Screen | 25 |
| 🔍 | Google | Data Harvest | 35 |
| 📦 | Amazon | Cloud Lock | 20 |
| 👓 | Meta | Privacy Void | 40 |
| ⚡ | Tesla | Self-Drive | 45 |
| 🎮 | Nvidia | GPU Meltdown | 50 |

### How to Win

1. **Block exploits** from agent-redteam to deal damage to MAG7
2. **Activate defenses** to protect your cluster
3. **Defeat all 7 heads** to win!

## 🏗️ Architecture

```
agent-blueteam/
├── k8s/
│   └── kustomize/
│       ├── base/           # Base Kubernetes manifests
│       ├── studio/         # Studio cluster overlay
│       └── pro/            # Production cluster overlay
├── src/
│   ├── defense_runner/     # Main handler code
│   │   ├── handler.py      # Defense logic
│   │   ├── main.py         # FastAPI server
│   │   └── Dockerfile
│   └── shared/             # Shared types and metrics
│       ├── types.py        # Data models
│       └── metrics.py      # Prometheus metrics
└── tests/                  # Unit tests
```

## 🚀 Quick Start

### Build and Deploy

```bash
# Build the image
make build

# Push to registry
make push

# Deploy to Kubernetes
make deploy
```

### Local Development

```bash
# Install dependencies
cd src && pip install -r requirements.txt

# Run locally
cd defense_runner && python main.py
```

## 📡 Events

### Events Received

| Event Type | Description |
|------------|-------------|
| `io.homelab.exploit.executed` | Exploit was executed - analyze threat |
| `io.homelab.exploit.success` | Exploit succeeded - CRITICAL alert |
| `io.homelab.exploit.blocked` | Exploit was blocked - log metrics |
| `io.homelab.mag7.attack` | MAG7 is attacking! |
| `io.homelab.demo.game.start` | Game started |

### Events Emitted

| Event Type | Description |
|------------|-------------|
| `io.homelab.threat.detected` | Threat was detected |
| `io.homelab.threat.blocked` | Threat was blocked |
| `io.homelab.defense.activated` | Defense was activated |
| `io.homelab.mag7.damage` | Damage dealt to MAG7 |
| `io.homelab.mag7.defeated` | MAG7 was defeated! |

## 📊 Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `blueteam_threats_detected_total` | Counter | Total threats detected |
| `blueteam_threats_blocked_total` | Counter | Total threats blocked |
| `blueteam_mag7_health` | Gauge | MAG7 dragon health |
| `blueteam_mag7_damage_dealt_total` | Counter | Damage dealt to MAG7 |
| `blueteam_game_score` | Gauge | Current game score |

## 🔗 Integration with Knative Lambda Operator

Agent Blueteam uses the [Knative Lambda Operator](../../infrastructure/knative-lambda-operator) to:

1. **Subscribe to events** via CloudEvents
2. **Scale to zero** when not in use
3. **Auto-scale** under load
4. **Forward events** to other agents

See the [LambdaAgent CRD](k8s/kustomize/base/lambdaagent.yaml) for configuration details.

## 📚 Related Projects

- [agent-redteam](../agent-redteam) - Offensive security agent
- [demo-mag7-battle](../demo-mag7-battle) - Web game demo
- [knative-lambda-operator](../../infrastructure/knative-lambda-operator) - Operator that powers the agents
