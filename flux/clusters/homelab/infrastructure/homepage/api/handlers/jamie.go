package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// JamieConfig holds configuration for the jamie service proxy
type JamieConfig struct {
	ServiceURL string
	Timeout    time.Duration
}

// NewJamieHandler creates a new handler for jamie proxy
func NewJamieHandler(config JamieConfig) *JamieHandler {
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second // 🧠 Longer timeout for AI responses
	}
	if config.ServiceURL == "" {
		// Default to Jamie Slack Bot service (REST API endpoint)
		config.ServiceURL = "http://jamie-slack-bot-service.jamie.svc.cluster.local:8080"
	}

	return &JamieHandler{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// JamieHandler handles proxying requests to jamie service
type JamieHandler struct {
	config JamieConfig
	client *http.Client
}

// CheckHealth implements DependencyChecker interface
// 🏥 Verifies that Jamie service is reachable and healthy
func (h *JamieHandler) CheckHealth() error {
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

// proxyRequest is a helper function to proxy requests to jamie
func (h *JamieHandler) proxyRequest(c *gin.Context, path string, method string) {
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
			"error": "Failed to create proxy request to Jamie",
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
			"error":   "Jamie service unavailable",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read response from Jamie",
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

// Chat handles POST /api/v1/jamie/chat
// 💬 Main chatbot endpoint for Homepage - talks to Jamie AI
func (h *JamieHandler) Chat(c *gin.Context) {
	h.proxyRequest(c, "/api/chat", http.MethodPost)
}

// CheckGoldenSignals handles POST /api/v1/jamie/golden-signals
// 📊 Check service health and golden signals via Jamie
func (h *JamieHandler) CheckGoldenSignals(c *gin.Context) {
	h.proxyRequest(c, "/api/golden-signals", http.MethodPost)
}

// QueryPrometheus handles POST /api/v1/jamie/prometheus/query
// 📈 Execute PromQL queries via Jamie
func (h *JamieHandler) QueryPrometheus(c *gin.Context) {
	h.proxyRequest(c, "/api/prometheus/query", http.MethodPost)
}

// GetPodLogs handles POST /api/v1/jamie/pod-logs
// 📝 Get Kubernetes pod logs via Jamie
func (h *JamieHandler) GetPodLogs(c *gin.Context) {
	h.proxyRequest(c, "/api/pod-logs", http.MethodPost)
}

// AnalyzeLogs handles POST /api/v1/jamie/analyze-logs
// 🔍 Analyze logs with AI via Jamie
func (h *JamieHandler) AnalyzeLogs(c *gin.Context) {
	h.proxyRequest(c, "/api/analyze-logs", http.MethodPost)
}

// Health handles GET /api/v1/jamie/health
// 🏥 Check Jamie service health
func (h *JamieHandler) Health(c *gin.Context) {
	h.proxyRequest(c, "/health", http.MethodGet)
}

// Ready handles GET /api/v1/jamie/ready
// ✅ Check Jamie service readiness
func (h *JamieHandler) Ready(c *gin.Context) {
	h.proxyRequest(c, "/ready", http.MethodGet)
}

// ChatRequest represents a chat message request
type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}

// ChatResponse represents a chat message response
type ChatResponse struct {
	Response  string `json:"response"`
	Timestamp string `json:"timestamp"`
}

// DirectChat handles direct chat without proxying - useful for custom logic
// 💬 Alternative chat endpoint with custom handling
func (h *JamieHandler) DirectChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: message is required",
		})
		return
	}

	// Forward to Jamie
	payload, _ := json.Marshal(req)
	httpReq, err := http.NewRequest(
		http.MethodPost,
		h.config.ServiceURL+"/api/chat",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create request to Jamie",
		})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "Jamie service unavailable",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to decode Jamie response",
		})
		return
	}

	c.JSON(http.StatusOK, chatResp)
}
