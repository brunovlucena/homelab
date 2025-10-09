# Bruno Site Makefile
# This Makefile manages the complete Bruno site system using Docker Compose

# Environment configuration
ENV ?= dev
DOCKER_COMPOSE_FILE = docker-compose.yml
REGISTRY ?= ghcr.io/brunovlucena/homelab

.PHONY: help start stop restart build build-push logs clean status api-logs frontend-logs db-logs psql redis-cli api-shell frontend-shell frontend-dev migrate migrate-k8s test-api test-api-unit test-frontend-unit test-e2e test-load test-coverage update-deps format lint pf-api pf-redis pf-postgres tp-intercept tp-intercept-with-mounts tp-stop tp-connect tp-disconnect tp-status tp-list restart-fresh reconcile optimize-images

# Default target
help:
	@echo "🚀 Bruno Site Management"
	@echo "========================"
	@echo ""
	@echo "Environment: $(ENV)"
	@echo ""
	@echo "Available commands:"
	@echo "  make up                    - Start all services (dev environment - Docker Compose)"
	@echo "  make down                  - Stop all services"
	@echo "  make restart               - Restart all services"
	@echo "  make build-dev             - Build all Docker images (dev environment)"
	@echo "  make logs                  - Show logs from all services"
	@echo "  make logs-api              - Show API logs only"
	@echo "  make logs-frontend         - Show frontend logs only"
	@echo "  make logs-postgres         - Show database logs only"
	@echo "  make status-services       - Show service status"
	@echo "  make clean                 - Stop and remove all containers/volumes"
	@echo "  make psql                  - Connect to PostgreSQL database"
	@echo "  make migrate               - Run database migration"
	@echo "  make migrate-k8s           - Run database migration in Kubernetes"
	@echo "  make redis-cli             - Connect to Redis CLI"
	@echo "  make api-shell             - Open shell in API container"
	@echo "  make frontend-shell        - Open shell in frontend container"
	@echo "  make up-frontend-dev       - Run frontend in development mode (hot reload)"
	@echo "  make restart-fresh         - Restart with fresh database (clean + start)"
	@echo "  make pf-api                - Port forward API service for local testing (Kubernetes)"
	@echo "  make tp-intercept          - Intercept both API and frontend services (no volume mounts)"
	@echo "  make tp-intercept-mounts   - Intercept with volume mounts (requires sshfs)"
	@echo "  make tp-stop               - Stop all active intercepts"
	@echo "  make tp-disconnect         - Disconnect from Kubernetes cluster"
	@echo "  make tp-status             - Show Telepresence status"
	@echo "  make tp-list               - List active intercepts"
	@echo "  make reconcile             - Reconcile Flux HelmRelease for bruno-site"
	@echo "  make test-api-endpoints    - Test API endpoints"
	@echo "  make test-k6               - Run k6 load tests for golden signals (Kubernetes)"
	@echo "  make test-k6-local         - Run k6 load tests locally (development)"
	@echo "  make test                  - Run all tests (API, frontend, E2E)"
	@echo "  make format                - Format code"
	@echo "  make lint                  - Lint code"
	@echo "  make optimize-images       - Optimize images for web performance"
	@echo ""

# Start services (development)
up:
	@echo "🚀 Starting Bruno Site (Development)..."
	@echo "Environment: $(ENV)"
	@docker-compose -f $(DOCKER_COMPOSE_FILE) up --build -d
	@echo "⏳ Waiting for database to be ready..."
	@timeout 60 bash -c 'until docker exec postgres pg_isready -U postgres -d bruno_site; do sleep 2; done' || true
	@echo "🗄️ Running database migrations..."
	@make migrate || echo "⚠️ Migration failed, but continuing..."
	@echo "✅ Bruno site is running!"
	@echo ""
	@echo "📋 Access Information:"
	@echo "  Frontend: http://localhost:3000"
	@echo "  API Health: http://localhost:8080/health"
	@echo "  API Projects: http://localhost:8080/api/v1/projects"
	@echo "  Grafana: http://localhost:3002 (admin/admin)"
	@echo "  Prometheus: http://localhost:9090"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  Redis: localhost:6379"

# Stop services
down:
	@echo "🛑 Stopping Bruno Site..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down
	@echo "✅ Services stopped"

# Restart services
restart: down up

# Restart with fresh database (clean and start)
restart-fresh: clean up

# Build all images (development)
build-dev:
	@echo "🏗️ Building Docker images (Development)..."
	@echo "Environment: $(ENV)"
	@cp frontend/Dockerfile.dev frontend/Dockerfile
	@docker-compose -f $(DOCKER_COMPOSE_FILE) build
	@echo "✅ Images built successfully"

