# SRE Agent Refactored - Makefile
# Separate MCP Server and Agent deployments

.PHONY: help build-mcp build-agent build-all deploy-mcp deploy-agent deploy-all start stop clean test health logs

# Default target
help:
	@echo "🚀 SRE Agent Refactored - Available Commands:"
	@echo ""
	@echo "  📦 Build Commands:"
	@echo "    make build-mcp          - Build MCP server Docker image"
	@echo "    make build-agent         - Build Agent Docker image"
	@echo "    make build-all           - Build both images"
	@echo ""
	@echo "  🚀 Deploy Commands:"
	@echo "    make deploy-mcp          - Deploy MCP server to Kubernetes"
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

# Build MCP Server
build-mcp:
	@echo "🔨 Building MCP Server Docker image..."
	docker build -f deployments/mcp-server/Dockerfile -t ghcr.io/brunovlucena/agent-sre-mcp-server:latest .
	@echo "✅ MCP Server image built successfully"

# Build Agent
build-agent:
	@echo "🔨 Building Agent Docker image..."
	docker build -f deployments/agent/Dockerfile -t ghcr.io/brunovlucena/agent-sre-agent:latest .
	@echo "✅ Agent image built successfully"

# Build all images
build-all: build-mcp build-agent
	@echo "✅ All images built successfully"

# Deploy MCP Server to Kubernetes
deploy-mcp:
	@echo "🚀 Deploying MCP Server to Kubernetes..."
	kubectl apply -f deployments/mcp-server/k8s-mcp-server.yaml
	@echo "✅ MCP Server deployed successfully"

# Deploy Agent to Kubernetes
deploy-agent:
	@echo "🚀 Deploying Agent to Kubernetes..."
	kubectl apply -f deployments/agent/k8s-agent.yaml
	@echo "✅ Agent deployed successfully"

# Deploy all services
deploy-all: deploy-mcp deploy-agent
	@echo "✅ All services deployed successfully"

# Start services with docker-compose
start:
	@echo "🐳 Starting services with docker-compose..."
	docker-compose up -d
	@echo "✅ Services started successfully"
	@echo "🌐 MCP Server: http://localhost:30120"
	@echo "🌐 Agent: http://localhost:8080"
	@echo "🌐 Nginx: http://localhost:80"

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

# Test MCP communication
test-mcp:
	@echo "🧪 Testing MCP communication..."
	@echo "Testing MCP chat..."
	curl -X POST http://localhost:8080/mcp/chat \
		-H "Content-Type: application/json" \
		-d '{"message": "How do I monitor Kubernetes pods?"}' | jq .
	@echo ""
	@echo "Testing MCP log analysis..."
	curl -X POST http://localhost:8080/mcp/analyze-logs \
		-H "Content-Type: application/json" \
		-d '{"logs": "ERROR: Database connection failed"}' | jq .

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

# Push images to registry
push-mcp:
	@echo "📤 Pushing MCP Server image to registry..."
	docker push ghcr.io/brunovlucena/agent-sre-mcp-server:latest
	@echo "✅ MCP Server image pushed successfully"

push-agent:
	@echo "📤 Pushing Agent image to registry..."
	docker push ghcr.io/brunovlucena/agent-sre-agent:latest
	@echo "✅ Agent image pushed successfully"

push-all: push-mcp push-agent
	@echo "✅ All images pushed successfully"

# Scale services
scale-mcp:
	@echo "📈 Scaling MCP Server..."
	kubectl scale deployment sre-agent-mcp-server --replicas=3 -n agent-sre
	@echo "✅ MCP Server scaled to 3 replicas"

scale-agent:
	@echo "📈 Scaling Agent..."
	kubectl scale deployment sre-agent --replicas=3 -n agent-sre
	@echo "✅ Agent scaled to 3 replicas"

# Get service URLs
urls:
	@echo "🌐 Service URLs:"
	@echo "MCP Server: http://localhost:30120"
	@echo "Agent: http://localhost:8080"
	@echo "Nginx: http://localhost:80"
	@echo ""
	@echo "Kubernetes NodePort URLs:"
	@echo "MCP Server: http://localhost:31120"
	@echo "Agent: http://localhost:31080"

# Monitor services
monitor:
	@echo "📊 Monitoring services..."
	watch -n 5 'curl -s http://localhost:30120/health | jq .status && curl -s http://localhost:8080/health | jq .status'
