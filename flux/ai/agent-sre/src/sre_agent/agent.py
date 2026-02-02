"""
SRE Agent - Core agent logic for health report generation.
"""
from typing import Dict, Any, Optional
import structlog

from .config import AgentConfig
from src.metrics_collector import MetricsCollector
from src.report_generator import ReportGenerator, HealthReport

logger = structlog.get_logger()


class SREAgent:
    """SRE Agent for generating health reports."""
    
    def __init__(
        self,
        metrics_collector: MetricsCollector,
        report_generator: ReportGenerator,
        config: AgentConfig
    ):
        self.metrics_collector = metrics_collector
        self.report_generator = report_generator
        self.config = config
    
    async def generate_component_report(self, component: str) -> HealthReport:
        """Generate health report for a specific component."""
        logger.info("Generating component report", component=component)
        
        # Collect metrics for component
        metrics = await self._collect_component_metrics(component)
        
        # Generate report using LLM
        report = await self.report_generator.generate_report(
            component=component,
            metrics=metrics,
            time_range=self.config.report_time_range
        )
        
        return report
    
    async def generate_full_report(self) -> HealthReport:
        """Generate comprehensive health report for all components."""
        logger.info("Generating full health report")
        
        # Collect metrics for all components
        all_metrics = {}
        for component in ["loki", "prometheus", "infrastructure", "observability"]:
            all_metrics[component] = await self._collect_component_metrics(component)
        
        # Generate comprehensive report
        report = await self.report_generator.generate_full_report(
            metrics=all_metrics,
            time_range=self.config.report_time_range
        )
        
        return report
    
    async def _collect_component_metrics(self, component: str) -> Dict[str, Any]:
        """Collect metrics for a specific component."""
        metrics = {}
        
        if component == "loki":
            metrics = await self.metrics_collector.collect_loki_metrics()
        elif component == "prometheus":
            metrics = await self.metrics_collector.collect_prometheus_metrics()
        elif component == "infrastructure":
            metrics = await self.metrics_collector.collect_infrastructure_metrics()
        elif component == "observability":
            metrics = await self.metrics_collector.collect_observability_metrics()
        
        return metrics

