# Prometheus Runbooks

This directory contains runbooks for troubleshooting Prometheus and AlertManager issues in the homelab infrastructure.

## 📚 Available Runbooks

### Critical Issues

- **[Prometheus Down](./prometheus-down.md)** - Prometheus server is completely down
- **[AlertManager Down](./alertmanager-down.md)** - AlertManager is unavailable, no alerts being sent

### Performance Issues

- **[High Memory Usage](./high-memory-usage.md)** - Prometheus consuming excessive memory
- **[Slow Queries](./slow-queries.md)** - Queries taking too long to execute
- **[Storage Full](./storage-full.md)** - Prometheus storage (PVC) running out of space

### Operational Issues

- **[Target Down](./target-down.md)** - One or more scrape targets are unavailable
- **[Rule Evaluation Failures](./rule-evaluation-failures.md)** - Recording or alerting rules failing to evaluate

## 🎯 Quick Reference

### Common Commands

```bash
# Check Prometheus status
kubectl get pods -n prometheus -l app.kubernetes.io/name=prometheus

# Access Prometheus UI
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-prometheus 9090:9090

# Access AlertManager UI
kubectl port-forward -n prometheus svc/prometheus-kube-prometheus-alertmanager 9093:9093

# Check logs
kubectl logs -n prometheus -l app.kubernetes.io/name=prometheus --tail=100

# Restart Prometheus
kubectl delete pod -n prometheus -l app.kubernetes.io/name=prometheus

# Force Flux reconciliation
flux reconcile helmrelease kube-prometheus-stack -n prometheus
```

### Useful Queries

```promql
# Check Prometheus health
up{job="prometheus-kube-prometheus-prometheus"}

# Check all targets
up

# Count down targets
count(up == 0)

# Check memory usage
container_memory_usage_bytes{pod=~"prometheus-prometheus-kube-prometheus-prometheus-.*"}

# Check storage usage
prometheus_tsdb_storage_blocks_bytes + prometheus_tsdb_head_chunks_storage_size_bytes

# Check series count
count({__name__=~".+"})

# Query duration
histogram_quantile(0.99, rate(prometheus_http_request_duration_seconds_bucket[5m]))
```

## 🔍 Troubleshooting Decision Tree

```
Is Prometheus responding?
├─ No → See prometheus-down.md
└─ Yes
    ├─ Are queries slow?
    │   └─ Yes → See slow-queries.md
    ├─ Is memory usage high?
    │   └─ Yes → See high-memory-usage.md
    ├─ Is storage filling up?
    │   └─ Yes → See storage-full.md
    ├─ Are some targets down?
    │   └─ Yes → See target-down.md
    ├─ Are rules failing?
    │   └─ Yes → See rule-evaluation-failures.md
    └─ Is AlertManager down?
        └─ Yes → See alertmanager-down.md
```

## 🚨 Alert Severity Levels

- **Critical**: Immediate action required, service down or at risk
- **Warning**: Action required soon, service degraded or at risk
- **Info**: Awareness only, no immediate action needed

## 📊 Monitoring Dashboard

Access the Prometheus dashboard at:
- Prometheus UI: `http://prometheus.homelab.local` or via port-forward
- Grafana Dashboards: `http://grafana.homelab.local`
  - Prometheus Stats
  - AlertManager Stats
  - Kubernetes / Compute Resources / Prometheus

## 🔗 Related Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
- [Kube-Prometheus-Stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)
- [AlertManager Documentation](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [PromQL Documentation](https://prometheus.io/docs/prometheus/latest/querying/basics/)

## 📝 Runbook Template

When creating new runbooks, follow this structure:

1. **Alert Information** - Name, severity, component, service
2. **Symptom** - What you observe
3. **Impact** - User, business, and data impact
4. **Diagnosis** - How to identify the issue
5. **Resolution Steps** - Step-by-step fixes
6. **Verification** - How to verify the fix worked
7. **Prevention** - How to prevent recurrence
8. **Related Alerts** - Connected issues
9. **Escalation** - When and how to escalate
10. **Additional Resources** - Links to docs

## 🤝 Contributing

When updating runbooks:
1. Keep commands up to date with current infrastructure
2. Test procedures before documenting
3. Include real examples from incidents
4. Update the "Last Updated" date
5. Version changes appropriately

## 📞 Escalation

If runbooks don't resolve the issue:
1. Check related runbooks
2. Review recent changes (Flux, Helm, config)
3. Check cluster-wide issues
4. Contact on-call engineer
5. Escalate to Prometheus expert if needed

---

**Last Updated:** 2025-10-15  
**Version:** 1.0

