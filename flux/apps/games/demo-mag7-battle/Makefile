# MAG7 Battle Demo - Makefile
# 🐉 Defeat the dragon!

.PHONY: help deploy undeploy url logs open clean

NAMESPACE := demo-mag7-battle

help:
	@echo "🐉 MAG7 Battle - Knative Lambda Operator Demo"
	@echo ""
	@echo "Commands:"
	@echo "  make deploy   - Deploy demo to Kubernetes"
	@echo "  make undeploy - Remove demo from Kubernetes"
	@echo "  make url      - Get game URL"
	@echo "  make logs     - Watch game server logs"
	@echo "  make open     - Open game in browser (local)"
	@echo "  make clean    - Clean up resources"

deploy:
	@echo "🚀 Deploying MAG7 Battle..."
	kubectl apply -k k8s/kustomize/studio
	@echo ""
	@echo "✅ Deployed! Run 'make url' to get the game URL"

undeploy:
	@echo "🗑️ Removing MAG7 Battle..."
	kubectl delete -k k8s/kustomize/studio --ignore-not-found

url:
	@echo "🔗 Game URLs:"
	@kubectl get ksvc -n $(NAMESPACE) -o jsonpath='{range .items[*]}{.metadata.name}: {.status.url}{"\n"}{end}' 2>/dev/null || \
		echo "Waiting for services to be ready..."

logs:
	@echo "📋 Watching game server logs..."
	kubectl logs -n $(NAMESPACE) -l app.kubernetes.io/name=mag7-game-server -f --tail=100

open:
	@echo "🌐 Opening game in browser..."
	@if [ -f static/index.html ]; then \
		open static/index.html; \
	else \
		echo "Static file not found. Deploy to Kubernetes with 'make deploy'"; \
	fi

clean: undeploy
	@echo "🧹 Cleanup complete"

# Deploy with agents for full demo
deploy-full: deploy
	@echo ""
	@echo "🛡️ Deploying agent-blueteam..."
	kubectl apply -k ../agent-blueteam/k8s/kustomize/studio
	@echo ""
	@echo "🔴 Deploying agent-redteam..."
	kubectl apply -k ../agent-redteam/k8s/kustomize/studio
	@echo ""
	@echo "🎮 Full demo deployed!"
