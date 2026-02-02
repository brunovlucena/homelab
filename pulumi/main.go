package main

// TODO: https://www.pulumi.com/docs/iac/guides/continuous-delivery/pulumi-kubernetes-operator/
import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/kustomize"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// Configuration constants
const (
	DefaultFluxNamespace = "flux-system"
)

// Config holds cluster configuration
type Config struct {
	StackName                  string
	ClusterName                string
	ClusterPath                string
	KindConfigPath             string
	KubeContext                string
	FluxClusterPath            string
	FluxBootstrapPath          string
	GitRepositoryPath          string
	HelmRepositoryPath         string
	SealedSecretsPath          string
	ExternalSecretsPath        string
	RepositoriesRootPath       string
	SealedSecretsBackupPattern string
}

// NewConfig creates and validates cluster configuration
func NewConfig(stackName string, clusterName string) (*Config, error) {
	cfg := &Config{
		StackName:   stackName,
		ClusterName: clusterName,
		KubeContext: fmt.Sprintf("homelab-%s", clusterName),
		ClusterPath: filepath.Join("..", "flux", "clusters", clusterName),
	}

	fluxPath := func(parts ...string) string {
		return filepath.Join(append([]string{"..", "flux"}, parts...)...)
	}

	// Validate cluster exists in flux/clusters/
	cfg.ClusterPath = fluxPath("clusters", cfg.ClusterName)
	if _, err := os.Stat(cfg.ClusterPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cluster '%s' not found in flux/clusters/. Available clusters can be found in flux/clusters/ directory", cfg.ClusterName)
	}

	cfg.KindConfigPath = filepath.Join(cfg.ClusterPath, "kind.yaml")
	cfg.FluxClusterPath = cfg.ClusterPath
	cfg.FluxBootstrapPath = fluxPath("infrastructure", "flux")
	cfg.RepositoriesRootPath = fluxPath("infrastructure")

	return cfg, nil
}

func camelToKebab(camel string) string {
	return strings.ToLower(strings.ReplaceAll(regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(camel, "${1}-${2}"), "_", "-"))
}

func getSecretsConfig() map[string][]string {
	return map[string][]string{
		"github-token": {"githubToken"},
		"homepage":     {"homepagePostgresPassword", "homepageRedisPassword"},
		"prometheus":   {"grafanaPassword", "grafanaApiKey", "pagerdutyUrl", "pagerdutyServiceKey", "slackWebhookUrl"},
		"logfire":      {"logfireToken"},
		"cloudflare":   {"cloudflareEmail", "cloudflareApiKey", "cloudflareApiToken", "cloudflareWarpToken", "cloudflareTunnelToken"},
		"twingate":     {"twingateAccessToken", "twingateRefreshToken"},
		"pulumi":       {"pulumiAccessToken"},
		// Flux CDEvents webhook tokens for Notification Controller receivers
		// Used to authenticate CloudEvents from Knative Lambda and agent-contracts
		"lambda-webhook-token":          {"fluxLambdaWebhookToken"},
		"agent-contracts-webhook-token": {"fluxAgentContractsWebhookToken"},
		// Linear API key for agent-sre and other agents
		"linear-api-key": {"linearApiKey"},
		// Jira credentials for agent-sre and other agents
		"jira-credentials": {"jiraUrl", "jiraEmail", "jiraApiToken"},
	}
}

