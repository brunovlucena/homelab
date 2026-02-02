package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// LLMService handles communication with Agent-Bruno (the AI chatbot lambda)
type LLMService struct {
	agentBrunoURL  string
	contextBuilder *ContextBuilder
	httpClient     *http.Client
}

// ChatRequest represents an incoming chat request
type ChatRequest struct {
	Message string `json:"message" binding:"required"`
	Context string `json:"context,omitempty"`
}

// ChatResponse represents the response from the chatbot
type ChatResponse struct {
	Response  string   `json:"response"`
	Sources   []string `json:"sources,omitempty"`
	Model     string   `json:"model"`
	Timestamp string   `json:"timestamp"`
}

// AgentBrunoRequest represents request format for Agent-Bruno API
type AgentBrunoRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
}

// AgentBrunoResponse represents response format from Agent-Bruno API
type AgentBrunoResponse struct {
	Response       string  `json:"response"`
	ConversationID string  `json:"conversation_id"`
	TokensUsed     int     `json:"tokens_used"`
	Model          string  `json:"model"`
	DurationMs     float64 `json:"duration_ms"`
}

// AgentBrunoHealthResponse represents the health check response
type AgentBrunoHealthResponse struct {
	Status          string `json:"status"`
	OllamaAvailable bool   `json:"ollama_available"`
}

