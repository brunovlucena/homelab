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

		// 🔧 Install Flux
		installFlux, err := local.NewCommand(ctx, "install-flux", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`cd ../scripts && ./install-flux.sh %s`, clusterName)),
		}, pulumi.DependsOn([]pulumi.Resource{createCluster}))
		if err != nil {
			return err
		}

		// 🔐 Create secrets
		createSecrets, err := local.NewCommand(ctx, "create-secrets", &local.CommandArgs{
			Create: pulumi.String(`cd ../scripts && ./create-secrets.sh`),
		}, pulumi.DependsOn([]pulumi.Resource{installFlux}))
		if err != nil {
			return err
		}

		// 🔗 Install Linkerd
		installLinkerd, err := local.NewCommand(ctx, "install-linkerd", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`cd ../scripts && ./install-linkerd.sh %s`, clusterName)),
		}, pulumi.DependsOn([]pulumi.Resource{createSecrets}))
		if err != nil {
			return err
		}

		// 📊 Install Linkerd Viz
		_, err = local.NewCommand(ctx, "install-linkerd-viz", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`cd ../scripts && ./install-linkerd-viz.sh %s`, clusterName)),
		}, pulumi.DependsOn([]pulumi.Resource{installLinkerd}))
		if err != nil {
			return err
		}

		// 📦 Apply GitRepository using Pulumi Kubernetes with Server-Side Apply
		gitRepo, err := apiextensions.NewCustomResource(ctx, "homelab-gitrepo", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("source.toolkit.fluxcd.io/v1"),
			Kind:       pulumi.String("GitRepository"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("homelab"),
				Namespace: pulumi.String("flux-system"),
				Annotations: pulumi.StringMap{
					"pulumi.com/patchForce": pulumi.String("true"),
				},
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"interval": pulumi.String("1m"),
					"url":      pulumi.String("https://github.com/brunovlucena/homelab"),
					"ref": pulumi.Map{
						"branch": pulumi.String("main"),
					},
				},
			},
		}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{installFlux}))
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
		ctx.Export("fluxInstalled", pulumi.String("installed"))
		ctx.Export("linkerdInstalled", pulumi.String("installed"))
		ctx.Export("linkerdVizInstalled", pulumi.String("installed"))
		ctx.Export("fluxRootKustomization", pulumi.String("applied"))

		return nil
	})
}
