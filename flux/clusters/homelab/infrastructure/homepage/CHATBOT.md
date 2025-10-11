# ⚠️ DEPRECATED - See JAMIE_INTEGRATION.md

> **Note:** This document has been superseded by [JAMIE_INTEGRATION.md](./JAMIE_INTEGRATION.md)
>
> The homepage chatbot architecture has evolved. The chatbot now connects to **Jamie** (AI-powered SRE assistant) instead of directly to agent-sre.

---

## 🤖 Current Architecture

For complete and up-to-date documentation, please see **[JAMIE_INTEGRATION.md](./JAMIE_INTEGRATION.md)**

### Quick Summary

```
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│   Homepage      │      │   Homepage      │      │   Jamie         │
│   Frontend      │─────▶│   API (Go)      │─────▶│   Slack Bot     │
│   (React)       │      │   /api/v1/jamie │      │   REST API      │
└─────────────────┘      └─────────────────┘      └─────────────────┘
                                                            │
                                                            ├─────▶ Ollama AI
                                                            │       (bruno-sre)
                                                            │
                                                            └─────▶ Agent-SRE
                                                                    MCP Tools
```

### Key Endpoints

All chatbot endpoints now use the `/api/v1/jamie/*` prefix:

- `POST /api/v1/jamie/chat` - Main chatbot endpoint
- `POST /api/v1/jamie/analyze-logs` - Log analysis
- `POST /api/v1/jamie/golden-signals` - Check service health
- `POST /api/v1/jamie/prometheus/query` - Execute PromQL queries
- `POST /api/v1/jamie/pod-logs` - Get Kubernetes pod logs
- `GET /api/v1/jamie/health` - Health check
- `GET /api/v1/jamie/ready` - Readiness check

### Migration Notes

**Old Architecture (Deprecated)**:
- Frontend → API (`/api/v1/agent-sre/*`) → Agent-SRE (direct)

**New Architecture (Current)**:
- Frontend → API (`/api/v1/jamie/*`) → Jamie → Agent-SRE MCP + Ollama

### Benefits of New Architecture

- 🧠 **Smarter Responses** - Jamie provides context-aware AI assistance
- 🔧 **Better Tooling** - Access to comprehensive MCP tools
- 📊 **Enhanced Monitoring** - Golden signals and infrastructure insights
- 🤖 **Unified Interface** - Single point of contact for all SRE tasks
- 🔒 **Improved Security** - Additional authentication layer

---

## 📚 Documentation Links

- **[JAMIE_INTEGRATION.md](./JAMIE_INTEGRATION.md)** - Complete Jamie integration guide (⭐ START HERE)
- **[Jamie README](../agent-jamie/README.md)** - Jamie service architecture
- **[Agent-SRE](../agent-sre/README.md)** - Backend SRE agent with MCP tools
- **[Homepage API](./api/README.md)** - Go API documentation
- **[Frontend](./frontend/README.md)** - React frontend documentation

---

**Status**: ⚠️ DEPRECATED (use JAMIE_INTEGRATION.md)  
**Last Updated**: 2025-10-10  
**Superseded By**: [JAMIE_INTEGRATION.md](./JAMIE_INTEGRATION.md)
