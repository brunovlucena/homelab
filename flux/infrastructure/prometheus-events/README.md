# Prometheus Events

**CloudEvents converter for Alertmanager alerts**

Converts Alertmanager webhook payloads to CloudEvents format and publishes them to Knative Eventing for agent consumption.

## Architecture

```
Alertmanager → Webhook → prometheus-events → CloudEvents → Knative Eventing → agent-sre
```

## Event Types

- `io.homelab.prometheus.alert.fired` - Alert fired
- `io.homelab.prometheus.alert.resolved` - Alert resolved

## Configuration

Alertmanager should be configured to send webhooks to this service:

```yaml
receivers:
  - name: 'prometheus-events'
    webhook_configs:
      - url: 'http://prometheus-events.prometheus.svc.cluster.local:8080/webhook'
```

