# 🐉 MAG7 Battle - Knative Lambda Operator Demo

**A Space Invaders-style game demonstrating the Knative Lambda Operator!**

![MAG7 Battle](https://img.shields.io/badge/Demo-MAG7%20Battle-blue)
![Knative Lambda](https://img.shields.io/badge/Powered%20by-Knative%20Lambda%20Operator-green)

## 🎮 The Game

MAG7 Battle is an interactive demo that showcases the capabilities of the Knative Lambda Operator through an engaging Space Invaders-style game.

### Story

The **MAG7 Dragon** 🐉 has awakened! This seven-headed beast represents the "Magnificent 7" tech giants, each head with unique attack powers:

| Head | Company | Special Attack | Damage |
|------|---------|---------------|--------|
| 🍎 | Apple | Walled Garden | 30 |
| 🪟 | Microsoft | Blue Screen | 25 |
| 🔍 | Google | Data Harvest | 35 |
| 📦 | Amazon | Cloud Lock | 20 |
| 👓 | Meta | Privacy Void | 40 |
| ⚡ | Tesla | Self-Drive | 45 |
| 🎮 | Nvidia | GPU Meltdown | 50 |

### How to Play

1. **agent-redteam** spawns exploits (falling enemies) using LambdaFunctions
2. **You control agent-blueteam** 🛡️ to defend against exploits
3. **Block exploits** to deal damage to MAG7
4. **Defeat all 7 heads** to win!

### Controls

- **← →** or **A D**: Move left/right
- **Space**: Shoot defense projectile
- **P**: Pause game

### Exploit Types

Each exploit is a real vulnerability that the Knative Lambda Operator defends against:

| Emoji | Exploit | Description |
|-------|---------|-------------|
| 🌐 | SSRF | Server-Side Request Forgery via go-git |
| 💉 | Injection | Go Template Code Injection |
| ⚡ | CMD Exec | Command Injection via git URL |
| 🔥 | Code Exec | Arbitrary inline code execution |
| 👑 | RBAC | Privilege escalation to cluster-admin |
| 🔑 | Token | Service Account token exposure |

## 🏗️ Architecture

```
demo-mag7-battle/
├── k8s/
│   └── kustomize/
│       ├── base/
│       │   ├── namespace.yaml
│       │   ├── kustomization.yaml
│       │   ├── lambdafunction-game-server.yaml    # Serves the game UI
│       │   ├── lambdafunction-exploit-spawner.yaml # Spawns exploits (RedTeam tool)
│       │   └── configmap-game-ui.yaml             # HTML5 game code
│       ├── studio/
│       └── pro/
├── static/                    # Standalone HTML for local testing
│   └── index.html
├── Makefile
└── README.md
```

### Components

1. **mag7-game-server** (LambdaFunction)
   - Serves the HTML5 game
   - Handles game state via REST API
   - Scales to zero when not in use

2. **mag7-exploit-spawner** (LambdaFunction)
   - Tool for agent-redteam
   - Generates exploit waves
   - Demonstrates LambdaFunction as agent tools

3. **agent-blueteam** (LambdaAgent)
   - Defensive agent
   - Monitors for threats
   - Blocks exploits and damages MAG7

4. **agent-redteam** (LambdaAgent)
   - Offensive agent
   - Uses LambdaFunctions as exploit tools
   - Spawns attack waves

## 🚀 Quick Start

### Local Development (Browser Only)

```bash
# Open the game directly in your browser
open static/index.html
```

### Deploy to Kubernetes

```bash
# Deploy the demo
make deploy

# Get the game URL
make url

# Watch the game in action
make logs
```

### Full Demo with Agents

```bash
# Deploy all components
kubectl apply -k k8s/kustomize/studio

# Deploy agent-blueteam
kubectl apply -k ../agent-blueteam/k8s/kustomize/studio

# Deploy agent-redteam (to spawn real exploits)
kubectl apply -k ../agent-redteam/k8s/kustomize/studio

# Start the game
curl -X POST http://mag7-game-server.demo-mag7-battle/game/start
```

## 📡 Events

The demo uses CloudEvents to communicate between components:

### Game Events

| Event Type | Description |
|------------|-------------|
| `io.homelab.demo.game.start` | Game started |
| `io.homelab.demo.game.wave` | New wave incoming |
| `io.homelab.demo.spawn.wave` | Request to spawn exploit wave |
| `io.homelab.demo.spawn.single` | Request to spawn single exploit |

### Integration Events

| Event Type | Description |
|------------|-------------|
| `io.homelab.exploit.blocked` | Exploit was blocked by blueteam |
| `io.homelab.mag7.damage` | MAG7 took damage |
| `io.homelab.mag7.defeated` | MAG7 was defeated |

## 📊 Metrics

The demo exposes Prometheus metrics:

| Metric | Description |
|--------|-------------|
| `demo_game_score` | Current game score |
| `demo_game_wave` | Current wave number |
| `demo_exploits_spawned_total` | Total exploits spawned |
| `demo_exploits_blocked_total` | Total exploits blocked |
| `demo_mag7_health` | MAG7 remaining health |

## 🔗 Related Projects

- [knative-lambda-operator](../../infrastructure/knative-lambda-operator) - The operator that powers everything
- [agent-redteam](../agent-redteam) - Offensive security agent
- [agent-blueteam](../agent-blueteam) - Defensive security agent
- [agent-bruno](../agent-bruno) - AI chatbot agent

## 📜 License

Part of the homelab project. MIT License.

---

**🎮 Have fun defeating MAG7! 🐉**
