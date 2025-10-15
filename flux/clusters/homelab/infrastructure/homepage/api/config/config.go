package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	DatabaseURL   string
	RedisURL      string
	CORSOrigin    string
	Port          string
	AgentBrunoURL string // 🤖 Agent Bruno - Homepage chatbot and knowledge assistant URL
	MinIO         MinIOConfig
	Cloudflare    CloudflareConfig
}

// MinIOConfig holds MinIO configuration
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

// CloudflareConfig holds Cloudflare configuration
type CloudflareConfig struct {
	ZoneID   string
	APIToken string
	Domain   string
	Enabled  bool
	CacheTTL int
}

// Load loads configuration from environment variables
func Load() *Config {
	// Construct DATABASE_URL programmatically from individual components
	// Instead of using a hardcoded DATABASE_URL, we build it from parts
	dbHost := getEnvOrDefault("POSTGRES_HOST", "localhost")
	dbPort := getEnvOrDefault("POSTGRES_PORT", "5432")
	dbUser := getEnvOrDefault("POSTGRES_USER", "postgres")
	dbPassword := getEnvOrDefault("POSTGRES_PASSWORD", "")
	dbName := getEnvOrDefault("POSTGRES_DB", "bruno_site")

	databaseURL := ""
	if dbPassword != "" {
		databaseURL = "postgresql://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName
	}

	return &Config{
		DatabaseURL:   databaseURL,
		RedisURL:      os.Getenv("REDIS_URL"),
		CORSOrigin:    getEnvOrDefault("CORS_ORIGIN", "*"),
		Port:          getEnvOrDefault("PORT", "8080"),
		AgentBrunoURL: getEnvOrDefault("AGENT_BRUNO_URL", "http://agent-bruno-service.agent-bruno.svc.cluster.local:8080"),
		MinIO: MinIOConfig{
			Endpoint:  getEnvOrDefault("MINIO_ENDPOINT", "minio-service.minio.svc.cluster.local:9000"),
			AccessKey: getEnvOrDefault("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnvOrDefault("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:    getEnvOrDefault("MINIO_USE_SSL", "false") == "true",
			Bucket:    getEnvOrDefault("MINIO_BUCKET", "homepage-assets"),
		},
		Cloudflare: CloudflareConfig{
			ZoneID:   getEnvOrDefault("CLOUDFLARE_ZONE_ID", ""),
			APIToken: getEnvOrDefault("CLOUDFLARE_API_TOKEN", ""),
			Domain:   getEnvOrDefault("CLOUDFLARE_DOMAIN", "lucena.cloud"),
			Enabled:  getEnvOrDefault("CLOUDFLARE_ENABLED", "false") == "true",
			CacheTTL: parseIntOrDefault("CLOUDFLARE_CACHE_TTL", 86400), // 24 hours default
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
