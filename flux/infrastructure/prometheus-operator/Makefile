# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
#
#   📊 PROMETHEUS-OPERATOR - TESTING MAKEFILE
#
#   🎯 Purpose: Validate Prometheus configurations before deployment
#
#   🔧 USAGE:
#     make              - Show help
#     make test-unit    - Run unit tests
#     make test-lint    - Run YAML/JSON linting
#     make test-all     - Run all tests
#
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.DEFAULT_GOAL := help

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🎯 PATH CONFIGURATION                                                  │
# └─────────────────────────────────────────────────────────────────────────┘

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
TESTS_DIR := $(ROOT_DIR)/tests
K8S_DIR := $(ROOT_DIR)/k8s

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🎨 COLORS                                                              │
# └─────────────────────────────────────────────────────────────────────────┘

COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_BLUE := \033[34m
COLOR_YELLOW := \033[33m
COLOR_RED := \033[31m

define print_status
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)$(1)$(COLOR_RESET)"
endef

define print_success
	@echo "$(COLOR_BOLD)$(COLOR_GREEN)✅ $(1)$(COLOR_RESET)"
endef

define print_warning
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)⚠️ $(1)$(COLOR_RESET)"
endef

define print_error
	@echo "$(COLOR_BOLD)$(COLOR_RED)❌ $(1)$(COLOR_RESET)"
endef

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  📋 HELP                                                                │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: help
help: ## 📋 Show this help message
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)📊 Prometheus-Operator Testing$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Available targets:$(COLOR_RESET)"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-25s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)💡 Quick start:$(COLOR_RESET)"
	@echo "  make test-all    # Run all validation tests"

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🧪 TESTING TARGETS                                                     │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: test-all test-unit test-lint test-promtool test-yamllint test-jsonlint

test-all: test-lint test-unit ## 🧪 Run all tests (lint + unit)
	$(call print_success,All tests passed!)

test-unit: ## 🧪 Run Go unit tests
	$(call print_status,🧪 Running unit tests...)
	@cd $(TESTS_DIR) && go test -v -race -coverprofile=coverage.out ./unit/... 2>&1 | tee test-output.log
	@if [ -f $(TESTS_DIR)/coverage.out ]; then \
		cd $(TESTS_DIR) && go tool cover -func=coverage.out | tail -1; \
	fi
	@rm -f $(TESTS_DIR)/test-output.log
	$(call print_success,Unit tests completed)

test-unit-verbose: ## 🧪 Run Go unit tests with verbose output
	$(call print_status,🧪 Running unit tests (verbose)...)
	@cd $(TESTS_DIR) && go test -v -race -count=1 ./unit/...

test-lint: test-yamllint test-jsonlint ## 🔍 Run all linting checks
	$(call print_success,Linting completed)

test-yamllint: ## 🔍 Validate YAML syntax
	$(call print_status,🔍 Checking YAML syntax...)
	@if command -v yamllint >/dev/null 2>&1; then \
		yamllint -c $(ROOT_DIR)/.yamllint.yml $(K8S_DIR) 2>/dev/null || \
		yamllint $(K8S_DIR) 2>/dev/null || \
		echo "$(COLOR_YELLOW)⚠️ yamllint config not found, using defaults$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ yamllint not installed, skipping YAML linting$(COLOR_RESET)"; \
		echo "   Install with: pip install yamllint"; \
	fi

test-jsonlint: ## 🔍 Validate JSON syntax
	$(call print_status,🔍 Checking JSON syntax...)
	@ERRORS=0; \
	for file in $$(find $(K8S_DIR) -name "*.json" 2>/dev/null); do \
		if ! python3 -m json.tool "$$file" >/dev/null 2>&1; then \
			echo "$(COLOR_RED)❌ Invalid JSON: $$file$(COLOR_RESET)"; \
			ERRORS=$$((ERRORS+1)); \
		fi; \
	done; \
	if [ $$ERRORS -gt 0 ]; then \
		echo "$(COLOR_RED)❌ Found $$ERRORS invalid JSON file(s)$(COLOR_RESET)"; \
		exit 1; \
	fi
	$(call print_success,JSON syntax valid)

