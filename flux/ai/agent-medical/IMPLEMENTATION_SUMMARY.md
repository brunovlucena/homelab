# Apple Watch Integration - Implementation Summary

## ✅ Completed Components

### 1. Integration Guide
- **File**: `APPLE_WATCH_INTEGRATION.md`
- Comprehensive architecture document with data flow, implementation phases, and configuration details

### 1.1. Performance Analysis
- **File**: `PERFORMANCE_ANALYSIS.md`
- SRE analysis of Lambda tool calling bottlenecks in gRPC chat
- Solutions for async tool calling, pre-warming, and optimization strategies

### 2. iOS HealthKit Service
- **File**: `ios-app/AgentChat/AgentChat/Services/HealthKitService.swift`
- Features:
  - HealthKit authorization management
  - Heart rate data reading (latest and time-range queries)
  - Background delivery setup for real-time monitoring
  - CloudEvent transmission to medical agent
  - Error handling and status tracking

### 3. Heart Rate Analyzer (Python Lambda Function)
- **File**: `src/medical_agent/heart_rate_analyzer.py`
- **Architecture**:
  - Lambda function called by agent-medical
  - Invoked via CloudEvents from mobile app (patient)
  - Chat interaction uses gRPC in the app
  - Tool calling uses CloudEvents
- Features:
  - Context-aware analysis (resting, active, exercise, sleep)
  - Normal range detection (60-100 bpm resting)
  - Tachycardia detection (>100 bpm)
  - Bradycardia detection (<60 bpm)
  - Severity classification (none, mild, moderate, severe)
  - Clinical recommendations
  - Baseline comparison support

### 4. Medical Agent Integration
- **File**: `src/medical_agent/main.py`
- New endpoints:
  - `POST /analyze/heart-rate` - HTTP endpoint for heart rate analysis
- Event handler:
  - Handles `io.homelab.medical.heart-rate.report` CloudEvents
  - Returns `io.homelab.medical.heart-rate.analysis` response
- WhatsApp integration:
  - `send_whatsapp_notification()` function
  - Sends CloudEvents to WhatsApp broker when anomalies detected
  - Automatic notification for moderate/severe anomalies

## 📋 Remaining Tasks

### 1. Xcode Project Configuration (Required)
Add HealthKit capability to the iOS app:

1. Open `ios-app/AgentChat/AgentChat.xcodeproj` in Xcode
2. Select the project target
3. Go to "Signing & Capabilities" tab
4. Click "+ Capability"
5. Add "HealthKit"
6. Add to `Info.plist`:
   ```xml
   <key>NSHealthShareUsageDescription</key>
   <string>This app needs access to your heart rate data from Apple Watch to monitor your health and send alerts if needed.</string>
   <key>NSHealthUpdateUsageDescription</key>
   <string>This app needs access to update your health data.</string>
   ```

### 2. iOS App UI Integration (Optional but Recommended)
- Add UI to request HealthKit authorization
- Display heart rate status in app
- Show last reading and analysis result
- Add button to manually sync heart rate data

### 3. Configuration
Set environment variables for medical agent:
```bash
# WhatsApp broker URL (for notifications)
WHATSAPP_BROKER_URL=http://default-broker.knative-lambda.svc.cluster.local:80

# Heart rate thresholds (optional, defaults provided)
HEART_RATE_NORMAL_MIN=60
HEART_RATE_NORMAL_MAX=100
HEART_RATE_TACHYCARDIA_THRESHOLD=100
HEART_RATE_BRADYCARDIA_THRESHOLD=60
```

### 4. Testing
- Test HealthKit authorization flow on real device
- Test heart rate data reading
- Test CloudEvent transmission
- Test heart rate analysis endpoint
- Test WhatsApp notification flow
- Test anomaly detection scenarios

### 5. WhatsApp Agent Integration (Required for Notifications)
The WhatsApp agent needs to handle the `io.homelab.whatsapp.send.message` event type. Currently, it only handles `messaging.message.received`. You may need to:
- Add handler for `io.homelab.whatsapp.send.message` events
- Or modify to send messages via the messaging service directly
- Or configure the broker routing to handle this event type

## 🔄 Data Flow

1. **Apple Watch** records heart rate → **HealthKit** stores data
2. **iOS App** (background) reads from HealthKit
3. **HealthKitService** sends CloudEvent to **Medical Agent**
4. **Medical Agent** analyzes heart rate:
   - Normal → Returns analysis, stores in DB
   - Anomaly (moderate/severe) → Returns analysis + sends WhatsApp notification
5. **WhatsApp Agent** receives notification → Sends message to doctor

## 📝 Usage Example

### iOS App
```swift
// Request authorization
try await HealthKitService.shared.requestAuthorization()

// Read latest heart rate
let reading = try await HealthKitService.shared.readLatestHeartRate()

// Send to agent
try await HealthKitService.shared.sendHeartRateToAgent(
    reading,
    patientId: "patient-123",
    agentBaseURL: agent.baseURL,
    userToken: user.token
)
```

### Medical Agent API
```bash
# Direct HTTP call
curl -X POST http://agent-medical:8080/analyze/heart-rate \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/json" \
  -d '{
    "patient_id": "patient-123",
    "heart_rate_bpm": 95,
    "context": "resting",
    "device": "Apple Watch Series 9"
  }'

# CloudEvent
curl -X POST http://agent-medical:8080/ \
  -H "ce-type: io.homelab.medical.heart-rate.report" \
  -H "ce-source: /ios-app" \
  -H "Authorization: Bearer token" \
  -H "Content-Type: application/cloudevents+json" \
  -d '{
    "specversion": "1.0",
    "type": "io.homelab.medical.heart-rate.report",
    "source": "/ios-app",
    "data": {
      "patient_id": "patient-123",
      "heart_rate_bpm": 120,
      "context": "resting"
    }
  }'
```

## 🔐 Security Considerations

- All health data encrypted in transit (HTTPS/TLS)
- Patient IDs hashed in audit logs (HIPAA compliance)
- RBAC enforced on all endpoints
- Explicit user consent required for HealthKit access
- Audit trail for all heart rate analyses

## 📊 Monitoring

The medical agent logs:
- `heart_rate_analysis_processed` - When analysis completes
- `whatsapp_notification_sent` - When notification sent
- `heart_rate_anomaly_detected` - When anomaly found

Metrics (via Prometheus):
- `agent_medical_requests_total{role, status}` - Total requests
- `agent_medical_access_denied_total{reason}` - Access denied count

## 🚀 Next Steps

1. Add HealthKit capability to Xcode project (see above)
2. Test on real iPhone + Apple Watch device
3. Configure WhatsApp broker URL
4. Set up WhatsApp agent to handle notifications
5. Add UI components for health monitoring
6. Test end-to-end flow
7. Deploy to production

