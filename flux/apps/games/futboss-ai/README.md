# FutBoss AI

Multiplayer football management game with AI agents running locally on Ollama.

**Developer:** Bruno Lucena (bruno@lucena.cloud)

## Features

- Multiplayer real-time matches via WebSocket
- AI agents with unique personality attributes (runs on Ollama)
- Token-based economy (FutCoin)
- PIX and Bitcoin payments
- Cross-platform: iOS, Android, Web, and CLI

## Stack

| Layer | Technology |
|-------|------------|
| CLI | Go + Bubble Tea (k9s-style TUI) |
| Mobile | React Native + TypeScript |
| Web | React + Vite + TypeScript |
| Backend | Python + FastAPI |
| Database | MongoDB |
| Realtime | WebSockets |
| AI | Ollama (local) |

## Quick Start

```bash
# Install dependencies
pnpm install

# Start infrastructure
pnpm docker:up

# Run backend
pnpm dev:api

# Run web
pnpm dev:web

# Build CLI
pnpm build:cli
./apps/cli/futboss play
```

## Testing

```bash
pnpm test           # All tests
pnpm test:api       # Backend tests
pnpm test:cli       # CLI tests
pnpm test:game-engine  # Game engine tests
```

## License

MIT - Bruno Lucena (bruno@lucena.cloud)

