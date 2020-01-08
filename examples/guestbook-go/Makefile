.PHONY: skaffold build

help: ## Help. 
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Builds a binary for service.
	@./helper.sh build app

check: ## Checks app binary.
	@./helper.sh check app

debug: ## Runs dlv.
	@./helper.sh debug

debug-tests: ## Runs dlv test to debug Tests.
	@./helper.sh debug-tests

skaffold: ## Uses skaffold during the development (make scaffold).
	@./helper.sh skaffold

run: ## Runs app locally.
	@./helper.sh run app

test: ## Runs Tests (make test).
	@./helper.sh test

check-pod-security: ## outputs some security information
	@./helper.sh check-pod-security ${label} ${namespace}

sniff: ## Sniffs comunication
	@./helper.sh sniff ${label} ${namespace}
