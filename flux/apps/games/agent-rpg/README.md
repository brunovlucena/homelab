# ⚔️ Agent-RPG: AI-Driven Adventure

**An AI-powered RPG inspired by Chrono Trigger & Breath of Fire**

![Agent-RPG](https://img.shields.io/badge/Game-Agent%20RPG-purple)
![Knative Lambda](https://img.shields.io/badge/Powered%20by-Knative%20Lambda%20Operator-green)
![AI Agents](https://img.shields.io/badge/AI-Ollama%20%7C%20OpenAI%20%7C%20Anthropic-blue)

```
   ╔══════════════════════════════════════════════════════════╗
   ║  ⚔️  AGENT-RPG: CHRONICLES OF THE CLOUD KINGDOM  ⚔️     ║
   ║                                                          ║
   ║   🏰  A world where AI agents live their own stories    ║
   ║   🎮  Take control of any character - or watch AI play  ║
   ║   ⚡  Powered by CloudEvents & Kubernetes               ║
   ╚══════════════════════════════════════════════════════════╝
```

## 🎮 Concept

In **Agent-RPG**, every character is a living AI agent running on Kubernetes. They have personalities, memories, and make their own decisions. You can:

- 🎭 **Watch AI Play** - Characters interact, quest, and battle autonomously
- 🕹️ **Take Control** - Assume any character at any time
- 🤝 **Hybrid Mode** - Control one character while AI plays others
- 🌍 **Living World** - NPCs and events continue even when you're away

## 🌟 Features

### 🧙 AI Characters

Each character is a `LambdaAgent` with:
- **Unique Personality** - System prompts define character traits
- **Memory** - Characters remember events and relationships
- **Decision Making** - AI chooses actions based on context
- **Emotions** - Mood affects dialogue and combat choices

### ⚔️ Combat System (ATB - Active Time Battle)

```
┌─────────────────────────────────────────────────────┐
│                 🐉 BOSS: Shadow Dragon              │
│                     HP: ████████░░ 2400/3000        │
├─────────────────────────────────────────────────────┤
│                                                     │
│     👤 Crono      ⚡ Lucca       💫 Marle          │
│   ████████░░    ██████████    ████░░░░░░          │
│   ATB: READY    ATB: READY    ATB: 40%            │
│                                                     │
│  ┌─────────────────────────────────────────────┐  │
│  │ > ⚔️ Attack    💫 Tech    🎒 Item   🏃 Run │  │
│  └─────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### 🗺️ World Events

CloudEvents drive the game world:
- `rpg.world.time.advance` - Day/night cycles
- `rpg.combat.encounter` - Random battles
- `rpg.story.trigger` - Story progression
- `rpg.character.emotion` - Mood changes

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        AGENT-RPG SYSTEM                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    CloudEvents    ┌─────────────────────────┐ │
│  │             │◄──────────────────►│                         │ │
│  │ iOS/Web App │                    │  🎮 game-master         │ │
│  │   Client    │                    │  (LambdaAgent)          │ │
│  │             │                    │  - World state          │ │
│  └─────────────┘                    │  - Story progression    │ │
│        ▲                            │  - Combat orchestration │ │
│        │ WebSocket                  └──────────▲──────────────┘ │
│        ▼                                       │                │
│  ┌─────────────┐                    ┌──────────┴──────────┐    │
│  │   Gateway   │◄──────────────────►│   RabbitMQ Broker   │    │
│  │  (Ingress)  │                    │   (Knative Eventing)│    │
│  └─────────────┘                    └──────────┬──────────┘    │
│                                                │                │
│        ┌───────────────────┬─────────────────┬┴────────────┐   │
│        ▼                   ▼                 ▼             ▼   │
│  ┌───────────┐       ┌───────────┐     ┌───────────┐ ┌────────┐│
│  │ 🗡️ Crono  │       │ ⚡ Lucca  │     │ 💫 Marle  │ │  ...   ││
│  │ (Agent)   │       │ (Agent)   │     │ (Agent)   │ │ (NPCs) ││
│  │           │       │           │     │           │ │        ││
│  │ Brave     │       │ Genius    │     │ Kind      │ │        ││
│  │ Leader    │       │ Inventor  │     │ Healer    │ │        ││
│  └───────────┘       └───────────┘     └───────────┘ └────────┘│
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    💾 world-state                        │   │
│  │  (Redis/PostgreSQL)                                     │   │
│  │  - Character stats & inventory                          │   │
│  │  - World state & flags                                  │   │
│  │  - Conversation history                                 │   │
│  │  - Save games                                           │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                 📊 Observability Stack                   │   │
│  │  Prometheus (metrics) | Loki (logs) | Tempo (traces)    │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## 👥 Characters

### 🗡️ Crono - The Silent Hero
```yaml
personality: brave, determined, protective
class: Warrior
element: Lightning ⚡
special: "X-Strike" (combo with Frog)
ai_behavior: "Rushes to protect allies, prioritizes threats"
```

### ⚡ Lucca - The Genius Inventor
```yaml
personality: analytical, creative, sarcastic
class: Mage/Engineer
element: Fire 🔥
special: "Flame Toss", "Hypno Wave"
ai_behavior: "Analyzes enemy weaknesses, uses tech strategically"
```

### 💫 Marle - The Compassionate Princess
```yaml
personality: kind, optimistic, rebellious
class: Healer/Support
element: Ice ❄️
special: "Aura", "Ice"
ai_behavior: "Monitors party health, heals proactively"
```

### 🐸 Frog - The Chivalrous Knight
```yaml
personality: honorable, melancholic, loyal
class: Paladin
element: Water 💧
special: "Slurp Slash", "Heal"
ai_behavior: "Protects the weak, challenges strongest foe"
```

### 🤖 Robo - The Gentle Machine
```yaml
personality: curious, logical, empathetic
class: Tank/Support
element: Shadow 🌑
special: "Rocket Punch", "Cure Beam"
ai_behavior: "Calculates optimal actions, protects efficiently"
```

### 🦖 Ayla - The Prehistoric Warrior
```yaml
personality: fierce, primal, loyal
class: Berserker
element: Physical 💪
special: "Cat Attack", "Charm"
ai_behavior: "Attacks strongest enemy, goes berserk when low HP"
```

## 📡 CloudEvents

### Game Events

| Event Type | Description | Payload |
|------------|-------------|---------|
| `rpg.game.start` | New game started | `{gameId, players}` |
| `rpg.game.save` | Save game request | `{gameId, slot}` |
| `rpg.game.load` | Load game request | `{slot}` |

### Character Events

| Event Type | Description | Payload |
|------------|-------------|---------|
| `rpg.character.action` | Character performs action | `{characterId, action, target}` |
| `rpg.character.speak` | Character dialogue | `{characterId, text, emotion}` |
| `rpg.character.move` | Character movement | `{characterId, x, y, zone}` |
| `rpg.character.control.request` | Player wants control | `{characterId, playerId}` |
| `rpg.character.control.release` | Player releases control | `{characterId}` |

### Combat Events

| Event Type | Description | Payload |
|------------|-------------|---------|
| `rpg.combat.start` | Battle begins | `{enemies, party}` |
| `rpg.combat.turn.ready` | ATB filled | `{characterId}` |
| `rpg.combat.action.execute` | Action performed | `{action, actor, target, damage}` |
| `rpg.combat.end` | Battle ends | `{result, exp, loot}` |

### World Events

| Event Type | Description | Payload |
|------------|-------------|---------|
| `rpg.world.time.tick` | Time passes | `{hour, day, weather}` |
| `rpg.world.zone.enter` | Enter new area | `{zone, characters}` |
| `rpg.story.flag.set` | Story progression | `{flag, value}` |
| `rpg.world.npc.spawn` | NPC appears | `{npcId, zone}` |

## 🚀 Quick Start

### Deploy to Kubernetes

```bash
# Deploy the game system
kubectl apply -k k8s/kustomize/studio

# Get the game URL
kubectl get ksvc -n agent-rpg

# Watch characters interact
kubectl logs -f -l app.kubernetes.io/part-of=agent-rpg -n agent-rpg
```

### Start a New Game

```bash
# Start game via CloudEvent
curl -X POST http://game-master.agent-rpg.svc/game/new \
  -H "Content-Type: application/json" \
  -d '{"playerId": "bruno", "difficulty": "normal"}'
```

### Take Control of a Character

```bash
# Request control of Crono
curl -X POST http://game-master.agent-rpg.svc/control \
  -H "Content-Type: application/json" \
  -d '{"playerId": "bruno", "characterId": "crono"}'
```

## 📱 iOS App (Future)

The iOS app will connect via WebSocket to:
- Receive real-time game state updates
- Send player commands
- View character perspectives
- Watch AI play

### SwiftUI Preview

```swift
struct GameView: View {
    @StateObject var gameState: GameState
    
    var body: some View {
        ZStack {
            // 16-bit style game world
            WorldView(zone: gameState.currentZone)
            
            // Character sprites
            ForEach(gameState.party) { character in
                CharacterSprite(character: character)
                    .position(character.position)
            }
            
            // Combat overlay when in battle
            if gameState.inCombat {
                CombatView(combat: gameState.combat)
            }
            
            // Dialogue box
            if let dialogue = gameState.activeDialogue {
                DialogueBox(dialogue: dialogue)
            }
        }
    }
}
```

## 🎨 Visual Style

Modern pixel art inspired by:
- Chrono Trigger (SNES)
- Breath of Fire II (SNES)
- Octopath Traveler (HD-2D)

```
┌───────────────────────────────────────────────┐
│     🌳🌳🌳    ☀️    🌳🌳🌳                    │
│   🌳      🌳      🌳      🌳                  │
│        🏠  🏠  🏠                              │
│     ═══════════════════                       │
│          🗡️ ⚡ 💫                             │
│        (Party walking)                        │
│     ═══════════════════                       │
│   🌲      🌲      🌲      🌲                  │
│     🌲🌲🌲    💧    🌲🌲🌲                    │
└───────────────────────────────────────────────┘
```

## 📊 Metrics

Prometheus metrics for game analytics:

| Metric | Description |
|--------|-------------|
| `rpg_battles_total` | Total battles fought |
| `rpg_character_deaths_total` | Character death count |
| `rpg_ai_decisions_total` | AI decisions made |
| `rpg_player_actions_total` | Player actions taken |
| `rpg_session_duration_seconds` | Play session length |
| `rpg_story_progress_percent` | Story completion |

## 🔗 Related Projects

- [knative-lambda-operator](../../infrastructure/knative-lambda-operator) - The operator powering agents
- [agent-bruno](../agent-bruno) - AI chatbot agent
- [demo-mag7-battle](../demo-mag7-battle) - Another game demo

## 📜 License

Part of the homelab project. MIT License.

---

**⚔️ Begin your adventure in the Cloud Kingdom! 🏰**
