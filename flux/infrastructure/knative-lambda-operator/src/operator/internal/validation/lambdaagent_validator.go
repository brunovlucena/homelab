/*
Copyright 2024 Bruno Lucena.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
*/

// Package validation provides admission webhook validation for LambdaAgent resources
package validation

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	lambdav1alpha1 "github.com/brunovlucena/knative-lambda-operator/api/v1alpha1"
)

// LambdaAgentValidator validates LambdaAgent resources
type LambdaAgentValidator struct{}

var _ webhook.CustomValidator = &LambdaAgentValidator{}

// SetupWebhookWithManager sets up the webhook with the Manager
func (v *LambdaAgentValidator) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&lambdav1alpha1.LambdaAgent{}).
		WithValidator(v).
		Complete()
}

// ValidateCreate implements webhook.CustomValidator
func (v *LambdaAgentValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	agent, ok := obj.(*lambdav1alpha1.LambdaAgent)
	if !ok {
		return nil, fmt.Errorf("expected LambdaAgent, got %T", obj)
	}

	var allErrs field.ErrorList
	allErrs = append(allErrs, v.validateSpec(agent)...)

	if len(allErrs) > 0 {
		return nil, allErrs.ToAggregate()
	}
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator
func (v *LambdaAgentValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	agent, ok := newObj.(*lambdav1alpha1.LambdaAgent)
	if !ok {
		return nil, fmt.Errorf("expected LambdaAgent, got %T", newObj)
	}

	var allErrs field.ErrorList
	allErrs = append(allErrs, v.validateSpec(agent)...)

	if len(allErrs) > 0 {
		return nil, allErrs.ToAggregate()
	}
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator
func (v *LambdaAgentValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// No validation needed for delete
	return nil, nil
}

// validateSpec validates the LambdaAgent spec
func (v *LambdaAgentValidator) validateSpec(agent *lambdav1alpha1.LambdaAgent) field.ErrorList {
	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	// Validate image
	allErrs = append(allErrs, v.validateImage(agent, specPath.Child("image"))...)

	// Validate AI configuration
	if agent.Spec.AI != nil {
		allErrs = append(allErrs, v.validateAI(agent.Spec.AI, specPath.Child("ai"))...)
	}

	// Validate scaling
	if agent.Spec.Scaling != nil {
		allErrs = append(allErrs, v.validateScaling(agent.Spec.Scaling, specPath.Child("scaling"))...)
	}

	// Validate eventing
	if agent.Spec.Eventing != nil {
		allErrs = append(allErrs, v.validateEventing(agent.Spec.Eventing, specPath.Child("eventing"))...)
	}

	// Validate resources
	if agent.Spec.Resources != nil {
		allErrs = append(allErrs, v.validateResources(agent.Spec.Resources, specPath.Child("resources"))...)
	}

	return allErrs
}

// validateImage validates image configuration
func (v *LambdaAgentValidator) validateImage(agent *lambdav1alpha1.LambdaAgent, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	image := agent.Spec.Image

	// Repository is required
	if image.Repository == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("repository"), "repository is required"))
	} else {
		// Validate repository format
		if !isValidImageRepository(image.Repository) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("repository"), image.Repository,
				"invalid image repository format"))
		}
	}

	// Validate port
	if image.Port < 0 || image.Port > 65535 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("port"), image.Port,
			"port must be between 0 and 65535"))
	}

	// Validate digest format if provided
	if image.Digest != "" && !isValidDigest(image.Digest) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("digest"), image.Digest,
			"invalid digest format, expected sha256:..."))
	}

	return allErrs
}

