package handlers

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AgentSREConfig holds configuration for the agent-sre proxy
type AgentSREConfig struct {
	ServiceURL string
	Timeout    time.Duration
}

// NewAgentSREHandler creates a new handler for agent-sre proxy
func NewAgentSREHandler(config AgentSREConfig) *AgentSREHandler {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.ServiceURL == "" {
		// Default to internal cluster service
		config.ServiceURL = "http://sre-agent-service.agent-sre.svc.cluster.local:8080"
	}

	return &AgentSREHandler{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// AgentSREHandler handles proxying requests to agent-sre service
type AgentSREHandler struct {
	config AgentSREConfig
	client *http.Client
}

// proxyRequest is a helper function to proxy requests to agent-sre
func (h *AgentSREHandler) proxyRequest(c *gin.Context, path string, method string) {
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
			"error":   "Agent-SRE service unavailable",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response from agent-sre",
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

// Chat handles POST /api/v1/agent-sre/chat
func (h *AgentSREHandler) Chat(c *gin.Context) {
	h.proxyRequest(c, "/chat", http.MethodPost)
}

// MCPChat handles POST /api/v1/agent-sre/mcp/chat
func (h *AgentSREHandler) MCPChat(c *gin.Context) {
	h.proxyRequest(c, "/mcp/chat", http.MethodPost)
}

// AnalyzeLogs handles POST /api/v1/agent-sre/analyze-logs
func (h *AgentSREHandler) AnalyzeLogs(c *gin.Context) {
	h.proxyRequest(c, "/analyze-logs", http.MethodPost)
}

// MCPAnalyzeLogs handles POST /api/v1/agent-sre/mcp/analyze-logs
func (h *AgentSREHandler) MCPAnalyzeLogs(c *gin.Context) {
	h.proxyRequest(c, "/mcp/analyze-logs", http.MethodPost)
}

// Health handles GET /api/v1/agent-sre/health
func (h *AgentSREHandler) Health(c *gin.Context) {
	h.proxyRequest(c, "/health", http.MethodGet)
}

// Ready handles GET /api/v1/agent-sre/ready
func (h *AgentSREHandler) Ready(c *gin.Context) {
	h.proxyRequest(c, "/ready", http.MethodGet)
}

// Status handles GET /api/v1/agent-sre/status
func (h *AgentSREHandler) Status(c *gin.Context) {
	h.proxyRequest(c, "/status", http.MethodGet)
}