# Build and push images to registry (latest tag only)
build-push-latest:
	@echo "🏗️ Building and pushing Docker images..."
	@echo "Registry: $(REGISTRY)"
	@echo "Tag: latest"
	@echo "🏗️ Building images..."
	@docker build -t $(REGISTRY)/homepage-api:latest -f api/Dockerfile api/
	@docker build -t $(REGISTRY)/homepage-frontend:latest -f frontend/Dockerfile frontend/
	@echo "🚀 Pushing images to registry..."
	@docker push $(REGISTRY)/homepage-api:latest
	@docker push $(REGISTRY)/homepage-frontend:latest
	@echo "✅ Images built and pushed successfully!"
	@echo "📋 Pushed images:"
	@echo "  API: $(REGISTRY)/homepage-api:latest"
	@echo "  Frontend: $(REGISTRY)/homepage-frontend:latest"

# Show logs from all services
logs:
	@echo "📋 Following logs from all services (Ctrl+C to stop):"
	@docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

# Show API logs
logs-api:
	@echo "📋 API logs:"
	@docker logs -f bruno-api --tail=50

# Show frontend logs
logs-frontend:
	@echo "📋 Frontend logs:"
	@docker logs -f bruno-frontend --tail=50

# Show database logs
logs-postgres:
	@echo "📋 Database logs:"
	@docker logs -f bruno-postgres --tail=50

# Show service status
status-services:
	@echo "📊 Service Status:"
	@docker-compose -f $(DOCKER_COMPOSE_FILE) ps

# Clean everything (stop and remove containers, volumes, networks)
clean:
	@echo "🧹 Cleaning up Bruno Site..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down -v --remove-orphans
	@docker system prune -f
	@echo "✅ Cleanup completed"

# Connect to PostgreSQL
psql:
	@echo "🗄️ Connecting to PostgreSQL..."
	@docker exec -it bruno-postgres psql -U postgres -d bruno_site

# Run database migration in Kubernetes
migrate-k8s:
	@echo "🗄️ Running database migration in Kubernetes..."
	@echo "💡 This will run the migration against the Kubernetes PostgreSQL instance"
	@echo "⏳ Setting up port forwarding to PostgreSQL..."
	@kubectl port-forward --address 0.0.0.0 -n bruno svc/bruno-site-postgres 5432:5432 &
	@echo "⏳ Waiting for port forwarding to establish..."
	@sleep 3
	@echo "🗄️ Running migration SQL..."
	@PGPASSWORD=secure-password psql -h 0.0.0.0 -p 5432 -U postgres -d bruno_site -f api/migrations/001_complete_schema.sql
	@echo "✅ Migration completed successfully!"
	@echo "🛑 Stopping port forwarding..."
	@pkill -f "kubectl port-forward.*bruno-site-postgres" || true

# Connect to Redis CLI
redis-cli:
	@echo "⚡ Connecting to Redis CLI..."
	@docker exec -it bruno-redis redis-cli

# Open shell in API container
api-shell:
	@echo "🔧 Opening shell in API container..."
	@docker exec -it bruno-api /bin/sh

# Open shell in frontend container
frontend-shell:
	@echo "🔧 Opening shell in frontend container..."
	@docker exec -it bruno-frontend /bin/sh

# Run frontend in development mode (hot reload)
up-frontend-dev:
	@echo "🚀 Starting frontend in development mode..."
	@echo "📋 This will run the frontend with hot reload on http://localhost:5173"
	@echo "🔗 It will connect to the API running in Docker on http://localhost:8080"
	@echo "⏳ Starting Vite dev server..."
	@cd frontend && npm run dev

# Port forward API service for local testing
pf-api:
	@echo "🚪 Port forwarding API service for local testing..."
	@echo "💡 Access API at http://localhost:8080"
	@echo "💡 Health check: http://localhost:8080/health"
	@echo "💡 API endpoints: http://localhost:8080/api/v1/*"
	@kubectl port-forward --address 0.0.0.0 -n bruno svc/bruno-site-api 8080:8080

# Port forward Redis service for local testing
pf-redis:
	@echo "🔴 Port forwarding Redis service for local testing..."
	@echo "💡 Access Redis at localhost:6379"
	@kubectl port-forward --address 0.0.0.0 -n bruno svc/bruno-site-redis 6379:6379

# Port forward both database services for local testing
pf-postgres:
	@echo "🗄️ Port forwarding database services for local testing..."
	@echo "💡 PostgreSQL: localhost:5432 (bruno_site/postgres/secure-password)"
	@echo "💡 Redis: localhost:6379"
	@kubectl port-forward --address 0.0.0.0 -n bruno svc/bruno-site-postgres 5432:5432 & \
	kubectl port-forward --address 0.0.0.0 -n bruno svc/bruno-site-redis 6379:6379 & \
	wait

