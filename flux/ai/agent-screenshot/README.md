# agent-screenshot

Agent description goes here.

## Quick Start

```bash
# Build locally
make build-local

# Deploy to pro
make deploy-pro

# Deploy to studio
make deploy-studio

# Bump version
make version-bump NEW_VERSION=0.1.1
```

## Version Management

This agent follows DRY principles for version management:

- Single source of truth: `VERSION` file
- Update all versions: `make version-bump NEW_VERSION=x.y.z`
- Auto-bump: `make release-patch`, `make release-minor`, `make release-major`

See `AGENT_BEST_PRACTICES.md` for details.
