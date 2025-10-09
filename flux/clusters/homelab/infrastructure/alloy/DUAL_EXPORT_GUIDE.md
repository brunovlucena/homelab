# Alloy Dual Export Guide: Tempo + Logfire

This guide explains how to send traces to **both** Tempo (local) and Logfire (SaaS) using Grafana Alloy.

## Architecture

```
Your App (with OTel SDK)
    │
    │ OTLP (port 4317)
    │
    ▼
┌─────────────────────────┐
│   Grafana Alloy         │
│  (OTel Collector)       │
│                         │
│  - Receives OTLP        │
│  - Processes traces     │
│  - Generates metrics    │
└────────┬────────────────┘
         │
         ├──────────────────┐
         │                  │
         ▼                  ▼
    ┌─────────┐      ┌──────────────┐
    │  Tempo  │      │   Logfire    │
    │ (Local) │      │    (SaaS)    │
    └─────────┘      └──────────────┘
```

## Current Status

✅ **Alloy is configured** to forward traces to **both destinations**
✅ **Tempo** is working (local traces)
⚠️  **Logfire** is ready but needs:
   - Logfire account signup
   - API token
   - Secret creation

## Option 1: Local Only (Default - Free)

**If you don't want to use Logfire SaaS**, simply comment out the Logfire exporter:

```yaml
# In helmrelease.yaml, line ~367
output {
  traces  = [
    otelcol.exporter.otlp.tempo.input,
    # otelcol.exporter.otlp.logfire.input,  # ← Commented out
  ]
}
```

Your traces will only go to **Tempo** (local, free, works offline).

## Option 2: Dual Export (Tempo + Logfire)

### Step 1: Sign Up for Logfire

1. Go to https://logfire.pydantic.dev
2. Create an account (check for free tier)
3. Get your API token from the dashboard

### Step 2: Create Secret

```bash
# Replace YOUR_TOKEN with your actual Logfire token
kubectl create secret generic logfire-secrets \
  --from-literal=token="Bearer YOUR_LOGFIRE_TOKEN_HERE" \
  -n alloy
```

Or using sealed-secrets (recommended for GitOps):

```bash
kubectl create secret generic logfire-secrets \
  --from-literal=token="Bearer YOUR_TOKEN_HERE" \
  --dry-run=client -o yaml | \
  kubeseal --format=yaml > logfire-sealed-secret.yaml

kubectl apply -f logfire-sealed-secret.yaml
```

### Step 3: Update Config

In `helmrelease.yaml`, change `optional: true` to `optional: false`:

```yaml
extraEnv:
- name: LOGFIRE_TOKEN
  valueFrom:
    secretKeyRef:
      name: logfire-secrets
      key: token
      optional: false  # ← Change to false
```

### Step 4: Restart Alloy

```bash
kubectl rollout restart deployment/alloy -n alloy
```

## How to Instrument Your Apps

### Go Example

```go
package main

import (
    "context"
    "log"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initTracer() (*sdktrace.TracerProvider, error) {
    // Send to Alloy (which forwards to Tempo + Logfire)
    exporter, err := otlptracegrpc.New(
        context.Background(),
        otlptracegrpc.WithEndpoint("alloy.alloy.svc.cluster.local:4317"),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    resource, err := resource.New(
        context.Background(),
        resource.WithAttributes(
            semconv.ServiceName("bruno-site-api"),
            semconv.ServiceVersion("1.0.0"),
            semconv.DeploymentEnvironment("production"),
        ),
    )
    if err != nil {
        return nil, err
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource),
        sdktrace.WithSampler(sdktrace.TraceIDRatioBased(1.0)), // Sample 100% for testing
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}

func main() {
    tp, err := initTracer()
    if err != nil {
        log.Fatal(err)
    }
    defer tp.Shutdown(context.Background())

    // Your app code here
    tracer := otel.Tracer("bruno-site-api")
    ctx, span := tracer.Start(context.Background(), "main")
    defer span.End()

    // ... your logic ...
}
```

### Python Example

```python
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

# Configure to send to Alloy
resource = Resource.create({
    "service.name": "bruno-site-api",
    "service.version": "1.0.0",
    "deployment.environment": "production",
})

provider = TracerProvider(resource=resource)

# Send to Alloy (which forwards to Tempo + Logfire)
otlp_exporter = OTLPSpanExporter(
    endpoint="alloy.alloy.svc.cluster.local:4317",
    insecure=True,
)

provider.add_span_processor(BatchSpanProcessor(otlp_exporter))
trace.set_tracer_provider(provider)

# Your app code
tracer = trace.get_tracer(__name__)

with tracer.start_as_current_span("my-operation"):
    # Your logic here
    pass
```

