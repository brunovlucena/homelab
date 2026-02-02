#!/usr/bin/env python3
"""
📊 Data Collector for TRM Fine-Tuning

Collects training data from:
1. Notifi-services codebase (C# files, configs, templates)
2. Observability data (Prometheus metrics, Loki logs) from last 30 days
3. Formats data for TRM recursive reasoning training
"""

import os
import json
import asyncio
from datetime import datetime, timedelta
from pathlib import Path
from typing import List, Dict, Any, Optional
import httpx
from dataclasses import dataclass, asdict


@dataclass
class TrainingExample:
    """Single training example for TRM."""
    problem: str  # Input question/problem
    initial_answer: str  # Initial answer (can be empty)
    solution: str  # Final solution
    reasoning_steps: List[str]  # Recursive reasoning steps
    metadata: Dict[str, Any]  # Source, timestamp, etc.


class NotifiServicesCollector:
    """Collects code and configs from notifi-services repository."""
    
    def __init__(self, repo_path: str):
        self.repo_path = Path(repo_path)
        self.supported_extensions = {'.cs', '.yaml', '.yml', '.json', '.mustache', '.md'}
    
    def collect_code_files(self) -> List[Dict[str, Any]]:
        """Collect all code files from notifi-services."""
        examples = []
        
        for ext in self.supported_extensions:
            for file_path in self.repo_path.rglob(f"*{ext}"):
                if self._should_include_file(file_path):
                    content = file_path.read_text(encoding='utf-8', errors='ignore')
                    examples.append({
                        "type": "code",
                        "file": str(file_path.relative_to(self.repo_path)),
                        "extension": ext,
                        "content": content,
                        "size": len(content),
                    })
        
        return examples
    
    def _should_include_file(self, file_path: Path) -> bool:
        """Filter files to include."""
        # Exclude test files, build artifacts, etc.
        exclude_patterns = [
            'bin/', 'obj/', '.git/', 'node_modules/',
            'Test', 'test', 'Tests', 'tests',
            '.vs/', '.idea/', '__pycache__/'
        ]
        
        path_str = str(file_path)
        return not any(pattern in path_str for pattern in exclude_patterns)
    
    def format_for_trm(self, code_files: List[Dict[str, Any]]) -> List[TrainingExample]:
        """Format code files into TRM training examples."""
        examples = []
        
        for file_data in code_files:
            # Create reasoning problem: "Understand this code file"
            problem = f"Analyze and understand the following {file_data['extension']} file:\n\n{file_data['file']}"
            initial_answer = ""  # Start with empty answer
            solution = f"File: {file_data['file']}\n\nContent:\n{file_data['content'][:2000]}"  # Truncate long files
            
            # Create recursive reasoning steps
            reasoning_steps = [
                f"Step 1: Identify file type: {file_data['extension']}",
                f"Step 2: Analyze file structure and purpose",
                f"Step 3: Extract key components and patterns",
                f"Step 4: Generate comprehensive understanding"
            ]
            
            examples.append(TrainingExample(
                problem=problem,
                initial_answer=initial_answer,
                solution=solution,
                reasoning_steps=reasoning_steps,
                metadata={
                    "source": "notifi-services",
                    "file": file_data['file'],
                    "type": "code_analysis",
                    "timestamp": datetime.now().isoformat()
                }
            ))
        
        return examples


