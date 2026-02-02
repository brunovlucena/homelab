"""
Audit logging system for HIPAA compliance.
"""
import os
from datetime import datetime
from typing import Optional, List, Dict, Any
from uuid import uuid4
import structlog

from .types import User, AuditLog
from .security import AccessControl
from .database import Database

logger = structlog.get_logger()
access_control = AccessControl()


class AuditLogger:
    """HIPAA-compliant audit logger."""
    
    def __init__(self, db: Optional[Database] = None):
        self.db = db
        self.enabled = os.getenv("HIPAA_MODE", "true").lower() == "true"
    
    async def log(
        self,
        user: User,
        action: str,
        patient_id: Optional[str] = None,
        query: Optional[str] = None,
        status: str = "success",
        ip_address: Optional[str] = None,
        audit_id: Optional[str] = None,
    ) -> str:
        """
        Log audit entry.
        
        Args:
            user: User performing the action
            action: Action performed (e.g., "read_record", "access_denied")
            patient_id: Patient ID (will be hashed)
            query: Query text (will be sanitized)
            status: Status (success, denied, error)
            ip_address: IP address (optional)
            audit_id: Audit ID (generated if not provided)
        
        Returns:
            Audit ID
        """
        if not self.enabled:
            return audit_id or str(uuid4())
        
        audit_id = audit_id or str(uuid4())
        # patient_id is already hashed if passed as patient_id_hash, otherwise hash it
        if patient_id and patient_id != "none" and len(patient_id) != 64:  # SHA256 hash is 64 chars
            patient_id_hash = access_control.hash_patient_id(patient_id)
        else:
            patient_id_hash = patient_id or "none"
        sanitized_query = access_control.sanitize_query(query) if query else None
        
        # Log to structured logging
        logger.info(
            "audit_log",
            audit_id=audit_id,
            user_id=user.id,
            user_role=user.role.value,
            patient_id_hash=patient_id_hash,
            action=action,
            query=sanitized_query,
            status=status,
            ip_address=ip_address,
            timestamp=datetime.utcnow().isoformat(),
        )
        
        # Store in database if available
        if self.db and self.db.db:
            try:
                await self.db.db.audit_logs.insert_one({
                    "_id": audit_id,
                    "timestamp": datetime.utcnow(),
                    "user_id": user.id,
                    "user_role": user.role.value,
                    "patient_id_hash": patient_id_hash,
                    "action": action,
                    "query": sanitized_query,
                    "status": status,
                    "ip_address": ip_address,
                    "audit_id": audit_id,
                })
            except Exception as e:
                logger.error("audit_log_db_failed", error=str(e))
        
        return audit_id
    
    async def query_audit_logs(
        self,
        user_id: Optional[str] = None,
        patient_id_hash: Optional[str] = None,
        action: Optional[str] = None,
        limit: int = 100,
    ) -> List[Dict[str, Any]]:
        """
        Query audit logs (admin only).
        
        Args:
            user_id: Filter by user ID
            patient_id_hash: Filter by patient ID hash
            action: Filter by action
            limit: Maximum number of results
        
        Returns:
            List of audit log entries
        """
        if not self.db:
            return []
        
        if not self.db or not self.db.db:
            return []
        
        try:
            query = {}
            if user_id:
                query["user_id"] = user_id
            if patient_id_hash:
                query["patient_id_hash"] = patient_id_hash
            if action:
                query["action"] = action
            
            cursor = self.db.db.audit_logs.find(query).sort("timestamp", -1).limit(limit)
            logs = await cursor.to_list(length=limit)
            
            # Convert _id to id
            for log in logs:
                log["id"] = str(log.pop("_id"))
            
            return logs
        
        except Exception as e:
            logger.error("audit_log_query_failed", error=str(e))
            return []
