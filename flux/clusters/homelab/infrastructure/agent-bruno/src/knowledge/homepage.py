"""
🏠 Homepage Knowledge Base

This module contains comprehensive knowledge about the homepage application.
"""

from typing import Dict, List, Any


class HomepageKnowledge:
    """Knowledge base for the homepage application"""

    def __init__(self):
        self.architecture = self._load_architecture()
        self.api_endpoints = self._load_api_endpoints()
        self.deployment = self._load_deployment_info()
        self.components = self._load_components()
        self.tech_stack = self._load_tech_stack()

    def _load_architecture(self) -> Dict[str, Any]:
        """Load architecture information"""
        return {
            "overview": "Modern cloud-native homepage with microservices architecture",
            "layers": {
                "frontend": {
                    "tech": "React 18 + TypeScript + Vite",
                    "description": "Static site served via Nginx",
                    "port": 80,
                    "deployment": "2 replicas with HPA"
                },
                "api": {
                    "tech": "Go 1.23 + Gin + GORM",
                    "description": "RESTful API with proxy to agent services",
                    "port": 8080,
                    "deployment": "2 replicas with HPA"
                },
                "database": {
                    "tech": "PostgreSQL 15",
                    "description": "Primary database for projects, skills, experiences",
                    "port": 5432,
                    "deployment": "1 replica with persistent volume"
                },
                "cache": {
                    "tech": "Redis 7",
                    "description": "Session storage and caching",
                    "port": 6379,
                    "deployment": "1 replica with persistent volume"
                },
                "storage": {
                    "tech": "MinIO",
                    "description": "S3-compatible object storage for assets",
                    "port": 9000,
                    "deployment": "Shared MinIO service"
                }
            },
            "flow": {
                "user_request": "User → Cloudflare CDN → Ingress → Frontend (Nginx) → API (Go) → Database/Services",
                "chatbot": "User → Frontend → API Proxy → Agent-SRE/Agent-Bruno → Ollama/LLM"
            }
        }

    def _load_api_endpoints(self) -> Dict[str, List[Dict[str, str]]]:
        """Load API endpoints information"""
        return {
            "content": [
                {"method": "GET", "path": "/api/v1/content", "description": "Get all dynamic content"},
                {"method": "GET", "path": "/api/v1/content/:id", "description": "Get specific content"},
                {"method": "POST", "path": "/api/v1/content", "description": "Create content (admin)"},
                {"method": "PUT", "path": "/api/v1/content/:id", "description": "Update content (admin)"},
                {"method": "DELETE", "path": "/api/v1/content/:id", "description": "Delete content (admin)"},
            ],
            "projects": [
                {"method": "GET", "path": "/api/v1/projects", "description": "Get all projects"},
                {"method": "GET", "path": "/api/v1/projects/:id", "description": "Get specific project"},
                {"method": "POST", "path": "/api/v1/projects", "description": "Create project"},
                {"method": "PUT", "path": "/api/v1/projects/:id", "description": "Update project"},
                {"method": "DELETE", "path": "/api/v1/projects/:id", "description": "Delete project"},
            ],
            "skills": [
                {"method": "GET", "path": "/api/v1/skills", "description": "Get all skills"},
                {"method": "GET", "path": "/api/v1/skills/:id", "description": "Get specific skill"},
                {"method": "POST", "path": "/api/v1/skills", "description": "Create skill"},
                {"method": "PUT", "path": "/api/v1/skills/:id", "description": "Update skill"},
                {"method": "DELETE", "path": "/api/v1/skills/:id", "description": "Delete skill"},
            ],
            "experiences": [
                {"method": "GET", "path": "/api/v1/experiences", "description": "Get all experiences"},
                {"method": "GET", "path": "/api/v1/experiences/:id", "description": "Get specific experience"},
            ],
            "agents": [
                {"method": "POST", "path": "/api/v1/agent-sre/chat", "description": "Direct chat with Agent-SRE"},
                {"method": "POST", "path": "/api/v1/agent-sre/mcp/chat", "description": "MCP chat with Agent-SRE"},
                {"method": "POST", "path": "/api/v1/agent-bruno/chat", "description": "Chat with Agent-Bruno"},
                {"method": "POST", "path": "/api/v1/jamie/chat", "description": "Chat with Jamie"},
            ],
            "health": [
                {"method": "GET", "path": "/health", "description": "Health check"},
                {"method": "GET", "path": "/ready", "description": "Readiness check"},
                {"method": "GET", "path": "/metrics", "description": "Prometheus metrics"},
            ],
            "assets": [
                {"method": "GET", "path": "/api/v1/assets/:filename", "description": "Get asset from MinIO"},
                {"method": "POST", "path": "/api/v1/assets", "description": "Upload asset to MinIO"},
            ],
            "cloudflare": [
                {"method": "POST", "path": "/api/v1/cloudflare/purge", "description": "Purge Cloudflare cache"},
                {"method": "GET", "path": "/api/v1/cloudflare/stats", "description": "Get Cloudflare stats"},
            ]
        }

    def _load_deployment_info(self) -> Dict[str, Any]:
        """Load deployment information"""
        return {
            "local_dev": {
                "method": "Docker Compose",
                "command": "docker-compose up -d",
                "ports": {
                    "frontend": 3000,
                    "api": 8080,
                    "postgres": 5432,
                    "redis": 6379
                }
            },
            "kubernetes": {
                "method": "Helm Chart",
                "namespace": "homepage",
                "command": "helm upgrade --install bruno-site ./chart --namespace homepage",
                "components": [
                    "frontend-deployment",
                    "api-deployment",
                    "postgres-deployment",
                    "redis-deployment",
                    "ingress",
                    "services",
                    "configmaps",
                    "secrets"
                ]
            },
            "ci_cd": {
                "platform": "GitHub Actions",
                "workflows": [
                    "homepage-tests.yml - Run on push/PR",
                    "homepage-pr-check.yml - Run on PR only",
                    "homepage-nightly-tests.yml - Run daily at 2 AM UTC"
                ]
            },
            "scaling": {
                "hpa_enabled": True,
                "min_replicas": 1,
                "max_replicas": 3,
                "target_cpu": "80%"
            }
        }

    def _load_components(self) -> Dict[str, Dict[str, Any]]:
        """Load component details"""
        return {
            "frontend": {
                "files": [
                    "src/App.tsx - Main React app",
                    "src/components/Chatbot.tsx - Chatbot UI",
                    "src/components/Header.tsx - Header component",
                    "src/services/chatbot.ts - Chatbot service with MCP/Direct modes",
                    "src/services/api.ts - API client",
                    "nginx.conf - Nginx configuration with proxy"
                ],
                "features": [
                    "Dynamic content loading",
                    "AI chatbot integration",
                    "Responsive design",
                    "TypeScript type safety"
                ]
            },
            "api": {
                "files": [
                    "main.go - Entry point with OpenTelemetry",
                    "config/config.go - Configuration management",
                    "database/database.go - Database initialization",
                    "database/redis.go - Redis initialization",
                    "handlers/agent_sre.go - Agent-SRE proxy",
                    "handlers/jamie.go - Jamie proxy",
                    "handlers/projects.go - Projects CRUD",
                    "handlers/skills.go - Skills CRUD",
                    "handlers/content.go - Content CRUD",
                    "router/router.go - Route definitions",
                    "storage/minio.go - MinIO client"
                ],
                "features": [
                    "OpenTelemetry tracing",
                    "CORS handling",
                    "Gzip compression",
                    "Health checks",
                    "Prometheus metrics"
                ]
            },
            "database": {
                "schema": {
                    "projects": "id, name, description, tech_stack, url, github_url, image_url, created_at, updated_at",
                    "skills": "id, name, category, proficiency, years, created_at, updated_at",
                    "experiences": "id, company, role, description, start_date, end_date, created_at, updated_at",
                    "content": "id, key, value, type, created_at, updated_at"
                },
                "migrations": [
                    "001_complete_schema.sql - Initial schema",
                    "002_performance_indexes.sql - Performance indexes"
                ]
            }
        }

    def _load_tech_stack(self) -> Dict[str, List[str]]:
        """Load technology stack"""
        return {
            "backend": [
                "Go 1.23",
                "Gin Web Framework",
                "GORM (ORM)",
                "Redis Client",
                "MinIO Client",
                "OpenTelemetry"
            ],
            "frontend": [
                "React 18",
                "TypeScript",
                "Vite",
                "Axios",
                "Tailwind CSS (optional)"
            ],
            "database": [
                "PostgreSQL 15",
                "Redis 7"
            ],
            "infrastructure": [
                "Kubernetes",
                "Helm",
                "Docker",
                "Nginx",
                "MinIO"
            ],
            "ai": [
                "Agent-SRE",
                "Agent-Bruno",
                "Jamie Slack Bot",
                "Ollama",
                "MCP Protocol"
            ],
            "observability": [
                "Prometheus",
                "Grafana",
                "Loki",
                "Tempo",
                "Alloy",
                "OpenTelemetry"
            ],
            "cicd": [
                "GitHub Actions",
                "Docker Registry",
                "Flux CD"
            ]
        }

    def get_info(self, category: str, subcategory: str = None) -> Any:
        """Get information from knowledge base"""
        categories = {
            "architecture": self.architecture,
            "api": self.api_endpoints,
            "deployment": self.deployment,
            "components": self.components,
            "tech": self.tech_stack
        }

        if category not in categories:
            return None

        data = categories[category]

        if subcategory and isinstance(data, dict):
            return data.get(subcategory)

        return data

    def search(self, query: str) -> List[Dict[str, Any]]:
        """Search knowledge base for query"""
        query_lower = query.lower()
        results = []

        # Search in all data structures
        for category in ["architecture", "api", "deployment", "components", "tech"]:
            data = self.get_info(category)
            if self._search_in_data(data, query_lower):
                results.append({
                    "category": category,
                    "data": data,
                    "relevance": self._calculate_relevance(data, query_lower)
                })

        # Sort by relevance
        results.sort(key=lambda x: x["relevance"], reverse=True)
        return results

    def _search_in_data(self, data: Any, query: str) -> bool:
        """Check if query exists in data"""
        if isinstance(data, str):
            return query in data.lower()
        elif isinstance(data, dict):
            return any(self._search_in_data(v, query) for v in data.values())
        elif isinstance(data, list):
            return any(self._search_in_data(item, query) for item in data)
        return False

    def _calculate_relevance(self, data: Any, query: str) -> int:
        """Calculate relevance score"""
        score = 0
        query_words = query.split()

        def count_matches(obj):
            nonlocal score
            if isinstance(obj, str):
                for word in query_words:
                    if word in obj.lower():
                        score += 1
            elif isinstance(obj, dict):
                for v in obj.values():
                    count_matches(v)
            elif isinstance(obj, list):
                for item in obj:
                    count_matches(item)

        count_matches(data)
        return score

    def get_summary(self) -> str:
        """Get a summary of the homepage application"""
        return """
🏠 Bruno's Homepage - Production-Ready Dynamic Site

**Architecture**: Modern cloud-native application with React frontend, Go API, PostgreSQL database, and Redis cache.

**Key Features**:
- Dynamic content management via API
- AI chatbot integration (Agent-SRE, Agent-Bruno, Jamie)
- Real-time updates
- Cloudflare CDN integration
- Comprehensive observability (Prometheus, Grafana, Loki, Tempo)

**Deployment**: Kubernetes with Helm, 2 replicas for frontend/API with HPA, persistent storage for database.

**Tech Stack**: Go 1.23, React 18, PostgreSQL 15, Redis 7, MinIO, OpenTelemetry

**Status**: ✅ Production Ready with 100% test coverage (50+ tests)
"""