# Telepresence intercept both services for local development
tp-intercept:
	@echo "🔗 Setting up Telepresence intercept for both API and frontend development..."
	@echo "💡 This will route traffic from K8s to your local services"
	@echo "💡 Make sure both services are running locally first"
	@echo "💡 API: make run-api (in one terminal)"
	@echo "💡 Frontend: make frontend-dev (in another terminal)"
	@echo "💡 Press Ctrl+C to stop Telepresence intercept"
	@echo "🔗 Starting API intercept..."
	@telepresence intercept bruno-site-api --port 8080:8080 --mechanism tcp --mount false &
	@echo "🔗 Starting Frontend intercept..."
	@telepresence intercept bruno-site-frontend --port 3000:80 --mechanism tcp --mount false &
	@echo "✅ Both intercepts started. Press Ctrl+C to stop all intercepts."
	@wait

# Telepresence intercept with volume mounts (requires sshfs)
tp-intercept-mounts:
	@echo "🔗 Setting up Telepresence intercept with volume mounts..."
	@echo "💡 This requires sshfs to be installed"
	@echo "💡 Make sure both services are running locally first"
	@echo "💡 API: make run-api (in one terminal)"
	@echo "💡 Frontend: make frontend-dev (in another terminal)"
	@echo "💡 Press Ctrl+C to stop Telepresence intercept"
	@echo "🔗 Starting API intercept with mounts..."
	@telepresence intercept bruno-site-api --port 8080:8080 --mechanism tcp &
	@echo "🔗 Starting Frontend intercept with mounts..."
	@telepresence intercept bruno-site-frontend --port 3000:80 --mechanism tcp &
	@echo "✅ Both intercepts started with volume mounts. Press Ctrl+C to stop all intercepts."
	@wait

# Stop all Telepresence intercepts
tp-stop:
	@echo "🛑 Stopping all Telepresence intercepts..."
	@telepresence leave bruno-site-api || true
	@telepresence leave bruno-site-frontend || true
	@echo "✅ All intercepts stopped"

# Telepresence connect to cluster
tp-connect:
	@echo "🔗 Connecting to Kubernetes cluster with Telepresence..."
	@telepresence connect

# Telepresence disconnect from cluster
tp-disconnect:
	@echo "🔗 Disconnecting from Kubernetes cluster..."
	@telepresence quit

# Telepresence status
tp-status:
	@echo "📊 Telepresence status:"
	@telepresence status

# List active Telepresence intercepts
tp-list:
	@echo "📋 Active Telepresence intercepts:"
	@telepresence list

# Reconcile Flux HelmRelease for bruno-site
reconcile:
	@echo "🔄 Reconciling Flux Git source for bruno-site..."
	@flux reconcile source git bruno-site -n flux-system
	@echo "✅ Git source reconciliation completed"
	@echo "🔄 Reconciling Flux HelmRelease for bruno-site..."
	@flux reconcile helmrelease bruno-site -n bruno
	@echo "✅ HelmRelease reconciliation completed"

# Test API endpoints
test-api-endpoints:
	@echo "🧪 Testing API endpoints..."
	@echo "Health check:"
	@curl -s http://localhost:8080/health | jq . || curl -s http://localhost:8080/health
	@echo ""
	@echo "Projects:"
	@curl -s http://localhost:8080/api/v1/projects | jq . || curl -s http://localhost:8080/api/v1/projects
	@echo ""
	@echo "About:"
	@curl -s http://localhost:8080/api/about | jq . || curl -s http://localhost:8080/api/about
	@echo ""
	@echo "Contact:"
	@curl -s http://localhost:8080/api/contact | jq . || curl -s http://localhost:8080/api/contact
	@echo "Chat Health:"
	@curl -s http://localhost:8080/api/chat/health | jq . || curl -s http://localhost:8080/api/chat/health
	@echo ""
	@echo "Chat (POST test):"
	@curl -s -X POST http://localhost:8080/api/chat -H "Content-Type: application/json" -d '{"message": "Hello"}' | jq . || curl -s -X POST http://localhost:8080/api/chat -H "Content-Type: application/json" -d '{"message": "Hello"}'

# Run k6 load tests for golden signals (Kubernetes)
test-k6:
	@echo "🎯 Running k6 golden signals test in Kubernetes..."
	@echo "📊 Testing: Latency, Traffic, Errors, and Saturation"
	@echo "🔧 Rendering only k6 templates..."
	@helm template bruno-site ./chart --set k6.enabled=true --show-only templates/k6-configmap.yaml --show-only templates/k6-test.yaml > /tmp/k6-only.yaml
	@echo "📋 Applying only k6 resources..."
	@kubectl apply -f /tmp/k6-only.yaml
	@echo "⏳ Waiting for k6 test to complete..."
	@kubectl wait --for=condition=complete --timeout=10m testrun/bruno-site-golden-signals-test -n bruno || true
	@echo "📋 Getting test results..."
	@kubectl logs -l app.kubernetes.io/component=load-test -n bruno --tail=100
	@rm -f /tmp/k6-only.yaml
	@echo "✅ k6 golden signals test completed"

