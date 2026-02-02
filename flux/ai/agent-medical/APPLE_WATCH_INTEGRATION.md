# 🏥 Apple Watch Integration with Agent-Medical

**Integration Guide: Apple Watch → iPhone → Homelab Cluster → WhatsApp Agent**

This document describes how to extract health data from Apple Watch, send it to the agent-medical running in the homelab cluster, analyze it, and trigger WhatsApp notifications when anomalies are detected.

## 📋 Architecture Overview

```
┌─────────────────┐
│  Apple Watch    │
│  (Heart Rate)   │
└────────┬────────┘
         │ HealthKit Sync
         ▼
┌─────────────────┐
│   iPhone 14     │
│  Agent-Medical  │
│     iOS App     │
│                 │
│  - HealthKit    │
│  - Background   │
│    Delivery     │
└────────┬────────┘
         │ CloudEvents (HTTPS)
         ▼
┌─────────────────┐
│  Homelab Cluster│
│  Agent-Medical  │
│                 │
│  - Heart Rate   │
│    Analysis     │
│  - Tool:        │
│    check_heart  │
└────────┬────────┘
         │ CloudEvents (if anomaly)
         ▼
┌─────────────────┐
│ WhatsApp Agent  │
│ (Rust Service)  │
│                 │
│  - Send Message │
│    to Doctor    │
└─────────────────┘
```

## 🔧 Implementation Components

### 1. iOS App - HealthKit Integration

**Location**: `ios-app/AgentChat/AgentChat/Services/HealthKitService.swift`

**Features**:
- Request HealthKit permissions
- Read heart rate data from Apple Watch
- Background delivery for real-time monitoring
- Send data to agent-medical via CloudEvents
- Handle authorization errors gracefully

**Key HealthKit Types**:
- `HKQuantityTypeIdentifierHeartRate` - Heart rate readings
- `HKQuantityTypeIdentifierHeartRateVariabilitySDNN` - HRV (optional)
- `HKQuantityTypeIdentifierRestingHeartRate` - Resting HR (optional)

### 2. Medical Agent - Heart Rate Analysis Tool

**Location**: `src/medical_agent/heart_rate_analyzer.py`

**Features**:
- New endpoint: `POST /analyze/heart-rate`
- Analysis logic:
  - Normal range: 60-100 bpm (resting)
  - Tachycardia: >100 bpm (resting)
  - Bradycardia: <60 bpm (resting)
  - Context-aware (exercise vs resting)
- Integration with patient records for baseline comparison
- Returns analysis result with recommendation

### 3. Medical Agent - WhatsApp Integration

**Location**: `src/medical_agent/main.py` (new function)

**Features**:
- Function: `send_whatsapp_notification()`
- Sends CloudEvent to WhatsApp agent broker
- Event type: `io.homelab.whatsapp.send.message`
- Payload includes: recipient, message, priority

### 4. CloudEvents Flow

#### iOS App → Medical Agent
```json
{
  "type": "io.homelab.medical.heart-rate.report",
  "source": "/ios-app/agent-chat",
  "data": {
    "patient_id": "patient-123",
    "heart_rate_bpm": 85,
    "timestamp": "2024-01-15T10:30:00Z",
    "context": "resting",
    "device": "Apple Watch Series 9"
  }
}
```

#### Medical Agent → WhatsApp Agent
```json
{
  "type": "io.homelab.whatsapp.send.message",
  "source": "/agent-medical/records",
  "data": {
    "recipient_id": "doctor-456",
    "conversation_id": "conv-789",
    "message": "Patient patient-123 has abnormal heart rate: 120 bpm (resting). Please review.",
    "priority": "high"
  }
}
```

## 🚀 Step-by-Step Implementation

### Phase 1: iOS HealthKit Integration

1. **Add HealthKit Capability**
   - Open Xcode project
   - Go to Signing & Capabilities
   - Add "HealthKit" capability
   - Add usage description to Info.plist

2. **Create HealthKitService**
   - Implement authorization request
   - Implement heart rate query
   - Implement background delivery
   - Implement data transmission to agent

3. **Integrate with AgentService**
   - Add method to send heart rate data
   - Handle CloudEvent response
   - Update UI to show health status

### Phase 2: Medical Agent - Heart Rate Analysis

1. **Create Heart Rate Analyzer**
   - Implement analysis logic
   - Add baseline comparison
   - Return structured analysis result

2. **Add Endpoint**
   - `POST /analyze/heart-rate`
   - Accept heart rate data
   - Call analyzer
   - Return analysis

3. **Update Event Handlers**
   - Handle `io.homelab.medical.heart-rate.report` event
   - Process heart rate data
   - Trigger WhatsApp if anomaly detected

### Phase 3: WhatsApp Integration

1. **Add WhatsApp Event Sender**
   - Function to send CloudEvents to WhatsApp broker
   - Configure broker URL
   - Handle errors

