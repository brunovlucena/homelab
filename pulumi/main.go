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

		// Create Kubernetes provider
		k8sProvider, err := kubernetes.NewProvider(ctx, "k8s-provider", &kubernetes.ProviderArgs{
			Context: pulumi.String(fmt.Sprintf("kind-%s", clusterName)),
		}, pulumi.DependsOn([]pulumi.Resource{createCluster}))
		if err != nil {
			return err
		}

		// 🔧 Install Flux
		installFlux, err := local.NewCommand(ctx, "install-flux", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`cd ../scripts && ./install-flux.sh %s`, clusterName)),
			// TODO: Substitute this script with a Pulumi Kubernetes resource
		}, pulumi.DependsOn([]pulumi.Resource{createCluster}))
		if err != nil {
			return err
		}

		// 🔐 Create secrets
		createSecrets, err := local.NewCommand(ctx, "create-secrets", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`cd ../scripts && ./create-secrets.sh kind-%s`, clusterName)),
			// TODO: Substitute this script with a Pulumi Kubernetes resource
		}, pulumi.DependsOn([]pulumi.Resource{installFlux}))
		if err != nil {
			return err
		}

		// 🔗 Install Linkerd
		installLinkerd, err := local.NewCommand(ctx, "install-linkerd", &local.CommandArgs{
			Create: pulumi.String(fmt.Sprintf(`cd ../scripts && ./install-linkerd.sh %s`, clusterName)),
			// TODO: Substitute this script with a Pulumi Kubernetes resource
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

		// 📦 Apply GitRepository using Pulumi Kubernetes
		_, err = apiextensions.NewCustomResource(ctx, "homelab-gitrepo", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("source.toolkit.fluxcd.io/v1"),
			Kind:       pulumi.String("GitRepository"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("homelab"),
				Namespace: pulumi.String("flux-system"),
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

		// 📋 Phase 1: Deploy Core Infrastructure (CRD providers and repositories)
		// This must be deployed first to install CRDs needed by subsequent phases
		phase1Kustomization, err := apiextensions.NewCustomResource(ctx, "homelab-phase1-kustomization", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("kustomize.toolkit.fluxcd.io/v1"),
			Kind:       pulumi.String("Kustomization"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("homelab-phase1-core"),
				Namespace: pulumi.String("flux-system"),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"interval": pulumi.String("10m"),
					"sourceRef": pulumi.Map{
						"kind": pulumi.String("GitRepository"),
						"name": pulumi.String("homelab"),
					},
					"path":  pulumi.String(fmt.Sprintf("./flux/clusters/%s/infrastructure/repositories", clusterName)),
					"prune": pulumi.Bool(true),
					"wait":  pulumi.Bool(true),
				},
			},
		}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{installFlux}))
		if err != nil {
			return err
		}

		// 📋 Phase 2: Deploy Prometheus Operator (provides CRDs)
		phase2Kustomization, err := apiextensions.NewCustomResource(ctx, "homelab-phase2-kustomization", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("kustomize.toolkit.fluxcd.io/v1"),
			Kind:       pulumi.String("Kustomization"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("homelab-phase2-prometheus"),
				Namespace: pulumi.String("flux-system"),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"interval": pulumi.String("10m"),
					"sourceRef": pulumi.Map{
						"kind": pulumi.String("GitRepository"),
						"name": pulumi.String("homelab"),
					},
					"path":  pulumi.String(fmt.Sprintf("./flux/clusters/%s/infrastructure/prometheus-operator", clusterName)),
					"prune": pulumi.Bool(true),
					"wait":  pulumi.Bool(true),
					"dependsOn": pulumi.Array{
						pulumi.Map{
							"name": pulumi.String("homelab-phase1-core"),
						},
					},
				},
			},
		}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{phase1Kustomization}))
		if err != nil {
			return err
		}

		// 📋 Phase 3: Deploy Knative Operator (provides Knative CRDs)
		phase3Kustomization, err := apiextensions.NewCustomResource(ctx, "homelab-phase3-kustomization", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("kustomize.toolkit.fluxcd.io/v1"),
			Kind:       pulumi.String("Kustomization"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("homelab-phase3-knative"),
				Namespace: pulumi.String("flux-system"),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"interval": pulumi.String("10m"),
					"sourceRef": pulumi.Map{
						"kind": pulumi.String("GitRepository"),
						"name": pulumi.String("homelab"),
					},
					"path":  pulumi.String(fmt.Sprintf("./flux/clusters/%s/infrastructure/knative-operator", clusterName)),
					"prune": pulumi.Bool(true),
					"wait":  pulumi.Bool(true),
					"dependsOn": pulumi.Array{
						pulumi.Map{
							"name": pulumi.String("homelab-phase2-prometheus"),
						},
					},
				},
			},
		}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{phase2Kustomization}))
		if err != nil {
			return err
		}

		// 📋 Phase 4: Deploy Everything Else (depends on all CRDs being available)
		_, err = apiextensions.NewCustomResource(ctx, "homelab-phase4-kustomization", &apiextensions.CustomResourceArgs{
			ApiVersion: pulumi.String("kustomize.toolkit.fluxcd.io/v1"),
			Kind:       pulumi.String("Kustomization"),
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String("homelab-phase4-apps"),
				Namespace: pulumi.String("flux-system"),
			},
			OtherFields: kubernetes.UntypedArgs{
				"spec": pulumi.Map{
					"interval": pulumi.String("10m"),
					"sourceRef": pulumi.Map{
						"kind": pulumi.String("GitRepository"),
						"name": pulumi.String("homelab"),
					},
					"path":  pulumi.String(fmt.Sprintf("./flux/clusters/%s/infrastructure", clusterName)),
					"prune": pulumi.Bool(true),
					"wait":  pulumi.Bool(true),
					"dependsOn": pulumi.Array{
						pulumi.Map{
							"name": pulumi.String("homelab-phase3-knative"),
						},
					},
				},
			},
		}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{phase3Kustomization}))
		if err != nil {
			return err
		}

		// Export outputs
		ctx.Export("clusterName", pulumi.String(clusterName))
		ctx.Export("fluxInstalled", pulumi.String("installed"))
		ctx.Export("linkerdInstalled", pulumi.String("installed"))
		ctx.Export("linkerdVizInstalled", pulumi.String("installed"))
		ctx.Export("fluxKustomizationApplied", pulumi.String("applied"))

		return nil
	})
}
