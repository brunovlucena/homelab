.PHONY: help

ARGS = $(filter-out $(firstword $(MAKECMDGOALS)), $(MAKECMDGOALS))

# Tool Variables
MINIKUBE_VERSION = v1.5.2

help: ## Help. 
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

pre-install: ## Pre-Installs tools (E.g: $ make pre-install).
	@./helper.sh pre-install minikube ${MINIKUBE_VERSION}
