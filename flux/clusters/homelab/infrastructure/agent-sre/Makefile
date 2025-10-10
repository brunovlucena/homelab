# SRE Agent Refactored - Makefile
# Separate MCP Server and Agent deployments

# Variables
REGISTRY := ghcr.io/brunovlucena

.PHONY: help build-agent build-all deploy-agent deploy-all start stop clean test health logs push-agent push-all scale-agent urls monitor

# Default target
help:
	@echo "🚀 SRE Agent Refactored - Available Commands:"
	@echo ""
	@echo "  📦 Build Commands:"
	@echo "    make build-agent         - Build Agent Docker image"
	@echo "    make build-all           - Build both images"
	@echo ""
	@echo "  🚀 Deploy Commands:"
	@echo "    make deploy-agent        - Deploy Agent to Kubernetes"
	@echo "    make deploy-all          - Deploy both services"
	@echo ""
	@echo "  🐳 Docker Commands:"
	@echo "    make start               - Start services with docker-compose"
	@echo "    make stop                - Stop services"
	@echo "    make logs                - Show logs"
	@echo ""
	@echo "  🧪 Test Commands:"
	@echo "    make test                - Run tests"
	@echo "    make health              - Check health of all services"
	@echo ""
	@echo "  🧹 Cleanup Commands:"
	@echo "    make clean               - Clean up resources"
	@echo ""

# Build Agent
build-agent:
	@echo "🔨 Building Agent Docker image..."
	docker build -f deployments/agent/Dockerfile -t $(REGISTRY)/agent-sre-agent:latest .
	@echo "✅ Agent image built successfully"

# Build all images
build-all: build-agent build-mcp-server
	@echo "✅ All images built successfully"

# Deploy Agent to Kubernetes
deploy-agent:
	@echo "🚀 Deploying Agent to Kubernetes..."
	kubectl apply -f deployments/agent/k8s-agent.yaml
	@echo "✅ Agent deployed successfully"

# Deploy all services
deploy-all: deploy-agent
	@echo "✅ All services deployed successfully"

# Start services with docker-compose
start:
	@echo "🐳 Starting services with docker-compose..."
	docker-compose up -d
	@echo "✅ Services started successfully"
	@echo "🌐 Agent: http://localhost:8080"
	
# Stop services
stop:
	@echo "🛑 Stopping services..."
	docker-compose down
	@echo "✅ Services stopped successfully"

# Show logs
logs:
	@echo "📋 Showing logs..."
	docker-compose logs -f

# Run tests
test:
	@echo "🧪 Running tests..."
	@echo "Testing MCP Server health..."
	curl -f http://localhost:30120/health || echo "❌ MCP Server health check failed"
	@echo "Testing Agent health..."
	curl -f http://localhost:8080/health || echo "❌ Agent health check failed"
	@echo "Testing MCP Server readiness..."
	curl -f http://localhost:30120/ready || echo "❌ MCP Server readiness check failed"
	@echo "Testing Agent readiness..."
	curl -f http://localhost:8080/ready || echo "❌ Agent readiness check failed"

# Check health of all services
health:
	@echo "🏥 Checking health of all services..."
	@echo ""
	@echo "📊 MCP Server Health:"
	curl -s http://localhost:30120/health | jq . || echo "❌ MCP Server not responding"
	@echo ""
	@echo "📊 Agent Health:"
	curl -s http://localhost:8080/health | jq . || echo "❌ Agent not responding"
	@echo ""
	@echo "📊 Agent Status:"
	curl -s http://localhost:8080/status | jq . || echo "❌ Agent status not available"
	@echo ""
	@echo "📊 MCP Server Status:"
	curl -s http://localhost:8080/mcp/status | jq . || echo "❌ MCP Server status not available"

# Clean up resources
clean:
	@echo "🧹 Cleaning up resources..."
	docker-compose down -v
	docker system prune -f
	@echo "✅ Cleanup completed"

# Test direct agent communication
test-agent:
	@echo "🧪 Testing direct agent communication..."
	@echo "Testing direct chat..."
	curl -X POST http://localhost:8080/chat \
		-H "Content-Type: application/json" \
		-d '{"message": "How do I monitor Kubernetes pods?"}' | jq .
	@echo ""
	@echo "Testing direct log analysis..."
	curl -X POST http://localhost:8080/analyze-logs \
		-H "Content-Type: application/json" \
		-d '{"logs": "ERROR: Database connection failed"}' | jq .

push-agent:
	@echo "📤 Pushing Agent image to registry..."
	docker push $(REGISTRY)/agent-sre-agent:latest
	@echo "✅ Agent image pushed successfully"

push-all: push-agent push-mcp-server
	@echo "✅ All images pushed successfully"

# ============================================================================
# MCP Server Targets
# ============================================================================

build-mcp-server:
	@echo "🔨 Building MCP server Docker image..."
	docker buildx build --platform linux/amd64,linux/arm64 -f deployments/mcp-server/Dockerfile -t ghcr.io/brunovlucena/agent-sre-mcp-server:latest .
	@echo "✅ MCP server image built successfully"

push-mcp-server:
	@echo "📤 Building and pushing MCP server Docker image..."
	docker buildx build --platform linux/amd64,linux/arm64 -f deployments/mcp-server/Dockerfile -t ghcr.io/brunovlucena/agent-sre-mcp-server:latest --push .
	@echo "✅ MCP server image pushed successfully"

deploy-mcp-server:
	@echo "🚀 Deploying MCP server to Kubernetes..."
	kubectl apply -k deployments/mcp-server/
	@echo "✅ MCP server deployed successfully"

restart-mcp-server:
	@echo "🔄 Restarting MCP server..."
	kubectl rollout restart deployment/agent-sre-mcp-server -n agent-sre
	@echo "✅ MCP server restarted successfully"

logs-mcp-server:
	@echo "📋 Viewing MCP server logs..."
	kubectl logs -f -l app=agent-sre-mcp-server -n agent-sre --tail=100

test-mcp-server:
	@echo "🧪 Testing MCP server..."
	@echo "Testing health endpoint..."
	kubectl run curl-test --image=curlimages/curl:latest --rm -i --restart=Never -- \
		curl -f http://agent-sre-mcp-server.agent-sre:3000/health
	@echo ""
	@echo "Testing tools list..."
	kubectl run curl-test --image=curlimages/curl:latest --rm -i --restart=Never -- \
		curl -f http://agent-sre-mcp-server.agent-sre:3000/tools
	@echo "✅ MCP server tests completed"