2. **Integrate with Heart Rate Analysis**
   - When anomaly detected, send WhatsApp notification
   - Include relevant patient data (sanitized)
   - Log notification in audit trail

### Phase 4: Testing & Monitoring

1. **Test HealthKit Integration**
   - Test on real device
   - Verify background delivery
   - Test data transmission

2. **Test Analysis Logic**
   - Test normal heart rate
   - Test tachycardia
   - Test bradycardia
   - Test edge cases

3. **Test WhatsApp Integration**
   - Verify CloudEvent delivery
   - Verify message received
   - Test error handling

## 🔐 Security & Privacy

### HIPAA Compliance

- **Encryption**: All health data encrypted in transit (HTTPS/TLS)
- **Audit Logging**: All heart rate analyses logged with patient ID hash
- **Access Control**: Only authorized users can access patient data
- **Data Minimization**: Only necessary data sent to cluster

### User Consent

- Explicit HealthKit authorization required
- Clear privacy policy displayed
- User can revoke permissions at any time
- Data transmission is opt-in

## 📊 Data Flow Example

1. **Apple Watch** records heart rate: 95 bpm
2. **HealthKit** syncs to iPhone
3. **iOS App** (background) detects new reading
4. **HealthKitService** queries latest heart rate
5. **AgentService** sends CloudEvent to medical agent:
   ```json
   {
     "type": "io.homelab.medical.heart-rate.report",
     "data": {
       "patient_id": "patient-123",
       "heart_rate_bpm": 95,
       "timestamp": "2024-01-15T10:30:00Z"
     }
   }
   ```
6. **Medical Agent** receives event, analyzes heart rate:
   - Normal range: ✓ (95 bpm is normal)
   - Returns: `{"status": "normal", "bpm": 95, "recommendation": "Continue monitoring"}`
7. **iOS App** receives response, updates UI

### Anomaly Flow

1-5. Same as above, but heart rate is **120 bpm**
6. **Medical Agent** analyzes:
   - Tachycardia detected: ✗ (>100 bpm resting)
   - Returns: `{"status": "tachycardia", "bpm": 120, "recommendation": "Consider contacting doctor"}`
7. **Medical Agent** sends CloudEvent to WhatsApp agent:
   ```json
   {
     "type": "io.homelab.whatsapp.send.message",
     "data": {
       "recipient_id": "doctor-456",
       "message": "Patient patient-123 has elevated heart rate: 120 bpm (resting). Review recommended."
     }
   }
   ```
8. **WhatsApp Agent** receives event, sends message to doctor
9. **iOS App** shows notification: "Heart rate elevated. Doctor notified."

## 📝 Configuration

### iOS App

Add to `Config.swift`:
```swift
enum HealthKit {
    static let heartRateType = HKQuantityTypeIdentifierHeartRate
    static let syncInterval: TimeInterval = 300 // 5 minutes
    static let backgroundDeliveryEnabled = true
}
```

### Medical Agent

Add environment variables:
```bash
WHATSAPP_BROKER_URL=http://default-broker.knative-lambda.svc.cluster.local:80
HEART_RATE_NORMAL_MIN=60
HEART_RATE_NORMAL_MAX=100
HEART_RATE_ALERT_THRESHOLD=100
```

## 🔍 Monitoring

### Metrics

- `agent_medical_heart_rate_checks_total` - Total heart rate checks
- `agent_medical_heart_rate_anomalies_total` - Anomalies detected
- `agent_medical_whatsapp_notifications_sent_total` - WhatsApp notifications sent
- `agent_medical_heart_rate_analysis_duration_seconds` - Analysis latency

### Audit Logs

All heart rate analyses logged with:
- Patient ID (hashed)
- Heart rate value
- Analysis result
- Timestamp
- User ID (hashed if patient)

## ✅ Checklist

- [ ] Add HealthKit capability to iOS app
- [ ] Implement HealthKitService
- [ ] Add heart rate query logic
- [ ] Implement background delivery
- [ ] Add CloudEvent transmission
- [ ] Create heart rate analyzer in medical agent
- [ ] Add `/analyze/heart-rate` endpoint
- [ ] Implement WhatsApp notification sender
- [ ] Add event handler for heart-rate.report
- [ ] Test on real device
- [ ] Test anomaly detection
- [ ] Test WhatsApp integration
- [ ] Add monitoring metrics
- [ ] Update documentation
- [ ] Add error handling
- [ ] Add retry logic for network failures

## 🚧 Future Enhancements

1. **Additional Metrics**: Blood oxygen, steps, sleep data
2. **Machine Learning**: Predictive analysis of heart rate patterns
3. **Baseline Learning**: Personalized normal ranges per patient
4. **Multi-Device Support**: Aggregate data from multiple sources
5. **Real-time Dashboard**: Live heart rate monitoring in web UI
6. **Alert Rules**: Configurable thresholds per patient
7. **Integration with Medical Records**: Compare with historical data

