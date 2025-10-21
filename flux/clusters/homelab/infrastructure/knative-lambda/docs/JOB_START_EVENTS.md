# 🚀 Job Start Events Flow

## Overview

This document describes the new job start event flow that uses RabbitMQ queues for improved scalability and reliability.

## 🔄 Event Flow Architecture

```
Build Event → network.notifi.lambda.job.start → RabbitMQ (kaniko-jobs) → Trigger → knative-lambda-builder
```

### Components

1. **Event Source**: External systems sending `network.notifi.lambda.job.start` events
2. **RabbitMQ Queue**: `kaniko-jobs` queue for event buffering
3. **Knative Trigger**: Routes events from queue to builder service
4. **knative-lambda-builder**: Processes events and creates Kaniko jobs asynchronously

## 📋 Event Structure

### Job Start Event Data

```json
{
  "specversion": "1.0",
  "id": "uuid",
  "source": "network.notifi.kaniko-jobs",
  "subject": "parser-id",
  "type": "network.notifi.lambda.job.start",
  "time": "2024-01-01T00:00:00Z",
  "data": {
    "third_party_id": "0307ea43639b4616b044d190310a26bd",
    "parser_id": "0197ad6c10b973b2b854a0e652155b7e",
    "correlation_id": "uuid",
    "job_name": "kaniko-job-0307ea43639b4616b044d190310a26bd-0197ad6c10b973b2b854a0e652155b7e",
    "parameters": {
      "priority": 1,
      "build_type": "container",
      "runtime": "nodejs18",
      "source_url": "https://github.com/notifi/parsers/0197ad6c10b973b2b854a0e652155b7e",
      "build_timeout": 1800,
      "environment": {
        "NODE_ENV": "production",
        "BUILD_MODE": "optimized"
      }
    },
    "priority": 1
  },
  "datacontenttype": "application/json"
}
```

## 🏗️ Infrastructure Components

### 1. RabbitMQ Queue

**Queue Name**: `kaniko-jobs`

**Configuration**:
- **Type**: Quorum queue (high availability)
- **Durability**: Durable
- **TTL**: 24 hours
- **Max Length**: 10,000 messages
- **Overflow**: Drop oldest messages
- **Dead Letter Exchange**: `dlx`

### 2. Knative Trigger

**Trigger Name**: `kaniko-jobs-trigger`

**Filtering**:
- **Event Type**: `network.notifi.lambda.job.start`
- **Source**: `network.notifi.kaniko-jobs`

**Sink**: `knative-lambda-builder` service

### 3. Dead Letter Queue

**Queue Name**: `kaniko-jobs.dlq`

**Purpose**: Collects failed events for manual inspection and retry

## ⚡ Processing Flow

### 1. Event Reception
- External system sends job start event to RabbitMQ broker
- Event is routed to `kaniko-jobs` queue based on routing key

### 2. Queue Processing
- RabbitMQ source listens to `kaniko-jobs` queue
- Events are delivered to Knative trigger

### 3. Event Filtering
- Trigger filters events by type and source
- Only matching events are forwarded to builder

### 4. Async Job Creation
- Builder receives event and parses job start data
- Converts to build request and queues async job creation
- Returns immediate response: `{"status": "queued"}`

### 5. Worker Processing
- Async job creator workers pick up job creation requests
- Create Kaniko jobs in parallel
- Handle retries and conflict resolution

## 🧪 Testing

### Send Test Events

Use the provided test script:

```bash
# Activate virtual environment
source .venv/bin/activate

# Send job start events
python3 tests/create-job-start-event.py
```

### Monitor Processing

1. **Check Queue Status**:
   ```bash
   kubectl get queue -n rabbitmq-dev
   ```

2. **Monitor Builder Logs**:
   ```bash
   kubectl logs -f deployment/knative-lambda-builder -n knative-lambda-dev
   ```

3. **Check Async Job Creator Stats**:
   ```bash
   curl http://knative-lambda-builder-dev.knative-lambda-dev.svc.cluster.local/async-jobs/stats
   ```

## 📊 Monitoring

### Metrics

- **Queue Depth**: `rabbitmq_queue_messages_ready`
- **Job Creation Rate**: `kaniko_job_creation_success_total`
- **Error Rate**: `kaniko_job_creation_failure_total`
- **Processing Duration**: `job_creation_duration_seconds`

### Alerts

- **High Queue Depth**: > 1,000 messages
- **High Error Rate**: > 5% failure rate
- **Long Processing Time**: > 30 seconds average

## 🔧 Configuration

### RabbitMQ Settings

```yaml
rabbitmq:
  clusterName: "rabbitmq-cluster"
  namespace: "rabbitmq-dev"
  connectionSecretName: "rabbitmq-connection"
  queues:
    kanikoJobs:
      name: "kaniko-jobs"
      durable: true
      autoDelete: false
      messageTtl: 86400000
      maxLength: 10000
      overflow: "drop-head"
```

### Async Job Creator Settings

```yaml
asyncJobCreator:
  workerCount: 5
  maxRetries: 3
  retryDelay: 100ms
  maxQueueSize: 100
```

## 🚨 Troubleshooting

### Common Issues

1. **Events Not Processing**
   - Check RabbitMQ queue depth
   - Verify trigger configuration
   - Check builder service logs

2. **High Error Rate**
   - Monitor dead letter queue
   - Check Kubernetes API connectivity
   - Review job creation logs

3. **Slow Processing**
   - Increase worker count
   - Check resource limits
   - Monitor queue depth

### Debug Commands

```bash
# Check queue status
kubectl get queue kaniko-jobs -n rabbitmq-dev -o yaml

# Check trigger status
kubectl get trigger kaniko-jobs-trigger -n knative-lambda-dev -o yaml

# Check builder service
kubectl get ksvc knative-lambda-builder-dev -n knative-lambda-dev

# Check async job creator stats
curl -s http://knative-lambda-builder-dev.knative-lambda-dev.svc.cluster.local/async-jobs/stats | jq
```

## 🔄 Migration

### From Direct Events to Queue-Based

1. **Phase 1**: Deploy new infrastructure
   - RabbitMQ queues
   - Knative triggers
   - Updated builder service

2. **Phase 2**: Dual Processing
   - Accept both direct and queue-based events
   - Monitor both flows

3. **Phase 3**: Switch to Queue-Only
   - Disable direct event processing
   - Monitor queue-based processing

## 📈 Benefits

1. **Scalability**: Queue buffers traffic spikes
2. **Reliability**: Dead letter queue for failed events
3. **Parallelism**: Async job creation with worker pool
4. **Monitoring**: Better observability and metrics
5. **Resilience**: Automatic retries and error handling
