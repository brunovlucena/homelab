# 🍽️ Agent-Restaurant: AI-Powered Fine Dining Experience

**AI agents that transform restaurant operations with personalized service**

![Agent-Restaurant](https://img.shields.io/badge/System-Agent%20Restaurant-gold)
![Knative Lambda](https://img.shields.io/badge/Powered%20by-Knative%20Lambda%20Operator-green)
![CloudEvents](https://img.shields.io/badge/Events-CloudEvents-blue)

```
   ╔══════════════════════════════════════════════════════════════════════╗
   ║  🍽️  AGENT-RESTAURANT: INTELLIGENT DINING EXPERIENCE  🍽️           ║
   ║                                                                      ║
   ║   👨‍🍳  AI Chef - Dishes & Kitchen Orchestration                      ║
   ║   🍷  AI Sommelier - Wine Pairing & Recommendations                  ║
   ║   👔  AI Waiter - Dish Presentation & Service                        ║
   ║   🎩  AI Host - Reservations & Seating                               ║
   ╚══════════════════════════════════════════════════════════════════════╝
```

## 🎯 Concept

**Agent-Restaurant** brings AI-powered agents to fine dining. Each agent has a personality and expertise, working together to create memorable dining experiences:

- 🍽️ **Personalized Recommendations** - AI learns customer preferences
- 🎭 **Theatrical Presentation** - Agents describe dishes with flair
- 🍷 **Perfect Pairings** - Wine suggestions based on orders
- ⏱️ **Optimized Flow** - Kitchen and service coordination

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    RESTAURANT COMMAND CENTER                                 │
│                      (Next.js Web Interface)                                 │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  🖥️ Dashboard                                                          │ │
│  │  - Live table status        - Kitchen queue                            │ │
│  │  - Active orders            - Revenue metrics                          │ │
│  │  - Agent activity           - Customer satisfaction                    │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                              ▲                                               │
│                              │ CloudEvents                                   │
│                   ┌──────────┴──────────┐                                   │
│                   │   RabbitMQ Broker   │                                   │
│                   └──────────┬──────────┘                                   │
└──────────────────────────────┼───────────────────────────────────────────────┘
                               │
         ┌─────────────────────┼─────────────────────┐
         │                     │                     │
         ▼                     ▼                     ▼
┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐
│   🎩 HOST       │   │   👔 WAITER     │   │   🍷 SOMMELIER  │
│   Agent         │   │   Agent         │   │   Agent         │
├─────────────────┤   ├─────────────────┤   ├─────────────────┤
│ - Greet guests  │   │ - Present menu  │   │ - Wine list     │
│ - Reservations  │   │ - Take orders   │   │ - Pairings      │
│ - Table assign  │   │ - Dish stories  │   │ - Sommelier     │
│ - Wait list     │   │ - Serve food    │   │   recommendations│
└─────────────────┘   └─────────────────┘   └─────────────────┘
                               │
                               ▼
                     ┌─────────────────┐
                     │   👨‍🍳 CHEF      │
                     │   Agent         │
                     ├─────────────────┤
                     │ - Kitchen queue │
                     │ - Dish timing   │
                     │ - Quality ctrl  │
                     │ - Special reqs  │
                     └─────────────────┘
```

## 👥 AI Agents

### 🎩 Host Agent - "Maximilian"
```yaml
personality: Elegant, warm, attentive
capabilities:
  - Guest greeting with personalization
  - Table assignment optimization
  - Wait time management
  - VIP recognition
ai_behavior: "Welcomes guests warmly, remembers regulars, optimizes seating"
voice_style: "Good evening, Mr. Santos! Your favorite table by the window awaits."
```

### 👔 Waiter Agent - "Pierre"
```yaml
personality: Knowledgeable, charming, theatrical
capabilities:
  - Menu presentation with storytelling
  - Dish recommendations based on preferences
  - Order taking and modifications
  - Dish presentation narration
ai_behavior: "Presents dishes with passion, tells ingredient stories, anticipates needs"
voice_style: "This evening's risotto features hand-harvested porcini from the Umbrian hills..."
```

### 🍷 Sommelier Agent - "Isabella"
```yaml
personality: Sophisticated, passionate, educational
capabilities:
  - Wine pairing recommendations
  - Wine list navigation
  - Tasting notes presentation
  - Budget-aware suggestions
ai_behavior: "Matches wines perfectly, educates without condescension, knows cellar by heart"
voice_style: "May I suggest a 2019 Barolo? Its earthy notes will beautifully complement your truffle risotto."
```

### 👨‍🍳 Chef Agent - "Marco"
```yaml
personality: Creative, precise, passionate
capabilities:
  - Kitchen queue management
  - Dish timing coordination
  - Special request handling
  - Quality assurance
ai_behavior: "Orchestrates kitchen flow, ensures perfect timing, handles dietary needs"
voice_style: "Table 7's lamb needs 3 more minutes for that perfect pink center."
```

## 📡 CloudEvents

### Guest Events

| Event Type | Description | Payload |
|------------|-------------|---------|
| `restaurant.guest.arrived` | Guest arrives | `{guestId, partySize, reservation}` |
| `restaurant.guest.seated` | Guest seated | `{tableId, guestId, server}` |
| `restaurant.guest.departed` | Guest leaves | `{guestId, totalSpent, rating}` |

### Order Events

| Event Type | Description | Payload |
|------------|-------------|---------|
| `restaurant.order.created` | New order | `{orderId, tableId, items}` |
| `restaurant.order.modified` | Order changed | `{orderId, changes}` |
| `restaurant.order.ready` | Ready to serve | `{orderId, items}` |
| `restaurant.order.served` | Dish served | `{orderId, dish, presentation}` |

### Kitchen Events

| Event Type | Description | Payload |
|------------|-------------|---------|
| `restaurant.kitchen.ticket.received` | New ticket | `{ticketId, items, priority}` |
| `restaurant.kitchen.dish.started` | Cooking started | `{ticketId, dish, station}` |
| `restaurant.kitchen.dish.ready` | Dish ready | `{ticketId, dish, quality}` |
| `restaurant.kitchen.alert` | Kitchen alert | `{type, message, urgency}` |

### Service Events

| Event Type | Description | Payload |
|------------|-------------|---------|
| `restaurant.service.presentation` | Dish presentation | `{tableId, dish, narrative}` |
| `restaurant.service.wine.poured` | Wine service | `{tableId, wine, pairing}` |
| `restaurant.service.feedback` | Guest feedback | `{tableId, rating, comments}` |

## 🍽️ Dish Presentation System

Each dish can have a personalized presentation:

```json
{
  "dish": "Risotto ai Porcini",
  "presentation": {
    "opening": "From Chef Marco's autumn collection...",
    "story": "Hand-foraged porcini mushrooms from the Apennine mountains, slowly stirred with aged Parmigiano Reggiano over 18 minutes to achieve this creamy perfection.",
    "pairing": "Isabella suggests the 2020 Gavi di Gavi, whose crisp minerality cuts through the richness beautifully.",
    "instruction": "I recommend starting from the center, where the truffle oil pools.",
    "closing": "Buon appetito!"
  }
}
```

## 🚀 Quick Start

### Deploy

```bash
# Deploy to studio
kubectl apply -k k8s/kustomize/studio

# Check agents
kubectl get lambdaagents -n agent-restaurant
```

### Run Command Center

```bash
cd web
npm install
npm run dev
# Open http://localhost:3000
```

## 📊 Command Center Features

### Dashboard
- **Live Floor Plan** - Real-time table status
- **Kitchen Queue** - Orders in progress
- **Agent Activity** - What each agent is doing
- **Revenue Metrics** - Tonight's performance

### Menu Management
- Add/edit dishes with presentation scripts
- Wine pairing suggestions
- Dietary information

### Guest Management
- Reservation system
- Guest preferences history
- VIP recognition

### Analytics
- Popular dishes
- Peak hours analysis
- Agent performance

## 🎨 Visual Design

The Command Center features a warm, elegant design inspired by fine dining:

- **Color Palette**: Burgundy, gold, cream, dark wood tones
- **Typography**: Elegant serif for headings, clean sans-serif for data
- **Animations**: Subtle, refined transitions

## 📱 Customer App (Future)

- View menu with AI descriptions
- See dish presentation videos
- Interact with AI sommelier
- Leave feedback

## 🔗 Related Projects

- [agent-pos-edge](../agent-pos-edge) - POS system for fast-food
- [knative-lambda-operator](../../infrastructure/knative-lambda-operator) - Powers the agents
- [agent-webinterface](../agent-webinterface) - Generic command center

## 📜 License

Part of the homelab project. MIT License.

---

**🍷 Elevate your dining experience with AI! 🍽️**
