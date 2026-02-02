# Grafana Dashboard Provisioning

This directory contains Grafana dashboards that are automatically provisioned via Kubernetes ConfigMaps.

## How It Works

1. **Dashboard JSON Files**: Place your Grafana dashboard JSON files in this directory
2. **Auto-Generation**: Run `generate-configmaps.py` to create ConfigMap YAML files
3. **GitOps Sync**: Flux syncs the ConfigMaps to Kubernetes
4. **Auto-Discovery**: Grafana's sidecar automatically discovers and provisions dashboards from ConfigMaps with the label `grafana_dashboard: "1"`

## Workflow

### Adding a New Dashboard

1. **Create Dashboard JSON**:
   ```bash
   # Export from Grafana UI or create manually
   # Save as: my-dashboard.json
   ```

2. **Generate ConfigMap**:
   ```bash
   cd dashboards
   python3 generate-configmaps.py
   ```

3. **Commit and Push**:
   ```bash
   git add dashboards/
   git commit -m "Add new dashboard: my-dashboard"
   git push
   ```

4. **Flux Syncs**: Flux will automatically apply the ConfigMap to Kubernetes

5. **Grafana Discovers**: Grafana sidecar will automatically load the dashboard

### Updating an Existing Dashboard

1. **Update JSON File**: Edit the `.json` file directly
2. **Regenerate ConfigMap**: Run `generate-configmaps.py` again
3. **Commit and Push**: GitOps will sync the changes

### Dashboard Requirements

- **UID**: Each dashboard must have a unique `uid` field in the JSON
- **Datasource UID**: Ensure datasource references use the correct UID (e.g., `"uid": "prometheus"`)
- **JSON Format**: Must be valid Grafana dashboard JSON (can export from Grafana UI)

## Current Dashboards

- **lambda-metrics-dashboard**: Knative Lambda metrics with dynamic service selection
- **lambda-rabbitmq-correlation-dashboard**: Correlation between Lambda and RabbitMQ metrics

## Script Details

The `generate-configmaps.py` script:

- Scans for all `*.json` files in the directory
- Generates ConfigMap YAML files with proper labels
- Updates `kustomization.yaml` automatically
- Preserves JSON formatting and structure

## Troubleshooting

### Dashboard Not Appearing in Grafana

1. Check ConfigMap exists:
   ```bash
   kubectl get configmap -n prometheus -l grafana_dashboard=1
   ```

2. Check Grafana sidecar logs:
   ```bash
   kubectl logs -n prometheus -l app.kubernetes.io/name=grafana -c grafana-sc-dashboard
   ```

3. Verify labels:
   ```bash
   kubectl get configmap <dashboard-name> -n prometheus --show-labels
   ```

### Regenerating All ConfigMaps

```bash
cd dashboards
# Remove old ConfigMap YAMLs (keep JSON files)
rm *-dashboard.yaml
# Regenerate
python3 generate-configmaps.py
```

## References

- [Grafana Dashboard Provisioning](https://grafana.com/docs/grafana/latest/administration/provisioning/#dashboards)
- [Grafana Sidecar Pattern](https://github.com/kiwigrid/k8s-sidecar)
