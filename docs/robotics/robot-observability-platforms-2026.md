# 🤖 Robot Observability Platforms 2026 - Homelab Integration Guide

> **Part of**: [Homelab Architecture](../architecture/CLUSTER_ARCHITECTURE.md)  
> **Related**: [Observability Architecture](../observability/README.md) | [MCP Observability](../architecture/mcp-observability.md)  
> **Last Updated**: January 2026

---

## 📋 Executive Summary

This document provides a comprehensive overview of **robot observability platforms available in 2026** and how to integrate them with the existing homelab observability infrastructure (Grafana, Prometheus, Loki, Tempo, Alloy).

**Key Findings**:
- **USA Commercial Platforms**: Ferronyx, Formant, Sift ($17.5M Series A), Ember Robotics (YC), WiseVision/WiseOS, Foxglove, Canonical
- **China Commercial Platforms**: Hongdian, KUKA iiQoT, Deep Robotics, Guozi Robotics, SEER Robotics, Inovance Technology
- **General Observability**: Datadog, Dynatrace, Honeycomb, Splunk, AppDynamics, Elastic, New Relic, Grafana Labs (adaptable for robotics)
- **Open Source Tools**: ROS2 tracing, OpenTelemetry integration, system metrics collectors
- **Research Frameworks**: RTAMT, LumiMAS, PEERNet, BlazeAIoT
- **Total Platforms Found**: 20+ companies/platforms globally
- **Integration Strategy**: Extend existing homelab stack with ROS2 exporters and robot-specific dashboards

---

## 📊 Platform Comparison Table

| Platform | Region | Type | Focus | ROS2 Support | Key Differentiator |
|----------|--------|------|-------|--------------|-------------------|
| **Ferronyx** | USA | Commercial | ROS2 Production | ✅ Native | Automated RCA (MTTR: 3-4h → 12-15min) |
| **Formant** | USA | Commercial | Fleet Management | ✅ ROS1/ROS2 | WebRTC teleoperation, on-demand ingestion |
| **Sift** | USA | Commercial | Hardware Sensors | ⚠️ Partial | $17.5M Series A, aerospace/defense focus |
| **Ember Robotics** | USA | Commercial | Edge Diagnostics | ✅ ROS | YC-backed, low-connectivity environments |
| **WiseVision/WiseOS** | Global | Open-Core | ROS2 Platform | ✅ Native | Complete ROS2 ops platform, InfluxDB backend |
| **Foxglove** | USA | Commercial | Visualization | ✅ Native | Live streaming, .mcap/.db3 replay |
| **Canonical** | Global | Beta | Ubuntu/ROS2 | ✅ Native | Prometheus/Grafana/Loki integration |
| **Hongdian** | China | Commercial | Industrial Arms | ⚠️ Gateway | 4G/5G gateway, RS232/RS485 support |
| **KUKA iiQoT** | China/Global | Commercial | Industrial Fleets | ⚠️ Partial | Condition monitoring, energy visualization |
| **Deep Robotics** | China | Commercial | Inspection | ⚠️ Custom | Multi-robot systems, infrastructure monitoring |
| **Guozi Robotics** | China | Commercial | Infrastructure | ⚠️ Custom | Tunnel/infrastructure inspection |
| **SEER Robotics** | China | Commercial | Logistics | ⚠️ Custom | Warehousing, navigation, scheduling |
| **Inovance** | China | Commercial | Industrial Auto | ⚠️ Control | Motion control, PLCs, servo drives |
| **Datadog** | USA | General | Cloud Observability | ⚠️ Adaptable | Cloud-scale, infrastructure monitoring |
| **Dynatrace** | Global | General | AI Monitoring | ⚠️ Adaptable | AI-assisted RCA, full-stack |
| **Honeycomb** | USA | General | Application Debug | ⚠️ Adaptable | Strong tracing, microservices |

**Total Platforms**: 20+ companies globally

---

## 🎯 Available Robot Observability Platforms (2026)

### 1. Commercial Platforms - USA

#### **Ferronyx** ⭐ Recommended for ROS2
- **Location**: USA
- **Purpose**: Real-time ROS2 observability with automated root cause analysis
- **Key Features**:
  - ROS2 nodes, topics, actions monitoring
  - Infrastructure metrics (CPU, memory, network)
  - Automated RCA (reduces MTTR from 3-4 hours to 12-15 minutes)
  - Sensor health and drift detection
  - Deployment/version tracking
