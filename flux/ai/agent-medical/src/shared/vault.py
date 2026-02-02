"""
Vault integration for secrets management.
"""
import os
from typing import Optional, Dict, Any
import structlog

try:
    import hvac
except ImportError:
    hvac = None

logger = structlog.get_logger()


class VaultClient:
    """Vault client for retrieving secrets."""
    
    def __init__(self):
        self.vault_addr = os.getenv("VAULT_ADDR", "http://vault.vault.svc.cluster.local:8200")
        self.vault_token = os.getenv("VAULT_TOKEN")
        self.client = None
        
        if hvac:
            try:
                self.client = hvac.Client(url=self.vault_addr)
                if self.vault_token:
                    self.client.token = self.vault_token
                logger.info("vault_client_initialized", vault_addr=self.vault_addr)
            except Exception as e:
                logger.warning("vault_client_init_failed", error=str(e))
        else:
            logger.warning("vault_client_not_available", note="hvac not installed")
    
    def get_secret(self, path: str, key: Optional[str] = None) -> Optional[Any]:
        """
        Get secret from Vault.
        
        Args:
            path: Secret path (e.g., "secret/data/medical-db")
            key: Optional key name (if None, returns entire secret)
        
        Returns:
            Secret value or None if not found
        """
        if not self.client:
            logger.warning("vault_client_not_available", path=path)
            return None
        
        try:
            # Vault KV v2 path format: secret/data/path
            if not path.startswith("secret/"):
                path = f"secret/data/{path}"
            
            response = self.client.secrets.kv.v2.read_secret_version(path=path)
            
            if response and "data" in response:
                data = response["data"].get("data", {})
                if key:
                    return data.get(key)
                return data
            
            logger.warning("vault_secret_not_found", path=path)
            return None
        
        except Exception as e:
            logger.error("vault_secret_read_failed", path=path, error=str(e))
            return None
    
    def get_db_credentials(self) -> Optional[Dict[str, str]]:
        """Get database credentials from Vault."""
        return {
            "url": self.get_secret("medical-db", "url") or os.getenv("MONGODB_URL"),
            "user": self.get_secret("medical-db", "user"),
            "password": self.get_secret("medical-db", "password"),
        }
    
    def get_sus_api_key(self) -> Optional[str]:
        """Get SUS Cloud API key from Vault."""
        return self.get_secret("sus-cloud", "api_key")
    
    def health_check(self) -> bool:
        """Check if Vault is accessible."""
        if not self.client:
            return False
        
        try:
            return self.client.sys.is_initialized()
        except Exception:
            return False


# Global vault client instance
_vault_client: Optional[VaultClient] = None


def get_vault_client() -> VaultClient:
    """Get or create Vault client instance."""
    global _vault_client
    if _vault_client is None:
        _vault_client = VaultClient()
    return _vault_client
