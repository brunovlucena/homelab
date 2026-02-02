# 📱 AgentChat iOS Client

**Private WhatsApp for AI Agents - iOS Application**

## Overview

The AgentChat iOS app allows users to chat with their personal AI agent assistants. Each user gets a dedicated agent that can:
- Generate images and videos on their behalf
- Send voice messages in their cloned voice  
- Alert contacts when they're nearby
- Have natural conversations and help with tasks

## Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Swift 5.9+ |
| UI Framework | SwiftUI |
| Architecture | MVVM + Clean Architecture |
| Networking | URLSession + Combine |
| WebSocket | Starscream |
| Local Storage | SwiftData (iOS 17+) |
| Push Notifications | APNs |
| Location | CoreLocation |
| Audio/Video | AVFoundation |
| Keychain | KeychainAccess |

## Project Structure

```
AgentChat/
├── App/
│   ├── AgentChatApp.swift           # App entry point
│   ├── AppDelegate.swift            # Push notifications, background tasks
│   └── ContentView.swift            # Root navigation
│
├── Core/
│   ├── Network/
│   │   ├── APIClient.swift          # REST API client
│   │   ├── WebSocketManager.swift   # Real-time messaging
│   │   ├── CloudEventsClient.swift  # CloudEvents handling
│   │   └── Endpoints.swift          # API endpoints
│   │
│   ├── Storage/
│   │   ├── SwiftDataModels.swift    # Local persistence
│   │   ├── KeychainManager.swift    # Secure storage
│   │   └── CacheManager.swift       # Media caching
│   │
│   ├── Location/
│   │   ├── LocationManager.swift    # CoreLocation wrapper
│   │   └── GeofenceManager.swift    # Proximity detection
│   │
│   ├── Audio/
│   │   ├── AudioRecorder.swift      # Voice recording
│   │   ├── AudioPlayer.swift        # Playback
│   │   └── VoiceCloneManager.swift  # Voice sample management
│   │
│   └── Services/
│       ├── AuthService.swift        # Authentication
│       ├── ChatService.swift        # Message handling
│       ├── MediaService.swift       # Image/video uploads
│       └── NotificationService.swift # Push notifications
│
├── Features/
│   ├── Auth/
│   │   ├── Views/
│   │   │   ├── LoginView.swift
│   │   │   ├── RegisterView.swift
│   │   │   └── OnboardingView.swift
│   │   └── ViewModels/
│   │       └── AuthViewModel.swift
│   │
│   ├── Chat/
│   │   ├── Views/
│   │   │   ├── ChatListView.swift
│   │   │   ├── ChatDetailView.swift
│   │   │   ├── MessageBubble.swift
│   │   │   └── MessageInputBar.swift
│   │   ├── ViewModels/
│   │   │   ├── ChatListViewModel.swift
│   │   │   └── ChatDetailViewModel.swift
│   │   └── Models/
│   │       ├── Chat.swift
│   │       └── Message.swift
│   │
│   ├── Voice/
│   │   ├── Views/
│   │   │   ├── VoiceRecorderView.swift
│   │   │   ├── VoiceCloneSetupView.swift
│   │   │   └── VoiceMessageView.swift
│   │   └── ViewModels/
│   │       └── VoiceViewModel.swift
│   │
│   ├── Media/
│   │   ├── Views/
│   │   │   ├── ImageGeneratorView.swift
│   │   │   ├── MediaGalleryView.swift
│   │   │   └── ImagePreviewView.swift
│   │   └── ViewModels/
│   │       └── MediaViewModel.swift
│   │
│   ├── Location/
│   │   ├── Views/
│   │   │   ├── LocationSettingsView.swift
│   │   │   ├── NearbyContactsView.swift
│   │   │   └── ProximityAlertView.swift
│   │   └── ViewModels/
│   │       └── LocationViewModel.swift
│   │
│   ├── Contacts/
│   │   ├── Views/
│   │   │   ├── ContactListView.swift
│   │   │   └── ContactDetailView.swift
│   │   └── ViewModels/
│   │       └── ContactsViewModel.swift
│   │
│   └── Settings/
│       ├── Views/
│       │   ├── SettingsView.swift
│       │   ├── AgentSettingsView.swift
│       │   ├── PrivacySettingsView.swift
│       │   └── VoiceSettingsView.swift
│       └── ViewModels/
│           └── SettingsViewModel.swift
│
├── Components/
│   ├── AgentAvatar.swift
│   ├── TypingIndicator.swift
│   ├── LoadingView.swift
│   ├── ErrorView.swift
│   └── GradientButton.swift
│
├── Extensions/
│   ├── Date+Extensions.swift
│   ├── String+Extensions.swift
│   ├── View+Extensions.swift
│   └── Color+Extensions.swift
│
└── Resources/
    ├── Assets.xcassets
    ├── LaunchScreen.storyboard
    └── Info.plist
```

