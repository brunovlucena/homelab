# Security Implementation with Observability

This package provides comprehensive security validation and protection mechanisms for the Knative Lambda service, enhanced with full observability capabilities including metrics, tracing, and structured logging.

## 🛡️ Features

### Security Validation
- **Input Validation**: Comprehensive validation of user inputs with threat detection
- **Image Name Validation**: Docker image name format and security validation
- **Namespace Validation**: Kubernetes namespace format and reserved namespace detection
- **Event Data Validation**: CloudEvent data security scanning
- **ID Validation**: Identifier format and security validation

### Threat Detection
- **SQL Injection Detection**: Pattern-based SQL injection attack detection
- **XSS Attack Detection**: Cross-site scripting attack pattern recognition
- **Path Traversal Detection**: Directory traversal attack prevention
- **Malicious Content Detection**: Generic malicious content pattern matching
- **Threat Level Classification**: Low, Medium, High, and Critical threat levels

### Observability Integration
- **Structured Logging**: JSON-formatted logs with correlation IDs and trace context
- **Prometheus Metrics**: Comprehensive security metrics collection
- **OpenTelemetry Tracing**: Distributed tracing with security context
- **Security Event Recording**: Detailed security event tracking
- **Performance Monitoring**: Security validation performance metrics

## 📊 Metrics

The security implementation exposes the following Prometheus metrics:

### Validation Metrics
- `security_validations_total`: Total number of security validations
  - Labels: `event_type`, `operation`, `status`, `threat_level`
- `security_validation_failures_total`: Total number of validation failures
  - Labels: `event_type`, `operation`, `threat_level`
- `security_validation_duration_seconds`: Validation duration histogram
  - Labels: `event_type`, `operation`, `status`

### Threat Detection Metrics
- `security_threats_detected_total`: Total number of threats detected
  - Labels: `event_type`, `operation`, `threat_level`
- `security_threats_by_level_total`: Threats grouped by threat level
  - Labels: `threat_level`
- `security_critical_threats_total`: Critical threats counter
  - Labels: `event_type`, `operation`

### Example Queries
```promql
# Security validation success rate
rate(security_validations_total{status="success"}[5m]) / rate(security_validations_total[5m])

# Threat detection rate by type
rate(security_threats_detected_total[5m])

# Validation performance (P95)
histogram_quantile(0.95, rate(security_validation_duration_seconds_bucket[5m]))

# Critical threats per minute
rate(security_critical_threats_total[5m]) * 60
```

## 🔍 Tracing

Each security validation operation creates OpenTelemetry spans with the following attributes:

### Span Attributes
- `security.input_length`: Length of input being validated
- `security.input_type`: Type of input (string, image_name, namespace, etc.)
- `security.validation_type`: Type of validation being performed
- `security.image_name`: Docker image name (for image validation)
- `security.namespace`: Kubernetes namespace (for namespace validation)
- `security.id_length`: Length of ID (for ID validation)
- `security.data_type`: Type of event data (for event data validation)

### Example Trace
```
security.validate_input
├── security.input_length: 25
├── security.input_type: string
└── security.validation_type: input_validation
    ├── containsSQLInjection
    ├── containsXSS
    └── containsPathTraversal
```

## 📝 Logging

Security events are logged with structured JSON format:

### Log Levels
- **Info**: Successful validations and general security events
- **Error**: Validation failures and security threats
- **Warn**: Reserved namespace usage and performance issues

### Log Fields
- `event_type`: Type of security event
- `operation`: Specific operation being performed
- `status`: Success, failure, or threat_detected
- `threat_level`: Low, Medium, High, or Critical
- `duration`: Validation duration
- `content_length`: Length of content being validated
- `error`: Error message for failures

## 🚨 Alerting

The security implementation includes comprehensive alerting rules:

### Critical Alerts
- **Critical Security Threats**: Any critical threat level detection
- **SQL Injection Attempts**: SQL injection pattern detection
- **XSS Attack Attempts**: Cross-site scripting detection
- **Security Validation Service Down**: No validations recorded

