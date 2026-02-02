package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/brunolucena/homelab/vaultwarden/internal/auth"
	"github.com/brunolucena/homelab/vaultwarden/internal/vault"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	vaultClient *vault.Client
	authService *auth.Service
}

func NewHandlers(vaultClient *vault.Client, authService *auth.Service) *Handlers {
	return &Handlers{
		vaultClient: vaultClient,
		authService: authService,
	}
}

// Login handles user authentication
func (h *Handlers) Login(c *gin.Context) {
	var req struct {
		GrantType string `form:"grant_type" json:"grant_type"`
		Username  string `form:"username" json:"username"`
		Password  string `form:"password" json:"password"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.GrantType != "password" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported grant_type"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user auth data from Vault
	authData, err := h.vaultClient.GetUserAuth(ctx, req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	hashedPassword, ok := authData["password_hash"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user data"})
		return
	}

	// Verify password
	if !h.authService.VerifyPassword(hashedPassword, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	userID, ok := authData["user_id"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user data"})
		return
	}

	// Generate JWT token
	token, err := h.authService.GenerateToken(userID, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   86400, // 24 hours
	})
}

// ListCiphers returns all password entries for the authenticated user
func (h *Handlers) ListCiphers(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ciphers, err := h.vaultClient.ListCiphers(ctx, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Data":   ciphers,
		"Object": "list",
	})
}

// CreateCipher creates a new password entry
func (h *Handlers) CreateCipher(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var cipherData map[string]interface{}
	if err := c.ShouldBindJSON(&cipherData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cipherID, ok := cipherData["Id"].(string)
	if !ok || cipherID == "" {
		cipherID = generateID()
		cipherData["Id"] = cipherID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.vaultClient.StoreCipher(ctx, userID.(string), cipherID, cipherData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cipherData)
}

// GetCipher retrieves a specific password entry
func (h *Handlers) GetCipher(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	cipherID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cipher, err := h.vaultClient.GetCipher(ctx, userID.(string), cipherID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cipher not found"})
		return
	}

	c.JSON(http.StatusOK, cipher)
}

// UpdateCipher updates an existing password entry
func (h *Handlers) UpdateCipher(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	cipherID := c.Param("id")

	var cipherData map[string]interface{}
	if err := c.ShouldBindJSON(&cipherData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cipherData["Id"] = cipherID

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.vaultClient.StoreCipher(ctx, userID.(string), cipherID, cipherData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cipherData)
}

// DeleteCipher deletes a password entry
func (h *Handlers) DeleteCipher(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	cipherID := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.vaultClient.DeleteCipher(ctx, userID.(string), cipherID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetProfile returns user profile information
func (h *Handlers) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := h.vaultClient.GetUser(ctx, userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile updates user profile
func (h *Handlers) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var userData map[string]interface{}
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.vaultClient.StoreUser(ctx, userID.(string), userData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userData)
}

// generateID generates a simple ID (in production, use UUID)
func generateID() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}