// validateAI validates AI configuration
func (v *LambdaAgentValidator) validateAI(ai *lambdav1alpha1.AgentAISpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// Validate temperature (should be between 0.0 and 2.0)
	if ai.Temperature != "" {
		temp, err := strconv.ParseFloat(ai.Temperature, 64)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("temperature"), ai.Temperature,
				"temperature must be a valid number"))
		} else if temp < 0.0 || temp > 2.0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("temperature"), ai.Temperature,
				"temperature must be between 0.0 and 2.0"))
		}
	}

	// Validate maxTokens
	if ai.MaxTokens < 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("maxTokens"), ai.MaxTokens,
			"maxTokens must be a positive integer"))
	}
	if ai.MaxTokens > 1000000 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("maxTokens"), ai.MaxTokens,
			"maxTokens exceeds reasonable limit (1000000)"))
	}

	// Validate endpoint URL if provided
	if ai.Endpoint != "" {
		if _, err := url.Parse(ai.Endpoint); err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("endpoint"), ai.Endpoint,
				"invalid endpoint URL"))
		}
	}

	// Validate API key required for cloud providers
	if ai.Provider == "openai" || ai.Provider == "anthropic" {
		if ai.APIKeySecretRef == nil {
			allErrs = append(allErrs, field.Required(fldPath.Child("apiKeySecretRef"),
				fmt.Sprintf("API key secret is required for provider %s", ai.Provider)))
		}
	}

	// Validate API key secret reference
	if ai.APIKeySecretRef != nil {
		if ai.APIKeySecretRef.Name == "" {
			allErrs = append(allErrs, field.Required(fldPath.Child("apiKeySecretRef", "name"),
				"secret name is required"))
		}
		if ai.APIKeySecretRef.Key == "" {
			allErrs = append(allErrs, field.Required(fldPath.Child("apiKeySecretRef", "key"),
				"secret key is required"))
		}
	}

	return allErrs
}

// validateScaling validates scaling configuration
func (v *LambdaAgentValidator) validateScaling(scaling *lambdav1alpha1.AgentScalingSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// Validate minReplicas
	if scaling.MinReplicas < 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("minReplicas"), scaling.MinReplicas,
			"minReplicas cannot be negative"))
	}

	// Validate maxReplicas
	if scaling.MaxReplicas < 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("maxReplicas"), scaling.MaxReplicas,
			"maxReplicas cannot be negative"))
	}

	// Validate min <= max
	if scaling.MinReplicas > scaling.MaxReplicas {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("minReplicas"), scaling.MinReplicas,
			fmt.Sprintf("minReplicas (%d) cannot be greater than maxReplicas (%d)",
				scaling.MinReplicas, scaling.MaxReplicas)))
	}

	// Validate targetConcurrency
	if scaling.TargetConcurrency < 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("targetConcurrency"), scaling.TargetConcurrency,
			"targetConcurrency cannot be negative"))
	}

	// Validate scaleToZeroGracePeriod format
	if scaling.ScaleToZeroGracePeriod != "" {
		if !isValidDuration(scaling.ScaleToZeroGracePeriod) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("scaleToZeroGracePeriod"),
				scaling.ScaleToZeroGracePeriod, "invalid duration format (e.g., 30s, 5m, 1h)"))
		}
	}

	return allErrs
}

// validateEventing validates eventing configuration
func (v *LambdaAgentValidator) validateEventing(eventing *lambdav1alpha1.AgentEventingSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// Validate subscriptions
	for i, sub := range eventing.Subscriptions {
		subPath := fldPath.Child("subscriptions").Index(i)
		if sub.EventType == "" {
			allErrs = append(allErrs, field.Required(subPath.Child("eventType"),
				"eventType is required for subscription"))
		} else if !isValidCloudEventType(sub.EventType) {
			allErrs = append(allErrs, field.Invalid(subPath.Child("eventType"), sub.EventType,
				"invalid CloudEvent type format"))
		}
	}

	// Validate forwards
	for i, fwd := range eventing.Forwards {
		fwdPath := fldPath.Child("forwards").Index(i)

		if len(fwd.EventTypes) == 0 {
			allErrs = append(allErrs, field.Required(fwdPath.Child("eventTypes"),
				"at least one eventType is required for forward"))
		}

		for j, eventType := range fwd.EventTypes {
			if !isValidCloudEventType(eventType) {
				allErrs = append(allErrs, field.Invalid(fwdPath.Child("eventTypes").Index(j),
					eventType, "invalid CloudEvent type format"))
			}
		}

		if fwd.TargetAgent == "" {
			allErrs = append(allErrs, field.Required(fwdPath.Child("targetAgent"),
				"targetAgent is required for forward"))
		}

		if fwd.TargetNamespace == "" {
			allErrs = append(allErrs, field.Required(fwdPath.Child("targetNamespace"),
				"targetNamespace is required for forward"))
		}
	}

	// Validate DLQ
	if eventing.DLQ != nil {
		if eventing.DLQ.RetryMaxAttempts < 0 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("dlq", "retryMaxAttempts"),
				eventing.DLQ.RetryMaxAttempts, "retryMaxAttempts cannot be negative"))
		}
		if eventing.DLQ.RetryMaxAttempts > 100 {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("dlq", "retryMaxAttempts"),
				eventing.DLQ.RetryMaxAttempts, "retryMaxAttempts exceeds reasonable limit (100)"))
		}
	}

	return allErrs
}