class ObservabilityCollector:
    """Collects observability data from Prometheus, Loki, and Tempo."""
    
    def __init__(
        self,
        prometheus_url: str,
        loki_url: str,
        tempo_url: str,
        days: int = 30
    ):
        self.prometheus_url = prometheus_url
        self.loki_url = loki_url
        self.tempo_url = tempo_url
        self.days = days
        self.end_time = datetime.now()
        self.start_time = self.end_time - timedelta(days=days)
    
    async def collect_prometheus_metrics(self) -> List[Dict[str, Any]]:
        """Collect Prometheus metrics from last N days."""
        examples = []
        
        # Key metrics to collect
        queries = [
            "up",  # Service availability
            "rate(http_requests_total[5m])",  # Request rate
            "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",  # P95 latency
            "rate(container_cpu_usage_seconds_total[5m])",  # CPU usage
            "container_memory_usage_bytes",  # Memory usage
            "kube_pod_status_phase",  # Pod status
        ]
        
        async with httpx.AsyncClient(timeout=60.0) as client:
            for query in queries:
                try:
                    # Query range for last 30 days
                    response = await client.get(
                        f"{self.prometheus_url}/api/v1/query_range",
                        params={
                            "query": query,
                            "start": self.start_time.timestamp(),
                            "end": self.end_time.timestamp(),
                            "step": "1h"  # 1 hour intervals
                        }
                    )
                    response.raise_for_status()
                    data = response.json()
                    
                    if data.get("status") == "success":
                        result = data.get("data", {}).get("result", [])
                        examples.append({
                            "type": "prometheus_metric",
                            "query": query,
                            "data": result,
                            "timestamp_range": {
                                "start": self.start_time.isoformat(),
                                "end": self.end_time.isoformat()
                            }
                        })
                except Exception as e:
                    print(f"Error querying Prometheus for {query}: {e}")
        
        return examples
    
    async def collect_loki_logs(self) -> List[Dict[str, Any]]:
        """Collect Loki logs from last N days."""
        examples = []
        
        # Key log queries
        log_queries = [
            '{namespace="knative-lambda"} |= "error"',  # Errors
            '{namespace="ai-agents"} | json',  # Agent logs
            '{service_name=~"lambda-.*"} | json',  # Lambda function logs
        ]
        
        async with httpx.AsyncClient(timeout=120.0) as client:
            for logql in log_queries:
                try:
                    response = await client.get(
                        f"{self.loki_url}/loki/api/v1/query_range",
                        params={
                            "query": logql,
                            "start": int(self.start_time.timestamp() * 1e9),  # Nanoseconds
                            "end": int(self.end_time.timestamp() * 1e9),
                            "limit": 1000  # Limit results
                        }
                    )
                    response.raise_for_status()
                    data = response.json()
                    
                    if data.get("status") == "success":
                        result = data.get("data", {}).get("result", [])
                        examples.append({
                            "type": "loki_logs",
                            "query": logql,
                            "entries": result,
                            "timestamp_range": {
                                "start": self.start_time.isoformat(),
                                "end": self.end_time.isoformat()
                            }
                        })
                except Exception as e:
                    print(f"Error querying Loki for {logql}: {e}")
        
        return examples
    
    async def collect_tempo_traces(self) -> List[Dict[str, Any]]:
        """Collect Tempo traces from last N days."""
        examples = []
        
        # Key trace queries - search by service tags
        trace_queries = [
            {"service.name": "knative-lambda-operator"},  # Lambda operator traces
            {"service.name": "agent-sre"},  # Agent-SRE traces
            {"service.name": "agent-bruno"},  # Agent-Bruno traces
            {"namespace": "ai-agents"},  # All AI agents
            {"namespace": "knative-lambda"},  # Lambda functions
        ]
        
        async with httpx.AsyncClient(timeout=120.0) as client:
            for tags in trace_queries:
                try:
                    # Build tags query string
                    tags_query = " ".join([f"{k}={v}" for k, v in tags.items()])
                    
                    # Query Tempo search API
                    response = await client.get(
                        f"{self.tempo_url}/api/search",
                        params={
                            "tags": tags_query,
                            "start": int(self.start_time.timestamp()),
                            "end": int(self.end_time.timestamp()),
                            "limit": 100  # Limit results per query
                        }
                    )
                    response.raise_for_status()
                    data = response.json()
                    
                    if data.get("traces"):
                        examples.append({
                            "type": "tempo_traces",
                            "tags": tags,
                            "traces": data.get("traces", []),
                            "total_traces": len(data.get("traces", [])),
                            "timestamp_range": {
                                "start": self.start_time.isoformat(),
                                "end": self.end_time.isoformat()
                            }
                        })
                except Exception as e:
                    print(f"Error querying Tempo for {tags}: {e}")
        
        # Also query for slow traces and error traces
        async with httpx.AsyncClient(timeout=120.0) as slow_client:
            try:
                # Query for slow traces (duration > 5s)
                response = await slow_client.get(
                    f"{self.tempo_url}/api/search",
                    params={
                        "tags": "duration>5s",
                        "start": int(self.start_time.timestamp()),
                        "end": int(self.end_time.timestamp()),
                        "limit": 50
                    }
                )
                if response.status_code == 200:
                    data = response.json()
                    if data.get("traces"):
                        examples.append({
                            "type": "tempo_traces",
                            "tags": {"filter": "slow_traces", "duration": ">5s"},
                            "traces": data.get("traces", []),
                            "total_traces": len(data.get("traces", [])),
                            "timestamp_range": {
                                "start": self.start_time.isoformat(),
                                "end": self.end_time.isoformat()
                            }
                        })
            except Exception as e:
                print(f"Error querying Tempo for slow traces: {e}")
        
        # Query for error traces
        async with httpx.AsyncClient(timeout=120.0) as error_client:
            try:
                response = await error_client.get(
                    f"{self.tempo_url}/api/search",
                    params={
                        "tags": "status.code=ERROR",
                        "start": int(self.start_time.timestamp()),
                        "end": int(self.end_time.timestamp()),
                        "limit": 50
                    }
                )
                if response.status_code == 200:
                    data = response.json()
                    if data.get("traces"):
                        examples.append({
                            "type": "tempo_traces",
                            "tags": {"filter": "error_traces", "status": "ERROR"},
                            "traces": data.get("traces", []),
                            "total_traces": len(data.get("traces", [])),
                            "timestamp_range": {
                                "start": self.start_time.isoformat(),
                                "end": self.end_time.isoformat()
                            }
                        })
            except Exception as e:
                print(f"Error querying Tempo for error traces: {e}")
        
        return examples
    
    def format_for_trm(
        self,
        metrics: List[Dict[str, Any]],
        logs: List[Dict[str, Any]],
        traces: List[Dict[str, Any]] = None
    ) -> List[TrainingExample]:
        """Format observability data into TRM training examples."""
        examples = []
        
        # Format metrics as reasoning problems
        for metric_data in metrics:
            problem = f"Analyze Prometheus metric: {metric_data['query']}\n\nTime range: {metric_data['timestamp_range']['start']} to {metric_data['timestamp_range']['end']}"
            initial_answer = ""
            
            # Extract key insights from metric data
            solution = self._extract_metric_insights(metric_data)
            
            reasoning_steps = [
                "Step 1: Understand metric type and purpose",
                "Step 2: Analyze time series patterns",
                "Step 3: Identify anomalies or trends",
                "Step 4: Generate actionable insights"
            ]
            
            examples.append(TrainingExample(
                problem=problem,
                initial_answer=initial_answer,
                solution=solution,
                reasoning_steps=reasoning_steps,
                metadata={
                    "source": "prometheus",
                    "query": metric_data['query'],
                    "type": "metric_analysis",
                    "timestamp": datetime.now().isoformat()
                }
            ))
        
        # Format logs as reasoning problems
        for log_data in logs:
            problem = f"Analyze logs matching: {log_data['query']}\n\nTime range: {log_data['timestamp_range']['start']} to {log_data['timestamp_range']['end']}"
            initial_answer = ""
            
            # Extract patterns from logs
            solution = self._extract_log_patterns(log_data)
            
            reasoning_steps = [
                "Step 1: Parse log entries and structure",
                "Step 2: Identify error patterns and frequencies",
                "Step 3: Correlate events and timelines",
                "Step 4: Generate diagnostic insights"
            ]
            
            examples.append(TrainingExample(
                problem=problem,
                initial_answer=initial_answer,
                solution=solution,
                reasoning_steps=reasoning_steps,
                metadata={
                    "source": "loki",
                    "query": log_data['query'],
                    "type": "log_analysis",
                    "timestamp": datetime.now().isoformat()
                }
            ))
        
        # Format traces as reasoning problems
        if traces:
            for trace_data in traces:
                tags_str = ", ".join([f"{k}={v}" for k, v in trace_data.get('tags', {}).items()])
                problem = f"Analyze distributed traces matching: {tags_str}\n\nTime range: {trace_data['timestamp_range']['start']} to {trace_data['timestamp_range']['end']}\nTotal traces: {trace_data.get('total_traces', 0)}"
                initial_answer = ""
                
                # Extract patterns from traces
                solution = self._extract_trace_patterns(trace_data)
                
                reasoning_steps = [
                    "Step 1: Understand trace structure and service topology",
                    "Step 2: Analyze span durations and relationships",
                    "Step 3: Identify bottlenecks and slow operations",
                    "Step 4: Correlate errors and failures across services",
                    "Step 5: Generate performance insights and recommendations"
                ]
                
                examples.append(TrainingExample(
                    problem=problem,
                    initial_answer=initial_answer,
                    solution=solution,
                    reasoning_steps=reasoning_steps,
                    metadata={
                        "source": "tempo",
                        "tags": trace_data.get('tags', {}),
                        "type": "trace_analysis",
                        "trace_count": trace_data.get('total_traces', 0),
                        "timestamp": datetime.now().isoformat()
                    }
                ))
        
        return examples
    
    def _extract_metric_insights(self, metric_data: Dict[str, Any]) -> str:
        """Extract insights from Prometheus metric data."""
        result = metric_data.get("data", [])
        if not result:
            return "No data available for this metric."
        
        insights = []
        for series in result[:5]:  # Limit to first 5 series
            metric = series.get("metric", {})
            values = series.get("values", [])
            if values:
                latest_value = values[-1][1]  # Last value
                insights.append(f"Series {metric.get('__name__', 'unknown')}: {latest_value}")
        
        return "\n".join(insights) if insights else "No insights extracted."
    
    def _extract_log_patterns(self, log_data: Dict[str, Any]) -> str:
        """Extract patterns from Loki log data."""
        result = log_data.get("entries", [])
        if not result:
            return "No log entries found."
        
        patterns = []
        error_count = 0
        for stream in result[:10]:  # Limit to first 10 streams
            entries = stream.get("values", [])
            for entry in entries:
                log_line = entry[1] if len(entry) > 1 else ""
                if "error" in log_line.lower() or "exception" in log_line.lower():
                    error_count += 1
        
        patterns.append(f"Total log streams: {len(result)}")
        patterns.append(f"Error/exception mentions: {error_count}")
        
        return "\n".join(patterns)
    
    def _extract_trace_patterns(self, trace_data: Dict[str, Any]) -> str:
        """Extract patterns from Tempo trace data."""
        traces = trace_data.get("traces", [])
        if not traces:
            return "No traces found for this query."
        
        patterns = []
        patterns.append(f"Total traces: {len(traces)}")
        
        # Analyze trace characteristics
        total_spans = 0
        total_duration = 0
        error_count = 0
        service_names = set()
        
        for trace in traces[:20]:  # Limit to first 20 traces for analysis
            trace_id = trace.get("traceID", "unknown")
            spans = trace.get("spans", [])
            total_spans += len(spans)
            
            for span in spans:
                duration = span.get("duration", 0)
                total_duration += duration
                
                # Extract service name
                tags = span.get("tags", {})
                service_name = tags.get("service.name") or tags.get("serviceName")
                if service_name:
                    service_names.add(service_name)
                
                # Check for errors
                status_code = tags.get("status.code") or tags.get("http.status_code")
                if status_code and (status_code >= 400 or status_code == "ERROR"):
                    error_count += 1
        
        patterns.append(f"Total spans analyzed: {total_spans}")
        if total_spans > 0:
            avg_duration = total_duration / total_spans
            patterns.append(f"Average span duration: {avg_duration:.2f}ms")
        patterns.append(f"Services involved: {', '.join(list(service_names)[:10])}")  # Limit to 10
        patterns.append(f"Error spans: {error_count}")
        
        return "\n".join(patterns)


