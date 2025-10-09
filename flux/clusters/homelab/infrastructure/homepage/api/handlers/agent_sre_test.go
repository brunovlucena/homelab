package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

func TestNewAgentSREHandler(t *testing.T) {
	tests := []struct {
		name   string
		config AgentSREConfig
		want   AgentSREConfig
	}{
		{
			name: "with custom config",
			config: AgentSREConfig{
				ServiceURL: "http://custom-agent:8080",
				Timeout:    10 * time.Second,
			},
			want: AgentSREConfig{
				ServiceURL: "http://custom-agent:8080",
				Timeout:    10 * time.Second,
			},
		},
		{
			name:   "with default timeout",
			config: AgentSREConfig{ServiceURL: "http://test:8080"},
			want: AgentSREConfig{
				ServiceURL: "http://test:8080",
				Timeout:    30 * time.Second,
			},
		},
		{
			name:   "with default service URL",
			config: AgentSREConfig{},
			want: AgentSREConfig{
				ServiceURL: "http://sre-agent-service.agent-sre.svc.cluster.local:8080",
				Timeout:    30 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewAgentSREHandler(tt.config)
			assert.NotNil(t, handler)
			assert.Equal(t, tt.want.ServiceURL, handler.config.ServiceURL)
			assert.Equal(t, tt.want.Timeout, handler.config.Timeout)
			assert.NotNil(t, handler.client)
		})
	}
}

func TestAgentSREHandler_Chat(t *testing.T) {
	// Create a mock agent-sre server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		// Verify request body
		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		assert.NoError(t, err)
		assert.Equal(t, "How do I debug pods?", reqBody["message"])

		// Return mock response
		response := map[string]interface{}{
			"response":  "Use kubectl logs to debug pods",
			"timestamp": time.Now().Format(time.RFC3339),
			"model":     "bruno-sre:latest",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create handler with mock server URL
	handler := NewAgentSREHandler(AgentSREConfig{
		ServiceURL: mockServer.URL,
		Timeout:    5 * time.Second,
	})

	// Create test request
	reqBody := map[string]interface{}{
		"message":   "How do I debug pods?",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/agent-sre/chat", bytes.NewBuffer(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call handler
	handler.Chat(c)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Use kubectl logs to debug pods", response["response"])
}

func TestAgentSREHandler_MCPChat(t *testing.T) {
	// Create a mock agent-sre server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/mcp/chat", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		response := map[string]interface{}{
			"response":  "MCP response about monitoring",
			"timestamp": time.Now().Format(time.RFC3339),
			"sources":   []string{"MCP Server", "Knowledge Base"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	handler := NewAgentSREHandler(AgentSREConfig{
		ServiceURL: mockServer.URL,
		Timeout:    5 * time.Second,
	})

	reqBody := map[string]interface{}{
		"message": "Tell me about monitoring",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/agent-sre/mcp/chat", bytes.NewBuffer(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.MCPChat(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "MCP response about monitoring", response["response"])
	assert.NotNil(t, response["sources"])
}

func TestAgentSREHandler_AnalyzeLogs(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/analyze-logs", r.URL.Path)

		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "ERROR: Connection timeout", reqBody["logs"])

		response := map[string]interface{}{
			"analysis":        "Database connectivity issue detected",
			"severity":        "high",
			"recommendations": []string{"Check network", "Verify credentials"},
			"timestamp":       time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	handler := NewAgentSREHandler(AgentSREConfig{
		ServiceURL: mockServer.URL,
		Timeout:    5 * time.Second,
	})

	reqBody := map[string]interface{}{
		"logs":    "ERROR: Connection timeout",
		"context": "Production API",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/agent-sre/analyze-logs", bytes.NewBuffer(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.AnalyzeLogs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "high", response["severity"])
}

func TestAgentSREHandler_Health(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		response := map[string]string{"status": "healthy"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	handler := NewAgentSREHandler(AgentSREConfig{
		ServiceURL: mockServer.URL,
		Timeout:    5 * time.Second,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/agent-sre/health", nil)

	handler.Health(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestAgentSREHandler_Status(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/status", r.URL.Path)

		response := map[string]interface{}{
			"status":    "healthy",
			"service":   "sre-agent",
			"timestamp": time.Now().Format(time.RFC3339),
			"mcp_server": map[string]interface{}{
				"status": "healthy",
				"url":    "http://mcp-server:30120",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	handler := NewAgentSREHandler(AgentSREConfig{
		ServiceURL: mockServer.URL,
		Timeout:    5 * time.Second,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/agent-sre/status", nil)

	handler.Status(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.NotNil(t, response["mcp_server"])
}

func TestAgentSREHandler_ServiceUnavailable(t *testing.T) {
	// Use an invalid URL to simulate service unavailable
	handler := NewAgentSREHandler(AgentSREConfig{
		ServiceURL: "http://invalid-service:9999",
		Timeout:    1 * time.Second,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/agent-sre/health", nil)

	handler.Health(c)

	assert.Equal(t, http.StatusBadGateway, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Agent-SRE service unavailable", response["error"])
}

func TestAgentSREHandler_HeaderForwarding(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers are forwarded
		assert.NotEmpty(t, r.Header.Get("X-Forwarded-For"))
		assert.NotEmpty(t, r.Header.Get("X-Forwarded-Host"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer mockServer.Close()

	handler := NewAgentSREHandler(AgentSREConfig{
		ServiceURL: mockServer.URL,
		Timeout:    5 * time.Second,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/agent-sre/health", nil)
	c.Request.Header.Set("Authorization", "Bearer test-token")

	handler.Health(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
