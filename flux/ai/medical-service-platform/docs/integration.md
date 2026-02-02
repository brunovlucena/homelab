# Integration Guide

## Overview

This guide explains how to integrate the medical-service-platform with existing `agent-medical` and `agents-whatsapp-rust` services.

## Architecture Flow

```
Doctor App (Web/Mobile)
    ↓ WebSocket
agents-whatsapp-rust (messaging-service)
    ↓ CloudEvent: messaging.message.received
agent-gateway
    ↓ CloudEvent: agent.message
medical-service
    ↓ CloudEvent: io.homelab.medical.query
agent-medical
    ↓ CloudEvent: io.homelab.medical.response
medical-service
    ↓ CloudEvent: messaging.message.received
agents-whatsapp-rust
    ↓ WebSocket
Doctor App
```

## Configuration

### 1. Update agent-gateway

Modify `agent-gateway/src/handlers.rs` to route doctor conversations to `agent-medical`:

```rust
// In handle_event function, update agent_id determination:
let agent_id = if let Some(conv) = conv_collection.find_one(conv_filter, None).await? {
    // Check if conversation is with agent-medical
    if conv.get_str("agent_id").map(|s| s.contains("agent-medical")).unwrap_or(false) {
        "agent-medical".to_string()
    } else {
        conv.get_str("agent_id")
            .map(String::from)
            .unwrap_or_else(|_| "agent-bruno".to_string())
    }
} else {
    // For new conversations, check if sender is a doctor
    // and route to agent-medical
    if sender_id.starts_with("doctor-") {
        "agent-medical".to_string()
    } else {
        "agent-bruno".to_string()
    }
};
```

### 2. Deploy medical-service

```bash
# Build Docker image
cd services/medical-service
docker build -t medical-service:latest .

# Apply Kubernetes manifests
kubectl apply -f ../../k8s/medical-service.yaml
```

### 3. Configure agent-medical

Ensure `agent-medical` is deployed and accessible at:
- URL: `http://agent-medical.agent-medical.svc.cluster.local:8080`
- CloudEvents: Subscribed to `io.homelab.medical.query`

### 4. Set up MongoDB collections

```javascript
// Create doctors collection
db.doctors.createIndex({ "_id": 1 });
db.doctors.createIndex({ "email": 1 });
db.doctors.createIndex({ "crm": 1 });

// Create consultations collection
db.consultations.createIndex({ "doctor_id": 1, "patient_id": 1 });
db.consultations.createIndex({ "status": 1 });
```

### 5. Environment Variables

Set these in your Kubernetes secrets:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: medical-service-secrets
  namespace: medical-service
type: Opaque
stringData:
  mongodb-uri: "mongodb://mongodb:27017"
  redis-uri: "redis://redis:6379"
  jwt-secret: "your-secret-key"
```

## Testing

### Test WebSocket Connection

```bash
# Connect to messaging service
wscat -c ws://localhost:8080/ws

# Send auth message
{"type":"auth","payload":{"user_id":"doctor-123","auth_token":"token","device_id":"test","platform":"web","app_version":"1.0.0"}}
```

### Test API Endpoints

```bash
# Summarize case
curl -X POST http://localhost:8080/summarize \
  -H "Content-Type: application/json" \
  -d '{"doctor_id":"doctor-123","patient_id":"patient-456"}'

# Correlate data
curl -X POST http://localhost:8080/correlate \
  -H "Content-Type: application/json" \
  -d '{"doctor_id":"doctor-123","patient_id":"patient-456","query":"lab results"}'
```

## Troubleshooting

### WebSocket Connection Issues

- Check messaging-service is running
- Verify WebSocket URL is correct
- Check authentication token

### Agent Not Responding

- Verify agent-medical is deployed
- Check CloudEvents are being published
- Review agent-medical logs

### Database Errors

- Verify MongoDB connection
- Check collection indexes exist
- Review access permissions
