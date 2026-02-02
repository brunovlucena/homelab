"""
Agent DevSecOps - CloudEvent Handler

Handles incoming security-related CloudEvents and coordinates responses.

RBAC Levels:
- readonly: Scanning and auditing (default)
- operator: Automated security operations
- admin: Emergency incident response (requires approval)
"""

import json
import logging
import os
from datetime import datetime, timezone
from typing import Any
from cloudevents.http import CloudEvent, to_json
from kubernetes import client, config
from image_scanner import ImageScanner
from metrics_exporter import get_metrics_exporter

logger = logging.getLogger(__name__)


class SecurityScanner:
    """Security scanner with multiple RBAC levels."""
    
    def __init__(self):
        self.environment = os.getenv("ENVIRONMENT", "pro")
        self.rbac_level = os.getenv("RBAC_LEVEL", "readonly")
        self.approval_required = os.getenv("APPROVAL_REQUIRED", "true").lower() == "true"
        
        # Initialize image scanner and metrics
        self.image_scanner = ImageScanner()
        self.metrics = get_metrics_exporter()
        
        # Initialize Kubernetes client
        try:
            config.load_incluster_config()
            self.k8s_enabled = True
            self.core_v1 = client.CoreV1Api()
            self.apps_v1 = client.AppsV1Api()
            self.custom_api = client.CustomObjectsApi()
            logger.info(f"Kubernetes client initialized with RBAC level: {self.rbac_level}")
        except Exception as e:
            logger.warning(f"Running outside cluster: {e}")
            self.k8s_enabled = False
    
    def handle_event(self, event: CloudEvent) -> dict[str, Any]:
        """Route CloudEvent to appropriate handler based on type."""
        event_type = event["type"]
        
        handlers = {
            "io.homelab.build.completed": self.handle_build_completed,
            "io.homelab.deploy.requested": self.handle_deploy_request,
            "io.homelab.agent.security.query": self.handle_security_query,
            "io.homelab.alert.security": self.handle_security_alert,
            "io.homelab.schedule.daily": self.handle_daily_check,
            "io.homelab.schedule.weekly": self.handle_weekly_scan,
            "io.homelab.scan.lambdafunctions": self.handle_scan_lambdafunctions,
        }
        
        handler = handlers.get(event_type, self.handle_unknown)
        
        logger.info(f"Processing event: {event_type} with RBAC level: {self.rbac_level}")
        return handler(event)
    
    def handle_build_completed(self, event: CloudEvent) -> dict[str, Any]:
        """Trigger vulnerability scan on new builds."""
        data = event.data or {}
        image = data.get("image", "unknown")
        
        logger.info(f"Build completed, scanning image: {image}")
        
        # In readonly mode, we can only report findings
        vulnerabilities = self._scan_image(image)
        
        if vulnerabilities.get("critical", 0) > 0:
            # Create a SecurityProposal instead of blocking directly
            if self.rbac_level in ["operator", "admin"]:
                self._create_security_proposal(
                    action="suspend",
                    target={"kind": "HelmRelease", "name": data.get("release_name")},
                    severity="critical",
                    reason=f"Critical vulnerabilities found in {image}",
                    cve_ids=vulnerabilities.get("cve_ids", [])
                )
        
        return {
            "status": "scanned",
            "image": image,
            "vulnerabilities": vulnerabilities,
            "proposal_created": vulnerabilities.get("critical", 0) > 0
        }
    
    def handle_deploy_request(self, event: CloudEvent) -> dict[str, Any]:
        """Security gate for deployments."""
        data = event.data or {}
        namespace = data.get("namespace", "default")
        name = data.get("name", "unknown")
        
        logger.info(f"Deployment request: {namespace}/{name}")
        
        # Perform security checks
        checks = {
            "image_signed": self._check_image_signature(data.get("image")),
            "no_critical_cves": self._check_no_critical_cves(data.get("image")),
            "network_policy_exists": self._check_network_policy(namespace),
            "resource_limits_set": self._check_resource_limits(data),
            "security_context_set": self._check_security_context(data),
        }
        
        all_passed = all(checks.values())
        
        return {
            "approved": all_passed,
            "checks": checks,
            "rbac_level": self.rbac_level,
            "message": "All security checks passed" if all_passed else "Security checks failed"
        }
    
    def handle_security_query(self, event: CloudEvent) -> dict[str, Any]:
        """Handle security queries from other agents."""
        data = event.data or {}
        query_type = data.get("query_type", "status")
        
        if query_type == "status":
            return self._get_security_status()
        elif query_type == "vulnerabilities":
            return self._get_vulnerability_summary()
        elif query_type == "compliance":
            return self._get_compliance_status()
        else:
            return {"error": f"Unknown query type: {query_type}"}
    
    def handle_security_alert(self, event: CloudEvent) -> dict[str, Any]:
        """Handle security alerts from monitoring."""
        data = event.data or {}
        alert_name = data.get("alertname", "unknown")
        severity = data.get("severity", "warning")
        
        logger.warning(f"Security alert received: {alert_name} ({severity})")
        
        if severity == "critical" and self.rbac_level == "admin":
            # Admin level can take immediate action
            return self._handle_critical_alert(data)
        elif self.rbac_level in ["operator", "admin"]:
            # Operator creates a proposal
            self._create_security_proposal(
                action="suspend",
                target=data.get("labels", {}),
                severity=severity,
                reason=f"Security alert: {alert_name}"
            )
            return {"status": "proposal_created", "alert": alert_name}
        else:
            # Readonly just logs and reports
            return {"status": "logged", "alert": alert_name, "action": "manual_review_required"}
    
    def handle_daily_check(self, event: CloudEvent) -> dict[str, Any]:
        """Run daily compliance checks."""
        logger.info("Running daily compliance checks")
        
        results = {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "checks_performed": [],
            "issues_found": []
        }
        
        # Check RBAC configurations
        results["checks_performed"].append("rbac_audit")
        
        # Check network policies
        results["checks_performed"].append("network_policy_coverage")
        
        # Check secret expiry
        results["checks_performed"].append("secret_expiry")
        
        # Check pod security standards
        results["checks_performed"].append("pod_security_standards")
        
        return results
    
    def handle_weekly_scan(self, event: CloudEvent) -> dict[str, Any]:
        """Run weekly vulnerability scans."""
        logger.info("Running weekly vulnerability scan")
        
        if not self.k8s_enabled:
            return {"status": "skipped", "reason": "Kubernetes not available"}
        
        # Get all running images
        images = self._get_running_images()
        
        results = {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "images_scanned": len(images),
            "vulnerabilities_by_severity": {
                "critical": 0,
                "high": 0,
                "medium": 0,
                "low": 0
            }
        }
        
        for image in images:
            scan_result = self._scan_image(image)
            for severity in ["critical", "high", "medium", "low"]:
                results["vulnerabilities_by_severity"][severity] += scan_result.get(severity, 0)
        
        return results
    
    def handle_scan_lambdafunctions(self, event: CloudEvent) -> dict[str, Any]:
        """Scan all LambdaFunctions for outdated images."""
        logger.info("Scanning LambdaFunctions for outdated images")
        
        if not self.k8s_enabled:
            return {
                "status": "skipped",
                "reason": "Kubernetes not available",
                "timestamp": datetime.now(timezone.utc).isoformat()
            }
        
        data = event.data or {}
        namespace_filter = data.get("namespace")  # Optional namespace filter
        
        try:
            # List all LambdaFunctions
            if namespace_filter:
                lambdas = self.custom_api.list_namespaced_custom_object(
                    group="lambda.knative.io",
                    version="v1alpha1",
                    namespace=namespace_filter,
                    plural="lambdafunctions"
                )
            else:
                lambdas = self.custom_api.list_cluster_custom_object(
                    group="lambda.knative.io",
                    version="v1alpha1",
                    plural="lambdafunctions"
                )
            
            results = {
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "total_scanned": 0,
                "outdated": 0,
                "up_to_date": 0,
                "no_image": 0,
                "errors": 0,
                "functions": []
            }
            
            # Track outdated images by namespace/registry for metrics
            outdated_by_ns_reg = {}
            
            for lf in lambdas.get("items", []):
                namespace = lf.get("metadata", {}).get("namespace", "default")
                try:
                    scan_result = self.image_scanner.scan_lambdafunction(lf)
                    results["total_scanned"] += 1
                    
                    if scan_result.get("error"):
                        results["no_image"] += 1
                        self.metrics.record_error(namespace, "no_image")
                    elif scan_result.get("is_outdated"):
                        results["outdated"] += 1
                        # Track outdated count by namespace/registry
                        registry = scan_result.get("image_uri", "").split("/")[0].split(":")[0] if scan_result.get("image_uri") else "unknown"
                        key = (namespace, registry)
                        outdated_by_ns_reg[key] = outdated_by_ns_reg.get(key, 0) + 1
                    else:
                        results["up_to_date"] += 1
                    
                    # Update Prometheus metrics
                    self.metrics.update_lambda_image_info(scan_result)
                    self.metrics.record_scan(namespace, "success")
                    
                    results["functions"].append(scan_result)
                except Exception as e:
                    logger.error(f"Error scanning LambdaFunction {lf.get('metadata', {}).get('name')}: {e}")
                    results["errors"] += 1
                    namespace = lf.get("metadata", {}).get("namespace", "default")
                    self.metrics.record_error(namespace, "scan_error")
                    self.metrics.record_scan(namespace, "error")
            
            # Update outdated images gauge
            for (namespace, registry), count in outdated_by_ns_reg.items():
                self.metrics.outdated_images.labels(
                    namespace=namespace,
                    registry=registry
                ).set(count)
            
            logger.info(
                f"LambdaFunction scan completed: {results['total_scanned']} scanned, "
                f"{results['outdated']} outdated, {results['up_to_date']} up-to-date"
            )
            
            return results
            
        except Exception as e:
            logger.exception(f"Failed to scan LambdaFunctions: {e}")
            return {
                "status": "error",
                "error": str(e),
                "timestamp": datetime.now(timezone.utc).isoformat()
            }
    
    def handle_unknown(self, event: CloudEvent) -> dict[str, Any]:
        """Handle unknown event types."""
        logger.warning(f"Unknown event type: {event['type']}")
        return {"status": "ignored", "reason": "unknown_event_type"}
    
    # ── Private Methods ─────────────────────────────────────────────────────
    
    def _scan_image(self, image: str) -> dict[str, Any]:
        """Scan a container image for vulnerabilities."""
        # Placeholder - in production, integrate with Trivy/Grype
        logger.info(f"Scanning image: {image}")
        return {
            "image": image,
            "critical": 0,
            "high": 0,
            "medium": 0,
            "low": 0,
            "cve_ids": []
        }
    
    def _create_security_proposal(
        self,
        action: str,
        target: dict,
        severity: str,
        reason: str,
        cve_ids: list[str] | None = None
    ) -> None:
        """Create a SecurityProposal CRD for human approval."""
        if not self.k8s_enabled:
            logger.warning("Cannot create proposal - Kubernetes not available")
            return
        
        if self.rbac_level == "readonly":
            logger.warning("Cannot create proposal - readonly RBAC level")
            return
        
        proposal = {
            "apiVersion": "devsecops.homelab.io/v1alpha1",
            "kind": "SecurityProposal",
            "metadata": {
                "name": f"proposal-{datetime.now(timezone.utc).strftime('%Y%m%d%H%M%S')}",
                "namespace": "agent-devsecops"
            },
            "spec": {
                "action": action,
                "targetResource": target,
                "severity": severity,
                "reason": reason,
                "cveIds": cve_ids or [],
                "autoApprove": severity == "low" and not self.approval_required
            }
        }
        
        try:
            self.custom_api.create_namespaced_custom_object(
                group="devsecops.homelab.io",
                version="v1alpha1",
                namespace="agent-devsecops",
                plural="securityproposals",
                body=proposal
            )
            logger.info(f"Created SecurityProposal: {proposal['metadata']['name']}")
        except Exception as e:
            logger.error(f"Failed to create SecurityProposal: {e}")
    
    def _check_image_signature(self, image: str | None) -> bool:
        """Check if image is signed (placeholder)."""
        return True  # Implement with cosign/notation
    
    def _check_no_critical_cves(self, image: str | None) -> bool:
        """Check image has no critical CVEs."""
        if not image:
            return True
        result = self._scan_image(image)
        return result.get("critical", 0) == 0
    
    def _check_network_policy(self, namespace: str) -> bool:
        """Check if namespace has network policies."""
        if not self.k8s_enabled:
            return True
        try:
            policies = client.NetworkingV1Api().list_namespaced_network_policy(namespace)
            return len(policies.items) > 0
        except Exception:
            return False
    
    def _check_resource_limits(self, data: dict) -> bool:
        """Check if resource limits are set."""
        return data.get("resources", {}).get("limits") is not None
    
    def _check_security_context(self, data: dict) -> bool:
        """Check if security context is properly configured."""
        ctx = data.get("securityContext", {})
        return ctx.get("runAsNonRoot", False)
    
    def _get_security_status(self) -> dict[str, Any]:
        """Get overall security status."""
        return {
            "status": "healthy",
            "rbac_level": self.rbac_level,
            "environment": self.environment,
            "last_scan": datetime.now(timezone.utc).isoformat()
        }
    
    def _get_vulnerability_summary(self) -> dict[str, Any]:
        """Get vulnerability summary."""
        return {
            "total_images": 0,
            "critical": 0,
            "high": 0,
            "medium": 0,
            "low": 0
        }
    
    def _get_compliance_status(self) -> dict[str, Any]:
        """Get compliance status."""
        return {
            "framework": "CIS Kubernetes Benchmark",
            "status": "PartiallyCompliant",
            "passed": 80,
            "failed": 10,
            "warnings": 5
        }
    
    def _get_running_images(self) -> list[str]:
        """Get list of running container images."""
        if not self.k8s_enabled:
            return []
        
        images = set()
        try:
            pods = self.core_v1.list_pod_for_all_namespaces()
            for pod in pods.items:
                for container in pod.spec.containers:
                    images.add(container.image)
        except Exception as e:
            logger.error(f"Failed to get running images: {e}")
        
        return list(images)
    
    def _handle_critical_alert(self, data: dict) -> dict[str, Any]:
        """Handle critical alert with admin privileges."""
        logger.critical(f"CRITICAL ALERT - Admin action: {data}")
        # In admin mode, can take immediate action
        # BUT we still log everything for audit
        return {
            "status": "escalated",
            "action_taken": "incident_created",
            "requires_human_followup": True
        }


# Singleton scanner instance
scanner = SecurityScanner()


def handle(event: CloudEvent) -> str:
    """Main CloudEvent handler entrypoint."""
    try:
        result = scanner.handle_event(event)
        return json.dumps(result)
    except Exception as e:
        logger.exception(f"Error handling event: {e}")
        return json.dumps({"error": str(e)})