func createSecrets(ctx *pulumi.Context, k8sProvider *kubernetes.Provider) error {
	conf := config.New(ctx, "")
	secrets := getSecretsConfig()

	for secretName, keys := range secrets {
		// Special handling for github-token: needs different keys in different namespaces
		if secretName == "github-token" {
			githubToken := conf.Get("githubToken")
			if githubToken == "" {
				return fmt.Errorf("githubToken is required")
				// continue
			}

			// Create secret in flux-system namespace with username and password keys
			_, err := corev1.NewSecret(ctx, fmt.Sprintf("%s-flux-system", secretName), &corev1.SecretArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(secretName),
					Namespace: pulumi.String(DefaultFluxNamespace),
				},
				StringData: pulumi.StringMap{
					"username": pulumi.String("git"),
					"password": pulumi.String(githubToken),
				},
			}, pulumi.Provider(k8sProvider))
			if err != nil {
				return fmt.Errorf("failed to create %s secret in flux-system: %w", secretName, err)
			}

			continue
		}

		// Special handling for Flux CDEvents webhook tokens
		// Flux Receiver expects the key to be "token" in the secret
		if secretName == "lambda-webhook-token" || secretName == "agent-contracts-webhook-token" {
			tokenValue := conf.Get(keys[0]) // First key is the token config
			if tokenValue == "" {
				// Skip if not configured (optional for local dev)
				continue
			}

			_, err := corev1.NewSecret(ctx, secretName, &corev1.SecretArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(secretName),
					Namespace: pulumi.String(DefaultFluxNamespace),
					Labels: pulumi.StringMap{
						"app.kubernetes.io/component": pulumi.String("cdevents-receiver"),
						"app.kubernetes.io/part-of":   pulumi.String("flux-system"),
					},
				},
				StringData: pulumi.StringMap{
					"token": pulumi.String(tokenValue),
				},
			}, pulumi.Provider(k8sProvider))
			if err != nil {
				return fmt.Errorf("failed to create %s secret: %w", secretName, err)
			}

			continue
		}

		// Special handling for linear-api-key: create in ai namespace
		if secretName == "linear-api-key" {
			apiKeyValue := conf.Get(keys[0]) // First key is the API key config
			if apiKeyValue == "" {
				// Skip if not configured (optional for local dev)
				continue
			}

			_, err := corev1.NewSecret(ctx, secretName, &corev1.SecretArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(secretName),
					Namespace: pulumi.String("ai"),
					Labels: pulumi.StringMap{
						"app.kubernetes.io/component": pulumi.String("linear-integration"),
						"app.kubernetes.io/part-of":   pulumi.String("homelab-ai"),
					},
				},
				StringData: pulumi.StringMap{
					"api-key": pulumi.String(apiKeyValue),
				},
			}, pulumi.Provider(k8sProvider))
			if err != nil {
				return fmt.Errorf("failed to create %s secret: %w", secretName, err)
			}

			continue
		}

		// Special handling for jira-credentials: create in ai namespace
		// Follows Jira recommendation: secret keys should be jira-email and jira-api-token
		if secretName == "jira-credentials" {
			jiraUrl := conf.Get(keys[0])      // jiraUrl
			jiraEmail := conf.Get(keys[1])    // jiraEmail
			jiraApiToken := conf.Get(keys[2]) // jiraApiToken

			// Skip if not all credentials are configured (optional for local dev)
			if jiraUrl == "" || jiraEmail == "" || jiraApiToken == "" {
				continue
			}

			_, err := corev1.NewSecret(ctx, secretName, &corev1.SecretArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(secretName),
					Namespace: pulumi.String("ai"),
					Labels: pulumi.StringMap{
						"app.kubernetes.io/component": pulumi.String("jira-integration"),
						"app.kubernetes.io/part-of":   pulumi.String("homelab-ai"),
					},
				},
				StringData: pulumi.StringMap{
					"jira-url":       pulumi.String(jiraUrl),
					"jira-email":     pulumi.String(jiraEmail),
					"jira-api-token": pulumi.String(jiraApiToken),
				},
			}, pulumi.Provider(k8sProvider))
			if err != nil {
				return fmt.Errorf("failed to create %s secret: %w", secretName, err)
			}

			continue
		}

		// Standard secret creation for other secrets
		stringData := pulumi.StringMap{}
		hasData := false

		for _, key := range keys {
			if value := conf.Get(key); value != "" {
				stringData[camelToKebab(key)] = pulumi.String(value)
				hasData = true
			}
		}

		if hasData {
			_, err := corev1.NewSecret(ctx, secretName, &corev1.SecretArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Name:      pulumi.String(secretName),
					Namespace: pulumi.String(DefaultFluxNamespace),
				},
				StringData: stringData,
			}, pulumi.Provider(k8sProvider))
			if err != nil {
				return fmt.Errorf("failed to create %s secret: %w", secretName, err)
			}
		}
	}
	return nil
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// ✅ Load and validate configuration
		// Extract cluster name from stack name (e.g., "studio-studio" -> "studio", "pro-pro" -> "pro")
		stackName := ctx.Stack()
		clusterName := stackName
		// If stack name contains a dash, take the first part (machine-specific stacks: machine-cluster)
		if parts := strings.Split(stackName, "-"); len(parts) > 1 {
			// Check if it's a machine-specific format (e.g., "studio-studio", "pro-pro")
			// In that case, both parts are the same, so use the first part
			if parts[0] == parts[1] {
				clusterName = parts[0]
			} else {
				// Otherwise, use the full stack name as cluster name
				clusterName = stackName
			}
		}
		cfg, err := NewConfig(stackName, clusterName)
		if err != nil {
			return err
		}

		// Create Kubernetes provider (cluster already exists from separate bootstrap)
		k8sProvider, err := kubernetes.NewProvider(ctx, "k8s-provider", &kubernetes.ProviderArgs{
			Context:               pulumi.String(cfg.KubeContext),
			EnableServerSideApply: pulumi.Bool(true),
			DeleteUnreachable:     pulumi.Bool(true),
		})
		if err != nil {
			return fmt.Errorf("failed to create k8s provider: %w", err)
		}

		// Create ai namespace (required for linear-api-key and jira-credentials secrets)
		_, err = corev1.NewNamespace(ctx, "ai-namespace", &corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String("ai"),
			},
		}, pulumi.Provider(k8sProvider))
		if err != nil {
			return fmt.Errorf("failed to create ai namespace: %w", err)
		}

		err = createSecrets(ctx, k8sProvider)
		if err != nil {
			return fmt.Errorf("failed to create secrets: %w", err)
		}

		// Apply flux-install-job.yaml directly (bootstrap - before Flux is running)
		fluxInstallJobPath := filepath.Join(cfg.FluxBootstrapPath, "flux-install-job.yaml")
		_, err = yaml.NewConfigFile(ctx, "flux-install-job", &yaml.ConfigFileArgs{
			File: fluxInstallJobPath,
		}, pulumi.Provider(k8sProvider))
		if err != nil {
			return fmt.Errorf("failed to deploy flux install job: %w", err)
		}

		// Apply Flux-managed resources (GitRepository, HelmRepository, etc.)
		_, err = kustomize.NewDirectory(ctx, "flux-bootstrap", kustomize.DirectoryArgs{
			Directory: pulumi.String(cfg.FluxBootstrapPath),
		}, pulumi.Provider(k8sProvider))
		if err != nil {
			return fmt.Errorf("failed to deploy flux bootstrap resources: %w", err)
		}

		_, err = kustomize.NewDirectory(ctx, "flux-cluster-workloads", kustomize.DirectoryArgs{
			Directory: pulumi.String(cfg.FluxClusterPath),
		}, pulumi.Provider(k8sProvider))
		if err != nil {
			return fmt.Errorf("failed to deploy cluster workloads via flux: %w", err)
		}

		// Export outputs
		ctx.Export("clusterName", pulumi.String(cfg.ClusterName))
		ctx.Export("kubeContext", pulumi.String(cfg.KubeContext))
		ctx.Export("fluxInstalled", pulumi.String("managed-by-pulumi"))
		ctx.Export("rootKustomization", pulumi.String(fmt.Sprintf("applied via clusters/%s directory", cfg.ClusterName)))
		ctx.Export("nextSteps", pulumi.String("Run: make observe"))

		return nil
	})
}
