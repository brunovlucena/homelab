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

// LLMService handles communication with Ollama
type LLMService struct {
	ollamaURL      string
	model          string
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

// OllamaRequest represents request format for Ollama Chat API
type OllamaRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaResponse represents response format from Ollama Chat API
type OllamaResponse struct {
	Message OllamaMessage `json:"message"`
	Done    bool          `json:"done"`
}

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NewLLMService creates a new LLM service
func NewLLMService(db *sql.DB) *LLMService {
	service := &LLMService{
		ollamaURL:      getEnv("OLLAMA_URL", "http://192.168.0.3:11434"),
		model:          getEnv("GEMMA_MODEL", "gemma3n:e4b"),
		contextBuilder: NewContextBuilder(db),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	log.Printf("🤖 LLM Service initialized")
	log.Printf("   📍 Ollama URL: %s", service.ollamaURL)
	log.Printf("   🎯 Model: %s", service.model)
	log.Printf("   ⏱️  Timeout: %v", service.httpClient.Timeout)

	// Test connection on startup
	go service.testConnectionOnStartup()

	return service
}

// testConnectionOnStartup tests the Ollama connection in background
func (llm *LLMService) testConnectionOnStartup() {
	log.Printf("🔍 Testing Ollama connection on startup...")

	// Wait a bit for the service to fully start
	time.Sleep(2 * time.Second)

	if err := llm.HealthCheck(); err != nil {
		log.Printf("❌ Ollama connection test failed: %v", err)
		log.Printf("💡 Troubleshooting tips:")
		log.Printf("   1. Check if Ollama is running: curl %s/api/tags", llm.ollamaURL)
		log.Printf("   2. Verify network connectivity to %s", llm.ollamaURL)
		log.Printf("   3. Check if model %s is available", llm.model)
		log.Printf("   4. Verify firewall settings")
	} else {
		log.Printf("✅ Ollama connection test successful")
	}
}

// ProcessChat handles a chat request and returns an AI response
func (llm *LLMService) ProcessChat(request ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()
	requestID := fmt.Sprintf("chat_%d", startTime.UnixNano())

	log.Printf("🚀 [%s] Starting chat processing", requestID)
	log.Printf("   📝 Message: %s", truncateString(request.Message, 100))
	log.Printf("   🎯 Model: %s", llm.model)
	log.Printf("   🌐 Ollama URL: %s", llm.ollamaURL)
	log.Printf("   🔧 Environment: OLLAMA_URL=%s", os.Getenv("OLLAMA_URL"))
	log.Printf("   🔧 Environment: GEMMA_MODEL=%s", os.Getenv("GEMMA_MODEL"))

	// Build context from PostgreSQL data
	log.Printf("🔧 [%s] Building context from database...", requestID)
	context, err := llm.contextBuilder.BuildContext(request.Message)
	if err != nil {
		log.Printf("❌ [%s] Context building failed: %v", requestID, err)
		log.Printf("   🔍 Database connection status: %v", llm.contextBuilder.db != nil)
		return nil, fmt.Errorf("failed to build context: %v", err)
	}
	log.Printf("✅ [%s] Context built successfully (%d chars)", requestID, len(context))
	log.Printf("   📄 Context preview: %s", truncateString(context, 200))

	// Generate response using Ollama
	log.Printf("🦙 [%s] Calling Ollama API...", requestID)
	log.Printf("   🔍 Testing Ollama connectivity first...")
	if err := llm.HealthCheck(); err != nil {
		log.Printf("❌ [%s] Ollama health check failed before API call: %v", requestID, err)
		log.Printf("   💡 This might indicate network connectivity issues")
	}

	response, err := llm.callOllama(context, requestID)

	if err != nil {
		log.Printf("❌ [%s] Ollama API call failed: %v", requestID, err)
		log.Printf("   🔍 Error type: %T", err)
		log.Printf("   🔍 Full error details: %+v", err)
		return nil, fmt.Errorf("LLM request failed: %v", err)
	}

	// Create response
	chatResponse := &ChatResponse{
		Response:  response,
		Model:     llm.model,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Sources:   []string{"PostgreSQL Database"},
	}

	duration := time.Since(startTime)
	log.Printf("✅ [%s] Chat processing completed in %v", requestID, duration)
	log.Printf("   📤 Response length: %d chars", len(response))
	log.Printf("   🎯 Model used: %s", llm.model)

	return chatResponse, nil
}

// callOllama sends request to Ollama API with enhanced logging
func (llm *LLMService) callOllama(prompt string, requestID string) (string, error) {
	log.Printf("🦙 [%s] Preparing Ollama request", requestID)
	log.Printf("   📍 URL: %s/api/chat", llm.ollamaURL)
	log.Printf("   🎯 Model: %s", llm.model)
	log.Printf("   📝 Prompt length: %d chars", len(prompt))
	log.Printf("   🔧 HTTP Client timeout: %v", llm.httpClient.Timeout)
	log.Printf("   🔧 HTTP Client transport: %T", llm.httpClient.Transport)

	requestBody := OllamaRequest{
		Model: llm.model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are a fact-based assistant. NEVER use greetings, introductions, or pleasantries. Answer questions immediately with facts only. Maximum 2 sentences. Start directly with the answer.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("❌ [%s] Failed to marshal request: %v", requestID, err)
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}
	log.Printf("📦 [%s] Request payload size: %d bytes", requestID, len(jsonData))

	// Log request details (without sensitive data)
	log.Printf("📤 [%s] Sending HTTP POST request", requestID)
	log.Printf("   🔗 URL: %s/api/chat", llm.ollamaURL)
	log.Printf("   📋 Headers: Content-Type=application/json")
	log.Printf("   ⏱️  Timeout: %v", llm.httpClient.Timeout)

	startTime := time.Now()
	resp, err := llm.httpClient.Post(
		fmt.Sprintf("%s/api/chat", llm.ollamaURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	requestDuration := time.Since(startTime)

	if err != nil {
		log.Printf("❌ [%s] HTTP request failed after %v: %v", requestID, requestDuration, err)
		log.Printf("💡 [%s] Connection troubleshooting:", requestID)
		log.Printf("   - Check if Ollama is running on %s", llm.ollamaURL)
		log.Printf("   - Verify network connectivity")
		log.Printf("   - Check firewall settings")
		log.Printf("   - Test with: curl -X POST %s/api/chat", llm.ollamaURL)
		log.Printf("   🔍 Error type: %T", err)
		log.Printf("   🔍 Network error details: %+v", err)
		log.Printf("   🔍 DNS resolution test: nslookup %s", strings.TrimPrefix(strings.TrimPrefix(llm.ollamaURL, "http://"), "https://"))
		return "", fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📥 [%s] Received response in %v", requestID, requestDuration)
	log.Printf("   📊 Status: %d %s", resp.StatusCode, resp.Status)
	log.Printf("   📋 Headers: %v", resp.Header)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ [%s] Ollama API error response:", requestID)
		log.Printf("   📊 Status: %d", resp.StatusCode)
		log.Printf("   📝 Body: %s", string(body))
		log.Printf("💡 [%s] Error troubleshooting:", requestID)
		log.Printf("   - Check if model '%s' is available", llm.model)
		log.Printf("   - Verify Ollama service status")
		log.Printf("   - Check Ollama logs for errors")
		return "", fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ [%s] Failed to read response body: %v", requestID, err)
		return "", fmt.Errorf("failed to read response body: %v", err)
	}
	log.Printf("📦 [%s] Response body size: %d bytes", requestID, len(body))

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		log.Printf("❌ [%s] Failed to unmarshal response: %v", requestID, err)
		log.Printf("   📝 Raw response: %s", string(body))
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	response := strings.TrimSpace(ollamaResp.Message.Content)
	response = strings.TrimSpace(response)

	log.Printf("✅ [%s] Ollama response processed successfully", requestID)
	log.Printf("   📝 Response length: %d chars", len(response))
	log.Printf("   🎯 Model: %s", llm.model)
	log.Printf("   ⏱️  Total time: %v", requestDuration)

	return response, nil
}

// HealthCheck checks if Ollama service is available with enhanced logging
func (llm *LLMService) HealthCheck() error {
	log.Printf("🏥 Starting Ollama health check...")
	log.Printf("   📍 URL: %s/api/tags", llm.ollamaURL)
	log.Printf("   ⏱️  Timeout: %v", llm.httpClient.Timeout)
	log.Printf("   🔧 Environment OLLAMA_URL: %s", os.Getenv("OLLAMA_URL"))
	log.Printf("   🔧 Service ollamaURL: %s", llm.ollamaURL)

	startTime := time.Now()
	resp, err := llm.httpClient.Get(fmt.Sprintf("%s/api/tags", llm.ollamaURL))
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("❌ Health check failed after %v: %v", duration, err)
		log.Printf("💡 Troubleshooting tips:")
		log.Printf("   - Check if Ollama is running: ollama serve")
		log.Printf("   - Verify URL is accessible: curl %s/api/tags", llm.ollamaURL)
		log.Printf("   - Check network connectivity")
		log.Printf("   - Verify firewall settings")
		log.Printf("🔍 Error type: %T", err)
		log.Printf("🔍 Network error details: %+v", err)
		log.Printf("🔍 DNS resolution test: nslookup %s", strings.TrimPrefix(strings.TrimPrefix(llm.ollamaURL, "http://"), "https://"))
		return fmt.Errorf("ollama health check failed: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("📥 Health check response received in %v", duration)
	log.Printf("   📊 Status: %d %s", resp.StatusCode, resp.Status)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Health check failed with status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("ollama health check failed with status: %d", resp.StatusCode)
	}

	// Try to parse the response to get model information
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("⚠️ Health check succeeded but failed to read response: %v", err)
		log.Printf("✅ Ollama is responding (status 200)")
		return nil
	}

	// Parse models list
	var modelsResponse struct {
		Models []struct {
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"models"`
	}

	if err := json.Unmarshal(body, &modelsResponse); err != nil {
		log.Printf("⚠️ Health check succeeded but failed to parse models: %v", err)
		log.Printf("✅ Ollama is responding (status 200)")
		return nil
	}

	log.Printf("✅ Ollama health check successful")
	log.Printf("   📋 Available models: %d", len(modelsResponse.Models))

	// Check if our model is available
	modelFound := false
	for _, model := range modelsResponse.Models {
		if model.Name == llm.model {
			modelFound = true
			log.Printf("   ✅ Required model '%s' found (%d bytes)", model.Name, model.Size)
			break
		}
	}

	if !modelFound {
		log.Printf("⚠️ Required model '%s' not found in available models", llm.model)
		log.Printf("   📋 Available models:")
		for _, model := range modelsResponse.Models {
			log.Printf("      - %s (%d bytes)", model.Name, model.Size)
		}
		log.Printf("💡 To install the model: ollama pull %s", llm.model)
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
