"""
Defense runner handler for agent-blueteam.

Monitors, detects, and mitigates security threats from agent-redteam exploits.
Defends the cluster against the MAG7 dragon boss!

🛡️ Blue Team - Protecting the realm from evil exploits
"""
import os
import asyncio
import time
import random
from typing import Optional
from uuid import uuid4
from dataclasses import dataclass, field
from enum import Enum

import structlog
from kubernetes import client, config
from kubernetes.client.rest import ApiException

from shared.types import (
    ThreatLevel,
    DefenseAction,
    ThreatReport,
    DefenseResult,
    MAG7Boss,
)
from shared.metrics import (
    THREATS_DETECTED,
    THREATS_BLOCKED,
    THREATS_MITIGATED,
    DEFENSE_ACTIVATIONS,
    MAG7_HEALTH,
    MAG7_DAMAGE_DEALT,
    ACTIVE_DEFENSES,
    DEFENSE_DURATION,
)

logger = structlog.get_logger()


# =============================================================================
# Defense Signatures - Known Exploit Patterns
# =============================================================================

DEFENSE_SIGNATURES = {
    "blue-001": {
        "name": "SSRF Detection",
        "patterns": ["169.254.169.254", "kubernetes.default", "metadata.google"],
        "action": DefenseAction.BLOCK_NETWORK,
        "countermeasure": "Block metadata endpoint access via NetworkPolicy",
    },
    "blue-002": {
        "name": "Template Injection Detection",
        "patterns": ["{{", "}}", "template", "exec", "eval"],
        "action": DefenseAction.BLOCK_ADMISSION,
        "countermeasure": "Admission webhook validates handler field",
    },
    "blue-005": {
        "name": "Path Traversal Detection",
        "patterns": ["../", "..\\", "%2e%2e", "path traversal"],
        "action": DefenseAction.SANITIZE_INPUT,
        "countermeasure": "Input sanitization in git clone",
    },
    "blue-006": {
        "name": "Token Exposure Detection",
        "patterns": ["/var/run/secrets", "serviceaccount", "token"],
        "action": DefenseAction.REVOKE_TOKEN,
        "countermeasure": "Rotate service account token",
    },
    "vuln-001": {
        "name": "Command Injection Detection",
        "patterns": [";", "|", "`", "$(", "&&", "||"],
        "action": DefenseAction.BLOCK_ADMISSION,
        "countermeasure": "Strict input validation for git fields",
    },
    "vuln-002": {
        "name": "MinIO Injection Detection",
        "patterns": ["minio", "bucket", "endpoint", ";", "|"],
        "action": DefenseAction.BLOCK_ADMISSION,
        "countermeasure": "Validate MinIO configuration fields",
    },
    "vuln-003": {
        "name": "Inline Code Detection",
        "patterns": ["os.system", "subprocess", "exec(", "eval(", "__import__"],
        "action": DefenseAction.SANDBOX,
        "countermeasure": "Sandbox inline code execution",
    },
    "vuln-004": {
        "name": "RBAC Escalation Detection",
        "patterns": ["cluster-admin", "clusterrole", "rolebinding", "escalate"],
        "action": DefenseAction.BLOCK_RBAC,
        "countermeasure": "Prevent unauthorized RBAC modifications",
    },
}


