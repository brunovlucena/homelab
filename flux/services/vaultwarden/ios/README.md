# AppAgentVault - iOS App

Native iOS password manager app built with SwiftUI, following the AgentApp pattern for CloudEvents-based communication.

## Features

- ✅ **AgentApp Pattern**: Uses CloudEvents protocol for communication
- ✅ **Chat Interface**: Natural language interaction with password manager
- ✅ **Traditional Vault View**: List and manage passwords directly
- ✅ **Biometric Authentication**: Face ID / Touch ID support
- ✅ **Self-Hosted**: Connects to your homelab Vaultwarden instance
- ✅ **Secure Storage**: Keychain integration for tokens

## Architecture

Following the AgentApp pattern:

```
AppAgentVault/
├── Models/
│   ├── Agent.swift          # Vault agent configuration
│   ├── CloudEvent.swift     # CloudEvents protocol models
│   └── User.swift           # User model
├── Services/
│   ├── AgentService.swift   # CloudEvents communication
│   ├── AuthService.swift    # Authentication
│   └── KeychainService.swift # Secure storage
├── ViewModels/
│   ├── AppViewModel.swift   # App-wide state
│   └── ChatViewModel.swift  # Chat state
└── Views/
    ├── AppAgentVaultApp.swift # App entry point
    ├── LoginView.swift      # Authentication
    ├── HomeView.swift       # Tab navigation
    ├── ChatView.swift       # AI chat interface
    ├── VaultView.swift      # Password list
    ├── CipherDetailView.swift # Password details
    ├── OnboardingView.swift # First launch
    └── SettingsView.swift   # App settings
```

## Requirements

- iOS 16.0+
- Xcode 15.0+
- Swift 5.9+
- VPN connection to homelab (or direct access)

## Setup

### 1. Open in Xcode

```bash
cd ios/AppAgentVault
# Create new Xcode project if needed, or open existing
```

### 2. Configure Agent URL

Update the agent base URL in `Models/Agent.swift`:

```swift
static let vault = Agent(
    name: "Password Manager",
    description: "Self-hosted password manager agent",
    baseURL: "https://vaultwarden.lucena.cloud", // Your server URL
    ...
)
```

### 3. Build and Run

1. Open project in Xcode
2. Select your device or simulator
3. Press `Cmd + R` to build and run

## Usage

### Chat Interface

Use natural language to interact with your password manager:

- "Save a password for github.com with username john@example.com"
- "Show me all my passwords"
- "What's the password for github.com?"
- "Generate a strong password"

### Vault View

Traditional password manager interface:
- Browse all saved passwords
- View password details
- Copy passwords to clipboard
- Edit and delete passwords

## CloudEvents Integration

The app sends CloudEvents to the backend:

```swift
// Example: Query passwords
let event = CloudEvent(
    specversion: "1.0",
    type: "io.homelab.vault.query",
    source: "/ios-app/vault",
    id: UUID().uuidString,
    time: ISO8601DateFormatter().string(from: Date()),
    data: CloudEventData(query: "show me my passwords")
)
```

## Backend Requirements

The backend should support:

1. **CloudEvents endpoints**:
   - `POST /api/vault/chat` - Natural language queries
   - `POST /api/vault/save` - Save password
   - `GET /api/vault/list` - List passwords

2. **REST API endpoints** (fallback):
   - `POST /api/identity/connect/token` - Authentication
   - `GET /api/ciphers` - List passwords
   - `GET /api/ciphers/:id` - Get password
   - `POST /api/ciphers` - Create password
   - `PUT /api/ciphers/:id` - Update password
   - `DELETE /api/ciphers/:id` - Delete password

## Security

- ✅ Client-side encryption (before sending to server)
- ✅ Keychain storage for tokens
- ✅ Biometric authentication
- ✅ HTTPS/TLS for all communication
- ✅ Token-based authentication

## Development

### Add New Features

1. **New Chat Commands**: Extend `AgentService.sendMessage()` to handle new query types
2. **New Views**: Add SwiftUI views in `Views/` directory
3. **New Models**: Add data models in `Models/` directory

### Testing

```bash
# Run tests (when tests are added)
xcodebuild test -scheme AppAgentVault -destination 'platform=iOS Simulator,name=iPhone 15'
```

## Notes

- The app uses both CloudEvents (for chat) and REST API (for traditional operations)
- Authentication uses REST API (standard OAuth2 token flow)
- CloudEvents are used for natural language queries
- Passwords are encrypted client-side before sending to server
