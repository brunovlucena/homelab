# 🎧 DJ Collab P2P Game

Multiplayer DJ collaboration game with P2P streaming, real-time synchronization, and gamification.

**Developer:** Bruno Lucena (bruno@lucena.cloud)

## Features

- P2P music streaming via WebRTC (no central server needed)
- Real-time collaboration between multiple DJs
- BPM and key synchronization
- Music analysis (BPM, key, waveform) - local processing
- Gamification system (points, achievements, leaderboards)
- Voice chat integration
- Set recording and sharing via IPFS
- Cross-platform: iOS, Android, Web, and Desktop

## Stack

| Layer | Technology |
|-------|------------|
| Desktop | Electron + React + TypeScript |
| Mobile | React Native + TypeScript |
| Web | Next.js + React + TypeScript |
| Backend | Go + Gin (WebSocket coordination) |
| P2P Streaming | WebRTC |
| Discovery | IPFS + DHT |
| Database | MongoDB |
| Cache | Redis |
| Music Analysis | librosa (Python) + WebAssembly |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              DJ Collab P2P Game Architecture                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌─────────────┐ │
│  │  📱 Mobile    │    │  💻 Desktop   │    │   🌐 Web     │ │
│  │  (React Nav) │    │  (Electron)   │    │  (Next.js)   │ │
│  │              │    │              │    │             │ │
│  │ • DJ Client  │    │ • DJ Client  │    │ • DJ Client │ │
│  │ • P2P Stream │    │ • P2P Stream │    │ • P2P Stream│ │
│  │ • Game UI    │    │ • Game UI    │    │ • Game UI    │ │
│  └──────┬───────┘    └──────┬───────┘    └──────┬────────┘ │
│         │                   │                   │          │
│         └───────────────────┼───────────────────┘          │
│                             │                              │
│                    ┌─────────▼─────────┐                  │
│                    │  🎧 P2P Network   │                  │
│                    │  (WebRTC + IPFS)  │                  │
│                    │                   │                  │
│                    │ • Streaming       │                  │
│                    │ • Discovery       │                  │
│                    │ • Sync            │                  │
│                    └─────────┬─────────┘                  │
│                               │                             │
│                    ┌──────────▼──────────┐                 │
│                    │  🌐 Coordination    │                 │
│                    │  (Go + WebSocket)  │                 │
│                    │                    │                 │
│                    │ • Session Mgmt    │                 │
│                    │ • State Sync      │                 │
│                    │ • Signaling       │                 │
│                    └──────────┬──────────┘                 │
│                               │                             │
│         ┌─────────────────────┼─────────────────────┐        │
│         │                     │                     │        │
│  ┌──────▼──────┐    ┌─────────▼─────────┐  ┌─────▼─────┐ │
│  │  🗄️ MongoDB │    │   🔴 Redis Cache  │  │  🎵 IPFS  │ │
│  │             │    │                   │  │           │ │
│  │ • Users     │    │ • Sessions        │  │ • Metadata│ │
│  │ • Sessions  │    │ • State Cache     │  │ • Sets    │ │
│  │ • Scores    │    │ • Real-time       │  │ • Content │ │
│  └─────────────┘    └───────────────────┘  └───────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

```bash
# Install dependencies
pnpm install

# Start infrastructure (MongoDB, Redis, IPFS)
pnpm docker:up

# Start signaling server (STUN/TURN)
pnpm dev:signaling

# Start coordination server
pnpm dev:server

# Run desktop app
pnpm dev:desktop

# Run web app
pnpm dev:web

# Run mobile (iOS)
pnpm dev:mobile:ios
```

## Development

### Backend (Go)

```bash
cd apps/server
go run main.go
```

### Desktop (Electron)

```bash
cd apps/desktop
pnpm dev
```

### Web (Next.js)

```bash
cd apps/web
pnpm dev
```

### Mobile (React Native)

```bash
cd apps/mobile
pnpm ios
# or
pnpm android
```

## Testing

```bash
pnpm test           # All tests
pnpm test:server    # Backend tests
pnpm test:desktop   # Desktop tests
pnpm test:mobile    # Mobile tests
pnpm test:p2p       # P2P integration tests
```

## Deployment

### Homelab (Flux)

```bash
# Apply Flux configuration
kubectl apply -k flux/dj-collab-p2p/

# Or use Flux CLI
flux reconcile source git dj-collab-p2p
```

### Manual Kubernetes

```bash
# Apply manifests
kubectl apply -f k8s/
```

## Configuration

### Environment Variables

```bash
# Signaling Server
SIGNALING_HOST=localhost
SIGNALING_PORT=3478
STUN_SERVERS=stun:stun.l.google.com:19302

# Coordination Server
COORDINATION_HOST=localhost
COORDINATION_PORT=8080
MONGODB_URI=mongodb://localhost:27017
REDIS_URI=redis://localhost:6379

# IPFS
IPFS_API_URL=http://localhost:5001
IPFS_GATEWAY_URL=http://localhost:8080

# P2P
P2P_ENABLED=true
P2P_ICE_SERVERS=stun:stun.l.google.com:19302
```

## Features in Detail

### P2P Streaming
- Direct peer-to-peer audio streaming via WebRTC
- No central server required for audio transmission
- Adaptive bitrate based on connection quality
- Multiple codec support (Opus, AAC)

### Real-time Collaboration
- Synchronized BPM and key detection
- Shared state management
- Low-latency updates via WebSocket
- Conflict resolution for simultaneous changes

### Music Analysis
- Local BPM detection (librosa)
- Key detection (chroma analysis)
- Waveform generation
- Energy and genre classification
- Results cached and shared via IPFS

### Gamification
- Points for smooth transitions
- Achievements for milestones
- Leaderboards (global and friends)
- Daily challenges
- Streak tracking

## License

MIT - Bruno Lucena (bruno@lucena.cloud)