// validateResources validates resource configuration
func (v *LambdaAgentValidator) validateResources(resources *lambdav1alpha1.AgentResourcesSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	// Validate requests
	if resources.Requests != nil {
		if resources.Requests.CPU != "" && !isValidResourceQuantity(resources.Requests.CPU) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("requests", "cpu"),
				resources.Requests.CPU, "invalid CPU quantity format"))
		}
		if resources.Requests.Memory != "" && !isValidResourceQuantity(resources.Requests.Memory) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("requests", "memory"),
				resources.Requests.Memory, "invalid memory quantity format"))
		}
	}

	// Validate limits
	if resources.Limits != nil {
		if resources.Limits.CPU != "" && !isValidResourceQuantity(resources.Limits.CPU) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("limits", "cpu"),
				resources.Limits.CPU, "invalid CPU quantity format"))
		}
		if resources.Limits.Memory != "" && !isValidResourceQuantity(resources.Limits.Memory) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("limits", "memory"),
				resources.Limits.Memory, "invalid memory quantity format"))
		}
	}

	return allErrs
}

// Helper functions for validation

// isValidImageRepository checks if the repository format is valid
func isValidImageRepository(repo string) bool {
	// Basic validation - not empty and doesn't contain invalid chars
	if repo == "" {
		return false
	}
	// Allow localhost, registry paths, etc.
	return !strings.ContainsAny(repo, " \t\n")
}

// isValidDigest checks if the digest format is valid
func isValidDigest(digest string) bool {
	// Digest should be in format sha256:hexstring
	if !strings.HasPrefix(digest, "sha256:") {
		return false
	}
	hexPart := strings.TrimPrefix(digest, "sha256:")
	if len(hexPart) != 64 {
		return false
	}
	for _, c := range hexPart {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// isValidDuration checks if the duration string is valid
func isValidDuration(duration string) bool {
	// Simple validation for common duration formats
	if duration == "" {
		return false
	}
	// Check for formats like "30s", "5m", "1h", "0s"
	validSuffixes := []string{"s", "m", "h"}
	hasValidSuffix := false
	for _, suffix := range validSuffixes {
		if strings.HasSuffix(duration, suffix) {
			hasValidSuffix = true
			break
		}
	}
	if !hasValidSuffix {
		return false
	}
	// Try to parse the number part
	numPart := duration[:len(duration)-1]
	_, err := strconv.ParseFloat(numPart, 64)
	return err == nil
}

// isValidCloudEventType checks if the CloudEvent type format is valid
func isValidCloudEventType(eventType string) bool {
	// CloudEvent types should be reverse-DNS format
	// e.g., "io.homelab.chat.message", "com.example.event"
	if eventType == "" {
		return false
	}
	// Must contain at least one dot
	if !strings.Contains(eventType, ".") {
		return false
	}
	// Cannot start or end with dot
	if strings.HasPrefix(eventType, ".") || strings.HasSuffix(eventType, ".") {
		return false
	}
	return true
}

// isValidResourceQuantity checks if the Kubernetes resource quantity is valid
func isValidResourceQuantity(qty string) bool {
	// Valid formats: 100m, 1, 1.5, 256Mi, 1Gi, etc.
	if qty == "" {
		return false
	}
	// Try to match common patterns
	validSuffixes := []string{"m", "Mi", "Gi", "Ki", "Ti", "Pi", "Ei", "M", "G", "K", "T", "P", "E", ""}
	for _, suffix := range validSuffixes {
		if strings.HasSuffix(qty, suffix) {
			numPart := strings.TrimSuffix(qty, suffix)
			if numPart == "" {
				continue
			}
			_, err := strconv.ParseFloat(numPart, 64)
			if err == nil {
				return true
			}
		}
	}
	return false
}
