package handlers

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AgentBrunoConfig holds configuration for the agent-bruno proxy
type AgentBrunoConfig struct {
	ServiceURL string
	Timeout    time.Duration
}

// NewAgentBrunoHandler creates a new handler for agent-bruno proxy
func NewAgentBrunoHandler(config AgentBrunoConfig) *AgentBrunoHandler {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.ServiceURL == "" {
		// Default to internal cluster service
		config.ServiceURL = "http://agent-bruno-service.agent-bruno.svc.cluster.local:8080"
	}

	return &AgentBrunoHandler{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// AgentBrunoHandler handles proxying requests to agent-bruno service
type AgentBrunoHandler struct {
	config AgentBrunoConfig
	client *http.Client
}

// CheckHealth implements DependencyChecker interface
// 🏥 Verifies that Agent Bruno service is reachable and healthy
func (h *AgentBrunoHandler) CheckHealth() error {
	req, err := http.NewRequest(http.MethodGet, h.config.ServiceURL+"/health", nil)
	if err != nil {
		return err
	}

	// Use a shorter timeout for health checks
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return http.ErrServerClosed
	}

	return nil
}

// proxyRequest is a helper function to proxy requests to agent-bruno
func (h *AgentBrunoHandler) proxyRequest(c *gin.Context, path string, method string) {
	// Build target URL
	targetURL := h.config.ServiceURL + path

	// Read request body
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
	}

	// Create new request
	req, err := http.NewRequest(method, targetURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create proxy request",
		})
		return
	}

	// Copy headers
	req.Header = c.Request.Header.Clone()
	req.Header.Set("X-Forwarded-For", c.ClientIP())
	req.Header.Set("X-Forwarded-Host", c.Request.Host)

	// Send request
	resp, err := h.client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "Agent-Bruno service unavailable",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response from agent-bruno",
		})
		return
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Return response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// Chat handles POST /api/v1/agent-bruno/chat
func (h *AgentBrunoHandler) Chat(c *gin.Context) {
	h.proxyRequest(c, "/chat", http.MethodPost)
}

// MCPChat handles POST /api/v1/agent-bruno/mcp/chat
func (h *AgentBrunoHandler) MCPChat(c *gin.Context) {
	h.proxyRequest(c, "/mcp/chat", http.MethodPost)
}

// GetMemory handles GET /api/v1/agent-bruno/memory/:ip
func (h *AgentBrunoHandler) GetMemory(c *gin.Context) {
	ip := c.Param("ip")
	h.proxyRequest(c, "/memory/"+ip, http.MethodGet)
}

// GetMemoryHistory handles GET /api/v1/agent-bruno/memory/:ip/history
func (h *AgentBrunoHandler) GetMemoryHistory(c *gin.Context) {
	ip := c.Param("ip")
	h.proxyRequest(c, "/memory/"+ip+"/history", http.MethodGet)
}

// ClearMemory handles DELETE /api/v1/agent-bruno/memory/:ip
func (h *AgentBrunoHandler) ClearMemory(c *gin.Context) {
	ip := c.Param("ip")
	h.proxyRequest(c, "/memory/"+ip, http.MethodDelete)
}

// GetKnowledgeSummary handles GET /api/v1/agent-bruno/knowledge/summary
func (h *AgentBrunoHandler) GetKnowledgeSummary(c *gin.Context) {
	h.proxyRequest(c, "/knowledge/summary", http.MethodGet)
}

// SearchKnowledge handles GET /api/v1/agent-bruno/knowledge/search
func (h *AgentBrunoHandler) SearchKnowledge(c *gin.Context) {
	h.proxyRequest(c, "/knowledge/search", http.MethodGet)
}

// GetStats handles GET /api/v1/agent-bruno/stats
func (h *AgentBrunoHandler) GetStats(c *gin.Context) {
	h.proxyRequest(c, "/stats", http.MethodGet)
}

// Health handles GET /api/v1/agent-bruno/health
func (h *AgentBrunoHandler) Health(c *gin.Context) {
	h.proxyRequest(c, "/health", http.MethodGet)
}

// Ready handles GET /api/v1/agent-bruno/ready
func (h *AgentBrunoHandler) Ready(c *gin.Context) {
	h.proxyRequest(c, "/ready", http.MethodGet)
}