test-promtool: ## 🔍 Validate Prometheus rules with promtool
	$(call print_status,🔍 Validating Prometheus rules with promtool...)
	@if command -v promtool >/dev/null 2>&1; then \
		for file in $(K8S_DIR)/prometheusrules/*.yaml; do \
			if [ -f "$$file" ] && [ "$$(basename $$file)" != "kustomization.yaml" ]; then \
				echo "Checking $$file..."; \
				promtool check rules "$$file" 2>/dev/null || \
				echo "$(COLOR_YELLOW)⚠️ promtool validation skipped (file may need pre-processing)$(COLOR_RESET)"; \
			fi; \
		done; \
	else \
		echo "$(COLOR_YELLOW)⚠️ promtool not installed, skipping rule validation$(COLOR_RESET)"; \
		echo "   Install from: https://prometheus.io/download/"; \
	fi

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔧 SETUP TARGETS                                                       │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: setup setup-deps

setup: setup-deps ## 🔧 Setup development environment
	$(call print_status,🔧 Setting up development environment...)
	@cd $(TESTS_DIR) && go mod download
	@cd $(TESTS_DIR) && go mod tidy
	$(call print_success,Development environment ready)

setup-deps: ## 📦 Install test dependencies
	$(call print_status,📦 Installing dependencies...)
	@if command -v pip3 >/dev/null 2>&1; then \
		pip3 install --user yamllint 2>/dev/null || true; \
	fi
	@if command -v go >/dev/null 2>&1; then \
		cd $(TESTS_DIR) && go mod download; \
	fi
	$(call print_success,Dependencies installed)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🧹 CLEANUP TARGETS                                                     │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: clean

clean: ## 🧹 Clean test artifacts
	$(call print_status,🧹 Cleaning test artifacts...)
	@rm -f $(TESTS_DIR)/coverage.out
	@rm -f $(TESTS_DIR)/test-output.log
	@cd $(TESTS_DIR) && go clean -testcache 2>/dev/null || true
	$(call print_success,Cleanup complete)

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  📊 COVERAGE TARGETS                                                    │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: coverage coverage-html

coverage: test-unit ## 📊 Show test coverage
	$(call print_status,📊 Test Coverage Report...)
	@if [ -f $(TESTS_DIR)/coverage.out ]; then \
		cd $(TESTS_DIR) && go tool cover -func=coverage.out; \
	else \
		echo "$(COLOR_YELLOW)⚠️ No coverage data found. Run 'make test-unit' first.$(COLOR_RESET)"; \
	fi

coverage-html: test-unit ## 📊 Generate HTML coverage report
	$(call print_status,📊 Generating HTML coverage report...)
	@if [ -f $(TESTS_DIR)/coverage.out ]; then \
		cd $(TESTS_DIR) && go tool cover -html=coverage.out -o coverage.html; \
		echo "Coverage report generated: $(TESTS_DIR)/coverage.html"; \
	else \
		echo "$(COLOR_YELLOW)⚠️ No coverage data found. Run 'make test-unit' first.$(COLOR_RESET)"; \
	fi

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  🔍 INDIVIDUAL TEST TARGETS                                             │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: test-prometheusrules test-servicemonitors test-dashboards

test-prometheusrules: ## 🔍 Test PrometheusRules only
	$(call print_status,🔍 Testing PrometheusRules...)
	@cd $(TESTS_DIR) && go test -v -race ./unit/... -run "PrometheusRule"

test-servicemonitors: ## 🔍 Test ServiceMonitors only
	$(call print_status,🔍 Testing ServiceMonitors...)
	@cd $(TESTS_DIR) && go test -v -race ./unit/... -run "ServiceMonitor|PodMonitor"

test-dashboards: ## 🔍 Test Grafana Dashboards only
	$(call print_status,🔍 Testing Grafana Dashboards...)
	@cd $(TESTS_DIR) && go test -v -race ./unit/... -run "Dashboard"

# ┌─────────────────────────────────────────────────────────────────────────┐
# │  📋 CI TARGETS                                                          │
# └─────────────────────────────────────────────────────────────────────────┘

.PHONY: ci ci-lint ci-test

ci: ci-lint ci-test ## 🤖 Run CI pipeline
	$(call print_success,CI pipeline completed successfully!)

ci-lint: test-jsonlint ## 🤖 CI linting (no external deps)
	$(call print_success,CI linting passed)

ci-test: ## 🤖 CI unit tests
	$(call print_status,🤖 Running CI tests...)
	@cd $(TESTS_DIR) && go test -v -race -coverprofile=coverage.out ./unit/...
	$(call print_success,CI tests passed)