### Warning Alerts
- **High Security Threat Rate**: Elevated threat detection rate
- **High Validation Failure Rate**: >10% validation failure rate
- **Performance Degradation**: P95 validation time >1 second
- **Unusual Activity Pattern**: Significant deviation from baseline

### Alert Configuration
Alerts are configured in `20-platform/services/prometheus/deploy/security-alerts.yaml` and include:
- Severity levels (critical, high, warning)
- Category labels for filtering
- Runbook URLs for incident response
- Appropriate thresholds and time windows

## 📈 Dashboard

A comprehensive Grafana dashboard is available at `20-platform/services/dashboards/deploy/dashboards/notifi/security-monitoring.json` with panels for:

### Key Metrics
- Security validations rate
- Threat detection rate
- Validation duration (P95)
- Validation failure rate
- Threats by threat level
- Critical threats rate

### Detailed Views
- Validations by type (input, image, namespace, event data, ID)
- Threats by attack type (SQL injection, XSS, path traversal, malicious content)
- Validation failure rate percentage
- Total validation counts

## 🧪 Testing

Comprehensive tests are included in `security_test.go`:

### Test Coverage
- **Unit Tests**: Individual validation method testing
- **Integration Tests**: Observability integration testing
- **Performance Tests**: Validation performance benchmarking
- **Concurrency Tests**: Thread-safe operation verification
- **Threat Detection Tests**: Attack pattern detection validation

### Running Tests
```bash
# Run all security tests
go test ./internal/security -v

# Run with coverage
go test ./internal/security -cover

# Run benchmarks
go test ./internal/security -bench=.

# Run specific test
go test ./internal/security -run TestSecurityValidator_ValidateInput_WithObservability
```

## 🔧 Usage

### Basic Usage
```go
// Create observability instance
obs, err := observability.New(observability.Config{
    ServiceName:    "my-service",
    ServiceVersion: "1.0.0",
    Environment:    "production",
    MetricsEnabled: true,
    TracingEnabled: true,
})

// Create security validator
validator := security.NewSecurityValidator(obs)

// Validate input with context
ctx := context.Background()
result := validator.ValidateInput(ctx, userInput)

if !result.Valid {
    // Handle validation failure
    log.Printf("Validation failed: %v", result.Error)
    return
}
```

### Advanced Usage
```go
// Validate multiple types
inputResult := validator.ValidateInput(ctx, userInput)
imageResult := validator.ValidateImageName(ctx, imageName)
namespaceResult := validator.ValidateNamespace(ctx, namespace)
eventResult := validator.ValidateEventData(ctx, eventData)
idResult := validator.ValidateID(ctx, id)

// Check for warnings
for _, warning := range namespaceResult.Warnings {
    log.Printf("Warning: %s", warning)
}

// Access validation details
if inputResult.Details != nil {
    // Process additional validation details
}
```

## 🏗️ Architecture

### Security Validation Flow
```
Input → Validation → Threat Detection → Result
  ↓         ↓              ↓            ↓
Logging → Metrics → Tracing → Observability
```

### Threat Detection Pipeline
```
Input → Pattern Matching → Threat Classification → Alerting
  ↓           ↓                    ↓                ↓
SQL/XSS → Path Traversal → Malicious Content → Critical/High/Medium/Low
```

### Observability Integration
```
Security Event → Structured Log → Prometheus Metric → OpenTelemetry Span
      ↓              ↓                ↓                    ↓
Event Type → JSON Fields → Counter/Histogram → Trace Attributes
```

## 🔒 Security Best Practices

### Input Validation
- Always validate all user inputs
- Use context-aware validation
- Implement defense in depth
- Log validation failures for analysis

### Threat Detection
- Use pattern-based detection
- Implement threat level classification
- Monitor for unusual patterns
- Alert on critical threats immediately

### Observability
- Log all security events
- Track validation performance
- Monitor threat trends
- Use correlation IDs for tracing

### Performance
- Optimize validation patterns
- Use efficient regex matching
- Implement caching where appropriate
- Monitor validation latency

## 📚 References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [Docker Security Best Practices](https://docs.docker.com/develop/security-best-practices/)
- [OpenTelemetry Security](https://opentelemetry.io/docs/concepts/security/)
- [Prometheus Security](https://prometheus.io/docs/operating/security/) 