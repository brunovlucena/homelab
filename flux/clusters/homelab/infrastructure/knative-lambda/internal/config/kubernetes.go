// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	☸️ KUBERNETES CONFIGURATION - Kubernetes client and resource configuration
//
//	🎯 Purpose: Kubernetes client settings, namespace configuration, RBAC settings
//	💡 Features: In-cluster config, kubeconfig path, service account settings
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
)

// ☸️ KubernetesConfig - "Kubernetes client and resource configuration"
type KubernetesConfig struct {
	// Cluster Configuration
	Namespace      string `envconfig:"KUBERNETES_NAMESPACE" default:"knative-lambda" validate:"required,min=1,max=63"`
	InCluster      bool   `envconfig:"IN_CLUSTER" default:"true"`
	KubeConfig     string `envconfig:"KUBECONFIG" default:""`
	ServiceAccount string `envconfig:"SERVICE_ACCOUNT" default:"knative-lambda-builder" validate:"required,min=1,max=63"`

	// Resource Configuration
	RunAsUser int64 `envconfig:"RUN_AS_USER" default:"65534" validate:"required,min=1"`

	// Job Configuration
	JobTTLSeconds int32 `envconfig:"JOB_TTL_SECONDS" default:"3600" validate:"required,min=60"`

	// Job Management Configuration
	JobDeletionWaitTimeout   time.Duration `envconfig:"JOB_DELETION_WAIT_TIMEOUT" default:"30s" validate:"required,min=1s"`
	JobDeletionCheckInterval time.Duration `envconfig:"JOB_DELETION_CHECK_INTERVAL" default:"500ms" validate:"required,min=100ms"`
}

// 🔧 NewKubernetesConfig - "Create Kubernetes configuration with defaults"
func NewKubernetesConfig() *KubernetesConfig {
	// Get environment from environment variable or default to "dev"
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "dev"
	}

	// Create environment-aware default namespace
	defaultNamespace := fmt.Sprintf("knative-lambda-%s", environment)

	return &KubernetesConfig{
		Namespace:                defaultNamespace,
		InCluster:                true,
		KubeConfig:               "",
		ServiceAccount:           constants.ServiceAccountDefault,
		RunAsUser:                constants.K8sRunAsUserDefault,
		JobTTLSeconds:            constants.K8sJobTTLSecondsDefault,
		JobDeletionWaitTimeout:   constants.JobDeletionWaitTimeoutDefault,
		JobDeletionCheckInterval: constants.JobDeletionCheckIntervalDefault,
	}
}

// 🔧 Validate - "Validate Kubernetes configuration"
func (c *KubernetesConfig) Validate() error {
	if !constants.IsValidNamespace(c.Namespace) {
		return errors.NewValidationError("namespace", c.Namespace, constants.ErrK8sNamespaceValid)
	}

	if !constants.IsValidName(c.ServiceAccount) {
		return errors.NewValidationError("service_account", c.ServiceAccount, constants.ErrK8sServiceAccountValid)
	}

	if !constants.IsValidUserID(c.RunAsUser) {
		return errors.NewValidationError("run_as_user", c.RunAsUser, constants.ErrK8sRunAsUserValid)
	}

	if c.JobTTLSeconds < 60 {
		return errors.NewValidationError("job_ttl_seconds", c.JobTTLSeconds, constants.ErrK8sJobTTLMin60Seconds)
	}

	if c.JobDeletionWaitTimeout < time.Second {
		return errors.NewValidationError("job_deletion_wait_timeout", c.JobDeletionWaitTimeout, constants.ErrK8sJobDeletionWaitMin1Second)
	}

	if c.JobDeletionCheckInterval < 100*time.Millisecond {
		return errors.NewValidationError("job_deletion_check_interval", c.JobDeletionCheckInterval, constants.ErrK8sJobDeletionCheckMin100ms)
	}

	return nil
}

// 🔧 GetKubernetesConfig - "Get Kubernetes client configuration"
func (c *KubernetesConfig) GetKubernetesConfig() (*rest.Config, error) {
	if c.InCluster {
		return rest.InClusterConfig()
	}

	kubeconfig := c.KubeConfig
	if kubeconfig == "" {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

// 🔧 GetNamespace - "Get target namespace"
func (c *KubernetesConfig) GetNamespace() string {
	return c.Namespace
}

// 🔧 GetServiceAccount - "Get service account name"
func (c *KubernetesConfig) GetServiceAccount() string {
	return c.ServiceAccount
}

// 🔧 IsInCluster - "Check if running in cluster"
func (c *KubernetesConfig) IsInCluster() bool {
	return c.InCluster
}
