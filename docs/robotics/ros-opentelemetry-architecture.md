# ROS2 OpenTelemetry Architecture Diagrams

> Based on: [ros-opentelemetry GitHub](https://github.com/szobov/ros-opentelemetry)

---

## 1. Overall Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    ROS2 ROBOT APPLICATION LAYER                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────────┐              ┌──────────────────┐              │
│  │  C++ ROS2 Node   │              │ Python ROS2 Node  │              │
│  │  (RobotControl)  │              │  (TaskProducer)   │              │
│  │                  │              │                   │              │
│  │  - MoveIt2       │              │  - Task Planning │              │
│  │  - Hardware Ctrl │              │  - Coordination   │              │
│  └────────┬─────────┘              └────────┬─────────┘              │
│           │                                  │                         │
│           │  ROS2 Topics/Actions/Services    │                         │
│           │  (with TraceMetadata)            │                         │
│           └──────────────┬───────────────────┘                         │
│                          │                                             │
└──────────────────────────┼─────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────────────────┐
│              ROS-OPENTELEMETRY INSTRUMENTATION LAYER                     │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────────────────┐    ┌──────────────────────────┐         │
│  │ ros_opentelemetry_cpp    │    │ ros_opentelemetry_py     │         │
│  │                          │    │                          │         │
│  │  - setup_tracer()       │    │  - setup_tracer()        │         │
│  │  - inject_trace_context()│   │  - inject_trace_context() │         │
│  │  - extract_trace_context()│  │  - wrap_logger()         │         │
│  │  - RCLCPP_*_TRACED()    │    │                          │         │
│  └──────────┬───────────────┘    └──────────┬───────────────┘         │
│              │                               │                         │
│              └───────────┬──────────────────┘                         │
│                          │                                             │
│              ┌───────────▼───────────┐                                │
│              │  OpenTelemetry SDK    │                                 │
│              │  (C++ & Python)      │                                 │
│              │                      │                                 │
│              │  - Tracer Provider   │                                 │
│              │  - Span Context      │                                 │
│              │  - Trace Propagation │                                 │
│              └───────────┬───────────┘                                │
│                          │                                             │
└──────────────────────────┼─────────────────────────────────────────────┘
                           │
                           │ OTLP (gRPC) Protocol
                           │ Traces + Logs
                           ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    OTLP COLLECTOR (OpenTelemetry)                        │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────┐     │
│  │  OTLP Receiver                                                │     │
│  │  - Receives traces via gRPC (port 4317)                       │     │
│  │  - Receives logs via filelog receiver                        │     │
│  └────────────────────┬──────────────────────────────────────┘     │
│                         │                                              │
│  ┌─────────────────────▼──────────────────────────────────────┐     │
│  │  Processors                                                 │     │
│  │  - Trace parser (extract trace_id, span_id from logs)      │     │
│  │  - Regex parser (parse log format)                          │     │
│  │  - Batch processor                                          │     │
│  └─────────────────────┬──────────────────────────────────────┘     │
│                         │                                              │
│  ┌─────────────────────▼──────────────────────────────────────┐     │
│  │  Exporters                                                   │     │
│  │  - OTLP Exporter → Backend                                   │     │
│  └─────────────────────┬──────────────────────────────────────┘     │
│                           │                                            │
└───────────────────────────┼───────────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    VISUALIZATION BACKEND                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│  │   SigNoz     │  │   Grafana    │  │   Jaeger     │                 │
│  │              │  │              │  │              │                 │
│  │  - Traces    │  │  - Traces    │  │  - Traces    │                 │
│  │  - Logs      │  │  - Logs      │  │  - Logs       │                 │
│  │  - Metrics   │  │  - Metrics   │  │              │                 │
│  └──────────────┘  └──────────────┘  └──────────────┘                 │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Trace Context Propagation Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    DISTRIBUTED TRACE PROPAGATION                         │
└─────────────────────────────────────────────────────────────────────────┘

STEP 1: Python Node (TaskProducer) Creates Trace
─────────────────────────────────────────────────

┌─────────────────────────────────────┐
│  Python ROS2 Node                   │
│  (TaskProducer)                     │
│                                     │
│  from ros_opentelemetry_py import   │
│      setup_tracer,                  │
│      inject_trace_context           │
│                                     │
│  setup_tracer("robot_task_producer")│
│                                     │
│  tracer = trace.get_tracer(__name__)│
│                                     │
│  with tracer.start_as_current_span(│
│      "create_task"):                │
│      # Create task message          │
│      task_msg = TaskMessage()       │
│      task_msg.trace_metadata =      │
│          inject_trace_context()     │◄─── Injects trace_id, span_id
│                                     │     into ROS2 message
│      publisher.publish(task_msg)   │
└──────────────┬──────────────────────┘
               │
               │ ROS2 Topic: /task_queue
               │ Message contains TraceMetadata field
               │
               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  ROS2 Message with TraceMetadata                                        │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  TaskMessage                                                  │     │
│  │  ├─ task_id: "task_123"                                      │     │
│  │  ├─ priority: HIGH                                           │     │
│  │  └─ trace_metadata: TraceMetadata                            │     │
│  │      ├─ trace_id: "a1b2c3d4e5f6..." (32 hex chars)          │     │
│  │      ├─ span_id: "f1e2d3c4..." (16 hex chars)                │     │
│  │      └─ trace_flags: 0x01                                    │     │
│  └──────────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────────────┘
               │
               │ ROS2 DDS Middleware
               │
               ▼
STEP 2: C++ Node (RobotControl) Receives and Extracts Context
───────────────────────────────────────────────────────────────

┌─────────────────────────────────────┐
│  C++ ROS2 Node                      │
│  (RobotControl)                     │
│                                     │
│  #include "ros_opentelemetry_cpp/   │
│      ros_opentelemetry_cpp.hpp"     │
│                                     │
│  setup_tracer("robot_control",      │
│      "collector:4317");             │
│                                     │
│  void callback(TaskMessage::SharedPtr msg) {                         │
│      // Extract trace context       │
│      auto extracted_ctx =          │
│          ros_opentelemetry_cpp::    │
│          extract_trace_context(     │
│              &msg->trace_metadata); │◄─── Extracts trace context
│                                     │     from message
│      auto ctx_token =               │
│          RuntimeContext::Attach(    │
│              extracted_ctx);        │
│                                     │
│      // Create child span           │
│      auto tracer = Provider::       │
│          GetTracerProvider()->      │
│          GetTracer("robot_control");│
│                                     │
│      auto span = tracer->           │
│          StartSpan("execute_task");  │◄─── Child span linked to parent
│                                     │
│      // Execute robot action        │
│      moveit_controller.execute();  │
│                                     │
│      span->End();                   │
│  }                                  │
└──────────────┬──────────────────────┘
               │
               │ Trace spans connected
               │
               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  Distributed Trace Structure                                            │
│                                                                         │
│  Trace ID: a1b2c3d4e5f6...                                             │
│  │                                                                     │
│  ├─ Span: create_task (TaskProducer)                                   │
│  │   ├─ Start: 10:00:00.000                                           │
│  │   ├─ Duration: 50ms                                                │
│  │   └─ Attributes: {task_id: "task_123"}                             │
│  │                                                                     │
│  └─ Span: execute_task (RobotControl) ◄─── Child span                 │
│      ├─ Start: 10:00:00.050                                           │
│      ├─ Duration: 200ms                                               │
│      ├─ Parent: create_task                                           │
│      └─ Attributes: {robot_id: "g1-001"}                              │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Log-Trace Connection

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    LOG-TRACE CORRELATION                                │
└─────────────────────────────────────────────────────────────────────────┘

C++ Node Logging with Trace Context
────────────────────────────────────

┌─────────────────────────────────────┐
│  C++ ROS2 Node                      │
│                                     │
│  RCLCPP_ERROR_TRACED(               │◄─── Traced logger automatically
│      this->get_logger(),            │     includes trace_id & span_id
│      "Failed to execute task"       │
│  );                                 │
│                                     │
│  Output:                            │
│  [ERROR] [1234567890.123] [node]:   │
│  [trace_id=a1b2c3d4...              │
│   span_id=f1e2d3c4...]              │
│  Failed to execute task             │
└──────────────┬──────────────────────┘
               │
               │ Log file: /opt/logs/robot.log
               │
               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  File Log Receiver (OTLP Collector)                                    │
│                                                                         │
│  receivers:                            │
│    filelog:                            │
│      include: ["/opt/logs/**/*.log"]   │
│      operators:                         │
│        - type: regex_parser             │
│          regex: '^\[(?P<level>\w+)\]   │
│            \[(?P<timestamp>\d+\.\d+)\]  │
│            \[(?P<source>[^\]]+)\]:      │
│            (?P<message>.*)$'            │
│        - type: regex_parser             │
│          parse_from: attributes.message  │
│          regex: '^\[trace_id=(?P<     │
│            trace_id>[0-9a-f]{32})      │
│            \s+span_id=(?P<span_id>     │
│            [0-9a-f]{16})\]\s*          │
│            (?P<body>.*)$'               │◄─── Extracts trace context
│        - type: trace_parser             │     from log message
│          trace_id:                      │
│            parse_from: attributes.trace_id │
│          span_id:                       │
│            parse_from: attributes.span_id  │
│        - type: move                     │
│          from: attributes.body          │
│          to: body                      │
└─────────────────────┬───────────────────┘
                      │
                      │ Correlated Log Entry
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  Visualization Backend (SigNoz/Grafana)                                 │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  Trace View                                                   │     │
│  │                                                               │     │
│  │  Trace: a1b2c3d4e5f6...                                      │     │
│  │  ├─ Span: create_task                                        │     │
│  │  │   └─ [View Logs] ◄─── Click to see related logs           │     │
│  │  └─ Span: execute_task                                       │     │
│  │       └─ [View Logs] ◄─── Shows: "Failed to execute task"     │     │
│  │                                                               │     │
│  └──────────────────────────────────────────────────────────────┘     │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │  Log View                                                      │     │
│  │                                                               │     │
│  │  [ERROR] 10:00:00.250 Failed to execute task                 │     │
│  │  └─ [View Trace] ◄─── Click to see full trace                │     │
│  │                                                               │     │
│  └──────────────────────────────────────────────────────────────┘     │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Component Interaction Sequence

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    COMPLETE EXECUTION FLOW                               │
└─────────────────────────────────────────────────────────────────────────┘

