# Homepage Grafana Dashboards

This directory contains two dashboards for monitoring the homepage application.

## Dashboards

### 1. Homepage Metrics Dashboard

**File**: `homepage-metrics-dashboard-configmap.yaml`  
**Dashboard UID**: `homepage-metrics`  
**Datasource**: Prometheus

#### Overview

Comprehensive metrics dashboard for the homepage API backend, showing application performance, business metrics, and infrastructure health.

#### Sections

##### Overview
- **Request Rate**: Total HTTP requests per second
- **P95 Latency**: 95th percentile request latency
- **Error Rate**: Percentage of 5xx errors
- **Project Views/sec**: Rate of project views
- **DB Connections**: Active database connections

##### HTTP Metrics
- **Request Rate by Endpoint**: Requests per second grouped by endpoint
- **P95 Latency by Endpoint**: Response time distribution by endpoint

##### Business Metrics
- **Project Views by ID**: Track which projects are most viewed
- **Chat Sessions**: Success/error rates for chat interactions

##### Database & Redis
- **Database Connections**: Active vs idle connections
- **Database Query Duration**: P95 latency for DB operations
- **Redis Operations**: Operation rates by type and status
- **Redis Operation Duration**: Performance of Redis operations

##### LLM/AI Metrics
- **LLM Requests by Model**: Usage of different AI models
- **LLM Request Duration**: Response times for AI requests
- **LLM Tokens Generated**: Token usage tracking

#### Available Metrics

```promql
# HTTP Metrics
http_requests_total{job="homepage-api"}
http_request_duration_seconds{job="homepage-api"}
http_request_size_bytes{job="homepage-api"}
http_response_size_bytes{job="homepage-api"}

# Database Metrics
db_connections_active{job="homepage-api"}
db_connections_idle{job="homepage-api"}
db_queries_total{job="homepage-api"}
db_query_duration_seconds{job="homepage-api"}

# Redis Metrics
redis_operations_total{job="homepage-api"}
redis_operation_duration_seconds{job="homepage-api"}

# LLM Metrics
llm_requests_total{job="homepage-api"}
llm_request_duration_seconds{job="homepage-api"}
llm_tokens_generated_total{job="homepage-api"}

# Business Metrics
project_views_total{job="homepage-api"}
chat_sessions_total{job="homepage-api"}
```

### 2. Google Analytics Dashboard

**File**: `google-analytics-dashboard-configmap.yaml`  
**Dashboard UID**: `google-analytics-homepage`  
**Datasource**: Google Analytics (blackcowmoo-googleanalytics-datasource)

#### Overview

Comprehensive analytics dashboard showing website traffic, user behavior, and engagement metrics from Google Analytics.

#### Sections

##### Overview
- **Active Users**: Current active users on the site
- **Page Views**: Total page views in selected period
- **Sessions**: Total sessions
- **Avg Session Duration**: Average time users spend on the site
- **Active Users Over Time**: Time series of active users

##### Traffic Sources
- **Users by Channel**: Distribution of traffic by channel (Organic, Direct, Social, etc.)
- **Top Traffic Sources**: Most significant traffic sources

##### Pages & Content
- **Top Pages**: Most visited pages with view counts
- **Page Views Over Time**: Page view trends

##### Audience
- **Users by Country**: Geographic distribution of visitors
- **Users by Device**: Mobile, Desktop, Tablet breakdown

##### Engagement
- **Bounce Rate**: Percentage of single-page sessions over time
- **User Engagement Duration**: Average time users engage with content

#### Metrics

The Google Analytics dashboard uses the following GA4 metrics:

- `activeUsers`: Number of active users
- `screenPageViews`: Page views
- `sessions`: Session count
- `averageSessionDuration`: Average time per session
- `bounceRate`: Percentage of bounced sessions
- `userEngagementDuration`: Time users spend engaging

#### Dimensions

- `date`: Time dimension
- `pagePath`: URL path of pages
- `sessionDefaultChannelGrouping`: Traffic channel (Organic, Direct, etc.)
- `sessionSource`: Specific traffic source
- `country`: User country
- `deviceCategory`: Device type (Desktop, Mobile, Tablet)

## Deployment

These dashboards are automatically deployed by Flux when you commit and push changes. The Grafana sidecar watches for ConfigMaps with the label `grafana_dashboard: "1"` and automatically provisions them.

### Manual Deployment

If you need to manually verify or deploy:

```bash
# Check if ConfigMaps exist
kubectl get configmap -n prometheus | grep dashboard

# Verify homepage metrics dashboard
kubectl get configmap homepage-metrics-dashboard -n prometheus

# Verify Google Analytics dashboard
kubectl get configmap google-analytics-dashboard -n prometheus

# Check Grafana sidecar logs
kubectl logs -n prometheus deployment/kube-prometheus-stack-grafana -c grafana-sc-dashboards --tail=50
```