## Key Features

### 1. Chat Interface

```swift
struct ChatDetailView: View {
    @StateObject var viewModel: ChatDetailViewModel
    @State private var messageText = ""
    
    var body: some View {
        VStack(spacing: 0) {
            // Messages
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(spacing: 12) {
                        ForEach(viewModel.messages) { message in
                            MessageBubble(message: message)
                                .id(message.id)
                        }
                    }
                    .padding()
                }
                .onChange(of: viewModel.messages.count) { _ in
                    withAnimation {
                        proxy.scrollTo(viewModel.messages.last?.id)
                    }
                }
            }
            
            // Input Bar
            MessageInputBar(
                text: $messageText,
                onSend: { viewModel.sendMessage(messageText) },
                onVoice: { viewModel.startVoiceRecording() },
                onMedia: { viewModel.openMediaPicker() }
            )
        }
        .navigationTitle(viewModel.agent.displayName)
        .toolbar {
            ToolbarItem(placement: .navigationBarTrailing) {
                AgentStatusIndicator(status: viewModel.agent.status)
            }
        }
    }
}
```

### 2. Voice Recording & Cloning

```swift
class VoiceViewModel: ObservableObject {
    @Published var isRecording = false
    @Published var voiceCloneStatus: VoiceCloneStatus = .notStarted
    @Published var recordingDuration: TimeInterval = 0
    
    private let audioRecorder = AudioRecorder()
    private let apiClient: APIClient
    
    // Record voice sample for cloning (30s minimum)
    func startRecording() async {
        isRecording = true
        do {
            let audioURL = try await audioRecorder.startRecording()
            // Timer updates recordingDuration
        } catch {
            // Handle error
        }
    }
    
    func stopRecording() async throws {
        let audioData = try await audioRecorder.stopRecording()
        isRecording = false
        
        // Validate minimum duration
        guard recordingDuration >= 30 else {
            throw VoiceError.tooShort
        }
        
        // Upload for voice cloning
        voiceCloneStatus = .processing
        try await apiClient.uploadVoiceSample(audioData)
    }
    
    // Request agent to speak with cloned voice
    func sendVoiceMessage(text: String) async throws {
        let event = CloudEvent(
            type: "io.agentchat.voice.message.request",
            data: VoiceMessageRequest(text: text, useClonedVoice: true)
        )
        try await apiClient.sendCloudEvent(event)
    }
}
```

### 3. Location Tracking

```swift
class LocationViewModel: ObservableObject {
    @Published var isTrackingEnabled = false
    @Published var nearbyContacts: [NearbyContact] = []
    @Published var currentCity: String?
    
    private let locationManager = LocationManager()
    private let webSocket: WebSocketManager
    
    func enableTracking() {
        locationManager.requestAuthorization()
        locationManager.startUpdating { [weak self] location in
            self?.updateLocation(location)
        }
        isTrackingEnabled = true
    }
    
    private func updateLocation(_ location: CLLocation) {
        // Send to backend via WebSocket
        webSocket.send(CloudEvent(
            type: "io.agentchat.location.updated",
            data: LocationData(
                latitude: location.coordinate.latitude,
                longitude: location.coordinate.longitude,
                accuracy: location.horizontalAccuracy,
                timestamp: Date()
            )
        ))
        
        // Reverse geocode for city name
        Task {
            currentCity = await reverseGeocode(location)
        }
    }
    
    // Handle proximity alerts from server
    func handleProximityAlert(_ alert: ProximityAlert) {
        // Show notification
        NotificationService.shared.showLocalNotification(
            title: "Contact Nearby!",
            body: "\(alert.contactName) is in \(alert.city) (\(alert.distance)km away)"
        )
        nearbyContacts.append(NearbyContact(from: alert))
    }
}
```

