"""Flux reconciliation capability for agent-sre."""
import subprocess
from typing import Optional, Dict, Any, List
import structlog

logger = structlog.get_logger()


class FluxReconciler:
    """Reconciles Flux resources based on PrometheusRule triggers."""
    
    def __init__(self, namespace: str = "flux-system"):
        self.namespace = namespace
    
    def reconcile_kustomization(self, name: str, namespace: Optional[str] = None) -> bool:
        """Reconcile a Kustomization resource."""
        ns = namespace or self.namespace
        try:
            result = subprocess.run(
                ["flux", "reconcile", "kustomization", name, "-n", ns],
                capture_output=True,
                text=True,
                timeout=60
            )
            
            if result.returncode == 0:
                logger.info(
                    "flux_kustomization_reconciled",
                    name=name,
                    namespace=ns
                )
                return True
            else:
                logger.error(
                    "flux_kustomization_reconcile_failed",
                    name=name,
                    namespace=ns,
                    error=result.stderr
                )
                return False
        except subprocess.TimeoutExpired:
            logger.error("flux_reconcile_timeout", name=name, namespace=ns)
            return False
        except Exception as e:
            logger.error("flux_reconcile_error", name=name, namespace=ns, error=str(e))
            return False
    
    def reconcile_gitrepository(self, name: str, namespace: Optional[str] = None) -> bool:
        """Reconcile a GitRepository resource."""
        ns = namespace or self.namespace
        try:
            result = subprocess.run(
                ["flux", "reconcile", "source", "git", name, "-n", ns],
                capture_output=True,
                text=True,
                timeout=60
            )
            
            if result.returncode == 0:
                logger.info(
                    "flux_gitrepository_reconciled",
                    name=name,
                    namespace=ns
                )
                return True
            else:
                logger.error(
                    "flux_gitrepository_reconcile_failed",
                    name=name,
                    namespace=ns,
                    error=result.stderr
                )
                return False
        except subprocess.TimeoutExpired:
            logger.error("flux_reconcile_timeout", name=name, namespace=ns)
            return False
        except Exception as e:
            logger.error("flux_reconcile_error", name=name, namespace=ns, error=str(e))
            return False
    
    def reconcile_helmrelease(self, name: str, namespace: Optional[str] = None) -> bool:
        """Reconcile a HelmRelease resource."""
        ns = namespace or self.namespace
        try:
            result = subprocess.run(
                ["flux", "reconcile", "helmrelease", name, "-n", ns],
                capture_output=True,
                text=True,
                timeout=60
            )
            
            if result.returncode == 0:
                logger.info(
                    "flux_helmrelease_reconciled",
                    name=name,
                    namespace=ns
                )
                return True
            else:
                logger.error(
                    "flux_helmrelease_reconcile_failed",
                    name=name,
                    namespace=ns,
                    error=result.stderr
                )
                return False
        except subprocess.TimeoutExpired:
            logger.error("flux_reconcile_timeout", name=name, namespace=ns)
            return False
        except Exception as e:
            logger.error("flux_reconcile_error", name=name, namespace=ns, error=str(e))
            return False
    
    def reconcile_from_prometheus_rule(
        self,
        prometheus_rule: str,
        alert_labels: Dict[str, Any]
    ) -> List[bool]:
        """
        Reconcile Flux resources based on PrometheusRule annotation.
        
        Expected format in PrometheusRule annotations:
        flux.reconcile/kustomizations: "name1,namespace1/name2,namespace2"
        flux.reconcile/gitrepositories: "name1,namespace1"
        flux.reconcile/helmreleases: "name1,namespace1"
        """
        results = []
        
        # Parse reconciliation targets from alert labels/annotations
        # For now, we'll use a simple mapping - can be enhanced
        reconcile_targets = alert_labels.get("flux_reconcile", "")
        
        if not reconcile_targets:
            logger.warning(
                "no_flux_reconcile_targets",
                prometheus_rule=prometheus_rule
            )
            return results
        
        # Parse targets (format: "kind:name:namespace" or "kind:name")
        targets = reconcile_targets.split(",")
        for target in targets:
            target = target.strip()
            if not target:
                continue
            
            parts = target.split(":")
            if len(parts) < 2:
                continue
            
            kind = parts[0].lower()
            name = parts[1]
            namespace = parts[2] if len(parts) > 2 else None
            
            if kind == "kustomization":
                results.append(self.reconcile_kustomization(name, namespace))
            elif kind == "gitrepository":
                results.append(self.reconcile_gitrepository(name, namespace))
            elif kind == "helmrelease":
                results.append(self.reconcile_helmrelease(name, namespace))
            else:
                logger.warning("unknown_flux_kind", kind=kind, target=target)
        
        return results

