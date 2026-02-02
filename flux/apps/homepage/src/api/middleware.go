package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// =============================================================================
// 🔧 MIDDLEWARE FUNCTIONS
// =============================================================================

// requestIDMiddleware adds a unique request ID to each request for tracing
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// requestLogger logs all incoming requests with request ID
func requestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		requestID, _ := param.Keys["request_id"].(string)
		return fmt.Sprintf("[%s] [%s] %s %s %d %s %s\n",
			param.TimeStamp.Format(time.RFC3339),
			requestID,
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
		)
	})
}

// errorHandler handles panics and errors
func errorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID, _ := c.Get("request_id")
		if err, ok := recovered.(string); ok {
			log.Printf("[%v] Panic recovered: %s", requestID, err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Internal server error",
			"request_id": requestID,
		})
		c.Abort()
	})
}

// =============================================================================
// 🚦 RATE LIMITING
// =============================================================================

// RateLimiter implements a simple in-memory rate limiter using token bucket algorithm
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

type visitor struct {
	tokens    int
	lastSeen  time.Time
	resetTime time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
	// Start cleanup goroutine
	go rl.cleanup()
	return rl
}

// cleanup removes old visitors periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	now := time.Now()

	if !exists {
		rl.visitors[ip] = &visitor{
			tokens:    rl.rate - 1,
			lastSeen:  now,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	// Reset tokens if window has passed
	if now.After(v.resetTime) {
		v.tokens = rl.rate - 1
		v.resetTime = now.Add(rl.window)
		v.lastSeen = now
		return true
	}

	// Check if tokens available
	if v.tokens > 0 {
		v.tokens--
		v.lastSeen = now
		return true
	}

	v.lastSeen = now
	return false
}

// RemainingTokens returns remaining tokens for an IP
func (rl *RateLimiter) RemainingTokens(ip string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if v, exists := rl.visitors[ip]; exists {
		return v.tokens
	}
	return rl.rate
}

// Global rate limiters
var (
	// General API rate limiter: 100 requests per minute
	apiRateLimiter = NewRateLimiter(100, time.Minute)
	// Chat rate limiter: 20 requests per minute (more expensive operation)
	chatRateLimiter = NewRateLimiter(20, time.Minute)
)

// rateLimitMiddleware applies rate limiting to requests
func rateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.Allow(ip) {
			requestID, _ := c.Get("request_id")
			log.Printf("[%v] Rate limit exceeded for IP: %s", requestID, ip)
			c.Header("Retry-After", "60")
			c.Header("X-RateLimit-Remaining", "0")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":      "Rate limit exceeded. Please try again later.",
				"request_id": requestID,
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limiter.RemainingTokens(ip)))
		c.Next()
	}
}

// =============================================================================
// 🔒 SECURITY HEADERS
// =============================================================================

// securityHeadersMiddleware adds security headers to responses
func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Next()
	}
}