## Accessing Dashboards

1. Go to [Grafana](https://grafana.lucena.cloud)
2. Navigate to **Dashboards** → **Browse**
3. Look for:
   - **Homepage Metrics** (tagged: homepage, api, prometheus)
   - **Google Analytics - Homepage** (tagged: google-analytics, homepage, traffic)

Or use direct URLs (after deployment):
- Homepage Metrics: `https://grafana.lucena.cloud/d/homepage-metrics`
- Google Analytics: `https://grafana.lucena.cloud/d/google-analytics-homepage`

## Customization

### Modifying Dashboards

1. Edit the JSON files:
   - `homepage-metrics-dashboard.json`
   - `google-analytics-dashboard.json`

2. Update the ConfigMaps (automatically done by script):
   ```bash
   # Regenerate ConfigMaps
   cd /Users/brunolucena/workspace/bruno/repos/homelab/flux/infrastructure/prometheus-operator/k8s/dashboards
   
   # For homepage metrics
   cat homepage-metrics-dashboard-configmap.yaml | head -9 > temp.yaml
   cat homepage-metrics-dashboard.json | sed 's/^/    /' >> temp.yaml
   mv temp.yaml homepage-metrics-dashboard-configmap.yaml
   
   # For Google Analytics
   cat google-analytics-dashboard-configmap.yaml | head -9 > temp.yaml
   cat google-analytics-dashboard.json | sed 's/^/    /' >> temp.yaml
   mv temp.yaml google-analytics-dashboard-configmap.yaml
   ```

3. Commit and push changes

### Adding New Panels

To add new panels to existing dashboards:

1. Edit the dashboard in Grafana UI
2. Export the dashboard JSON (Settings → JSON Model)
3. Save to the respective `.json` file
4. Regenerate the ConfigMap
5. Commit and push

## Troubleshooting

### Dashboard Not Appearing

1. **Check ConfigMap exists**:
   ```bash
   kubectl get configmap -n prometheus | grep -E "homepage-metrics|google-analytics"
   ```

2. **Check labels**:
   ```bash
   kubectl get configmap homepage-metrics-dashboard -n prometheus -o yaml | grep grafana_dashboard
   ```

3. **Check Grafana sidecar logs**:
   ```bash
   kubectl logs -n prometheus deployment/kube-prometheus-stack-grafana -c grafana-sc-dashboards --tail=100
   ```

4. **Force sidecar refresh**:
   ```bash
   kubectl rollout restart deployment/kube-prometheus-stack-grafana -n prometheus
   ```

### No Data in Homepage Metrics

1. **Verify ServiceMonitor exists**:
   ```bash
   kubectl get servicemonitor homepage-api -n homepage
   ```

2. **Check if Prometheus is scraping**:
   - Go to Prometheus UI: `https://prometheus.lucena.cloud`
   - Check Targets: Status → Targets
   - Look for `homepage-api` endpoint

3. **Test metrics endpoint directly**:
   ```bash
   kubectl port-forward -n homepage svc/homepage-api 8080:80
   curl http://localhost:8080/metrics
   ```

### No Data in Google Analytics

1. **Verify datasource is configured**:
   ```bash
   kubectl get configmap grafana-datasource-google-analytics -n prometheus
   ```

2. **Check datasource in Grafana**:
   - Go to Connections → Data sources
   - Click on "Google Analytics"
   - Click "Save & Test"

3. **Verify service account permissions**:
   - In Google Analytics: Admin → Account Access Management
   - Ensure `homelab@homelab-481500.iam.gserviceaccount.com` has "Viewer" or "Analyst" role

4. **Check Grafana logs**:
   ```bash
   kubectl logs -n prometheus deployment/kube-prometheus-stack-grafana -c grafana --tail=100 | grep -i "google\|analytics"
   ```

## Metrics Reference

### Homepage Metrics Source

All metrics are collected from the homepage API at `/metrics` endpoint:
- Exposed on port `80` (http)
- Scraped by Prometheus every 30 seconds
- ServiceMonitor: `homepage-api` in `homepage` namespace

### Google Analytics Source

Data is fetched via Google Analytics Data API:
- Property ID configured in datasource
- Authentication via service account
- Data freshness: ~24-48 hours (GA processing delay)

## Next Steps

- **Add Alerts**: Create PrometheusRules for critical metrics
- **Create Custom Dashboards**: Build role-specific views (developer, product, business)
- **Export Reports**: Use Grafana reporting plugin for automated reports
- **Link Dashboards**: Cross-reference between Prometheus and GA data
