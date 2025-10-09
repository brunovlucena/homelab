package cdn

import (
	"fmt"
	"strings"
)

// CDNManager handles CDN configuration (Golden Rule #7: CDN)
type CDNManager struct {
	baseURL  string
	enabled  bool
	cacheTTL int
}

// NewCDNManager creates a new CDN manager
func NewCDNManager(baseURL string, enabled bool, cacheTTL int) *CDNManager {
	return &CDNManager{
		baseURL:  baseURL,
		enabled:  enabled,
		cacheTTL: cacheTTL,
	}
}

// GetAssetURL returns the CDN URL for a static asset
func (c *CDNManager) GetAssetURL(assetPath string) string {
	if !c.enabled || c.baseURL == "" {
		return assetPath
	}

	// Ensure asset path starts with /
	if !strings.HasPrefix(assetPath, "/") {
		assetPath = "/" + assetPath
	}

	// Remove leading slash from baseURL if present
	baseURL := strings.TrimSuffix(c.baseURL, "/")

	return fmt.Sprintf("%s%s", baseURL, assetPath)
}

// GetImageURL returns optimized image URL with CDN
func (c *CDNManager) GetImageURL(imagePath string, width, height int, format string) string {
	if !c.enabled {
		return imagePath
	}

	// For image optimization services like Cloudinary, ImageKit, etc.
	// Example: https://cdn.example.com/w_300,h_200,f_webp/image.jpg
	if width > 0 && height > 0 && format != "" {
		optimizedPath := fmt.Sprintf("w_%d,h_%d,f_%s%s", width, height, format, imagePath)
		return c.GetAssetURL(optimizedPath)
	}

	return c.GetAssetURL(imagePath)
}

// GetCacheHeaders returns CDN cache headers
func (c *CDNManager) GetCacheHeaders() map[string]string {
	if !c.enabled {
		return map[string]string{
			"Cache-Control": "public, max-age=3600",
		}
	}

	return map[string]string{
		"Cache-Control":     fmt.Sprintf("public, max-age=%d", c.cacheTTL),
		"CDN-Cache-Control": fmt.Sprintf("public, max-age=%d", c.cacheTTL),
	}
}

// IsEnabled returns whether CDN is enabled
func (c *CDNManager) IsEnabled() bool {
	return c.enabled
}

// GetBaseURL returns the CDN base URL
func (c *CDNManager) GetBaseURL() string {
	return c.baseURL
}
