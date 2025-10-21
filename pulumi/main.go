package main

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	apiextensions "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stack := ctx.Stack()

		var clusterName string

		switch stack {
		case "studio":
			clusterName = "studio"
		case "homelab":
			clusterName = "homelab"
		default:
			return fmt.Errorf("unsupported stack: %s", stack)
		}

		// 🚀 Create Kind cluster (never delete!)
		createCluster, err := local.NewCommand(ctx, fmt.Sprintf("create-kind-cluster-%s", clusterName), &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`cd ../scripts && ./create-kind-cluster.sh %s`, clusterName)),
		})
		if err != nil {
			return err
		}

		// Create Kubernetes provider with Server-Side Apply enabled
		k8sProvider, err := kubernetes.NewProvider(ctx, "k8s-provider", &kubernetes.ProviderArgs{
			Context:               pulumi.String(fmt.Sprintf("kind-%s", clusterName)),
			EnableServerSideApply: pulumi.Bool(true),
			DeleteUnreachable:     pulumi.Bool(true),
		}, pulumi.DependsOn([]pulumi.Resource{createCluster}))
		if err != nil {
			return err
		}

		// 🔧 Install Flux via GitOps Job
		// Apply flux-bootstrap manifests directly using kubectl
		installFlux, err := local.NewCommand(ctx, "install-flux", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(
				`cd .. && kubectl apply -k flux/clusters/%s/infrastructure/flux-bootstrap --context kind-%s`,
				clusterName, clusterName)),
			Update: pulumi.String(fmt.Sprintf(
				`cd .. && kubectl apply -k flux/clusters/%s/infrastructure/flux-bootstrap --context kind-%s`,
				clusterName, clusterName)),
		}, pulumi.DependsOn([]pulumi.Resource{createCluster}))
		if err != nil {
			return err
		}

		// ⏳ Wait for flux-install Job to complete
		waitForFluxJob, err := local.NewCommand(ctx, "wait-for-flux-job", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(
				`kubectl wait --for=condition=complete --timeout=300s job/flux-install -n flux-system --context kind-%s`,
				clusterName)),
		}, pulumi.DependsOn([]pulumi.Resource{installFlux}))
		if err != nil {
			return err
		}

		// 🔐 Create secrets
		createSecrets, err := local.NewCommand(ctx, "create-secrets", &local.CommandArgs{
			Create: pulumi.String(`cd ../scripts && ./create-secrets.sh`),
		}, pulumi.DependsOn([]pulumi.Resource{waitForFluxJob}))
		if err != nil {
			return err
		}

		// 🔗 Linkerd is now managed by Flux via Jobs (see infrastructure/linkerd)
		// ✅ Flux is now also self-managed via Jobs (see infrastructure/flux-bootstrap)
		// No need to install via scripts - everything is GitOps!

		// 📚 Create GitRepository first (this was previously only in git.yaml causing chicken-egg problem)
		gitRepo, err := apiextensions.NewCustomResource(ctx, "homelab-gitrepository", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("source.toolkit.fluxcd.io/v1beta2"),
			Kind:       pulumi.String("GitRepository"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("homelab"),
				Namespace: pulumi.String("flux-system"),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"interval": pulumi.String("1m0s"),
					"url":      pulumi.String("https://github.com/brunovlucena/homelab.git"),
					"ref": pulumi.Map{
						"branch": pulumi.String("main"),
					},
				},
			},
		}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{createSecrets}))
		if err != nil {
			return err
		}

		// 📋 Create root Kustomization - applies kustomization.yaml which includes all phase Kustomizations
		_, err = apiextensions.NewCustomResource(ctx, "homelab-root-kustomization", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("kustomize.toolkit.fluxcd.io/v1"),
			Kind:       pulumi.String("Kustomization"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("homelab-root"),
				Namespace: pulumi.String("flux-system"),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"interval": pulumi.String("5m"),
					"sourceRef": pulumi.Map{
						"kind": pulumi.String("GitRepository"),
						"name": pulumi.String("homelab"),
					},
					"path":  pulumi.String(fmt.Sprintf("./flux/clusters/%s", clusterName)),
					"prune": pulumi.Bool(true),
					"wait":  pulumi.Bool(false), // Don't wait for all phases to complete
				},
			},
		}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{gitRepo}))
		if err != nil {
			return err
		}

		// Export outputs
		ctx.Export("clusterName", pulumi.String(clusterName))
		ctx.Export("fluxInstalled", pulumi.String("gitops-job"))
		ctx.Export("fluxRootKustomization", pulumi.String("applied"))
		ctx.Export("installation", pulumi.String("All components managed via Flux GitOps jobs"))

		return nil
	})
}
