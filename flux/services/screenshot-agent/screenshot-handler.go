package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// ScreenshotHandler handles screenshot uploads from browser extension
func ScreenshotHandler(c *gin.Context) {
	// Receber arquivo
	file, err := c.FormFile("screenshot")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No screenshot file provided", "details": err.Error()})
		return
	}

	// Receber metadados
	url := c.PostForm("url")
	title := c.PostForm("title")
	timestamp := c.PostForm("timestamp")

	log.Printf("📸 Screenshot recebido: %s (title: %s, url: %s)", file.Filename, title, url)

	// Opção 1: Salvar localmente (para desenvolvimento)
	saveDir := os.Getenv("SCREENSHOT_SAVE_DIR")
	if saveDir == "" {
		saveDir = "./screenshots"
	}

	// Criar diretório se não existir
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		log.Printf("⚠️  Erro ao criar diretório: %v", err)
	} else {
		// Salvar arquivo
		filename := filepath.Join(saveDir, file.Filename)
		if err := c.SaveUploadedFile(file, filename); err != nil {
			log.Printf("⚠️  Erro ao salvar arquivo: %v", err)
		} else {
			log.Printf("✅ Screenshot salvo: %s", filename)
		}
	}

	// Opção 2: Enviar para MinIO (se configurado)
	minioEnabled := os.Getenv("MINIO_ENABLED") == "true"
	if minioEnabled {
		// TODO: Implementar upload para MinIO
		log.Printf("📤 Upload para MinIO (não implementado ainda)")
	}

	// Opção 3: Enviar para agente de análise (se configurado)
	agentURL := os.Getenv("SCREENSHOT_AGENT_URL")
	if agentURL != "" {
		// TODO: Implementar chamada para agente de análise
		log.Printf("🤖 Enviar para agente: %s (não implementado ainda)", agentURL)
	}

	// Resposta de sucesso
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"url":       url,
		"title":     title,
		"timestamp": timestamp,
		"filename":  file.Filename,
		"size":      file.Size,
		"message":   "Screenshot recebido com sucesso",
		"received_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// Exemplo de como adicionar ao mobile-api/main.go:
//
// No grupo "/api/v1", adicione:
//   api.POST("/screenshots", ScreenshotHandler)
//
// Ou use este arquivo como referência para criar um handler próprio.