// NewLLMService creates a new LLM service
func NewLLMService(db *sql.DB) *LLMService {
	service := &LLMService{
		// Default to in-cluster agent-bruno service URL
		agentBrunoURL:  getEnv("AGENT_BRUNO_URL", "http://agent-bruno.agent-bruno.svc.cluster.local"),
		contextBuilder: NewContextBuilder(db),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	log.Printf("🤖 LLM Service initialized (using Agent-Bruno)")
	log.Printf("   📍 Agent-Bruno URL: %s", service.agentBrunoURL)
	log.Printf("   ⏱️  Timeout: %v", service.httpClient.Timeout)

	// Test connection on startup
	go service.testConnectionOnStartup()

	return service
}

// testConnectionOnStartup tests the Agent-Bruno connection in background
func (llm *LLMService) testConnectionOnStartup() {
	log.Printf("🔍 Testing Agent-Bruno connection on startup...")

	// Wait a bit for the service to fully start
	time.Sleep(2 * time.Second)

	if err := llm.HealthCheck(); err != nil {
		log.Printf("❌ Agent-Bruno connection test failed: %v", err)
		log.Printf("💡 Troubleshooting tips:")
		log.Printf("   1. Check if Agent-Bruno is running: kubectl get pods -n agent-bruno")
		log.Printf("   2. Verify service is accessible: curl %s/health", llm.agentBrunoURL)
		log.Printf("   3. Check Agent-Bruno logs for errors")
	} else {
		log.Printf("✅ Agent-Bruno connection test successful")
	}
}

// ProcessChat handles a chat request and returns an AI response via Agent-Bruno
func (llm *LLMService) ProcessChat(request ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()
	requestID := fmt.Sprintf("chat_%d", startTime.UnixNano())

	log.Printf("🚀 [%s] Starting chat processing via Agent-Bruno", requestID)
	log.Printf("   📝 Message: %s", truncateString(request.Message, 100))
	log.Printf("   🌐 Agent-Bruno URL: %s", llm.agentBrunoURL)

	// Build context from PostgreSQL data
	log.Printf("🔧 [%s] Building context from database...", requestID)
	context, err := llm.contextBuilder.BuildContext(request.Message)
	if err != nil {
		log.Printf("❌ [%s] Context building failed: %v", requestID, err)
		return nil, fmt.Errorf("failed to build context: %v", err)
	}
	log.Printf("✅ [%s] Context built successfully (%d chars)", requestID, len(context))

	// Combine user message with context for Agent-Bruno
	enrichedMessage := fmt.Sprintf("%s\n\nContext about Bruno:\n%s", request.Message, context)

	// Call Agent-Bruno
	log.Printf("🤖 [%s] Calling Agent-Bruno...", requestID)
	response, model, err := llm.callAgentBruno(enrichedMessage, requestID)

	if err != nil {
		log.Printf("❌ [%s] Agent-Bruno API call failed: %v", requestID, err)
		return nil, fmt.Errorf("Agent-Bruno request failed: %v", err)
	}

	// Create response
	chatResponse := &ChatResponse{
		Response:  response,
		Model:     model,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Sources:   []string{"Agent-Bruno", "PostgreSQL Database"},
	}

	duration := time.Since(startTime)
	log.Printf("✅ [%s] Chat processing completed in %v", requestID, duration)
	log.Printf("   📤 Response length: %d chars", len(response))
	log.Printf("   🎯 Model used: %s", model)

	return chatResponse, nil
}

// callAgentBruno sends request to Agent-Bruno API
func (llm *LLMService) callAgentBruno(message string, requestID string) (string, string, error) {
	log.Printf("🤖 [%s] Preparing Agent-Bruno request", requestID)
	log.Printf("   📍 URL: %s/chat", llm.agentBrunoURL)
	log.Printf("   📝 Message length: %d chars", len(message))

	requestBody := AgentBrunoRequest{
		Message: message,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("❌ [%s] Failed to marshal request: %v", requestID, err)
		return "", "", fmt.Errorf("failed to marshal request: %v", err)
	}
	log.Printf("📦 [%s] Request payload size: %d bytes", requestID, len(jsonData))

	startTime := time.Now()
	resp, err := llm.httpClient.Post(
		fmt.Sprintf("%s/chat", llm.agentBrunoURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	requestDuration := time.Since(startTime)

	if err != nil {
		log.Printf("❌ [%s] HTTP request failed after %v: %v", requestID, requestDuration, err)
		log.Printf("💡 [%s] Connection troubleshooting:", requestID)
		log.Printf("   - Check if Agent-Bruno is running: kubectl get pods -n agent-bruno")
		log.Printf("   - Check service: kubectl get svc -n agent-bruno")
		log.Printf("   - Test with: curl -X POST %s/chat", llm.agentBrunoURL)
		return "", "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	log.Printf("📥 [%s] Received response in %v", requestID, requestDuration)
	log.Printf("   📊 Status: %d %s", resp.StatusCode, resp.Status)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ [%s] Agent-Bruno API error response:", requestID)
		log.Printf("   📊 Status: %d", resp.StatusCode)
		log.Printf("   📝 Body: %s", string(body))
		return "", "", fmt.Errorf("Agent-Bruno API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ [%s] Failed to read response body: %v", requestID, err)
		return "", "", fmt.Errorf("failed to read response body: %v", err)
	}
	log.Printf("📦 [%s] Response body size: %d bytes", requestID, len(body))

	var agentResponse AgentBrunoResponse
	if err := json.Unmarshal(body, &agentResponse); err != nil {
		log.Printf("❌ [%s] Failed to unmarshal response: %v", requestID, err)
		log.Printf("   📝 Raw response: %s", string(body))
		return "", "", fmt.Errorf("failed to decode response: %v", err)
	}

	response := strings.TrimSpace(agentResponse.Response)

	log.Printf("✅ [%s] Agent-Bruno response processed successfully", requestID)
	log.Printf("   📝 Response length: %d chars", len(response))
	log.Printf("   🎯 Model: %s", agentResponse.Model)
	log.Printf("   🔢 Tokens used: %d", agentResponse.TokensUsed)
	log.Printf("   ⏱️  Agent processing time: %.2fms", agentResponse.DurationMs)

	return response, agentResponse.Model, nil
}

// HealthCheck checks if Agent-Bruno service is available
func (llm *LLMService) HealthCheck() error {
	log.Printf("🏥 Starting Agent-Bruno health check...")
	log.Printf("   📍 URL: %s/health", llm.agentBrunoURL)

	startTime := time.Now()
	resp, err := llm.httpClient.Get(fmt.Sprintf("%s/health", llm.agentBrunoURL))
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("❌ Health check failed after %v: %v", duration, err)
		log.Printf("💡 Troubleshooting tips:")
		log.Printf("   - Check if Agent-Bruno is running: kubectl get pods -n agent-bruno")
		log.Printf("   - Verify URL is accessible: curl %s/health", llm.agentBrunoURL)
		log.Printf("   - Check network connectivity between namespaces")
		return fmt.Errorf("Agent-Bruno health check failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	log.Printf("📥 Health check response received in %v", duration)
	log.Printf("   📊 Status: %d %s", resp.StatusCode, resp.Status)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Health check failed with status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("Agent-Bruno health check failed with status: %d", resp.StatusCode)
	}

	// Parse health response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("⚠️ Health check succeeded but failed to read response: %v", err)
		return nil
	}

	var healthResponse AgentBrunoHealthResponse
	if err := json.Unmarshal(body, &healthResponse); err != nil {
		log.Printf("⚠️ Health check succeeded but failed to parse response: %v", err)
		return nil
	}

	log.Printf("✅ Agent-Bruno health check successful")
	log.Printf("   📊 Status: %s", healthResponse.Status)
	log.Printf("   🦙 Ollama available: %v", healthResponse.OllamaAvailable)

	if !healthResponse.OllamaAvailable {
		log.Printf("⚠️ Agent-Bruno reports Ollama is not available")
	}

	return nil
}

// Helper function to get environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function to truncate strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
