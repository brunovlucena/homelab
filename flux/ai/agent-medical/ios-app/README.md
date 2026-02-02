# 📱 Agent Chat - iOS App

**Reusable iOS app for communicating with homelab AI agents**

A native SwiftUI app optimized for iPhone 14 Pro that connects to any CloudEvents-compatible agent running in your homelab via VPN.

## 🎯 Features

- **🔌 Multi-Agent Support**: Connect to any agent (medical, assistant, code, custom)
- **☁️ CloudEvents Protocol**: Full CloudEvents 1.0 specification support
- **🔐 Authentication**: Token-based auth with role-based access
- **💬 Modern Chat UI**: Beautiful, native iOS chat interface
- **📊 Response Metadata**: View model info, tokens used, latency
- **⚙️ Configurable**: Add custom agents, change themes, adjust settings
- **🔄 Offline Ready**: Conversations saved locally

## 📋 Requirements

- iOS 17.0+
- iPhone 14 Pro or later (recommended)
- VPN connection to homelab cluster
- Xcode 15.0+ (for development)

## 🚀 Quick Start

### Open in Xcode

```bash
cd ios-app/AgentChat
open AgentChat.xcodeproj
```

### Build & Run

1. Open `AgentChat.xcodeproj` in Xcode
2. Select your iPhone 14 Pro (device or simulator)
3. Press `Cmd + R` to build and run

### Connect to Agent

1. Complete the onboarding flow
2. Enter your agent's URL (e.g., `http://agent-medical.agent-medical.svc.cluster.local:8080`)
3. Or use the pre-configured Medical Agent

## 🏗️ Architecture

```
AgentChat/
├── Models/
│   ├── Agent.swift          # Agent configuration model
│   ├── Message.swift        # Chat message models
│   └── CloudEvent.swift     # CloudEvents protocol types
├── Services/
│   ├── AgentService.swift   # Network layer (CloudEvents API)
│   └── StorageService.swift # Local persistence (UserDefaults)
├── ViewModels/
│   ├── ChatViewModel.swift  # Chat logic & state
│   └── AppViewModel.swift   # App-wide state management
├── Views/
│   ├── ChatView.swift       # Main chat interface
│   ├── HomeView.swift       # Home screen with conversations
│   ├── AgentPickerView.swift # Agent selection
│   ├── SettingsView.swift   # App settings
│   └── OnboardingView.swift # Initial setup flow
├── Components/
│   ├── MessageBubble.swift  # Reusable chat bubble
│   ├── ChatInputBar.swift   # Text input component
│   └── AgentStatusBadge.swift # Status indicator
├── Config/
│   └── Config.swift         # App configuration
└── Extensions/
    └── View+Extensions.swift # SwiftUI helpers
```

## 🔧 Configuration

### Default Agents

The app comes pre-configured with:

| Agent | URL | Description |
|-------|-----|-------------|
| Medical | `agent-medical.agent-medical.svc.cluster.local:8080` | HIPAA-compliant medical records |
| Assistant | `agent-assistant.agents.svc.cluster.local:8080` | General purpose AI assistant |
| Code | `agent-code.agents.svc.cluster.local:8080` | Programming assistant |

### Adding Custom Agents

1. Tap the CPU icon in the navigation bar
2. Tap "Add Custom Agent"
3. Enter:
   - **Name**: Display name for the agent
   - **Description**: What the agent does
   - **Base URL**: Full URL to the agent endpoint
   - **Icon**: Choose from SF Symbols
   - **Color**: Pick a theme color
4. Test the connection
5. Save

### User Roles

| Role | Access |
|------|--------|
| Doctor | Full access to all patient records |
| Nurse | Access to assigned patients |
| Patient | Own records only |
| Admin | Administrative access |
| User | Generic role for non-medical agents |

## 📡 API Integration

### CloudEvents Request

```http
POST / HTTP/1.1
Host: agent-medical.agent-medical.svc.cluster.local:8080
Content-Type: application/json
ce-specversion: 1.0
ce-type: io.homelab.medical.query
ce-source: /ios-app/agent-chat
ce-id: <uuid>
Authorization: Bearer <token>

{
  "query": "Show my lab results",
  "patient_id": "patient-123",
  "conversation_id": "<uuid>"
}
```

### CloudEvents Response

```json
{
  "specversion": "1.0",
  "type": "io.homelab.medical.response",
  "source": "/agent-medical/records",
  "data": {
    "agent": "agent-medical",
    "response": "Your recent lab results show...",
    "patient_id": "patient-123",
    "records": [...],
    "model": "llama3.2:3b",
    "tokens_used": 256,
    "duration_ms": 1234.5,
    "audit_id": "audit-789"
  }
}
```

## 🎨 Customization

### Themes

- System (follows iOS dark/light mode)
- Light
- Dark

### Font Sizes

- Small (0.9x)
- Medium (1.0x default)
- Large (1.15x)

### Settings

- Show response metadata (model, tokens, latency)
- Auto-scroll to bottom
- Haptic feedback

## 🔐 Security

- All communication over HTTPS (when using proper certs)
- Token-based authentication
- Credentials stored in UserDefaults (consider Keychain for production)
- VPN required for cluster access

## 🛠️ Development

### SwiftUI Previews

All views support SwiftUI Previews. Use `Cmd + Option + P` to resume previews.

### Testing Locally

1. Run the agent locally:
   ```bash
   cd ../src/medical_agent
   uvicorn main:app --host 0.0.0.0 --port 8080
   ```

2. Update agent URL to `http://localhost:8080`

3. Run the iOS app on Simulator

### Adding New Agents

The architecture is designed for easy extension:

1. Add agent to `Agent.swift` static properties
2. Configure event types in `Config.swift`
3. Update `OnboardingView.swift` quick options (optional)

## 📱 Device Support

Optimized for:
- iPhone 14 Pro
- iPhone 14 Pro Max
- iPhone 15 series
- iPad Pro (with adaptations)

Minimum:
- Any device running iOS 17.0+

## 🐛 Troubleshooting

### Agent Offline

1. Check VPN connection
2. Verify agent is running: `kubectl get pods -n agent-medical`
3. Test health endpoint: `curl http://<agent-url>/health`

### Authentication Failed

1. Verify token is correct
2. Check user role matches required permissions
3. Review agent logs for details

### Network Timeout

1. LLM inference can take 30-120 seconds
2. Check agent resource limits
3. Verify cluster network connectivity

## 📄 License

Part of the homelab project. MIT License.

---

**📱 Chat with your AI agents from anywhere! 📱**