### 4. Image Generation

```swift
class MediaViewModel: ObservableObject {
    @Published var isGenerating = false
    @Published var generatedImage: UIImage?
    @Published var generationProgress: String?
    
    private let apiClient: APIClient
    
    func generateImage(prompt: String) async throws {
        isGenerating = true
        generationProgress = "Analyzing prompt..."
        
        // Send request to agent
        let event = CloudEvent(
            type: "io.agentchat.media.image.request",
            data: ImageGenerationRequest(prompt: prompt)
        )
        
        let response = try await apiClient.sendCloudEvent(event)
        
        generationProgress = "Generating image..."
        
        // Wait for completion event
        for await imageEvent in apiClient.waitForEvent(type: "io.agentchat.media.image.generated") {
            if let url = imageEvent.data.imageUrl {
                generationProgress = "Downloading..."
                let imageData = try await apiClient.downloadMedia(url)
                generatedImage = UIImage(data: imageData)
            }
            break
        }
        
        isGenerating = false
        generationProgress = nil
    }
}
```

## CloudEvents Integration

The app communicates with the backend using CloudEvents over WebSocket:

```swift
struct CloudEvent<T: Codable>: Codable {
    let specversion: String = "1.0"
    let id: String
    let source: String
    let type: String
    let subject: String?
    let time: String
    let datacontenttype: String = "application/json"
    let data: T
    
    init(type: String, data: T, subject: String? = nil) {
        self.id = UUID().uuidString
        self.source = "/agentchat/ios/\(DeviceInfo.deviceId)"
        self.type = type
        self.subject = subject
        self.time = ISO8601DateFormatter().string(from: Date())
        self.data = data
    }
}

class WebSocketManager: ObservableObject {
    private var socket: WebSocket?
    private var eventSubject = PassthroughSubject<CloudEvent<Data>, Never>()
    
    var events: AnyPublisher<CloudEvent<Data>, Never> {
        eventSubject.eraseToAnyPublisher()
    }
    
    func connect(token: String) {
        var request = URLRequest(url: URL(string: "\(Config.wsEndpoint)/ws")!)
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        
        socket = WebSocket(request: request)
        socket?.delegate = self
        socket?.connect()
    }
    
    func send<T: Codable>(_ event: CloudEvent<T>) {
        let data = try! JSONEncoder().encode(event)
        socket?.write(data: data)
    }
}
```

## Security

| Feature | Implementation |
|---------|----------------|
| Authentication | JWT tokens stored in Keychain |
| Transport | TLS 1.3 for all connections |
| Local Storage | SwiftData with encryption |
| Voice Data | Encrypted before upload |
| Location | User consent required, opt-in |
| Biometric | Face ID/Touch ID for app unlock |

## Requirements

- iOS 17.0+
- Xcode 15.0+
- Swift 5.9+

## Build Instructions

```bash
# Clone repository
git clone https://github.com/yourusername/agentchat-ios.git

# Open Xcode project
open AgentChat.xcodeproj

# Configure signing
# 1. Select your team in Signing & Capabilities
# 2. Update bundle identifier

# Build and run
# Cmd + R
```

## Configuration

Create `Config.swift` with your environment:

```swift
struct Config {
    static let apiEndpoint = "https://api.agentchat.example.com"
    static let wsEndpoint = "wss://ws.agentchat.example.com"
    static let apnsTopic = "com.yourcompany.agentchat"
}
```

## Push Notifications Setup

1. Enable Push Notifications capability in Xcode
2. Create APNs key in Apple Developer Portal
3. Upload key to backend (configured in Kubernetes secrets)
4. Handle notifications in `AppDelegate`:

```swift
func application(_ application: UIApplication, 
                 didReceiveRemoteNotification userInfo: [AnyHashable: Any],
                 fetchCompletionHandler: @escaping (UIBackgroundFetchResult) -> Void) {
    // Handle proximity alerts, new messages, etc.
    NotificationService.shared.handle(userInfo)
    fetchCompletionHandler(.newData)
}
```

## License

MIT License - see LICENSE for details.
