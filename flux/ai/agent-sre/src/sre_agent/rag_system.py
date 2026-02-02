"""
RAG (Retrieval Augmented Generation) - Phase 3: Vector Store and Similarity Search

Provides historical context and similar past incidents to improve AI decision-making.
"""
from typing import Dict, Any, List, Optional
import json
import structlog
from datetime import datetime
import hashlib

try:
    from sentence_transformers import SentenceTransformer
    SENTENCE_TRANSFORMERS_AVAILABLE = True
except ImportError:
    SENTENCE_TRANSFORMERS_AVAILABLE = False
    logger = structlog.get_logger()
    logger.warning("sentence_transformers not available, RAG will use simple text matching")

logger = structlog.get_logger()


class AlertEmbedding:
    """Represents an alert with its embedding."""
    
    def __init__(
        self,
        alertname: str,
        labels: Dict[str, Any],
        annotations: Dict[str, Any],
        lambda_function: Optional[str] = None,
        parameters: Optional[Dict[str, Any]] = None,
        success: Optional[bool] = None,
        embedding: Optional[List[float]] = None,
        timestamp: Optional[datetime] = None
    ):
        self.alertname = alertname
        self.labels = labels
        self.annotations = annotations
        self.lambda_function = lambda_function
        self.parameters = parameters
        self.success = success
        self.embedding = embedding
        self.timestamp = timestamp or datetime.utcnow()
        self.id = self._generate_id()
    
    def _generate_id(self) -> str:
        """Generate unique ID for this alert."""
        content = f"{self.alertname}:{json.dumps(self.labels, sort_keys=True)}"
        return hashlib.sha256(content.encode()).hexdigest()[:16]
    
    def to_text(self) -> str:
        """Convert alert to text representation for embedding."""
        parts = [
            f"Alert: {self.alertname}",
            f"Labels: {json.dumps(self.labels)}",
        ]
        
        if self.lambda_function:
            parts.append(f"Remediation: {self.lambda_function}")
            parts.append(f"Parameters: {json.dumps(self.parameters or {})}")
        
        if self.success is not None:
            parts.append(f"Success: {self.success}")
        
        return " | ".join(parts)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for storage."""
        return {
            "id": self.id,
            "alertname": self.alertname,
            "labels": self.labels,
            "annotations": self.annotations,
            "lambda_function": self.lambda_function,
            "parameters": self.parameters,
            "success": self.success,
            "timestamp": self.timestamp.isoformat(),
            "embedding": self.embedding
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "AlertEmbedding":
        """Create from dictionary."""
        return cls(
            alertname=data["alertname"],
            labels=data["labels"],
            annotations=data.get("annotations", {}),
            lambda_function=data.get("lambda_function"),
            parameters=data.get("parameters"),
            success=data.get("success"),
            embedding=data.get("embedding"),
            timestamp=datetime.fromisoformat(data.get("timestamp", datetime.utcnow().isoformat()))
        )


class SimpleVectorStore:
    """Simple in-memory vector store (can be replaced with Chroma/Qdrant later)."""
    
    def __init__(self, embedding_model: Optional[str] = None):
        """
        Initialize vector store.
        
        Args:
            embedding_model: Name of sentence transformer model (default: all-MiniLM-L6-v2)
        """
        self.embeddings: List[AlertEmbedding] = []
        self.embedding_model_name = embedding_model or "all-MiniLM-L6-v2"
        self.encoder = None
        
        if SENTENCE_TRANSFORMERS_AVAILABLE:
            try:
                self.encoder = SentenceTransformer(self.embedding_model_name)
                logger.info("embedding_model_loaded", model=self.embedding_model_name)
            except Exception as e:
                logger.warning("failed_to_load_embedding_model", error=str(e))
        else:
            logger.warning("sentence_transformers_not_available", fallback="simple_matching")
    
    def _encode(self, text: str) -> List[float]:
        """Encode text to embedding vector."""
        if self.encoder:
            return self.encoder.encode(text, convert_to_numpy=False).tolist()
        else:
            # Fallback: simple hash-based "embedding" (not semantic, but works)
            hash_val = hash(text)
            # Create a simple 128-dim vector from hash
            return [(hash_val >> i) & 1 for i in range(128)]
    
    def _cosine_similarity(self, vec1: List[float], vec2: List[float]) -> float:
        """Calculate cosine similarity between two vectors."""
        if len(vec1) != len(vec2):
            return 0.0
        
        dot_product = sum(a * b for a, b in zip(vec1, vec2))
        magnitude1 = sum(a * a for a in vec1) ** 0.5
        magnitude2 = sum(b * b for b in vec2) ** 0.5
        
        if magnitude1 == 0 or magnitude2 == 0:
            return 0.0
        
        return dot_product / (magnitude1 * magnitude2)
    
    def add_alert(
        self,
        alertname: str,
        labels: Dict[str, Any],
        annotations: Optional[Dict[str, Any]] = None,
        lambda_function: Optional[str] = None,
        parameters: Optional[Dict[str, Any]] = None,
        success: Optional[bool] = None
    ) -> str:
        """Add an alert to the vector store."""
        alert = AlertEmbedding(
            alertname=alertname,
            labels=labels,
            annotations=annotations or {},
            lambda_function=lambda_function,
            parameters=parameters,
            success=success
        )
        
        # Generate embedding
        text = alert.to_text()
        alert.embedding = self._encode(text)
        
        self.embeddings.append(alert)
        
        # Keep only recent alerts (last 5000)
        if len(self.embeddings) > 5000:
            self.embeddings = sorted(
                self.embeddings,
                key=lambda x: x.timestamp,
                reverse=True
            )[:5000]
        
        logger.debug("alert_added_to_vector_store", alert_id=alert.id, alertname=alertname)
        
        return alert.id
    
    def similarity_search(
        self,
        alertname: str,
        labels: Dict[str, Any],
        top_k: int = 5,
        min_similarity: float = 0.3,
        only_successful: bool = True
    ) -> List[AlertEmbedding]:
        """
        Find similar alerts using vector similarity.
        
        Args:
            alertname: Name of the alert
            labels: Alert labels
            top_k: Number of results to return
            min_similarity: Minimum similarity threshold (0-1)
            only_successful: Only return successful remediations
            
        Returns:
            List of similar alerts, sorted by similarity
        """
        # Create query embedding
        query_alert = AlertEmbedding(alertname=alertname, labels=labels, annotations={})
        query_text = query_alert.to_text()
        query_embedding = self._encode(query_text)
        
        # Calculate similarities
        candidates = []
        
        for alert in self.embeddings:
            # Filter by success if requested
            if only_successful and alert.success is False:
                continue
            
            # Calculate cosine similarity
            similarity = self._cosine_similarity(query_embedding, alert.embedding)
            
            if similarity >= min_similarity:
                candidates.append((similarity, alert))
        
        # Sort by similarity (descending)
        candidates.sort(key=lambda x: x[0], reverse=True)
        
        # Return top K
        results = [alert for _, alert in candidates[:top_k]]
        
        logger.debug(
            "similarity_search_completed",
            query_alertname=alertname,
            results_count=len(results),
            top_similarity=candidates[0][0] if candidates else 0.0
        )
        
        return results


class RemediationRAG:
    """RAG system for remediation selection."""
    
    def __init__(
        self,
        vector_store: Optional[SimpleVectorStore] = None,
        embedding_model: Optional[str] = None
    ):
        """
        Initialize RAG system.
        
        Args:
            vector_store: Vector store instance (creates new if None)
            embedding_model: Embedding model name
        """
        self.vector_store = vector_store or SimpleVectorStore(embedding_model=embedding_model)
    
    async def find_similar_alerts(
        self,
        alert_data: Dict[str, Any],
        top_k: int = 5
    ) -> List[Dict[str, Any]]:
        """
        Find similar past alerts and their remediation.
        
        Args:
            alert_data: Current alert data
            top_k: Number of similar alerts to return
            
        Returns:
            List of similar alerts with remediation info
        """
        labels = alert_data.get("labels", {})
        alertname = labels.get("alertname", "unknown")
        
        similar = self.vector_store.similarity_search(
            alertname=alertname,
            labels=labels,
            top_k=top_k,
            only_successful=True
        )
        
        return [
            {
                "alertname": alert.alertname,
                "labels": alert.labels,
                "lambda_function": alert.lambda_function,
                "parameters": alert.parameters,
                "success": alert.success,
                "similarity": 0.8  # Would be calculated in real implementation
            }
            for alert in similar
        ]
    
    def create_rag_prompt(
        self,
        alert_data: Dict[str, Any],
        similar_alerts: List[Dict[str, Any]]
    ) -> str:
        """Create prompt with RAG context."""
        prompt = f"Current Alert: {json.dumps(alert_data, indent=2)}\n\n"
        
        if similar_alerts:
            prompt += "Similar Past Incidents:\n"
            for i, alert in enumerate(similar_alerts, 1):
                prompt += f"\n{i}. Alert: {alert['alertname']}\n"
                prompt += f"   Labels: {json.dumps(alert['labels'], indent=2)}\n"
                if alert.get('lambda_function'):
                    prompt += f"   Remediation: {alert['lambda_function']}\n"
                    prompt += f"   Parameters: {json.dumps(alert['parameters'], indent=2)}\n"
                prompt += f"   Success: {alert.get('success', 'Unknown')}\n"
            
            prompt += "\nBased on these similar incidents, select the appropriate remediation for the current alert.\n"
        else:
            prompt += "No similar past incidents found. Use your knowledge to select the appropriate remediation.\n"
        
        return prompt
    
    def index_alert(
        self,
        alert_data: Dict[str, Any],
        lambda_function: Optional[str] = None,
        parameters: Optional[Dict[str, Any]] = None,
        success: Optional[bool] = None
    ):
        """Index an alert in the vector store."""
        labels = alert_data.get("labels", {})
        annotations = alert_data.get("annotations", {})
        alertname = labels.get("alertname", "unknown")
        
        self.vector_store.add_alert(
            alertname=alertname,
            labels=labels,
            annotations=annotations,
            lambda_function=lambda_function,
            parameters=parameters,
            success=success
        )


# Global RAG instance
_rag_instance: Optional[RemediationRAG] = None


def get_rag_instance(embedding_model: Optional[str] = None) -> RemediationRAG:
    """Get or create global RAG instance."""
    global _rag_instance
    
    if _rag_instance is None:
        _rag_instance = RemediationRAG(embedding_model=embedding_model)
    
    return _rag_instance

