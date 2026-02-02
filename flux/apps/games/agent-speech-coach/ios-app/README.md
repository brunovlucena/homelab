# 📱 Speech Coach iOS App

**Native iOS app for speech development coaching**

A native SwiftUI app that connects to the Speech Coach agent for helping autistic children develop speech skills through interactive games and exercises.

## 🎯 Features

- **🔌 Speech Coach Agent**: Connect to speech development agent
- **☁️ CloudEvents Protocol**: Full CloudEvents 1.0 specification support
- **📸 Face Recognition**: Track engagement using device camera
- **🎙️ Speech Recognition**: Use iOS Speech framework for speech exercises
- **🎨 Customizable Themes**: Child-friendly skins and colors
- **📊 Progress Tracking**: View speech development progress
- **💬 Modern Chat UI**: Beautiful, native iOS chat interface
- **🔐 Private & Secure**: All data stays on-device and in homelab

## 📋 Requirements

- iOS 17.0+
- iPhone with front-facing camera (for face recognition)
- VPN connection to homelab cluster (or direct access)
- Xcode 15.0+ (for development)

## 🚀 Quick Start

### Open in Xcode

```bash
cd ios-app/SpeechCoach
open SpeechCoach.xcodeproj
```

### Build & Run

1. Open `SpeechCoach.xcodeproj` in Xcode
2. Select your iPhone (device or simulator)
3. Press `Cmd + R` to build and run

### Configure Agent

The app is pre-configured to connect to the Speech Coach agent via the mobile-api gateway.

## 🏗️ Architecture

```
SpeechCoach/
├── Models/
│   ├── Agent.swift          # Speech Coach agent configuration
│   ├── Message.swift        # Chat message models
│   ├── Exercise.swift       # Exercise types and models
│   └── Progress.swift       # Progress tracking models
├── Services/
│   ├── AgentService.swift   # CloudEvents communication
│   ├── StorageService.swift # Local persistence
│   ├── FaceRecognitionService.swift # Face recognition
│   └── SpeechRecognitionService.swift # Speech recognition
├── ViewModels/
│   ├── ChatViewModel.swift  # Chat logic & state
│   └── AppViewModel.swift   # App-wide state management
├── Views/
│   ├── ChatView.swift       # Main chat interface
│   ├── ExerciseView.swift   # Exercise/game interface
│   ├── ProgressView.swift   # Progress tracking
│   └── ThemeView.swift      # Theme customization
└── Components/
    ├── FaceRecognitionView.swift
    └── SpeechInputView.swift
```

## 🔐 Privacy & Security

- Face recognition runs entirely on-device
- Speech recognition uses iOS native framework (on-device)
- No data leaves the homelab
- All conversations encrypted in transit
