#!/usr/bin/env python3
"""
Training Data Collection Script - Phase 4

Collects historical alert → remediation pairs from Prometheus and Agent-SRE logs
to create a fine-tuning dataset for FunctionGemma.
"""
import argparse
import json
import sys
from pathlib import Path
from datetime import datetime, timedelta
from typing import List, Dict, Any, Optional
import httpx
import structlog

# Add src to path
sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

logger = structlog.get_logger()


class TrainingDataCollector:
    """Collects training data from various sources."""
    
    def __init__(
        self,
        prometheus_url: str,
        loki_url: Optional[str] = None,
        lookback_days: int = 90
    ):
        self.prometheus_url = prometheus_url.rstrip("/")
        self.loki_url = loki_url.rstrip("/") if loki_url else None
        self.lookback_days = lookback_days
        self.client = httpx.AsyncClient(timeout=30.0)
    
    async def collect_from_prometheus(self) -> List[Dict[str, Any]]:
        """Collect alert history from Prometheus."""
        logger.info("collecting_from_prometheus", url=self.prometheus_url)
        
        # Query for alert history
        # Note: This is a simplified version - real implementation would query
        # Prometheus Alertmanager API or Loki for alert history
        
        end_time = datetime.utcnow()
        start_time = end_time - timedelta(days=self.lookback_days)
        
        # Example query structure (adjust based on your Prometheus setup)
        query = {
            "query": 'ALERTS{alertstate="firing"}',
            "start": start_time.timestamp(),
            "end": end_time.timestamp(),
            "step": "1h"
        }
        
        try:
            response = await self.client.get(
                f"{self.prometheus_url}/api/v1/query_range",
                params=query
            )
            response.raise_for_status()
            data = response.json()
            
            # Parse results
            alerts = []
            if "data" in data and "result" in data["data"]:
                for result in data["data"]["result"]:
                    metric = result.get("metric", {})
                    alerts.append({
                        "alertname": metric.get("alertname", "unknown"),
                        "labels": metric,
                        "timestamp": result.get("values", [])[0][0] if result.get("values") else None
                    })
            
            logger.info("prometheus_data_collected", count=len(alerts))
            return alerts
            
        except Exception as e:
            logger.warning("failed_to_collect_from_prometheus", error=str(e))
            return []
    
    async def collect_from_loki(self) -> List[Dict[str, Any]]:
        """Collect remediation history from Loki logs."""
        if not self.loki_url:
            logger.info("loki_not_configured")
            return []
        
        logger.info("collecting_from_loki", url=self.loki_url)
        
        end_time = datetime.utcnow()
        start_time = end_time - timedelta(days=self.lookback_days)
        
        # Query Loki for Agent-SRE remediation logs
        query = {
            "query": '{app="agent-sre"} |= "remediation"',
            "start": start_time.timestamp(),
            "end": end_time.timestamp(),
            "limit": 1000
        }
        
        try:
            response = await self.client.get(
                f"{self.loki_url}/loki/api/v1/query_range",
                params=query
            )
            response.raise_for_status()
            data = response.json()
            
            remediations = []
            if "data" in data and "result" in data["data"]:
                for stream in data["data"]["result"]:
                    for entry in stream.get("values", []):
                        # Parse log entry
                        log_line = entry[1]
                        # Extract remediation info (simplified - adjust based on your log format)
                        if "lambda_function" in log_line:
                            remediations.append({
                                "log": log_line,
                                "timestamp": entry[0]
                            })
            
            logger.info("loki_data_collected", count=len(remediations))
            return remediations
            
        except Exception as e:
            logger.warning("failed_to_collect_from_loki", error=str(e))
            return []
    
    def create_training_examples(
        self,
        alerts: List[Dict[str, Any]],
        remediations: List[Dict[str, Any]]
    ) -> List[Dict[str, Any]]:
        """Create training examples from collected data."""
        examples = []
        
        # Match alerts with remediations (simplified matching)
        for alert in alerts:
            alertname = alert.get("alertname", "unknown")
            labels = alert.get("labels", {})
            
            # Find matching remediation
            matching_remediation = None
            for remediation in remediations:
                log = remediation.get("log", "")
                if alertname in log or any(label_value in log for label_value in labels.values()):
                    matching_remediation = remediation
                    break
            
            if matching_remediation:
                # Extract lambda function and parameters from log
                # This is simplified - real implementation would parse structured logs
                lambda_function = self._extract_lambda_function(matching_remediation["log"])
                parameters = self._extract_parameters(matching_remediation["log"], labels)
                
                if lambda_function:
                    # Create training example
                    prompt = self._create_prompt(alertname, labels)
                    completion = self._create_completion(lambda_function, parameters)
                    
                    examples.append({
                        "prompt": prompt,
                        "completion": completion
                    })
        
        logger.info("training_examples_created", count=len(examples))
        return examples
    
    def _extract_lambda_function(self, log: str) -> Optional[str]:
        """Extract lambda function name from log."""
        # Simplified extraction - adjust based on your log format
        for func in [
            "flux-reconcile-kustomization",
            "flux-reconcile-gitrepository",
            "flux-reconcile-helmrelease",
            "pod-restart",
            "pod-check-status",
            "scale-deployment",
            "check-pvc-status"
        ]:
            if func in log:
                return func
        return None
    
    def _extract_parameters(self, log: str, labels: Dict[str, Any]) -> Dict[str, Any]:
        """Extract parameters from log and labels."""
        # Use labels as base parameters
        parameters = {}
        
        if "name" in labels:
            parameters["name"] = labels["name"]
        elif "pod" in labels:
            parameters["name"] = labels["pod"]
        elif "deployment" in labels:
            parameters["name"] = labels["deployment"]
        
        if "namespace" in labels:
            parameters["namespace"] = labels["namespace"]
        else:
            parameters["namespace"] = "flux-system"  # Default
        
        # Extract type for pod-restart
        if "pod-restart" in log:
            parameters["type"] = "pod"
        
        # Extract replicas for scale-deployment
        if "scale-deployment" in log and "replicas" in labels:
            try:
                parameters["replicas"] = int(labels["replicas"])
            except (ValueError, TypeError):
                pass
        
        return parameters
    
    def _create_prompt(self, alertname: str, labels: Dict[str, Any]) -> str:
        """Create prompt for training example."""
        return f"""Alert: {alertname}
Labels: {json.dumps(labels)}
Select remediation:"""
    
    def _create_completion(
        self,
        lambda_function: str,
        parameters: Dict[str, Any]
    ) -> str:
        """Create completion for training example."""
        reasoning = f"Selected {lambda_function} based on alert context"
        
        return json.dumps({
            "lambda_function": lambda_function,
            "parameters": parameters,
            "reasoning": reasoning
        })
    
    async def collect_all(self) -> List[Dict[str, Any]]:
        """Collect all training data."""
        logger.info("starting_data_collection", lookback_days=self.lookback_days)
        
        alerts = await self.collect_from_prometheus()
        remediations = await self.collect_from_loki()
        
        examples = self.create_training_examples(alerts, remediations)
        
        await self.client.aclose()
        
        return examples
    
    async def close(self):
        """Close HTTP client."""
        await self.client.aclose()