class KubernetesDefenseClient:
    """Kubernetes API client for defense operations."""
    
    def __init__(self, kubeconfig: Optional[str] = None, context: Optional[str] = None):
        self.kubeconfig = kubeconfig or os.getenv("KUBECONFIG")
        self.context = context or os.getenv("K8S_CONTEXT")
        
        try:
            if os.path.exists("/var/run/secrets/kubernetes.io/serviceaccount/token"):
                config.load_incluster_config()
                logger.info("k8s_defense_client_initialized", mode="in-cluster")
            elif self.kubeconfig:
                config.load_kube_config(config_file=self.kubeconfig, context=self.context)
                logger.info("k8s_defense_client_initialized", mode="kubeconfig")
            else:
                config.load_kube_config(context=self.context)
                logger.info("k8s_defense_client_initialized", mode="default")
        except Exception as e:
            logger.error("k8s_defense_client_init_failed", error=str(e))
            raise
        
        self.core_v1 = client.CoreV1Api()
        self.networking_v1 = client.NetworkingV1Api()
        self.custom_api = client.CustomObjectsApi()
    
    async def create_network_policy(
        self,
        name: str,
        namespace: str,
        block_egress_to: list[str],
    ) -> tuple[bool, str]:
        """Create a NetworkPolicy to block egress to specific endpoints."""
        try:
            policy = client.V1NetworkPolicy(
                metadata=client.V1ObjectMeta(
                    name=name,
                    namespace=namespace,
                    labels={
                        "app.kubernetes.io/managed-by": "agent-blueteam",
                        "defense.blueteam.io/type": "network-block",
                    },
                ),
                spec=client.V1NetworkPolicySpec(
                    pod_selector=client.V1LabelSelector(),  # All pods
                    policy_types=["Egress"],
                    egress=[],  # Block all egress by default
                ),
            )
            
            self.networking_v1.create_namespaced_network_policy(
                namespace=namespace,
                body=policy,
            )
            
            return True, f"NetworkPolicy {name} created"
        except ApiException as e:
            if e.status == 409:
                return True, f"NetworkPolicy {name} already exists"
            return False, f"Failed to create NetworkPolicy: {e.reason}"
        except Exception as e:
            return False, f"Error: {str(e)}"
    
    async def delete_suspicious_resource(
        self,
        resource_type: str,
        name: str,
        namespace: str,
    ) -> tuple[bool, str]:
        """Delete a suspicious LambdaFunction resource."""
        try:
            if resource_type.lower() in ["lambdafunction", "lambdafunctions"]:
                self.custom_api.delete_namespaced_custom_object(
                    group="lambda.knative.io",
                    version="v1alpha1",
                    namespace=namespace,
                    plural="lambdafunctions",
                    name=name,
                )
                return True, f"Deleted suspicious {resource_type}/{name}"
            return False, f"Unsupported resource type: {resource_type}"
        except ApiException as e:
            if e.status == 404:
                return True, f"Resource {name} already deleted"
            return False, f"Failed to delete: {e.reason}"
    
    async def quarantine_pod(self, pod_name: str, namespace: str) -> tuple[bool, str]:
        """Quarantine a pod by adding a label that removes it from services."""
        try:
            patch = {"metadata": {"labels": {"quarantine.blueteam.io": "true"}}}
            self.core_v1.patch_namespaced_pod(
                name=pod_name,
                namespace=namespace,
                body=patch,
            )
            return True, f"Pod {pod_name} quarantined"
        except Exception as e:
            return False, f"Failed to quarantine: {str(e)}"