# Run k6 load tests locally (development)
test-k6-local:
	@echo "🎯 Running k6 golden signals test locally..."
	@echo "💡 Make sure the API is running (make up) before running this test"
	@echo "📊 Testing: Latency, Traffic, Errors, and Saturation"
	@kubectl get configmap bruno-site-k6-test -n bruno -o jsonpath='{.data.golden-signals-test\.js}' > /tmp/golden-signals-test.js
	@k6 run --out json=golden-signals-results.json /tmp/golden-signals-test.js
	@rm -f /tmp/golden-signals-test.js
	@echo "✅ k6 golden signals test completed"
	@echo "📋 Results saved to: golden-signals-results.json"

# Run all tests
test: test-api-unit test-frontend-unit test-e2e test-k6

# Run API unit tests
test-api-unit:
	@echo "🧪 Running API unit tests..."
	@cd api && go test -v -race -coverprofile=coverage.out ./...
	@cd api && go tool cover -html=coverage.out -o coverage.html
	@echo "✅ API unit tests completed"

# Run frontend unit tests
test-frontend-unit:
	@echo "🧪 Running frontend unit tests..."
	@cd frontend && npm install --legacy-peer-deps && npm run test -- --run --coverage
	@echo "✅ Frontend unit tests completed"

# Run E2E tests
test-e2e:
	@echo "🧪 Running E2E tests..."
	@cd frontend && npm install --legacy-peer-deps && npm run test:e2e
	@echo "✅ E2E tests completed"

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@cd api && go test -v -race -coverprofile=coverage.out ./...
	@cd api && go tool cover -func=coverage.out
	@cd frontend && npm run test:coverage
	@echo "✅ Coverage reports generated"

# Update dependencies
update-deps:
	@echo "📦 Updating dependencies..."
	@cd api && go mod tidy
	@cd frontend && npm update
	@echo "✅ Dependencies updated"

# Format code
format:
	@echo "🎨 Formatting code..."
	@cd api && go fmt ./...
	@cd frontend && npm run format 2>/dev/null || echo "No format script found in frontend"
	@echo "✅ Code formatted"

# Lint code
lint:
	@echo "🔍 Linting code..."
	@cd api && go vet ./...
	@cd frontend && npm run lint 2>/dev/null || echo "No lint script found in frontend"
	@echo "✅ Code linted"

# Cloudflare CDN Management
cloudflare-setup:
	@echo "🛡️ Setting up Cloudflare CDN..."
	@chmod +x scripts/cloudflare-setup.sh
	@./scripts/cloudflare-setup.sh

cloudflare-purge:
	@echo "🧹 Purging Cloudflare cache..."
	@if [ -f .env.cloudflare ]; then \
		source .env.cloudflare; \
		curl -X POST "https://api.cloudflare.com/client/v4/zones/$$CLOUDFLARE_ZONE_ID/purge_cache" \
			-H "Authorization: Bearer $$CLOUDFLARE_API_TOKEN" \
			-H "Content-Type: application/json" \
			-d '{"purge_everything": true}'; \
		echo "✅ Cache purged successfully"; \
	else \
		echo "❌ .env.cloudflare file not found. Run 'make cloudflare-setup' first."; \
	fi

cloudflare-deploy:
	@echo "🚀 Deploying to Cloudflare..."
	@echo "Building frontend..."
	@cd frontend && npm run build
	@echo "Purging cache..."
	@make cloudflare-purge
	@echo "✅ Deployment completed"

cloudflare-status:
	@echo "📊 Cloudflare status..."
	@if [ -f .env.cloudflare ]; then \
		source .env.cloudflare; \
		echo "Domain: $$CLOUDFLARE_DOMAIN"; \
		echo "API: https://api.$$CLOUDFLARE_DOMAIN"; \
		echo "WWW: https://www.$$CLOUDFLARE_DOMAIN"; \
		curl -s -I "https://$$CLOUDFLARE_DOMAIN" | head -1 || echo "Domain not accessible"; \
	else \
		echo "❌ .env.cloudflare file not found. Run 'make cloudflare-setup' first."; \
	fi

# Optimize images for web performance
optimize-images:
	@echo "🖼️ Optimizing images for web performance..."
	@cd scripts && npm install
	@cd scripts && npm run optimize-images
	@echo "✅ Image optimization completed!"