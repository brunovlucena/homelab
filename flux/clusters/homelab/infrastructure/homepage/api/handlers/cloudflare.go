package handlers

import (
	"net/http"
	"strconv"

	"bruno-site/cdn"

	"github.com/gin-gonic/gin"
)

// CloudflareHandler handles Cloudflare CDN operations
type CloudflareHandler struct {
	cdn *cdn.CloudflareCDN
}

// NewCloudflareHandler creates a new Cloudflare handler
func NewCloudflareHandler(cloudflareCDN *cdn.CloudflareCDN) *CloudflareHandler {
	return &CloudflareHandler{
		cdn: cloudflareCDN,
	}
}

// PurgeCacheRequest represents a cache purge request
type PurgeCacheRequest struct {
	Files []string `json:"files,omitempty"`
	Tags  []string `json:"tags,omitempty"`
	All   bool     `json:"all,omitempty"`
}

// PurgeCacheResponse represents the response from cache purge
type PurgeCacheResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Files   int    `json:"files_purged,omitempty"`
}

// PurgeCache purges Cloudflare cache
func (h *CloudflareHandler) PurgeCache(c *gin.Context) {
	if !h.cdn.IsEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"message": "Cloudflare CDN is not enabled",
		})
		return
	}

	var req PurgeCacheRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request format",
		})
		return
	}

	var err error
	var filesPurged int

	if req.All {
		// Purge all cache
		err = h.cdn.PurgeAllCache()
		filesPurged = -1 // -1 indicates all files
	} else if len(req.Files) > 0 {
		// Purge specific files
		err = h.cdn.PurgeFiles(req.Files)
		filesPurged = len(req.Files)
	} else if len(req.Tags) > 0 {
		// Purge by tags
		err = h.cdn.PurgeByTags(req.Tags)
		filesPurged = len(req.Tags)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "No purge target specified",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to purge cache: " + err.Error(),
		})
		return
	}

	response := PurgeCacheResponse{
		Success: true,
		Message: "Cache purged successfully",
		Files:   filesPurged,
	}

	c.JSON(http.StatusOK, response)
}

// GetCDNStatus returns Cloudflare CDN status
func (h *CloudflareHandler) GetCDNStatus(c *gin.Context) {
	if !h.cdn.IsEnabled() {
		c.JSON(http.StatusOK, gin.H{
			"enabled": false,
			"message": "Cloudflare CDN is disabled",
		})
		return
	}

	// Test health
	err := h.cdn.Health()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"enabled": true,
			"healthy": false,
			"message": "Cloudflare CDN is enabled but unhealthy: " + err.Error(),
			"domain":  h.cdn.GetDomain(),
			"zone_id": h.cdn.GetZoneID(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled": true,
		"healthy": true,
		"message": "Cloudflare CDN is healthy",
		"domain":  h.cdn.GetDomain(),
		"zone_id": h.cdn.GetZoneID(),
	})
}

// GetAssetURL returns CDN URL for an asset
func (h *CloudflareHandler) GetAssetURL(c *gin.Context) {
	assetPath := c.Query("path")
	if assetPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Asset path is required",
		})
		return
	}

	// Optional query parameters for image optimization
	width := c.Query("width")
	height := c.Query("height")
	format := c.Query("format")

	var cdnURL string
	if width != "" && height != "" && format != "" {
		// Parse dimensions
		widthInt := 0
		heightInt := 0
		if w, err := strconv.Atoi(width); err == nil {
			widthInt = w
		}
		if h, err := strconv.Atoi(height); err == nil {
			heightInt = h
		}
		cdnURL = h.cdn.GetImageURL(assetPath, widthInt, heightInt, format)
	} else {
		cdnURL = h.cdn.GetAssetURL(assetPath)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"url":     cdnURL,
		"path":    assetPath,
	})
}

// GetCacheHeaders returns recommended cache headers
func (h *CloudflareHandler) GetCacheHeaders(c *gin.Context) {
	headers := h.cdn.GetCacheHeaders()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"headers": headers,
	})
}
