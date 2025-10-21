package config

import (
	"fmt"
	"knative-lambda-new/internal/constants"
	"os"
	"strconv"
	"time"
)

// SidecarConfig configuration for the sidecar
type SidecarConfig struct {
	// Build monitoring configuration
	KanikoNamespace     string
	KanikoPodName       string
	KanikoContainerName string
	PollInterval        time.Duration
	BuildTimeout        time.Duration

	// Event configuration
	JobName       string
	ImageURI      string
	ThirdPartyID  string
	ParserID      string
	ContentHash   string // New: content hash for unique image tagging
	CorrelationID string

	// Knative broker configuration
	BrokerURL string

	// Logging configuration
	LogLevel    string
	LogFormat   string
	ServiceName string

	// Security configuration
	TLSEnabled  bool
	TLSCertPath string
	TLSKeyPath  string
	RunAsUser   int64
	RunAsGroup  int64

	// Metrics configuration
	MetricsEnabled bool
	MetricsPort    int
	MetricsPath    string
}

// LoadSidecarConfig loads configuration from environment variables with comprehensive defaults
func LoadSidecarConfig() (*SidecarConfig, error) {
	config := &SidecarConfig{
		// Default values - updated to match examples
		KanikoContainerName: constants.ContainerNameKaniko,
		PollInterval:        5 * time.Second,
		BuildTimeout:        constants.BuildTimeoutDefault,
		LogLevel:            constants.LogLevelInfo,
		LogFormat:           "json",
		ServiceName:         constants.ContainerNameSidecar,
		TLSEnabled:          false,
		TLSCertPath:         "",
		TLSKeyPath:          "",
		RunAsUser:           int64(constants.K8sRunAsUserDefault),
		RunAsGroup:          int64(constants.K8sRunAsUserDefault),
		MetricsEnabled:      true,
		MetricsPort:         constants.K8sMetricsPortDefault,
		MetricsPath:         constants.MetricsPath,
	}

	// Required environment variables
	config.KanikoNamespace = getEnvOrDefault("KANIKO_NAMESPACE", "")
	if config.KanikoNamespace == "" {
		return nil, fmt.Errorf(constants.ErrKanikoNamespaceRequired)
	}

	config.KanikoPodName = getEnvOrDefault("KANIKO_POD_NAME", "")
	if config.KanikoPodName == "" {
		return nil, fmt.Errorf(constants.ErrKanikoPodNameRequired)
	}

	config.JobName = getEnvOrDefault("BUILD_JOB_NAME", "")
	if config.JobName == "" {
		return nil, fmt.Errorf(constants.ErrBuildJobNameRequired)
	}

	config.ImageURI = getEnvOrDefault("IMAGE_URI", "")
	if config.ImageURI == "" {
		return nil, fmt.Errorf(constants.ErrImageURIRequired)
	}

	config.ThirdPartyID = getEnvOrDefault("THIRD_PARTY_ID", "")
	if config.ThirdPartyID == "" {
		return nil, fmt.Errorf(constants.ErrThirdPartyIDRequired)
	}

	config.ParserID = getEnvOrDefault("PARSER_ID", "")
	if config.ParserID == "" {
		return nil, fmt.Errorf(constants.ErrParserIDRequired)
	}

	config.ContentHash = getEnvOrDefault("CONTENT_HASH", "") // New: load content hash from environment

	config.CorrelationID = getEnvOrDefault("CORRELATION_ID", "")
	if config.CorrelationID == "" {
		return nil, fmt.Errorf(constants.ErrCorrelationIDRequired)
	}

	config.BrokerURL = getEnvOrDefault("KNATIVE_BROKER_URL", "")
	if config.BrokerURL == "" {
		return nil, fmt.Errorf(constants.ErrKnativeBrokerURLRequired)
	}

	// Optional environment variables with defaults
	config.KanikoContainerName = getEnvOrDefault("KANIKO_CONTAINER_NAME", config.KanikoContainerName)
	config.LogLevel = getEnvOrDefault("LOG_LEVEL", config.LogLevel)
	config.LogFormat = getEnvOrDefault("LOG_FORMAT", config.LogFormat)
	config.ServiceName = getEnvOrDefault("SERVICE_NAME", config.ServiceName)
	config.TLSCertPath = getEnvOrDefault("TLS_CERT_PATH", config.TLSCertPath)
	config.TLSKeyPath = getEnvOrDefault("TLS_KEY_PATH", config.TLSKeyPath)
	config.MetricsPath = getEnvOrDefault("METRICS_PATH", config.MetricsPath)

	// Parse durations
	if pollIntervalStr := os.Getenv("MONITOR_INTERVAL"); pollIntervalStr != "" {
		pollInterval, err := time.ParseDuration(pollIntervalStr)
		if err != nil {
			return nil, fmt.Errorf(constants.ErrInvalidMonitorInterval, err)
		}
		config.PollInterval = pollInterval
	}

	if buildTimeoutStr := os.Getenv("BUILD_TIMEOUT"); buildTimeoutStr != "" {
		buildTimeout, err := time.ParseDuration(buildTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf(constants.ErrInvalidBuildTimeout, err)
		}
		config.BuildTimeout = buildTimeout
	}

	// Parse boolean values
	if tlsEnabledStr := os.Getenv("TLS_ENABLED"); tlsEnabledStr != "" {
		tlsEnabled, err := strconv.ParseBool(tlsEnabledStr)
		if err != nil {
			return nil, fmt.Errorf(constants.ErrInvalidTLSEnabled, err)
		}
		config.TLSEnabled = tlsEnabled
	}

	if metricsEnabledStr := os.Getenv("METRICS_ENABLED"); metricsEnabledStr != "" {
		metricsEnabled, err := strconv.ParseBool(metricsEnabledStr)
		if err != nil {
			return nil, fmt.Errorf(constants.ErrInvalidMetricsEnabled, err)
		}
		config.MetricsEnabled = metricsEnabled
	}

	// Parse integer values
	if runAsUserStr := os.Getenv("RUN_AS_USER"); runAsUserStr != "" {
		runAsUser, err := strconv.ParseInt(runAsUserStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf(constants.ErrInvalidRunAsUser, err)
		}
		config.RunAsUser = runAsUser
	}

	if runAsGroupStr := os.Getenv("RUN_AS_GROUP"); runAsGroupStr != "" {
		runAsGroup, err := strconv.ParseInt(runAsGroupStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf(constants.ErrInvalidRunAsGroup, err)
		}
		config.RunAsGroup = runAsGroup
	}

	if metricsPortStr := os.Getenv("METRICS_PORT"); metricsPortStr != "" {
		metricsPort, err := strconv.ParseInt(metricsPortStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf(constants.ErrInvalidMetricsPort, err)
		}
		config.MetricsPort = int(metricsPort)
	}

	return config, nil
}

// Validate validates the sidecar configuration
func (c *SidecarConfig) Validate() error {
	if c.KanikoNamespace == "" {
		return fmt.Errorf(constants.ErrKanikoNamespaceRequired)
	}

	if c.KanikoPodName == "" {
		return fmt.Errorf(constants.ErrKanikoPodNameRequired)
	}

	if c.JobName == "" {
		return fmt.Errorf(constants.ErrBuildJobNameRequired)
	}

	if c.ImageURI == "" {
		return fmt.Errorf(constants.ErrImageURIRequired)
	}

	if c.MetricsPort < constants.K8sMinPort || c.MetricsPort > constants.K8sMaxPort {
		return fmt.Errorf(constants.ErrInvalidPort, c.MetricsPort)
	}

	if c.TLSEnabled && (c.TLSCertPath == "" || c.TLSKeyPath == "") {
		return fmt.Errorf(constants.ErrTLSCertAndKeyRequired)
	}

	return nil
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
