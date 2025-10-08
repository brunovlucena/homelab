.PHONY: help apply delete logs status port-forward test config-backup config-restore config-view config-edit config-export

help:
	@echo "Docker MCP Gateway - Makefile Commands"
	@echo "======================================"
	@echo "apply          - Apply the manifests to the cluster"
	@echo "delete         - Delete the deployment from the cluster"
	@echo "logs           - Tail logs from the gateway"
	@echo "status         - Check deployment status"
	@echo "port-forward   - Forward port 8000 to localhost"
	@echo "test           - Test the gateway endpoint"
	@echo ""
	@echo "Configuration Management:"
	@echo "config-backup  - Backup current configuration"
	@echo "config-restore - Restore configuration from backup"
	@echo "config-view    - View configuration files"
	@echo "config-edit    - Edit configuration files"
	@echo "config-export  - Export all configurations"
	@echo ""
	@echo "For more config options: ./manage-config.sh"

apply:
	kubectl apply -k .

delete:
	kubectl delete -k .

logs:
	kubectl logs -n mcp-gateway -l app.kubernetes.io/name=mcp-gateway -f

status:
	@echo "=== Pods ==="
	kubectl get pods -n mcp-gateway
	@echo ""
	@echo "=== Services ==="
	kubectl get svc -n mcp-gateway
	@echo ""
	@echo "=== Endpoints ==="
	kubectl get endpoints -n mcp-gateway

port-forward:
	kubectl port-forward -n mcp-gateway svc/mcp-gateway 8000:8000

test:
	@echo "Testing gateway via NodePort (localhost:31140)..."
	@curl -f http://localhost:31140/health && echo "✓ NodePort test passed" || echo "✗ NodePort test failed"
	@echo ""
	@echo "Testing gateway via port-forward (localhost:8000)..."
	@curl -f http://localhost:8000/health && echo "✓ Port-forward test passed" || echo "✗ Port-forward test failed"

config-backup:
	./manage-config.sh backup

config-restore:
	@echo "Usage: ./manage-config.sh restore <backup-directory>"

config-view:
	@echo "Available files: docker-mcp.yaml, registry.yaml, config.yaml, tools.yaml"
	@echo "Usage: ./manage-config.sh view <filename>"

config-edit:
	@echo "Available files: docker-mcp.yaml, registry.yaml, config.yaml, tools.yaml"
	@echo "Usage: ./manage-config.sh edit <filename>"

config-export:
	./manage-config.sh export