### JavaScript/Node.js Example

```javascript
const { NodeSDK } = require('@opentelemetry/sdk-node');
const { OTLPTraceExporter } = require('@opentelemetry/exporter-trace-otlp-grpc');
const { Resource } = require('@opentelemetry/resources');
const { SemanticResourceAttributes } = require('@opentelemetry/semantic-conventions');

const sdk = new NodeSDK({
  resource: new Resource({
    [SemanticResourceAttributes.SERVICE_NAME]: 'bruno-site-api',
    [SemanticResourceAttributes.SERVICE_VERSION]: '1.0.0',
  }),
  traceExporter: new OTLPTraceExporter({
    url: 'grpc://alloy.alloy.svc.cluster.local:4317',
  }),
});

sdk.start();

// Your app code here
const tracer = require('@opentelemetry/api').trace.getTracer('bruno-site-api');

const span = tracer.startSpan('my-operation');
// Your logic
span.end();
```

## Kubernetes Deployment Example

Add these environment variables to your app deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bruno-site-api
spec:
  template:
    spec:
      containers:
      - name: api
        image: your-image:latest
        env:
        # OpenTelemetry standard environment variables
        - name: OTEL_SERVICE_NAME
          value: "bruno-site-api"
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "http://alloy.alloy.svc.cluster.local:4317"
        - name: OTEL_EXPORTER_OTLP_PROTOCOL
          value: "grpc"
        - name: OTEL_TRACES_SAMPLER
          value: "traceidratio"
        - name: OTEL_TRACES_SAMPLER_ARG
          value: "1.0"  # Sample 100% (adjust for production)
```

## Viewing Your Traces

### Tempo (Local)
1. Open Grafana: `http://grafana.lucena.cloud`
2. Go to **Explore**
3. Select **Tempo** datasource
4. Query your traces

### Logfire (SaaS)
1. Go to https://logfire.pydantic.dev
2. Login to your account
3. View traces in the web UI (with better UX than Grafana)

## Benefits of Dual Export

✅ **Local Development**: Tempo works offline, no internet required
✅ **Production Debugging**: Logfire provides better UI and insights
✅ **Backup**: If Logfire is down, you still have Tempo
✅ **Cost Control**: Use Tempo for high-volume testing, Logfire for production
✅ **Gradual Migration**: Try Logfire without removing Tempo

## Cost Considerations

| Aspect | Tempo (Local) | Logfire (SaaS) |
|--------|---------------|----------------|
| **Cost** | Free (just infrastructure) | Paid (check pricing) |
| **Storage** | Your K8s cluster | Managed by Logfire |
| **Retention** | Configure as needed | Limited by plan |
| **UI** | Grafana | Better/Modern UI |
| **Availability** | 24/7 offline | Requires internet |

## Disable Logfire (Go Back to Local Only)

If you want to stop using Logfire:

1. Comment out the exporter in `helmrelease.yaml`:
   ```yaml
   output {
     traces  = [
       otelcol.exporter.otlp.tempo.input,
       # otelcol.exporter.otlp.logfire.input,
     ]
   }
   ```

2. Apply changes:
   ```bash
   kubectl apply -f helmrelease.yaml
   kubectl rollout restart deployment/alloy -n alloy
   ```

3. (Optional) Delete the secret:
   ```bash
   kubectl delete secret logfire-secrets -n alloy
   ```

## Troubleshooting

### Check Alloy Logs
```bash
kubectl logs -n alloy -l app.kubernetes.io/name=alloy --tail=100 -f
```

### Test OTLP Endpoint
```bash
# Port-forward Alloy
kubectl port-forward -n alloy svc/alloy 4317:4317

# Send test trace (requires grpcurl)
grpcurl -plaintext localhost:4317 list
```

### Verify Traces in Tempo
```bash
# Port-forward Tempo
kubectl port-forward -n tempo svc/tempo 3200:3200

# Query traces
curl http://localhost:3200/api/traces
```

### Check Logfire Connection
```bash
# Check Alloy logs for errors
kubectl logs -n alloy -l app.kubernetes.io/name=alloy | grep -i logfire
```

## Next Steps

1. **Instrument your homepage API** with OpenTelemetry
2. **Deploy and test** - traces should appear in both Tempo and Logfire
3. **Set up alerting** in Prometheus (already configured with PagerDuty)
4. **Monitor costs** if using Logfire SaaS

## References

- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [OpenTelemetry Python](https://opentelemetry.io/docs/instrumentation/python/)
- [Grafana Alloy Docs](https://grafana.com/docs/alloy/latest/)
- [Tempo Docs](https://grafana.com/docs/tempo/latest/)
- [Logfire Docs](https://logfire.pydantic.dev/docs/)