- **Integration**: Can export to Prometheus/Grafana
- **Use Case**: Production ROS2-based robots, fleet management
- **Source**: [ROS Discourse](https://discourse.openrobotics.org/t/ferronyx-real-time-ros2-observability-automated-rca/51747)

#### **Formant**
- **Location**: USA
- **Purpose**: Cloud-based fleet observability and teleoperation
- **Key Features**:
  - Real-time telemetry and video streams
  - ROS1 & ROS2 support via Formant Agent
  - Fleet-level dashboards
  - WebRTC teleoperation
  - On-demand data ingestion (reduces costs by ~80%)
- **Integration**: Agent-based, can forward to external systems
- **Use Case**: Field robots, outdoor systems, teleoperation needs
- **Source**: [Formant Docs](https://docs.formant.io/docs/fleet-observability)

#### **Sift** 💰 Series A $17.5M
- **Location**: El Segundo, CA, USA
- **Purpose**: Unified observability for hardware sensor data
- **Key Features**:
  - Performance optimization
  - Safety assurance
  - Predictive maintenance
  - Multi-modal data (LiDAR, IMU, video)
  - Hardware lifecycle from design through operations
  - High-frequency telemetry, root-cause, anomaly detection
- **Customers**: Astranis, K2 Space (aerospace & robotics)
- **Use Case**: Aerospace, defense, robotics, energy systems
- **Source**: [Sift Stack](https://www.siftstack.com/industry/robotics) | [PR Newswire](https://www.prnewswire.com/news-releases/sift-raises-17-5m-series-a-to-propel-the-future-of-machine-innovation-302180575.html)

#### **Ember Robotics** 🚀 Y Combinator
- **Location**: San Francisco, CA, USA
- **Purpose**: Edge-first diagnostics for robotics and embedded systems
- **Key Features**:
  - Edge-first diagnostics, sensor health
  - Fleet-wide observability in low/no connectivity
  - ROS integration
  - Camera support (RealSense, Luxonis, etc.)
  - Designed to scale from prototype → fleets
  - Works under tough field conditions
- **Use Case**: Field robots, embedded systems, rugged environments
- **Source**: [Y Combinator](https://www.ycombinator.com/launches/MlX-ember-robotics-datadog-for-robots) | [Ember Robotics](https://www.emberrobotics.com/)

#### **WiseVision / WiseOS** ⭐ Open-Core
- **Location**: USA/Global
- **Purpose**: Complete ROS2-based operational platform
- **Key Features**:
  - Continuous telemetry ("Data Black Box") powered by InfluxDB
  - Grafana-based dashboards, real-time alerts
  - Bridges ROS2, IoT sensors, LoRaWAN
  - Role-based dashboards, audit logs
- **Licensing**: Open-core (MPL-2.0 core, commercial Pro/enterprise)
- **Integration**: Native ROS2, Grafana, InfluxDB
- **Use Case**: ROS2 fleet management, IoT integration
- **Source**: [WiseVision](https://www.wisevision.tech/wiseos)

#### **Foxglove**
- **Location**: USA
- **Purpose**: Visualization & observability for robotics developers
- **Key Features**:
  - Live streaming and replay of `.mcap`, `.db3`, ROS-bag files
  - Ros-Foxglove bridge for live ROS2 data over WebSocket
  - Supports complex data types (images, point clouds)
- **Licensing**: Free for individuals/academics, paid for teams/enterprise
- **Use Case**: Development, debugging, data visualization
- **Source**: [ROS Docs](https://docs.ros.org/en/rolling/Related-Projects/Visualizing-ROS-2-Data-With-Foxglove.html)

#### **Canonical (Ubuntu) Observability Stack**
- **Location**: USA/Global
- **Purpose**: Observability stack for ROS2 devices via Ubuntu
- **Key Features**:
  - Integration with Prometheus, Grafana, Loki
  - Fleet ROS2 robots remote monitoring
  - System logs, alerts (battery, network)
  - Device onboarding & self-hosted infrastructure
- **Status**: Beta open-source stack
- **Use Case**: Ubuntu-based ROS2 deployments
- **Source**: [Ubuntu Blog](https://ubuntu.com/blog/roscon-2025)

### 2. Commercial Platforms - China 🇨🇳

#### **Hongdian Industrial Robot Solution**
- **Location**: China
- **Purpose**: Industrial robot arm monitoring & gateway + cloud backend
- **Key Features**:
  - Computing gateway (4G/5G)
  - Collects field data (RS232/RS485/RJ45)
  - Custom models support
  - Predictive/preventive maintenance
  - Anomaly detection
  - Centralized data aggregation
- **Use Case**: Industrial robot arms, manufacturing
- **Source**: [Hongdian](https://www.hongdian.com/en/solutions/industrial-robot.html)

#### **KUKA iiQoT** (China / Global)
- **Location**: China/Global
- **Purpose**: Industrial IoT / cloud software for robot fleets
- **Key Features**:
  - Condition monitoring, predictive maintenance
  - Diagnostics, edge-gateway to cloud architecture
  - Remote fleet visibility
  - Cycle-time / energy visualization
  - Alarms & fault detection
- **Use Case**: Industrial robot fleets, manufacturing
- **Source**: [KUKA China](https://www.kuka.cn/en-cn/products/robotics-systems/software/cloud-software/iiqot-robot-condition-monitoring)

#### **Deep Robotics**
- **Location**: China
- **Purpose**: Autonomous & inspection robotics + multi-robot systems
- **Key Features**:
  - Inspection robots (robot dogs)
  - Real-time monitoring of infrastructure (utility tunnels)
  - Durable hardware (IP67, wide temp range)
  - Collaborative dispatch from centralized platform
  - Continuous monitoring, alerting, remote control
- **Use Case**: Infrastructure inspection, utility monitoring
- **Source**: [Access Newswire](https://www.accessnewswire.com/newsroom/en/computers-technology-and-internet/deep-robotics-unveils-multi-robot-collaborative-system-ushering-a-1069948)

#### **Guozi Robotics**
- **Location**: China
- **Purpose**: Mobile/inspection robotics for infrastructure
- **Key Features**:
  - Deployed in Shenzhen-Zhongshan Link tunnels
  - Inspection robots monitoring safety, structural integrity
  - Central management platform for scheduling and response
  - Real-world inspection observability
  - Sensors, environmental monitoring
- **Use Case**: Infrastructure inspection, tunnel monitoring
- **Source**: [PR Newswire](https://www.prnewswire.com/news-releases/guozi-robotics-powers-the-shenzhen-zhongshan-link-a-milestone-in-smart-infrastructure-maintenance-302216493.html)

#### **SEER Robotics** (Shanghai XianGong Intelligent)
- **Location**: Shanghai, China
- **Purpose**: Mobile robots, control systems, navigation & robot software platforms
- **Key Features**:
  - Robots + controllers + software/digital platforms
  - Navigation, mapping, scheduling & dispatch
  - Platform component functions as observability & management layer
  - Tracks robot performance, location, status
  - Manages fleet & tasks
- **Use Case**: Warehousing, logistics, industrial settings
- **Source**: [Wikipedia](https://zh.wikipedia.org/wiki/%E4%BB%99%E5%B7%A5%E6%99%BA%E8%83%BD)

#### **Inovance Technology**
- **Location**: China
- **Purpose**: Industrial automation, robotics control systems
- **Key Features**:
  - Motion control, PLCs, servo drives, robots
  - "Smart software" for robotics & automation
  - Control hardware + software stack enables observability
  - Data access + control
- **Use Case**: Industrial automation, manufacturing
- **Source**: [Wikipedia](https://en.wikipedia.org/wiki/Inovance)

### 3. General Observability Platforms (Adaptable for Robotics)

#### **Datadog**
- **Location**: USA
- **Purpose**: Cloud-scale observability
- **Features**: Infrastructure, logs, traces, performance metrics
- **Use Case**: Cloud-hybrid robotics systems
- **Source**: [Wikipedia](https://en.wikipedia.org/wiki/Datadog)

#### **Dynatrace**
- **Location**: USA/Global
- **Purpose**: AI-assisted monitoring, root cause analysis
- **Features**: Full-stack observability
- **Use Case**: Enterprise robotics deployments

#### **Honeycomb**
- **Location**: USA
- **Purpose**: Observability and debugging for live applications
- **Features**: Strong in tracing, metrics, microservices
- **Use Case**: Complex robotics architectures

#### **Splunk, AppDynamics (Cisco), Elastic NV, New Relic, Grafana Labs**
- **Location**: USA/Global
- **Purpose**: Logs, APM, infrastructure monitoring
- **Features**: Dashboard/reporting, ML/AI-based anomaly detection
- **Use Case**: General infrastructure monitoring (can be adapted for robotics)

---

### 4. Open Source ROS2 Observability Tools

#### **ros2_tracing** (Official ROS2)
- **Purpose**: Distributed tracing for ROS2 applications
- **Features**:
  - Low-overhead tracing (LTTng-based)
  - Message flow analysis
  - Callback timing
  - CPU time tracking
- **Integration**: Exports CTF traces, can be processed by tracetools_analysis
- **Source**: [ROS Index](https://index.ros.org/r/ros2_tracing/)

#### **system_metrics_collector**
- **Purpose**: OS-level metrics for ROS2 processes
- **Features**:
  - CPU, memory, message age
  - Statistics (min, max, std dev)
  - ROS2 process monitoring
- **Integration**: Can export to Prometheus
- **Source**: [ROS Index](https://index.ros.org/r/system_metrics_collector/)

#### **ros-opentelemetry** (Community)
- **Purpose**: ROS2 integration with OpenTelemetry
- **Features**:
  - Standard OTEL traces, metrics, logs
  - C++ and Python support
  - Integration with Jaeger, Prometheus, Grafana
- **Integration**: Native OTEL → Tempo/Loki/Prometheus
- **Source**: [GitHub](https://github.com/szobov/ros-opentelemetry)

#### **tracetools_analysis**
- **Purpose**: Post-mortem trace analysis
- **Features**:
  - Callback duration extraction
  - Message flow patterns
  - Per-thread/CPU usage
- **Integration**: Processes CTF traces from ros2_tracing
- **Source**: [ROS Index](https://index.ros.org/r/tracetools_analysis/)

---

### 5. Research Frameworks

#### **RTAMT** (Runtime Robustness Monitors)
- **Purpose**: Temporal logic monitoring (STL)
- **Features**:
  - Real-time property verification
  - Safety/invariant violation detection
  - ROS and MATLAB/Simulink integration
- **Use Case**: Safety-critical systems, formal verification
- **Source**: [arXiv:2501.18608](https://arxiv.org/abs/2501.18608)

#### **LumiMAS** (Multi-Agent Observability)
- **Purpose**: Observability for multi-agent systems
- **Features**:
  - Agent action logging
  - Cross-agent anomaly detection
  - Failure explanation across agents
- **Use Case**: Collaborative robots, swarm systems
- **Source**: [arXiv:2508.12412](https://arxiv.org/abs/2508.12412)

#### **PEERNet** (Performance Profiling)
- **Purpose**: Networked robotic systems profiling
- **Features**:
  - Sensor/network/ML pipeline latency
  - Data rate measurement
  - Performance bottlenecks
- **Use Case**: Distributed robot systems
- **Source**: [arXiv:2409.06078](https://arxiv.org/abs/2409.06078)

#### **BlazeAIoT** (Edge/Fog/Cloud)
- **Purpose**: Distributed robotics monitoring
- **Features**:
  - Edge/fog/cloud integration
  - Dynamic service allocation
  - Multi-tier monitoring
- **Use Case**: Edge robotics, IoT integration
- **Source**: [arXiv:2601.06344](https://arxiv.org/abs/2601.06344)

---

## 🏗️ Homelab Integration Architecture

### Current Homelab Observability Stack

```
┌─────────────────────────────────────────────────────────────────┐
│              HOMELAB OBSERVABILITY STACK                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  METRICS (Prometheus)                                          │
│  ├─ Node Exporter (system metrics)                             │
│  ├─ Application metrics                                        │
│  └─ Custom exporters                                           │
│                                                                 │
│  LOGS (Loki)                                                   │
│  ├─ Alloy DaemonSet (log collection)                           │
│  ├─ Structured JSON logs                                       │
│  └─ Kubernetes events                                          │
│                                                                 │
│  TRACES (Tempo)                                                │
│  ├─ OpenTelemetry traces                                       │
│  ├─ Distributed tracing                                         │
│  └─ Exemplar linking                                           │
│                                                                 │
│  VISUALIZATION (Grafana)                                       │
│  ├─ Unified dashboards                                         │
│  ├─ Alerting                                                   │
│  └─ MCP integration                                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Proposed Robot Observability Extension

```
┌─────────────────────────────────────────────────────────────────┐
│           ROBOT OBSERVABILITY EXTENSION                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ROBOT AGENTS (Edge/On-Robot)                                  │
│  ├─ ROS2 Nodes (ros2_tracing enabled)                         │
│  ├─ ros-opentelemetry (OTEL instrumentation)                  │
│  ├─ system_metrics_collector (OS metrics)                     │
│  └─ Robot-specific exporters                                   │
│                                                                 │
│  DATA COLLECTION LAYER                                         │
│  ├─ ROS2 Topic Exporter → Prometheus                          │
│  ├─ ROS2 Trace Exporter → Tempo (via OTEL)                     │
│  ├─ Robot Logs → Loki (via Alloy)                              │
│  └─ Sensor Data → Time Series DB                              │
│                                                                 │
│  HOMELAB OBSERVABILITY STACK (Existing)                        │
│  ├─ Prometheus (metrics storage)                               │
│  ├─ Loki (log aggregation)                                     │
│  ├─ Tempo (trace storage)                                      │
│  └─ Grafana (visualization)                                    │
│                                                                 │
│  ROBOT-SPECIFIC DASHBOARDS                                     │
│  ├─ ROS2 Node Health                                          │
│  ├─ Topic Message Rates                                        │
│  ├─ Callback Latency                                           │
│  ├─ Sensor Health                                              │
│  ├─ Battery/Energy Monitoring                                 │
│  └─ Fleet Overview                                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔧 Integration Strategy

### Phase 1: ROS2 Metrics Collection

#### 1.1 Deploy ROS2 Prometheus Exporter

```yaml
# flux/clusters/studio/deploy/08-observability/ros2-exporter.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: ros2-prometheus-exporter
  namespace: observability
spec:
  selector:
    matchLabels:
      app: ros2-exporter
  template:
    metadata:
      labels:
        app: ros2-exporter
    spec:
      containers:
      - name: exporter
        image: ros2-prometheus-exporter:latest
        ports:
        - containerPort: 9090
          name: metrics
        env:
        - name: ROS_DOMAIN_ID
          value: "0"
        - name: PROMETHEUS_PORT
          value: "9090"
        volumeMounts:
        - name: ros2-socket
          mountPath: /tmp/ros2
      volumes:
      - name: ros2-socket
        hostPath:
          path: /tmp/ros2
```

#### 1.2 ROS2 Metrics to Collect

```python
# Example metrics from ROS2 exporter
ros2_topic_message_rate{node="robot_controller", topic="/cmd_vel"} 10.5
ros2_topic_latency{node="robot_controller", topic="/cmd_vel"} 0.025
ros2_callback_duration{node="robot_controller", callback="cmd_vel_callback"} 0.015
ros2_node_cpu_usage{node="robot_controller"} 45.2
ros2_node_memory_usage{node="robot_controller"} 128.5
ros2_topic_queue_depth{node="robot_controller", topic="/cmd_vel"} 2
```

#### 1.3 Prometheus Scrape Configuration

```yaml
# Add to Prometheus scrape configs
- job_name: 'ros2-robots'
  scrape_interval: 5s
  static_configs:
    - targets:
      - 'robot-1.local:9090'
      - 'robot-2.local:9090'
      - 'robot-3.local:9090'
      labels:
        robot_type: 'unitree_g1'
        environment: 'homelab'
```

---

### Phase 2: ROS2 Tracing Integration

#### 2.1 Deploy ros-opentelemetry Collector

```yaml
# flux/clusters/studio/deploy/08-observability/ros2-otel-collector.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: ros2-otel-collector
  namespace: observability
spec:
  selector:
    matchLabels:
      app: ros2-otel-collector
  template:
    metadata:
      labels:
        app: ros2-otel-collector
    spec:
      containers:
      - name: collector
        image: ros-opentelemetry-collector:latest
        env:
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "http://alloy.observability.svc.cluster.local:4317"
        - name: ROS_DOMAIN_ID
          value: "0"
```

#### 2.2 Configure Alloy to Forward ROS2 Traces

```yaml
# Update Alloy config to receive ROS2 traces
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

exporters:
  otlp/tempo:
    endpoint: tempo.observability.svc.cluster.local:4317
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [otlp/tempo]
```

---

### Phase 3: Robot-Specific Dashboards

#### 3.1 Grafana Dashboard: ROS2 Node Health

```json
{
  "dashboard": {
    "title": "ROS2 Robot Node Health",
    "panels": [
      {
        "title": "Node CPU Usage",
        "targets": [
          {
            "expr": "ros2_node_cpu_usage{robot=\"$robot\"}",
            "legendFormat": "{{node}}"
          }
        ]
      },
      {
        "title": "Topic Message Rates",
        "targets": [
          {
            "expr": "rate(ros2_topic_message_rate{robot=\"$robot\"}[5m])",
            "legendFormat": "{{topic}}"
          }
        ]
      },
      {
        "title": "Callback Latency (P95)",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, ros2_callback_duration_bucket{robot=\"$robot\"})",
            "legendFormat": "{{callback}}"
          }
        ]
      }
    ]
  }
}
```

#### 3.2 Grafana Dashboard: Robot Fleet Overview

```json
{
  "dashboard": {
    "title": "Robot Fleet Overview",
    "panels": [
      {
        "title": "Active Robots",
        "targets": [
          {
            "expr": "count(up{job=\"ros2-robots\"} == 1)"
          }
        ]
      },
      {
        "title": "Robot Status by Type",
        "targets": [
          {
            "expr": "count by (robot_type) (up{job=\"ros2-robots\"})"
          }
        ]
      },
      {
        "title": "Average Battery Level",
        "targets": [
          {
            "expr": "avg(robot_battery_level{robot_type=\"unitree_g1\"})"
          }
        ]
      }
    ]
  }
}
```

---

### Phase 4: Sensor and Hardware Metrics

#### 4.1 Custom Robot Metrics Exporter

```python
# robot-metrics-exporter.py
import rospy
from prometheus_client import Gauge, Counter, Histogram
import time

# Robot-specific metrics
battery_level = Gauge('robot_battery_level', 'Battery level percentage', ['robot_id'])
motor_temperature = Gauge('robot_motor_temperature', 'Motor temperature', ['robot_id', 'motor'])
sensor_health = Gauge('robot_sensor_health', 'Sensor health status', ['robot_id', 'sensor_type'])
movement_distance = Counter('robot_movement_distance', 'Total distance traveled', ['robot_id'])

def publish_metrics():
    # Collect from ROS2 topics
    battery_data = rospy.wait_for_message('/battery_status', BatteryStatus)
    battery_level.labels(robot_id='g1-001').set(battery_data.percentage)
    
    # Motor temperatures
    motor_data = rospy.wait_for_message('/motor_status', MotorStatus)
    for motor in motor_data.motors:
        motor_temperature.labels(robot_id='g1-001', motor=motor.name).set(motor.temperature)

if __name__ == '__main__':
    rospy.init_node('robot_metrics_exporter')
    # Start Prometheus HTTP server
    start_http_server(9090)
    while not rospy.is_shutdown():
        publish_metrics()
        time.sleep(1)
```

---

## 📊 Key Metrics to Monitor

### ROS2-Specific Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `ros2_topic_message_rate` | Messages per second per topic | < 1 Hz for critical topics |
| `ros2_topic_latency` | End-to-end message latency | > 100ms for real-time topics |
| `ros2_callback_duration` | Callback execution time | P95 > 50ms |
| `ros2_node_cpu_usage` | CPU usage per node | > 80% sustained |
| `ros2_topic_queue_depth` | Pending messages in queue | > 10 messages |
| `ros2_node_memory_usage` | Memory usage per node | > 512MB per node |

### Robot Hardware Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `robot_battery_level` | Battery percentage | < 20% |
| `robot_motor_temperature` | Motor temperature | > 70°C |
| `robot_sensor_health` | Sensor status (0=healthy, 1=degraded, 2=failed) | > 0 |
| `robot_movement_distance` | Total distance traveled | N/A (counter) |
| `robot_network_latency` | Network latency to homelab | > 100ms |

### Fleet-Level Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `robot_fleet_availability` | Percentage of robots online | < 90% |
| `robot_fleet_error_rate` | Error rate across fleet | > 5% |
| `robot_fleet_avg_battery` | Average battery level | < 30% |

---

## 🚨 Alerting Rules

### Prometheus Alert Rules

```yaml
# flux/clusters/studio/deploy/08-observability/robot-alerts.yaml
groups:
- name: robot_alerts
  interval: 30s
  rules:
  - alert: RobotOffline
    expr: up{job="ros2-robots"} == 0
    for: 2m
    annotations:
      summary: "Robot {{ $labels.instance }} is offline"
      description: "Robot has been offline for more than 2 minutes"

  - alert: HighCallbackLatency
    expr: histogram_quantile(0.95, ros2_callback_duration_bucket) > 0.1
    for: 5m
    annotations:
      summary: "High callback latency on {{ $labels.node }}"
      description: "P95 callback latency is {{ $value }}s"

  - alert: LowBattery
    expr: robot_battery_level < 20
    for: 1m
    annotations:
      summary: "Low battery on {{ $labels.robot_id }}"
      description: "Battery level is {{ $value }}%"

  - alert: MotorOverheating
    expr: robot_motor_temperature > 70
    for: 2m
    annotations:
      summary: "Motor overheating on {{ $labels.robot_id }}"
      description: "Motor {{ $labels.motor }} temperature is {{ $value }}°C"

  - alert: SensorDegraded
    expr: robot_sensor_health > 0
    for: 5m
    annotations:
      summary: "Sensor degraded on {{ $labels.robot_id }}"
      description: "Sensor {{ $labels.sensor_type }} health is {{ $value }}"
```

---

## 🔄 Data Flow Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    ROBOT OBSERVABILITY DATA FLOW                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ROBOT (Edge Device)                                           │
│  ├─ ROS2 Nodes (publishing topics)                             │
│  ├─ ros2_tracing (LTTng traces)                               │
│  ├─ ros-opentelemetry (OTEL spans)                             │
│  └─ robot-metrics-exporter (Prometheus metrics)               │
│       │                                                         │
│       ├─ Metrics ────────────────────────────────────────────┐
│       ├─ Traces ───────────────────────────────────────────────┤
│       └─ Logs ────────────────────────────────────────────────┤
│                                                                 │
│  COLLECTION LAYER (Homelab)                                    │
│  ├─ Prometheus (scrapes robot:9090/metrics)                   │
│  ├─ Alloy (receives OTEL traces)                              │
│  └─ Alloy (collects robot logs)                               │
│       │                                                         │
│       ├─ Metrics ────────────────────────────────────────────┐
│       ├─ Traces ───────────────────────────────────────────────┤
│       └─ Logs ────────────────────────────────────────────────┤
│                                                                 │
│  STORAGE LAYER (Homelab)                                       │
│  ├─ Prometheus (metrics storage)                               │
│  ├─ Tempo (trace storage)                                      │
│  └─ Loki (log storage)                                         │
│       │                                                         │
│       └─ Unified Access ───────────────────────────────────────┐
│                                                                 │
│  VISUALIZATION (Grafana)                                      │
│  ├─ Robot Dashboards                                           │
│  ├─ Fleet Overview                                             │
│  ├─ Alerting                                                   │
│  └─ MCP Integration (AI agents can query)                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🛠️ Implementation Steps

### Step 1: Setup ROS2 Exporter (Week 1)

1. **Create ROS2 Prometheus Exporter**
   ```bash
   # Build custom exporter or use existing
   git clone https://github.com/ros2/ros2_metrics_exporter
   cd ros2_metrics_exporter
   docker build -t ros2-prometheus-exporter:latest .
   ```

2. **Deploy to Homelab**
   ```bash
   # Apply Kubernetes manifests
   kubectl apply -f flux/clusters/studio/deploy/08-observability/ros2-exporter.yaml
   ```

3. **Configure Prometheus Scraping**
   ```bash
   # Update Prometheus config
   kubectl edit configmap prometheus-server -n observability
   ```

### Step 2: Integrate ROS2 Tracing (Week 2)

1. **Deploy ros-opentelemetry**
   ```bash
   # Install on robot
   pip install ros-opentelemetry
   ```

2. **Configure OTEL Export**
   ```bash
   # Set environment variables on robot
   export OTEL_EXPORTER_OTLP_ENDPOINT=http://homelab-gateway:4317
   export ROS_DOMAIN_ID=0
   ```

3. **Update Alloy Configuration**
   ```bash
   # Add ROS2 receiver to Alloy
   kubectl edit configmap alloy-config -n observability
   ```

### Step 3: Create Dashboards (Week 3)

1. **Import Grafana Dashboards**
   ```bash
   # Use Grafana API or UI
   grafana-cli dashboard import robot-node-health.json
   grafana-cli dashboard import robot-fleet-overview.json
   ```

2. **Configure Alerting**
   ```bash
   # Apply alert rules
   kubectl apply -f flux/clusters/studio/deploy/08-observability/robot-alerts.yaml
   ```

### Step 4: Robot-Specific Metrics (Week 4)

1. **Deploy Custom Metrics Exporter**
   ```bash
   # Build and deploy robot-specific exporter
   docker build -t robot-metrics-exporter:latest .
   kubectl apply -f robot-metrics-exporter.yaml
   ```

2. **Add Hardware Monitoring**
   - Battery levels
   - Motor temperatures
   - Sensor health
   - Network connectivity

---

## 📚 References

### Commercial Platforms - USA
- **Ferronyx**: [ROS Discourse](https://discourse.openrobotics.org/t/ferronyx-real-time-ros2-observability-automated-rca/51747)
- **Formant**: [Formant Docs](https://docs.formant.io/docs/fleet-observability)
- **Sift**: [Sift Stack](https://www.siftstack.com/industry/robotics) | [Series A Announcement](https://www.prnewswire.com/news-releases/sift-raises-17-5m-series-a-to-propel-the-future-of-machine-innovation-302180575.html)
- **Ember Robotics**: [Y Combinator](https://www.ycombinator.com/launches/MlX-ember-robotics-datadog-for-robots) | [Website](https://www.emberrobotics.com/)
- **WiseVision/WiseOS**: [WiseVision](https://www.wisevision.tech/wiseos)
- **Foxglove**: [ROS Docs](https://docs.ros.org/en/rolling/Related-Projects/Visualizing-ROS-2-Data-With-Foxglove.html)
- **Canonical Observability Stack**: [Ubuntu Blog](https://ubuntu.com/blog/roscon-2025)

### Commercial Platforms - China
- **Hongdian**: [Hongdian Industrial Robot Solution](https://www.hongdian.com/en/solutions/industrial-robot.html)
- **KUKA iiQoT**: [KUKA China](https://www.kuka.cn/en-cn/products/robotics-systems/software/cloud-software/iiqot-robot-condition-monitoring)
- **Deep Robotics**: [Access Newswire](https://www.accessnewswire.com/newsroom/en/computers-technology-and-internet/deep-robotics-unveils-multi-robot-collaborative-system-ushering-a-1069948)
- **Guozi Robotics**: [PR Newswire](https://www.prnewswire.com/news-releases/guozi-robotics-powers-the-shenzhen-zhongshan-link-a-milestone-in-smart-infrastructure-maintenance-302216493.html)
- **SEER Robotics**: [Wikipedia](https://zh.wikipedia.org/wiki/%E4%BB%99%E5%B7%A5%E6%99%BA%E8%83%BD)
- **Inovance Technology**: [Wikipedia](https://en.wikipedia.org/wiki/Inovance)

### Open Source Tools
- **ros2_tracing**: [ROS Index](https://index.ros.org/r/ros2_tracing/)
- **system_metrics_collector**: [ROS Index](https://index.ros.org/r/system_metrics_collector/)
- **ros-opentelemetry**: [GitHub](https://github.com/szobov/ros-opentelemetry)
- **tracetools_analysis**: [ROS Index](https://index.ros.org/r/tracetools_analysis/)

### Research Papers
- **RTAMT**: [arXiv:2501.18608](https://arxiv.org/abs/2501.18608)
- **LumiMAS**: [arXiv:2508.12412](https://arxiv.org/abs/2508.12412)
- **PEERNet**: [arXiv:2409.06078](https://arxiv.org/abs/2409.06078)
- **BlazeAIoT**: [arXiv:2601.06344](https://arxiv.org/abs/2601.06344)

### Homelab Documentation
- [Observability Architecture](../observability/README.md)
- [MCP Observability](../architecture/mcp-observability.md)
- [AI Agent Architecture](../architecture/ai-agent-architecture.md)

---

## 🎯 Next Steps

1. **Evaluate Platforms**: Test Ferronyx or Formant for production use
2. **POC Implementation**: Deploy ROS2 exporter to one robot
3. **Dashboard Creation**: Build initial Grafana dashboards
4. **Alert Configuration**: Set up critical alerts
5. **Fleet Expansion**: Scale to multiple robots
6. **AI Integration**: Connect robot observability to MCP for AI agent queries

---

**Last Updated**: January 2026  
**Maintained by**: SRE Team (Bruno Lucena)
