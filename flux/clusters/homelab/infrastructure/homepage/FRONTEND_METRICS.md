# 📊 Frontend Prometheus Metrics

This document explains how Prometheus metrics are exposed from the homepage frontend.

## 🏗️ Architecture

The frontend uses a **sidecar pattern** to expose Prometheus metrics:

```
┌─────────────────────────────────────────────────────────┐
│                    Frontend Pod                          │
│                                                          │
│  ┌─────────────────┐         ┌────────────────────┐    │
│  │  Nginx          │         │ nginx-prometheus-  │    │
│  │  Container      │────────▶│ exporter           │    │
│  │                 │ :8080/  │ (sidecar)          │    │
│  │  Port 8080      │ stub_   │                    │    │
│  │                 │ status  │ Port 9113          │    │
│  │                 │         │ /metrics           │    │
│  └─────────────────┘         └────────────────────┘    │
│         │                              ▲                │
│         │ proxy /metrics               │                │
│         └──────────────────────────────┘                │
│                                                          │
└─────────────────────────────────────────────────────────┘
                       │
                       │ scrapes /metrics every 30s
                       ▼
            ┌──────────────────┐
            │   Prometheus     │
            │  (ServiceMonitor)│
            └──────────────────┘
```

## 🔧 Components

### 1. Nginx Configuration
- **Location:** `frontend/nginx.conf` and `chart/templates/frontend-configmap.yaml`
- **Endpoints:**
  - `/health` - Health check endpoint
  - `/stub_status` - Nginx internal metrics (localhost only)
  - `/metrics` - Proxied to nginx-exporter sidecar

### 2. Nginx Prometheus Exporter
- **Image:** `nginx/nginx-prometheus-exporter:1.4.0`
- **Port:** 9113
- **Source:** Reads from `http://127.0.0.1:8080/stub_status`
- **Resources:**
  - CPU: 50m request, 100m limit
  - Memory: 64Mi request, 128Mi limit

### 3. ServiceMonitor
- **Location:** `chart/templates/servicemonitor.yaml`
- **Scrape Interval:** 30s
- **Scrape Timeout:** 10s
- **Path:** `/metrics`
- **Port:** `http` (8080)

## 📈 Available Metrics

The nginx-prometheus-exporter exposes the following metrics:

### Connection Metrics
- `nginx_connections_active` - Current active client connections
- `nginx_connections_reading` - Connections currently reading client requests
- `nginx_connections_writing` - Connections currently writing responses to clients
- `nginx_connections_waiting` - Idle client connections waiting for a request

### Request Metrics
- `nginx_http_requests_total` - Total HTTP requests processed
- `nginx_connections_accepted` - Total accepted client connections
- `nginx_connections_handled` - Total handled client connections

### Build Info
- `nginx_up` - Whether nginx is up (1) or down (0)
- `nginxexporter_build_info` - Exporter build information

## 🎯 Example Queries

### Request Rate
```promql
# Requests per second
rate(nginx_http_requests_total{app_kubernetes_io_component="frontend"}[5m])

# Requests per minute
rate(nginx_http_requests_total{app_kubernetes_io_component="frontend"}[1m]) * 60
```

### Connection Metrics
```promql
# Active connections
nginx_connections_active{app_kubernetes_io_component="frontend"}

# Connection rate
rate(nginx_connections_accepted{app_kubernetes_io_component="frontend"}[5m])

# Connection handling efficiency (should be close to 1.0)
rate(nginx_connections_handled{app_kubernetes_io_component="frontend"}[5m]) 
/ 
rate(nginx_connections_accepted{app_kubernetes_io_component="frontend"}[5m])
```

### Health Status
```promql
# Is nginx up?
nginx_up{app_kubernetes_io_component="frontend"}
```

## 🚀 Usage

### Local Development
Access metrics locally:
```bash
# If running with Docker Compose
curl http://localhost:8080/metrics

# Get nginx stub_status directly
curl http://localhost:8080/stub_status
```

### Kubernetes
The ServiceMonitor automatically scrapes metrics. Check in Prometheus:
```bash
# Port-forward to Prometheus
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-prometheus 9090:9090

# Open browser
open http://localhost:9090
```

## 🔍 Troubleshooting

### Metrics Not Appearing
1. Check if the nginx-exporter sidecar is running:
```bash
kubectl get pods -n homepage -l app.kubernetes.io/component=frontend
kubectl logs -n homepage <pod-name> -c nginx-exporter
```

2. Check ServiceMonitor:
```bash
kubectl get servicemonitor -n homepage bruno-site-frontend -o yaml
```

3. Test metrics endpoint:
```bash
kubectl exec -n homepage <pod-name> -c homepage-frontend -- curl http://localhost:8080/metrics
```

### Nginx Stub Status Not Working
```bash
# Check nginx config
kubectl exec -n homepage <pod-name> -c homepage-frontend -- cat /etc/nginx/conf.d/default.conf

# Test stub_status
kubectl exec -n homepage <pod-name> -c homepage-frontend -- curl http://localhost:8080/stub_status
```

### Exporter Not Scraping
```bash
# Check exporter logs
kubectl logs -n homepage <pod-name> -c nginx-exporter

# Check if exporter can reach nginx
kubectl exec -n homepage <pod-name> -c nginx-exporter -- wget -O- http://127.0.0.1:8080/stub_status
```

## 📊 Grafana Dashboard

Example dashboard queries:

### Panel: Request Rate
- **Query:** `rate(nginx_http_requests_total{namespace="homepage",app_kubernetes_io_component="frontend"}[5m])`
- **Type:** Graph
- **Unit:** requests/sec

### Panel: Active Connections
- **Query:** `nginx_connections_active{namespace="homepage",app_kubernetes_io_component="frontend"}`
- **Type:** Gauge
- **Unit:** connections

### Panel: Connection States
- **Query 1:** `nginx_connections_reading{namespace="homepage",app_kubernetes_io_component="frontend"}`
- **Query 2:** `nginx_connections_writing{namespace="homepage",app_kubernetes_io_component="frontend"}`
- **Query 3:** `nginx_connections_waiting{namespace="homepage",app_kubernetes_io_component="frontend"}`
- **Type:** Graph (stacked)
- **Unit:** connections

## 🔗 References

- [nginx-prometheus-exporter](https://github.com/nginxinc/nginx-prometheus-exporter)
- [Nginx stub_status module](http://nginx.org/en/docs/http/ngx_http_stub_status_module.html)
- [Prometheus ServiceMonitor](https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.ServiceMonitor)

## 📝 Notes

- The exporter is lightweight (50m CPU, 64Mi RAM) [[memory:7609072]]
- Metrics are scraped every 30s by default
- The `/stub_status` endpoint is only accessible from localhost for security
- In Kubernetes, `/metrics` proxies to the exporter sidecar
- In local dev, `/metrics` directly exposes stub_status (simpler setup)

---

**Last Updated:** 2025-10-15  
**Status:** Production Ready

