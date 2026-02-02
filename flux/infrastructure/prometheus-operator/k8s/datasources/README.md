# Grafana Datasources for Prometheus Operator

This directory contains Grafana datasource configurations that are automatically deployed via ConfigMaps.

## Datasource Discovery

Datasources are automatically discovered by Grafana via the sidecar container. The sidecar watches for ConfigMaps with the label `grafana_datasource: "1"` in the `prometheus` namespace and automatically imports them into Grafana.

## Available Datasources

### Loki Datasource

**File**: `loki-datasource.yaml`  
**ConfigMap**: `grafana-datasource-loki`  
**UID**: `loki`

#### Configuration

- **Type**: Loki
- **URL**: `http://loki.loki:3100`
- **Access**: Proxy
- **Features**:
  - Max lines: 1000
  - Derived fields configured to link trace IDs to Tempo

### Tempo Datasource

**File**: `tempo-datasource.yaml`  
**ConfigMap**: `grafana-datasource-tempo`  
**UID**: `tempo`

#### Configuration

- **Type**: Tempo
- **URL**: `http://tempo.tempo:3200`
- **Access**: Proxy
- **Features**:
  - Traces to logs integration with Loki
  - Service map integration with Prometheus
  - Node graph enabled
  - Span bar configured to show HTTP status codes

### Google Analytics Datasource

**File**: `google-analytics-datasource.yaml`  
**ConfigMap**: `grafana-datasource-google-analytics`  
**UID**: `google-analytics`

#### Configuration

- **Type**: Google Analytics (blackcowmoo-googleanalytics-datasource plugin)
- **Access**: Proxy
- **Authentication**: JWT (Service Account)
- **Service Account**: `homelab@homelab-481500.iam.gserviceaccount.com`
- **Project**: `homelab-481500`

#### Setup Requirements

1. **Download Service Account Key**:
   - Go to IAM & Admin > Service Accounts in Google Cloud Console
   - Select the "homelab" service account
   - Go to Keys tab > Add key > Create new key (JSON format)
   - Download the JSON file

2. **Update ConfigMap**:
   - Open `google-analytics-datasource.yaml`
   - Replace the placeholder values with actual values from the downloaded JSON:
     - `REPLACE_WITH_PRIVATE_KEY_ID_FROM_JSON` → `private_key_id` from JSON
     - `REPLACE_WITH_PRIVATE_KEY_FROM_JSON` → `private_key` from JSON (keep `\n` characters)
     - `REPLACE_WITH_CLIENT_ID_FROM_JSON` → `client_id` from JSON

3. **Enable APIs**:
   - Enable Google Analytics Admin API
   - Enable Google Analytics Data API

4. **Grant Permissions**:
   - In Google Analytics, go to Admin > Account Access Management
   - Add the service account email with "Read & Analyze" role

**Note**: The ConfigMap contains sensitive JWT data. The `secureJsonData` field is required by Grafana's datasource provisioning system.

## Deployment

Datasources are deployed automatically when:
1. The ConfigMap is created/updated in the `prometheus` namespace
2. The ConfigMap has the label `grafana_datasource: "1"`
3. The Grafana sidecar is running and watching for ConfigMaps

The sidecar configuration is in `helmrelease.yaml`:
```yaml
sidecar:
  datasources:
    enabled: true
    label: grafana_datasource
    labelValue: "1"
```

## Adding New Datasources

1. Create a YAML ConfigMap file in this directory
2. Use the following structure:
   ```yaml
   ---
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: grafana-datasource-<name>
     namespace: prometheus
     labels:
       grafana_datasource: "1"
   data:
     <name>-datasource.json: |
       {
         "apiVersion": 1,
         "datasources": [
           {
             "name": "<Display Name>",
             "type": "<datasource-type>",
             "uid": "<unique-id>",
             "access": "proxy",
             "url": "http://<service>.<namespace>:<port>",
             "jsonData": {
               // datasource-specific configuration
             },
             "editable": true
           }
         ]
       }
   ```
3. Add the ConfigMap to `kustomization.yaml`
4. Commit and push - Flux will deploy it automatically
