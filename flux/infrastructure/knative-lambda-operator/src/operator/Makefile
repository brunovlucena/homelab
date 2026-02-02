# Knative Lambda Operator Makefile - KISS
# Version is read from ../../../VERSION file (single source of truth)
VERSION ?= $(shell cat ../../VERSION 2>/dev/null || echo "v0.0.0")
IMG ?= localhost:5001/knative-lambda-operator:$(VERSION)

# Go settings
GOCMD ?= go
GOTEST ?= $(GOCMD) test
GOFLAGS ?= -v

.PHONY: build push deploy restart logs clean help test test-unit test-cover test-ci lint fmt vet

## ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 🧪 TESTING
## ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## test: Run all unit tests
test:
	$(GOTEST) $(GOFLAGS) ./...

## test-unit: Run unit tests only (fast, no integration)
test-unit:
	$(GOTEST) $(GOFLAGS) -short ./...

## test-cover: Run tests with coverage
test-cover:
	$(GOTEST) $(GOFLAGS) -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

## test-ci: Run tests for CI/CD pipeline (GitHub Actions)
test-ci:
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "✅ All tests passed"

## test-json: Run tests with JSON output (for CI parsing)
test-json:
	$(GOTEST) -v -json ./... | tee test-results.json

## ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 🔍 CODE QUALITY
## ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format code
fmt:
	$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	$(GOCMD) vet ./...

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test

## ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 🐳 BUILD & DEPLOY
## ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## release: Update all version references from VERSION file, build and deploy
release: update-version build apply
	@echo "✅ Released $(VERSION)"

## update-version: Update version in all manifests from VERSION file
update-version:
	@echo "📝 Updating version to $(VERSION) in all manifests..."
	@sed -i '' 's|localhost:5001/knative-lambda-operator:v[0-9]*\.[0-9]*\.[0-9]*|localhost:5001/knative-lambda-operator:$(VERSION)|g' ../../k8s/base/deployment.yaml
	@sed -i '' 's|value: "v[0-9]*\.[0-9]*\.[0-9]*"|value: "$(VERSION)"|g' ../../k8s/base/deployment.yaml
	@sed -i '' 's|newTag: v[0-9]*\.[0-9]*\.[0-9]*|newTag: $(VERSION)|g' ../../k8s/base/kustomization.yaml
	@echo "✅ Updated to $(VERSION)"

## build: Build and push image (no cache)
build:
	@echo "🐳 Building $(IMG)..."
	docker build --no-cache -t $(IMG) . && docker push $(IMG)

## apply: Apply kustomization to cluster
apply:
	@echo "🚀 Applying to cluster..."
	kubectl apply -k ../../k8s/base/
	kubectl rollout status deployment/knative-lambda-operator -n knative-lambda --timeout=90s

## deploy: Build, push, and restart (without version update)
deploy: build
	kubectl rollout restart deployment/knative-lambda-operator -n knative-lambda
	kubectl rollout status deployment/knative-lambda-operator -n knative-lambda --timeout=60s

## restart: Restart operator
restart:
	kubectl rollout restart deployment/knative-lambda-operator -n knative-lambda

## logs: Show logs
logs:
	kubectl logs -f deployment/knative-lambda-operator -n knative-lambda

## clean: Clean images and test artifacts
clean:
	-docker rmi $(IMG) 2>/dev/null || true
	-rm -f coverage.out coverage.html test-results.json

## ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
## 📖 HELP
## ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## help: Show help
help:
	@echo "VERSION=$(VERSION) IMG=$(IMG)"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