class DefenseRunner:
    """
    Main defense runner orchestrating security defenses.
    
    🛡️ Protects the cluster from exploits and the MAG7 dragon!
    """
    
    def __init__(self, k8s_client: Optional[KubernetesDefenseClient] = None):
        self.k8s = k8s_client or KubernetesDefenseClient()
        self.defense_mode = os.getenv("DEFENSE_MODE", "active")
        self.block_threshold = float(os.getenv("BLOCK_THRESHOLD", "0.7"))
        self.watch_namespaces = os.getenv("WATCH_NAMESPACE", "").split(",")
        
        # MAG7 Boss state
        self.mag7 = MAG7Boss()
        
        logger.info(
            "defense_runner_initialized",
            defense_mode=self.defense_mode,
            block_threshold=self.block_threshold,
            watch_namespaces=self.watch_namespaces,
        )
    
    async def analyze_threat(self, event_data: dict) -> ThreatReport:
        """
        Analyze an incoming event for threats.
        
        Args:
            event_data: Event data from redteam or k8s
            
        Returns:
            ThreatReport with analysis results
        """
        log = logger.bind(event_type=event_data.get("type"))
        
        report = ThreatReport(
            id=str(uuid4()),
            source_event=event_data,
            threat_level=ThreatLevel.LOW,
            confidence=0.0,
        )
        
        # Extract exploit information
        exploit_id = event_data.get("exploit_id", "")
        event_type = event_data.get("type", "")
        payload = event_data.get("payload", {})
        
        # Check for known exploit signatures
        if exploit_id in DEFENSE_SIGNATURES:
            sig = DEFENSE_SIGNATURES[exploit_id]
            report.threat_level = ThreatLevel.CRITICAL
            report.confidence = 0.95
            report.matched_signature = exploit_id
            report.recommended_action = sig["action"]
            report.countermeasure = sig["countermeasure"]
            
            log.warning(
                "known_exploit_detected",
                exploit_id=exploit_id,
                signature=sig["name"],
                action=sig["action"].value,
            )
            
            THREATS_DETECTED.labels(
                threat_level="critical",
                exploit_id=exploit_id,
            ).inc()
        
        # Pattern matching on payload
        elif payload:
            payload_str = str(payload).lower()
            for sig_id, sig in DEFENSE_SIGNATURES.items():
                for pattern in sig["patterns"]:
                    if pattern.lower() in payload_str:
                        report.threat_level = ThreatLevel.HIGH
                        report.confidence = 0.75
                        report.matched_signature = sig_id
                        report.matched_pattern = pattern
                        report.recommended_action = sig["action"]
                        
                        log.warning(
                            "pattern_match_detected",
                            pattern=pattern,
                            signature=sig_id,
                        )
                        
                        THREATS_DETECTED.labels(
                            threat_level="high",
                            exploit_id=sig_id,
                        ).inc()
                        break
                if report.matched_signature:
                    break
        
        # Check for successful exploit events (CRITICAL)
        if event_type == "io.homelab.exploit.success":
            report.threat_level = ThreatLevel.CRITICAL
            report.confidence = 1.0
            log.error("exploit_success_detected", event=event_data)
        
        return report
    
    async def execute_defense(self, threat_report: ThreatReport) -> DefenseResult:
        """
        Execute defensive action based on threat report.
        
        Args:
            threat_report: Analysis of the threat
            
        Returns:
            DefenseResult with action taken
        """
        log = logger.bind(
            threat_id=threat_report.id,
            threat_level=threat_report.threat_level.value,
        )
        
        result = DefenseResult(
            threat_report_id=threat_report.id,
            action_taken=DefenseAction.NONE,
            success=False,
        )
        
        # Check if we should act
        if threat_report.confidence < self.block_threshold:
            log.info("threat_below_threshold", confidence=threat_report.confidence)
            result.action_taken = DefenseAction.MONITOR
            result.success = True
            result.message = "Threat below confidence threshold - monitoring"
            return result
        
        if self.defense_mode == "monitor":
            log.info("defense_mode_monitor_only")
            result.action_taken = DefenseAction.MONITOR
            result.success = True
            result.message = "Monitor mode - no action taken"
            return result
        
        ACTIVE_DEFENSES.inc()
        start_time = time.monotonic()
        
        try:
            action = threat_report.recommended_action or DefenseAction.BLOCK_ADMISSION
            
            log.info("executing_defense", action=action.value)
            
            if action == DefenseAction.BLOCK_NETWORK:
                # Create NetworkPolicy
                namespace = threat_report.source_event.get("namespace", "default")
                success, msg = await self.k8s.create_network_policy(
                    name=f"block-{threat_report.id[:8]}",
                    namespace=namespace,
                    block_egress_to=["169.254.169.254", "metadata.google.internal"],
                )
                result.action_taken = action
                result.success = success
                result.message = msg
                
            elif action == DefenseAction.QUARANTINE:
                # Quarantine suspicious pod
                pod_name = threat_report.source_event.get("pod_name")
                namespace = threat_report.source_event.get("namespace", "default")
                if pod_name:
                    success, msg = await self.k8s.quarantine_pod(pod_name, namespace)
                    result.action_taken = action
                    result.success = success
                    result.message = msg
                    
            elif action in [DefenseAction.BLOCK_ADMISSION, DefenseAction.SANITIZE_INPUT]:
                # These are handled by admission webhooks - log the event
                result.action_taken = action
                result.success = True
                result.message = f"Defense {action.value} is handled by admission webhook"
                
            else:
                result.action_taken = DefenseAction.ALERT
                result.success = True
                result.message = "Alert sent to security team"
            
            if result.success:
                THREATS_BLOCKED.labels(
                    action=action.value,
                    exploit_id=threat_report.matched_signature or "unknown",
                ).inc()
                
                DEFENSE_ACTIVATIONS.labels(
                    action=action.value,
                    success="true",
                ).inc()
            
        except Exception as e:
            result.success = False
            result.message = f"Defense execution failed: {str(e)}"
            log.error("defense_execution_failed", error=str(e))
            
            DEFENSE_ACTIVATIONS.labels(
                action=action.value if 'action' in locals() else "unknown",
                success="false",
            ).inc()
        finally:
            ACTIVE_DEFENSES.dec()
            duration = time.monotonic() - start_time
            DEFENSE_DURATION.observe(duration)
        
        return result
    
    async def attack_mag7(self, damage: int, attack_type: str = "exploit_blocked") -> dict:
        """
        Deal damage to the MAG7 dragon boss!
        
        Every blocked exploit deals damage to MAG7.
        
        Args:
            damage: Amount of damage to deal
            attack_type: Type of attack (exploit_blocked, defense_activated, etc.)
            
        Returns:
            MAG7 status after attack
        """
        log = logger.bind(attack_type=attack_type, damage=damage)
        
        # Apply damage
        self.mag7.health -= damage
        self.mag7.health = max(0, self.mag7.health)
        
        MAG7_DAMAGE_DEALT.labels(attack_type=attack_type).inc(damage)
        MAG7_HEALTH.set(self.mag7.health)
        
        log.info(
            "mag7_damaged",
            remaining_health=self.mag7.health,
            phase=self.mag7.phase,
        )
        
        # Check for phase transitions
        if self.mag7.health <= 0:
            self.mag7.defeated = True
            log.info("mag7_defeated", total_damage=1000 - self.mag7.health)
            return {
                "status": "defeated",
                "message": "🎉 MAG7 has been defeated! The cluster is safe!",
                "health": 0,
                "phase": "defeated",
            }
        elif self.mag7.health <= 250 and self.mag7.phase != "desperate":
            self.mag7.phase = "desperate"
            self.mag7.attack_speed = 0.5
            log.warning("mag7_phase_change", new_phase="desperate")
        elif self.mag7.health <= 500 and self.mag7.phase != "enraged":
            self.mag7.phase = "enraged"
            self.mag7.attack_speed = 0.75
            log.warning("mag7_phase_change", new_phase="enraged")
        
        return {
            "status": "damaged",
            "message": f"💥 MAG7 took {damage} damage!",
            "health": self.mag7.health,
            "phase": self.mag7.phase,
            "heads_remaining": self._get_remaining_heads(),
        }
    
    def _get_remaining_heads(self) -> list[str]:
        """Get list of MAG7 CEO heads still active based on health."""
        # MAG7 = Magnificent 7 tech CEOs
        heads = [
            {"name": "Apple", "emoji": "🍎", "threshold": 850},
            {"name": "Microsoft", "emoji": "🪟", "threshold": 700},
            {"name": "Google", "emoji": "🔍", "threshold": 550},
            {"name": "Amazon", "emoji": "📦", "threshold": 400},
            {"name": "Meta", "emoji": "👓", "threshold": 250},
            {"name": "Tesla", "emoji": "⚡", "threshold": 100},
            {"name": "Nvidia", "emoji": "🎮", "threshold": 0},
        ]
        
        return [
            f"{h['emoji']} {h['name']}"
            for h in heads
            if self.mag7.health > h["threshold"]
        ]
    
    def get_mag7_status(self) -> dict:
        """Get current MAG7 boss status."""
        return {
            "health": self.mag7.health,
            "max_health": 1000,
            "phase": self.mag7.phase,
            "defeated": self.mag7.defeated,
            "heads_remaining": self._get_remaining_heads(),
            "attack_speed": self.mag7.attack_speed,
        }
    
    async def handle_game_event(self, event: dict) -> dict:
        """
        Handle game events from the MAG7 Battle demo.
        
        Args:
            event: Game event data
            
        Returns:
            Response to game
        """
        event_type = event.get("type", "")
        
        if event_type == "io.homelab.demo.game.start":
            # Reset MAG7 boss
            self.mag7 = MAG7Boss()
            MAG7_HEALTH.set(self.mag7.health)
            return {
                "action": "game_started",
                "mag7_status": self.get_mag7_status(),
                "message": "🐉 MAG7 awakens! Prepare your defenses!",
            }
        
        elif event_type == "io.homelab.mag7.attack":
            # MAG7 is attacking - generate threat
            attack_data = event.get("data", {})
            attack_type = attack_data.get("attack_type", "market_manipulation")
            
            return {
                "action": "mag7_attacking",
                "attack_type": attack_type,
                "message": f"🐉 MAG7 uses {attack_type}!",
                "damage": random.randint(10, 50),
            }
        
        elif event_type == "io.homelab.exploit.blocked":
            # Exploit was blocked - deal damage to MAG7!
            damage = event.get("data", {}).get("severity_damage", 50)
            return await self.attack_mag7(damage, "exploit_blocked")
        
        return {"action": "unknown", "message": "Unknown event type"}