async def main():
    parser = argparse.ArgumentParser(description="Collect training data for Agent-SRE fine-tuning")
    parser.add_argument(
        "--prometheus-url",
        default="http://prometheus.monitoring.svc:9090",
        help="Prometheus URL"
    )
    parser.add_argument(
        "--loki-url",
        default=None,
        help="Loki URL (optional)"
    )
    parser.add_argument(
        "--lookback-days",
        type=int,
        default=90,
        help="Number of days to look back"
    )
    parser.add_argument(
        "--output",
        default="training_data.jsonl",
        help="Output file path"
    )
    
    args = parser.parse_args()
    
    collector = TrainingDataCollector(
        prometheus_url=args.prometheus_url,
        loki_url=args.loki_url,
        lookback_days=args.lookback_days
    )
    
    try:
        examples = await collector.collect_all()
        
        # Write to JSONL file
        output_path = Path(args.output)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        
        with open(output_path, "w") as f:
            for example in examples:
                f.write(json.dumps(example) + "\n")
        
        logger.info(
            "training_data_collected",
            count=len(examples),
            output=str(output_path)
        )
        
        print(f"✅ Collected {len(examples)} training examples")
        print(f"📁 Saved to: {output_path}")
        
    finally:
        await collector.close()


if __name__ == "__main__":
    import asyncio
    asyncio.run(main())

