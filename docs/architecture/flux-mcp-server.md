# Flux MCP Server - AI-Assisted GitOps

This document describes how to set up and use the Flux Operator MCP Server for AI-assisted GitOps operations in the homelab.

## Overview

The [Flux Operator MCP Server](https://fluxcd.io/blog/2025/05/ai-assisted-gitops/) bridges AI assistants (Claude, Cursor, Windsurf, GitHub Copilot) with your Kubernetes clusters, enabling:

- **Natural language GitOps**: Ask questions and perform operations conversationally
- **End-to-end debugging**: Trace issues from Flux resources to application logs
- **Cross-cluster comparisons**: Compare configurations between environments
- **Pipeline visualization**: Generate diagrams of Flux dependencies
- **Documentation search**: Get accurate, up-to-date Flux guidance

## Installation

### macOS (Homebrew)

```bash
brew install controlplaneio-fluxcd/tap/flux-operator-mcp
```

### Manual Installation

Download pre-built binaries from the [Flux Operator releases](https://github.com/controlplaneio-fluxcd/flux-operator/releases).

### Verify Installation

```bash
flux-operator-mcp --version
```

## Configuration

### Cursor IDE

The MCP server is already configured in `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "flux-operator-mcp": {
      "command": "flux-operator-mcp",
      "args": ["serve"],
      "env": {
        "KUBECONFIG": "${HOME}/.kube/config"
      }
    }
  }
}
```

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "flux-operator-mcp": {
      "command": "flux-operator-mcp",
      "args": ["serve"],
      "env": {
        "KUBECONFIG": "/path/to/.kube/config"
      }
    }
  }
}
```

### Multi-Cluster Setup

For multiple clusters, the MCP server uses your kubeconfig contexts. Switch contexts before operations:

```bash
# List available contexts
kubectl config get-contexts

# Switch to pro cluster
kubectl config use-context kind-pro

# Switch to studio cluster
kubectl config use-context kind-studio
```

## Usage Examples

### 1. Health Assessment

> "Analyze the Flux installation in my current cluster and report the status of all components and ResourceSets."

### 2. Pipeline Visualization

> "List the Flux Kustomizations and draw a Mermaid diagram for the depends-on relationship."

### 3. Cross-Cluster Comparison

> "Compare the homepage HelmRelease between pro and studio clusters."

### 4. Root Cause Analysis

> "Perform a root cause analysis of the last failed Helm release in the homepage namespace."

### 5. GitOps Operations

> "Resume all suspended Flux resources in the current cluster and verify their status."

### 6. Kubernetes Operations

> "Create a namespace called test-app, then copy the redis HelmRelease and its source there."

## Homelab-Specific Context

When working with this homelab, the AI assistant should know:

### Clusters
| Cluster | Type | Purpose |
|---------|------|---------|
| pro | kind | Production-like environment |
| studio | kind | Development environment |

### Key Namespaces
- `flux-system`: Flux controllers and sources
- `homepage`: Portfolio website (Go API + React frontend)
- `knative-lambda`: Serverless functions operator
- `prometheus`: Monitoring stack
- `loki`: Log aggregation
- `tempo`: Distributed tracing

### Flux Structure
```
flux/
├── clusters/           # Cluster-specific configs
│   ├── pro/    # Production cluster
│   └── studio/ # Development cluster
├── infrastructure/     # Shared infrastructure components
│   ├── homepage/
│   ├── knative-lambda-operator/
│   ├── prometheus-operator/
│   └── ...
└── ai/                # AI-related deployments
```

## Security Considerations

The MCP server operates with your kubeconfig permissions:

- **Read-only mode**: Use `--read-only` flag for observation without mutations
- **Service account impersonation**: Limit access with `--as=<service-account>`
- **Secret masking**: Sensitive values in Secrets are automatically masked

### Read-Only Mode

For safer exploration:

```json
{
  "mcpServers": {
    "flux-operator-mcp": {
      "command": "flux-operator-mcp",
      "args": ["serve", "--read-only"],
      "env": {
        "KUBECONFIG": "${HOME}/.kube/config"
      }
    }
  }
}
```

## Troubleshooting

### MCP Server Not Connecting

1. Verify installation: `which flux-operator-mcp`
2. Check kubeconfig: `kubectl config current-context`
3. Test connectivity: `kubectl get ns`
4. Restart Cursor IDE

### Cluster Not Found

Ensure your kubeconfig has the correct context:

```bash
kubectl config get-contexts
kubectl config use-context <your-context>
```

### Permission Errors

The MCP server uses your kubeconfig credentials. Ensure your user/service account has appropriate RBAC permissions.

## References

- [Flux MCP Server Blog Post](https://fluxcd.io/blog/2025/05/ai-assisted-gitops/)
- [Flux Operator Documentation](https://fluxcd.io/flux/operator/)
- [Model Context Protocol (MCP)](https://modelcontextprotocol.io/)
- [Flux Installation Guide](https://fluxcd.io/flux/installation/)