class DataCollector:
    """Main data collector orchestrator."""
    
    def __init__(
        self,
        notifi_services_path: str,
        prometheus_url: str,
        loki_url: str,
        tempo_url: str,
        days: int = 30
    ):
        self.notifi_collector = NotifiServicesCollector(notifi_services_path)
        self.obs_collector = ObservabilityCollector(
            prometheus_url=prometheus_url,
            loki_url=loki_url,
            tempo_url=tempo_url,
            days=days
        )
    
    async def collect_all(self) -> List[TrainingExample]:
        """Collect all training data."""
        print("📊 Collecting notifi-services code...")
        code_files = self.notifi_collector.collect_code_files()
        code_examples = self.notifi_collector.format_for_trm(code_files)
        print(f"✅ Collected {len(code_examples)} code examples")
        
        print("📊 Collecting Prometheus metrics...")
        metrics = await self.obs_collector.collect_prometheus_metrics()
        print(f"✅ Collected {len(metrics)} metric queries")
        
        print("📊 Collecting Loki logs...")
        logs = await self.obs_collector.collect_loki_logs()
        print(f"✅ Collected {len(logs)} log queries")
        
        print("📊 Collecting Tempo traces...")
        traces = await self.obs_collector.collect_tempo_traces()
        print(f"✅ Collected {len(traces)} trace queries")
        
        print("📊 Formatting observability data...")
        obs_examples = self.obs_collector.format_for_trm(metrics, logs, traces)
        print(f"✅ Formatted {len(obs_examples)} observability examples")
        
        all_examples = code_examples + obs_examples
        print(f"🎉 Total training examples: {len(all_examples)}")
        
        return all_examples
    
    def save_to_jsonl(self, examples: List[TrainingExample], output_path: str):
        """Save training examples to JSONL format."""
        output_path = Path(output_path)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        
        with open(output_path, 'w') as f:
            for example in examples:
                f.write(json.dumps(asdict(example)) + '\n')
        
        print(f"💾 Saved {len(examples)} examples to {output_path}")


async def main():
    """Main entry point for data collection."""
    notifi_path = os.getenv("NOTIFI_SERVICES_PATH", "/workspace/notifi/repos/notifi-services")
    prometheus_url = os.getenv("PROMETHEUS_URL", "http://prometheus.monitoring.svc:9090")
    loki_url = os.getenv("LOKI_URL", "http://loki.monitoring.svc:3100")
    tempo_url = os.getenv("TEMPO_URL", "http://tempo.tempo.svc:3200")
    days = int(os.getenv("DATA_DAYS", "30"))
    output_path = os.getenv("OUTPUT_PATH", "./data/training_data.jsonl")
    
    collector = DataCollector(
        notifi_services_path=notifi_path,
        prometheus_url=prometheus_url,
        loki_url=loki_url,
        tempo_url=tempo_url,
        days=days
    )
    
    examples = await collector.collect_all()
    collector.save_to_jsonl(examples, output_path)


if __name__ == "__main__":
    asyncio.run(main())

