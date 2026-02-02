# 🎯 Speech Coach Agent - Autism Speech Development

A personal AI agent designed to help autistic children develop speech skills through interactive games and exercises.

## Overview

The Speech Coach Agent provides:
- **Speech Development Exercises**: Structured games that encourage verbal communication
- **Progress Tracking**: Monitor speech development milestones and improvements
- **Face Recognition**: Use device camera for engagement and feedback
- **Customizable Themes**: Child-friendly skins and personalization
- **Private & Secure**: All data stays on-device and in your homelab

## Features

- 🎮 Interactive speech games and exercises
- 📊 Progress monitoring and analytics
- 📸 Face recognition for engagement tracking
- 🎨 Customizable themes and skins
- 🔒 Private and secure (local processing preferred)
- 📱 Mobile-first design with AgentApp framework
- 🍓 Raspberry Pi web client support
- 🧠 SLM-powered for fast, on-device responses
- 🖥️ Connects to agent on studio cluster server

## Architecture

```
┌─────────────────────────────────────┐
│     📱 iOS App (AgentApp)           │
│  • Face Recognition (AVFoundation)  │
│  • Speech Recognition (Speech)      │
│  • UI with customizable themes      │
└──────────────┬──────────────────────┘
               │ CloudEvents
               ▼
┌─────────────────────────────────────┐
│   🌐 Mobile API (Router)            │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│   🤖 Speech Coach Agent             │
│  • Exercise management              │
│  • Progress tracking                │
│  • Game logic & suggestions         │
│  • SLM for natural interactions     │
└─────────────────────────────────────┘
```

## Requirements

- iOS 17.0+
- AgentApp framework
- Face recognition capabilities
- Speech recognition access

## Deployment

Deployed as a LambdaAgent (Knative service) for scale-to-zero capabilities.

See [Deployment Guide](k8s/kustomize/base/README.md) for details.
