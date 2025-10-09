package cdn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CloudflareCDN handles Cloudflare CDN configuration (Golden Rule #7: CDN)
type CloudflareCDN struct {
	zoneID     string
	apiToken   string
	domain     string
	enabled    bool
	cacheTTL   int
	httpClient *http.Client
}

// CloudflarePurgeRequest represents a cache purge request
type CloudflarePurgeRequest struct {
	PurgeEverything bool     `json:"purge_everything,omitempty"`
	Files           []string `json:"files,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	Hosts           []string `json:"hosts,omitempty"`
}

// CloudflarePurgeResponse represents the response from Cloudflare
type CloudflarePurgeResponse struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	Messages []interface{} `json:"messages"`
	Result   struct {
		ID string `json:"id"`
	} `json:"result"`
}

// NewCloudflareCDN creates a new Cloudflare CDN manager
func NewCloudflareCDN(zoneID, apiToken, domain string, enabled bool, cacheTTL int) *CloudflareCDN {
	return &CloudflareCDN{
		zoneID:   zoneID,
		apiToken: apiToken,
		domain:   domain,
		enabled:  enabled,
		cacheTTL: cacheTTL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAssetURL returns the Cloudflare CDN URL for a static asset
func (c *CloudflareCDN) GetAssetURL(assetPath string) string {
	if !c.enabled || c.domain == "" {
		return assetPath
	}

	// Ensure asset path starts with /
	if !strings.HasPrefix(assetPath, "/") {
		assetPath = "/" + assetPath
	}

	// Cloudflare automatically serves from the domain
	return fmt.Sprintf("https://%s%s", c.domain, assetPath)
}

// GetImageURL returns optimized image URL with Cloudflare
func (c *CloudflareCDN) GetImageURL(imagePath string, width, height int, format string) string {
	if !c.enabled {
		return imagePath
	}

	// Cloudflare Image Resizing (if enabled)
	if width > 0 && height > 0 && format != "" {
		// Format: https://domain.com/cdn-cgi/image/width=300,height=200,format=webp/path/to/image.jpg
		optimizedPath := fmt.Sprintf("/cdn-cgi/image/width=%d,height=%d,format=%s%s", width, height, format, imagePath)
		return c.GetAssetURL(optimizedPath)
	}

	return c.GetAssetURL(imagePath)
}

// GetCacheHeaders returns Cloudflare-optimized cache headers
func (c *CloudflareCDN) GetCacheHeaders() map[string]string {
	if !c.enabled {
		return map[string]string{
			"Cache-Control": "public, max-age=3600",
		}
	}

	return map[string]string{
		"Cache-Control":          fmt.Sprintf("public, max-age=%d", c.cacheTTL),
		"CF-Cache-Status":        "HIT", // This will be set by Cloudflare
		"CF-Ray":                 "",    // This will be set by Cloudflare
		"CF-IPCountry":           "",    // This will be set by Cloudflare
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}
}

// PurgeCache purges Cloudflare cache
func (c *CloudflareCDN) PurgeCache(request *CloudflarePurgeRequest) error {
	if !c.enabled || c.zoneID == "" || c.apiToken == "" {
		return fmt.Errorf("cloudflare CDN not properly configured")
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/purge_cache", c.zoneID)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal purge request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create purge request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute purge request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read purge response: %w", err)
	}

	var purgeResp CloudflarePurgeResponse
	if err := json.Unmarshal(body, &purgeResp); err != nil {
		return fmt.Errorf("failed to unmarshal purge response: %w", err)
	}

	if !purgeResp.Success {
		if len(purgeResp.Errors) > 0 {
			return fmt.Errorf("cloudflare purge failed: %s", purgeResp.Errors[0].Message)
		}
		return fmt.Errorf("cloudflare purge failed: unknown error")
	}

	return nil
}

// PurgeAllCache purges all cached content
func (c *CloudflareCDN) PurgeAllCache() error {
	return c.PurgeCache(&CloudflarePurgeRequest{
		PurgeEverything: true,
	})
}

// PurgeFiles purges specific files
func (c *CloudflareCDN) PurgeFiles(files []string) error {
	return c.PurgeCache(&CloudflarePurgeRequest{
		Files: files,
	})
}

// PurgeByTags purges cache by tags
func (c *CloudflareCDN) PurgeByTags(tags []string) error {
	return c.PurgeCache(&CloudflarePurgeRequest{
		Tags: tags,
	})
}

// IsEnabled returns whether Cloudflare CDN is enabled
func (c *CloudflareCDN) IsEnabled() bool {
	return c.enabled
}

// GetDomain returns the Cloudflare domain
func (c *CloudflareCDN) GetDomain() string {
	return c.domain
}

// GetZoneID returns the Cloudflare zone ID
func (c *CloudflareCDN) GetZoneID() string {
	return c.zoneID
}

// Health checks if Cloudflare API is accessible
func (c *CloudflareCDN) Health() error {
	if !c.enabled {
		return nil
	}

	if c.zoneID == "" || c.apiToken == "" {
		return fmt.Errorf("cloudflare not configured")
	}

	// Test API access
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s", c.zoneID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cloudflare API health check failed with status: %d", resp.StatusCode)
	}

	return nil
}
