"""
Image Scanner for LambdaFunctions

Scans LambdaFunction resources for outdated images and exposes Prometheus metrics.
"""

import logging
import re
from datetime import datetime, timezone
from typing import Any, Optional
from packaging import version as pkg_version

logger = logging.getLogger(__name__)


class ImageScanner:
    """Scans container images for version information and outdated status."""
    
    # Minimum expected versions by image prefix/pattern
    MIN_VERSIONS = {
        "ghcr.io/brunovlucena": "1.0.0",
        "localhost:5001": "0.1.0",
        "knative-lambdas": "0.1.0",
    }
    
    def __init__(self):
        self.min_versions = self.MIN_VERSIONS.copy()
    
    def extract_image_tag(self, image_uri: str) -> Optional[str]:
        """Extract tag from image URI (e.g., 'registry/image:tag' -> 'tag')."""
        if not image_uri:
            return None
        
        # Handle image URIs with tags
        if ":" in image_uri:
            tag = image_uri.split(":")[-1]
            # Remove digest if present (tag@sha256:...)
            if "@" in tag:
                tag = tag.split("@")[0]
            return tag if tag else None
        
        # Default to 'latest' if no tag
        return "latest"
    
    def extract_version_from_tag(self, tag: str) -> Optional[str]:
        """Extract semantic version from tag (e.g., 'v1.2.3' -> '1.2.3')."""
        if not tag:
            return None
        
        # Remove 'v' prefix if present
        tag = tag.lstrip("v")
        
        # Try to match semantic version pattern (MAJOR.MINOR.PATCH)
        version_pattern = r"^(\d+)\.(\d+)\.(\d+)(?:-[\w\.-]+)?(?:\+[\w\.-]+)?$"
        match = re.match(version_pattern, tag)
        if match:
            return tag.split("-")[0].split("+")[0]  # Remove pre-release and build metadata
        
        # Try to extract version from patterns like "1.2.3-dev", "v1.2.3-beta"
        version_match = re.search(r"(\d+\.\d+\.\d+)", tag)
        if version_match:
            return version_match.group(1)
        
        return None
    
    def is_outdated(self, image_uri: str, min_version: Optional[str] = None) -> tuple[bool, Optional[str], Optional[str]]:
        """
        Check if image is outdated.
        
        Returns:
            (is_outdated, current_version, min_version)
        """
        tag = self.extract_image_tag(image_uri)
        if not tag:
            return False, None, None
        
        current_version = self.extract_version_from_tag(tag)
        if not current_version:
            # Can't determine version, assume not outdated
            return False, None, None
        
        # Determine minimum version
        if not min_version:
            min_version = self._get_min_version_for_image(image_uri)
        
        if not min_version:
            # No minimum version defined, assume not outdated
            return False, current_version, None
        
        try:
            # Compare versions
            current = pkg_version.parse(current_version)
            minimum = pkg_version.parse(min_version)
            is_outdated = current < minimum
            return is_outdated, current_version, min_version
        except Exception as e:
            logger.warning(f"Failed to compare versions {current_version} vs {min_version}: {e}")
            return False, current_version, min_version
    
    def _get_min_version_for_image(self, image_uri: str) -> Optional[str]:
        """Get minimum expected version for an image based on registry/prefix."""
        for prefix, min_ver in self.min_versions.items():
            if prefix in image_uri:
                return min_ver
        return None
    
    def scan_lambdafunction(self, lf: dict[str, Any]) -> dict[str, Any]:
        """
        Scan a LambdaFunction resource for image information.
        
        Args:
            lf: LambdaFunction resource dict from Kubernetes API
        
        Returns:
            Dict with scan results
        """
        name = lf.get("metadata", {}).get("name", "unknown")
        namespace = lf.get("metadata", {}).get("namespace", "default")
        
        # Get image URI from status (built image) or spec (source image)
        image_uri = (
            lf.get("status", {}).get("buildStatus", {}).get("imageURI") or
            lf.get("spec", {}).get("source", {}).get("image") or
            None
        )
        
        if not image_uri:
            return {
                "name": name,
                "namespace": namespace,
                "image_uri": None,
                "tag": None,
                "version": None,
                "is_outdated": False,
                "min_version": None,
                "error": "No image URI found"
            }
        
        tag = self.extract_image_tag(image_uri)
        version = self.extract_version_from_tag(tag) if tag else None
        is_outdated, detected_version, min_version = self.is_outdated(image_uri)
        
        return {
            "name": name,
            "namespace": namespace,
            "image_uri": image_uri,
            "tag": tag,
            "version": detected_version or version,
            "is_outdated": is_outdated,
            "min_version": min_version,
            "error": None
        }
