package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/brunovlucena/homelab/homepage-api/storage"

	"github.com/gin-gonic/gin"
)

// ProxyAsset proxies assets from MinIO to the client
// This allows internal MinIO access without exposing it to the internet
func ProxyAsset(minioClient *storage.MinIOClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the asset path from URL (remove leading slash)
		path := strings.TrimPrefix(c.Param("path"), "/")

		if path == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "asset path is required"})
			return
		}

		log.Printf("📦 Proxying asset: %s", path)

		// Get object from MinIO
		reader, size, contentType, err := minioClient.GetObject(c.Request.Context(), path)
		if err != nil {
			log.Printf("❌ Failed to get asset %s: %v", path, err)
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("asset not found: %s", path)})
			return
		}
		defer reader.Close()

		// Set response headers
		c.Header("Content-Type", contentType)
		c.Header("Content-Length", fmt.Sprintf("%d", size))
		c.Header("Cache-Control", "public, max-age=31536000") // Cache for 1 year
		c.Header("ETag", fmt.Sprintf("\"%s\"", path))

		// Stream the content
		if _, err := io.Copy(c.Writer, reader); err != nil {
			log.Printf("❌ Failed to stream asset %s: %v", path, err)
			return
		}

		log.Printf("✅ Successfully proxied asset: %s (%d bytes, %s)", path, size, contentType)
	}
}