Time →
│
│  ┌──────────────┐         ┌──────────────┐         ┌──────────────┐
│  │ Python Node  │         │  C++ Node    │         │ OTLP Collector│
│  │(TaskProducer)│         │(RobotControl)│         │              │
│  └──────┬───────┘         └──────┬───────┘         └──────┬───────┘
│         │                        │                        │
│   1. setup_tracer()              │                        │
│         │                        │                        │
│         │─── Tracer Init ────────┼───────────────────────►│
│         │                        │                        │
│   2. Create span                 │                        │
│      "create_task"               │                        │
│         │                        │                        │
│   3. Create message              │                        │
│      + inject_trace_context()    │                        │
│         │                        │                        │
│         │─── Publish ────────────►│                        │
│         │   /task_queue          │                        │
│         │   (with TraceMetadata) │                        │
│         │                        │                        │
│         │                   4. Receive message           │
│         │                        │                        │
│         │                   5. extract_trace_context()    │
│         │                        │                        │
│         │                   6. Create child span         │
│         │                      "execute_task"             │
│         │                        │                        │
│         │                   7. Execute robot action       │
│         │                        │                        │
│         │                   8. Log error                  │
│         │                      (with trace context)        │
│         │                        │                        │
│         │                   9. End span                   │
│         │                        │                        │
│         │                   10. Export trace ─────────────►│
│         │                        │                        │
│         │                        │                   11. Parse log
│         │                        │                      (extract trace_id)
│         │                        │                        │
│         │                        │                   12. Correlate
│         │                        │                      log to trace
│         │                        │                        │
│         │                        │                   13. Export to
│         │                        │                      backend
│         │                        │                        │
│   14. End span                   │                        │
│         │                        │                        │
│   15. Export trace ──────────────┼───────────────────────►│
│         │                        │                        │
│         │                        │                   16. Forward to
│         │                        │                      SigNoz/Grafana
│         │                        │                        │
│  ┌──────▼───────┐         ┌──────▼───────┐         ┌──────▼───────┐
│  │ Trace sent   │         │ Trace sent   │         │ Data stored  │
│  └──────────────┘         └──────────────┘         └──────────────┘
```

---

## 5. Library Package Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    ROS-OPENTELEMETRY PACKAGES                           │
└─────────────────────────────────────────────────────────────────────────┘

ros-opentelemetry/
│
├── src/
│   ├── ros_opentelemetry_cpp/          ◄─── C++ Package
│   │   ├── CMakeLists.txt
│   │   ├── package.xml
│   │   └── include/
│   │       └── ros_opentelemetry_cpp/
│   │           └── ros_opentelemetry_cpp.hpp
│   │               ├── setup_tracer()
│   │               ├── inject_trace_context()
│   │               └── extract_trace_context()
│   │
│   ├── ros_opentelemetry_py/           ◄─── Python Package
│   │   ├── setup.py
│   │   ├── package.xml
│   │   └── ros_opentelemetry_py/
│   │       └── __init__.py
│   │           ├── setup_tracer()
│   │           ├── inject_trace_context()
│   │           └── wrap_logger()
│   │
│   └── ros_opentelemetry_interfaces/   ◄─── Message Definitions
│       ├── CMakeLists.txt
│       ├── package.xml
│       └── msg/
│           └── TraceMetadata.msg
│               ├── string trace_id
│               ├── string span_id
│               └── uint8 trace_flags
│
├── conanfile.txt                       ◄─── C++ Dependencies
│   └── opentelemetry-cpp/[~1.9.0]
│
├── pyproject.toml                      ◄─── Python Dependencies
│   └── opentelemetry-sdk
│
└── docker/                             ◄─── Example Setup
    ├── Dockerfile
    └── docker-compose.yml
```

