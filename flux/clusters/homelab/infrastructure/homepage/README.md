# 🏠 Homepage - Bruno Site

Production-ready dynamic homepage with AI-powered chatbot integration, built with modern cloud-native technologies.

## 🚀 Quick Start

```bash
# Local development
docker-compose up -d

# Production deployment
helm upgrade --install bruno-site ./chart \
  --namespace homepage \
  --values chart/values.yaml

# Run tests
./tests/run-all-tests.sh
```

## 📋 Overview

This homepage system features:
- **Dynamic Content Management** - Real-time updates via API
- **AI Chatbot** - Agent-SRE integration with fallback
- **Production Infrastructure** - Kubernetes with auto-scaling
- **CI/CD Pipeline** - GitHub Actions with comprehensive testing

## 🏗️ Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed system design.

```
Frontend (React) → API (Go) → Database (PostgreSQL)
     ↓                ↓             ↓
  Chatbot      Agent-SRE       Redis Cache
```

## 🤖 AI Chatbot

See [CHATBOT.md](./CHATBOT.md) for complete chatbot integration details.

- **Agent-SRE Integration** - Connect to SRE agent service
- **Dual Modes** - MCP and Direct communication
- **Automatic Fallback** - Resilient error handling
- **50+ Tests** - 100% coverage

## 🔐 Security

See [SECURITY.md](./SECURITY.md) for security details.

- **Proxy Pattern** - No direct service exposure
- **Automated Scanning** - Trivy, govulncheck, npm audit
- **Sealed Secrets** - Kubernetes secrets management
- **HTTPS/TLS** - Cloudflare integration

## 🛠️ Technology Stack

**Backend:** Go 1.23, Gin, GORM, Redis  
**Frontend:** React, TypeScript, Vite  
**Database:** PostgreSQL 15  
**AI:** Agent-SRE, Ollama, MCP Protocol  
**Infrastructure:** Kubernetes, Helm, Docker  
**CI/CD:** GitHub Actions

## 📊 Testing

```bash
# Backend tests
cd api && go test -v ./...

# Frontend tests  
cd frontend && npm test

# Integration tests
cd tests/integration && ./test-agent-sre-integration.sh
```

**Test Coverage:** 100% (50+ tests)

## 📚 Documentation

- [ARCHITECTURE.md](./ARCHITECTURE.md) - System architecture and design
- [CHATBOT.md](./CHATBOT.md) - AI chatbot integration guide
- [SECURITY.md](./SECURITY.md) - Security implementation
- [tests/TEST_README.md](./tests/TEST_README.md) - Testing guide
- [API Documentation](./api/README.md) - API reference
- [Frontend Guide](./frontend/README.md) - Frontend development

## 📈 Status

| Component | Status | Tests | Coverage |
|-----------|--------|-------|----------|
| Backend API | ✅ Running | 10/10 | 100% |
| Frontend | ✅ Running | 25/25 | 100% |
| Chatbot Integration | ✅ Working | 15/15 | 100% |
| GitHub Actions | ✅ Active | 3 workflows | All passing |

## 🤝 Contributing

1. Create feature branch
2. Make changes
3. Run tests locally
4. Submit PR (CI/CD runs automatically)
5. Merge when all checks pass

## 📞 Contact

- **GitHub:** [brunovlucena](https://github.com/brunovlucena)
- **LinkedIn:** [Bruno Lucena](https://www.linkedin.com/in/bvlucena)

---

**Version:** 1.0.0  
**Status:** ✅ Production Ready  
**Last Updated:** 2025-10-08