---

## 6. Data Flow Summary

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    DATA FLOW SUMMARY                                     │
└─────────────────────────────────────────────────────────────────────────┘

ROS2 Application Code
        │
        ├─► Instrument with OpenTelemetry SDK
        │   ├─ C++: ros_opentelemetry_cpp
        │   └─ Python: ros_opentelemetry_py
        │
        ├─► Create Traces (Spans)
        │   ├─ StartSpan("operation_name")
        │   └─ EndSpan()
        │
        ├─► Propagate Trace Context
        │   ├─ Inject into ROS2 messages (TraceMetadata)
        │   └─ Extract from ROS2 messages
        │
        ├─► Generate Traced Logs
        │   ├─ C++: RCLCPP_*_TRACED()
        │   └─ Python: wrap_logger()
        │
        │
        ▼
OTLP Protocol (gRPC)
        │
        ├─► Traces ──────────────┐
        │                         │
        └─► Logs (via filelog) ───┤
                                   │
                                   ▼
                        OTLP Collector
                                   │
                                   ├─► Parse & Process
                                   │   ├─ Extract trace_id from logs
                                   │   ├─ Correlate logs to traces
                                   │   └─ Batch & export
                                   │
                                   ▼
                        Visualization Backend
                                   │
                                   ├─► SigNoz
                                   ├─► Grafana
                                   └─► Jaeger
```

---

## Key Features

1. **Dual Language Support**: Both C++ and Python ROS2 nodes
2. **Trace Context Propagation**: Via ROS2 message metadata
3. **Log-Trace Correlation**: Automatic linking of logs to traces
4. **Backend Agnostic**: Works with any OTLP-compatible backend
5. **Production Ready**: Used in real robotics applications

---

**Source**: [ros-opentelemetry GitHub Repository](https://github.com/szobov/ros-opentelemetry)
