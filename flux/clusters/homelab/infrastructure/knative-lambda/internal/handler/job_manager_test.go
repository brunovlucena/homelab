// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🧪 JOB MANAGER TESTS - Comprehensive unit tests for JobManager
//
//	🎯 Purpose: Test JobManager functionality, especially registry mirror configuration
//	💡 Features: Mock Kubernetes client, test registry mirror args, validate job creation
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package handler

import (
	"fmt"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes/fake"

	"knative-lambda-new/internal/config"
	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/observability"
	"knative-lambda-new/internal/resilience"
	"knative-lambda-new/pkg/builds"
)

// 🧪 TestJobManager - "Test JobManager functionality"
func TestJobManager(t *testing.T) {
	t.Run("createKanikoArgs_RegistryMirror_NotEmpty", testCreateKanikoArgsRegistryMirrorNotEmpty)
	t.Run("createKanikoArgs_SkipTLSVerifyRegistry_NotEmpty", testCreateKanikoArgsSkipTLSVerifyRegistryNotEmpty)
	t.Run("createKanikoArgs_BaseImages_Correct", testCreateKanikoArgsBaseImagesCorrect)
	t.Run("createKanikoArgs_AllArguments_Present", testCreateKanikoArgsAllArgumentsPresent)
	t.Run("createKanikoArgs_CompleteValidation", testCreateKanikoArgsCompleteValidation)
	t.Run("createKanikoArgs_EmptyValuesDetection", testCreateKanikoArgsEmptyValuesDetection)
	t.Run("createKanikoArgs_EnvironmentVariableSimulation", testCreateKanikoArgsEnvironmentVariableSimulation)
	t.Run("createKanikoContainer_EnvironmentVariables", testCreateKanikoContainerEnvironmentVariables)
	t.Run("createKanikoContainer_EnvironmentVariablesMissing", testCreateKanikoContainerEnvironmentVariablesMissing)
	t.Run("NewJobManager_ValidConfig_Success", testNewJobManagerValidConfigSuccess)
	t.Run("NewJobManager_NilK8sClient_Error", testNewJobManagerNilK8sClientError)
	t.Run("NewJobManager_NilK8sConfig_Error", testNewJobManagerNilK8sConfigError)
	t.Run("NewJobManager_NilBuildConfig_Error", testNewJobManagerNilBuildConfigError)
	t.Run("NewJobManager_NilObservability_Error", testNewJobManagerNilObservabilityError)
	t.Run("GenerateJobName_ValidInputs_Success", testGenerateJobNameValidInputsSuccess)
	t.Run("IsJobRunning_ActiveJob_True", testIsJobRunningActiveJobTrue)
	t.Run("IsJobRunning_NoActiveJob_False", testIsJobRunningNoActiveJobFalse)
	t.Run("IsJobFailed_FailedJob_True", testIsJobFailedFailedJobTrue)
	t.Run("IsJobFailed_SucceededJob_False", testIsJobFailedSucceededJobFalse)
	t.Run("IsJobSucceeded_SucceededJob_True", testIsJobSucceededSucceededJobTrue)
	t.Run("IsJobSucceeded_FailedJob_False", testIsJobSucceededFailedJobFalse)
	t.Run("generateImageURI_ValidInputs_Success", testGenerateImageURISuccess)
	t.Run("generateImageURI_NilAWSConfig_Fallback", testGenerateImageURINilAWSConfigFallback)
}

// 🧪 testCreateKanikoArgsRegistryMirrorNotEmpty - "Test that registry mirror argument is not empty"
func testCreateKanikoArgsRegistryMirrorNotEmpty(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	buildRequest := createTestBuildRequest()
	destinationImageURI := "test-registry/test-image:latest"

	// Log the AWS config to verify values
	t.Logf("🔍 AWS Config RegistryMirror: '%s'", jobManager.awsConfig.RegistryMirror)
	t.Logf("🔍 AWS Config SkipTLSVerifyRegistry: '%s'", jobManager.awsConfig.SkipTLSVerifyRegistry)

	// Act
	args := jobManager.createKanikoArgs(buildRequest, destinationImageURI)

	// Log ALL arguments for debugging
	t.Logf("🔍 ALL Kaniko arguments:")
	for i, arg := range args {
		t.Logf("  [%d] %s", i, arg)
	}

	// Assert - Registry Mirror
	registryMirrorArg := findArg(args, "--registry-mirror=")
	if registryMirrorArg == "" {
		t.Fatal("❌ Registry mirror argument is missing!")
	}

	// Extract the value after the equals sign
	registryMirrorValue := extractArgValue(registryMirrorArg, "--registry-mirror=")
	if registryMirrorValue == "" {
		t.Fatal("❌ Registry mirror value is empty! Expected non-empty value")
	}

	// Validate the value is correct
	if registryMirrorValue != "docker.io" {
		t.Fatalf("❌ Registry mirror value is incorrect! Expected 'docker.io', got '%s'", registryMirrorValue)
	}

	t.Logf("✅ Registry mirror argument found: %s", registryMirrorArg)
	t.Logf("✅ Registry mirror value: %s", registryMirrorValue)
	t.Logf("✅ Registry mirror validation PASSED!")
}

// 🧪 testCreateKanikoArgsCompleteValidation - "Comprehensive test of ALL Kaniko arguments"
func testCreateKanikoArgsCompleteValidation(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	buildRequest := createTestBuildRequest()
	destinationImageURI := "test-registry/test-image:latest"

	// Log the AWS config to verify values
	t.Logf("🔍 AWS Config RegistryMirror: '%s'", jobManager.awsConfig.RegistryMirror)
	t.Logf("🔍 AWS Config SkipTLSVerifyRegistry: '%s'", jobManager.awsConfig.SkipTLSVerifyRegistry)
	t.Logf("🔍 AWS Config NodeBaseImage: '%s'", jobManager.awsConfig.NodeBaseImage)
	t.Logf("🔍 AWS Config PythonBaseImage: '%s'", jobManager.awsConfig.PythonBaseImage)
	t.Logf("🔍 AWS Config GoBaseImage: '%s'", jobManager.awsConfig.GoBaseImage)
	t.Logf("🔍 AWS Config S3TempBucket: '%s'", jobManager.awsConfig.S3TempBucket)

	// Act
	args := jobManager.createKanikoArgs(buildRequest, destinationImageURI)

	// Log ALL arguments for debugging
	t.Logf("🔍 ALL Kaniko arguments:")
	for i, arg := range args {
		t.Logf("  [%d] %s", i, arg)
	}

	// Assert - Check that we have exactly 8 arguments
	if len(args) != 8 {
		t.Fatalf("❌ Expected 8 arguments, got %d", len(args))
	}

	// Validate each argument individually
	expectedArgs := map[string]string{
		"--context=":                     fmt.Sprintf("s3://%s/build-context/%s/context.tar.gz", jobManager.awsConfig.S3TempBucket, buildRequest.ParserID),
		"--destination=":                 destinationImageURI,
		"--dockerfile=":                  constants.DefaultDockerfilePath,
		"--registry-mirror=":             "docker.io",
		"--skip-tls-verify-registry=":    "docker.io",
		"--build-arg=NODE_BASE_IMAGE=":   "docker.io/library/node:22-alpine",
		"--build-arg=PYTHON_BASE_IMAGE=": "docker.io/library/python:3.11-alpine",
		"--build-arg=GO_BASE_IMAGE=":     "docker.io/library/golang:1.21-alpine",
	}

	for expectedPrefix, expectedValue := range expectedArgs {
		arg := findArg(args, expectedPrefix)
		if arg == "" {
			t.Fatalf("❌ Required argument '%s' is missing!", expectedPrefix)
		}

		actualValue := extractArgValue(arg, expectedPrefix)
		if actualValue != expectedValue {
			t.Fatalf("❌ Argument '%s' has wrong value! Expected '%s', got '%s'", expectedPrefix, expectedValue, actualValue)
		}

		t.Logf("✅ Argument '%s' = '%s' ✓", expectedPrefix, actualValue)
	}

	// Special validation for registry mirror arguments
	registryMirrorArg := findArg(args, "--registry-mirror=")
	registryMirrorValue := extractArgValue(registryMirrorArg, "--registry-mirror=")
	if registryMirrorValue == "" {
		t.Fatal("❌ Registry mirror value is EMPTY! This is the main issue!")
	}

	skipTLSArg := findArg(args, "--skip-tls-verify-registry=")
	skipTLSValue := extractArgValue(skipTLSArg, "--skip-tls-verify-registry=")
	if skipTLSValue == "" {
		t.Fatal("❌ Skip TLS verify registry value is EMPTY! This is the main issue!")
	}

	t.Logf("✅ Registry mirror: '%s'", registryMirrorValue)
	t.Logf("✅ Skip TLS verify registry: '%s'", skipTLSValue)
	t.Logf("✅ ALL Kaniko arguments validation PASSED!")
}

// 🧪 testCreateKanikoArgsEmptyValuesDetection - "Test that reproduces the EXACT Kaniko error"
func testCreateKanikoArgsEmptyValuesDetection(t *testing.T) {
	// Arrange - Create JobManager with EMPTY registry values to reproduce the REAL error
	emptyAWSConfig := &config.AWSConfig{
		AWSRegion:             "us-west-2",
		AWSAccountID:          "339954290315",
		ECRRegistry:           "339954290315.dkr.ecr.us-west-2.amazonaws.com",
		ECRRepositoryName:     "knative-lambdas",
		S3SourceBucket:        "test-source-bucket",
		S3TempBucket:          "test-temp-bucket",
		RegistryMirror:        "", // 🔥 EMPTY VALUE - This causes the REAL error!
		SkipTLSVerifyRegistry: "", // 🔥 EMPTY VALUE - This causes the REAL error!
		NodeBaseImage:         "docker.io/library/node:22-alpine",
		PythonBaseImage:       "docker.io/library/python:3.11-alpine",
		GoBaseImage:           "docker.io/library/golang:1.21-alpine",
		UseEKSPodIdentity:     true,
		PodIdentityRole:       "test-pod-identity-role",
	}

	jobManager := &JobManagerImpl{
		awsConfig:     emptyAWSConfig,
		storageConfig: createTestStorageConfig(),
	}

	buildRequest := createTestBuildRequest()
	destinationImageURI := "test-registry/test-image:latest"

	// Act
	args := jobManager.createKanikoArgs(buildRequest, destinationImageURI)

	// Log ALL arguments for debugging
	t.Logf("🔍 ALL Kaniko arguments with EMPTY registry values:")
	for i, arg := range args {
		t.Logf("  [%d] %s", i, arg)
	}

	// Assert - Check for empty registry mirror value (verify that empty values ARE generated when config is empty)
	registryMirrorArg := findArg(args, "--registry-mirror=")
	registryMirrorValue := extractArgValue(registryMirrorArg, "--registry-mirror=")
	if registryMirrorValue != "" {
		t.Fatalf("❌ Expected empty registry mirror value, got: '%s'", registryMirrorValue)
	}
	t.Logf("✅ DETECTED: Registry mirror value is EMPTY as expected (config has empty RegistryMirror)")
	t.Logf("   This correctly reproduces: 'kaniko error building image: strict validation requires the registry to be explicitly defined'")
	t.Logf("   AWS Config RegistryMirror: '%s'", emptyAWSConfig.RegistryMirror)
	t.Logf("   Generated argument: '%s'", registryMirrorArg)

	// Assert - Check for empty skip TLS verify registry value (verify that empty values ARE generated when config is empty)
	skipTLSArg := findArg(args, "--skip-tls-verify-registry=")
	skipTLSValue := extractArgValue(skipTLSArg, "--skip-tls-verify-registry=")
	if skipTLSValue != "" {
		t.Fatalf("❌ Expected empty skip TLS value, got: '%s'", skipTLSValue)
	}
	t.Logf("✅ DETECTED: Skip TLS verify registry value is EMPTY as expected (config has empty SkipTLSVerifyRegistry)")
	t.Logf("   This correctly reproduces: 'kaniko error building image: strict validation requires the registry to be explicitly defined'")
	t.Logf("   AWS Config SkipTLSVerifyRegistry: '%s'", emptyAWSConfig.SkipTLSVerifyRegistry)
	t.Logf("   Generated argument: '%s'", skipTLSArg)

	t.Logf("✅ Empty value detection test PASSED!")
}

// 🧪 testCreateKanikoArgsEnvironmentVariableSimulation - "Simulate environment variable loading"
func testCreateKanikoArgsEnvironmentVariableSimulation(t *testing.T) {
	// Arrange - Simulate environment variables being loaded
	envVars := map[string]string{
		"REGISTRY_MIRROR":          "docker.io",
		"SKIP_TLS_VERIFY_REGISTRY": "docker.io",
		"NODE_BASE_IMAGE":          "docker.io/library/node:22-alpine",
		"PYTHON_BASE_IMAGE":        "docker.io/library/python:3.11-alpine",
		"GO_BASE_IMAGE":            "docker.io/library/golang:1.21-alpine",
	}

	// Create AWS config that simulates environment variable loading
	simulatedAWSConfig := &config.AWSConfig{
		AWSRegion:             "us-west-2",
		AWSAccountID:          "339954290315",
		ECRRegistry:           "339954290315.dkr.ecr.us-west-2.amazonaws.com",
		ECRRepositoryName:     "knative-lambdas",
		S3SourceBucket:        "test-source-bucket",
		S3TempBucket:          "test-temp-bucket",
		RegistryMirror:        envVars["REGISTRY_MIRROR"],
		SkipTLSVerifyRegistry: envVars["SKIP_TLS_VERIFY_REGISTRY"],
		NodeBaseImage:         envVars["NODE_BASE_IMAGE"],
		PythonBaseImage:       envVars["PYTHON_BASE_IMAGE"],
		GoBaseImage:           envVars["GO_BASE_IMAGE"],
		UseEKSPodIdentity:     true,
		PodIdentityRole:       "test-pod-identity-role",
	}

	jobManager := &JobManagerImpl{
		awsConfig:     simulatedAWSConfig,
		storageConfig: createTestStorageConfig(),
	}

	buildRequest := createTestBuildRequest()
	destinationImageURI := "test-registry/test-image:latest"

	// Log environment variable simulation
	t.Logf("🔍 Environment Variable Simulation:")
	for key, value := range envVars {
		t.Logf("  %s = '%s'", key, value)
	}

	// Act
	args := jobManager.createKanikoArgs(buildRequest, destinationImageURI)

	// Log ALL arguments for debugging
	t.Logf("🔍 ALL Kaniko arguments from environment simulation:")
	for i, arg := range args {
		t.Logf("  [%d] %s", i, arg)
	}

	// Assert - Validate each argument matches environment variables
	expectedArgs := map[string]string{
		"--registry-mirror=":             envVars["REGISTRY_MIRROR"],
		"--skip-tls-verify-registry=":    envVars["SKIP_TLS_VERIFY_REGISTRY"],
		"--build-arg=NODE_BASE_IMAGE=":   envVars["NODE_BASE_IMAGE"],
		"--build-arg=PYTHON_BASE_IMAGE=": envVars["PYTHON_BASE_IMAGE"],
		"--build-arg=GO_BASE_IMAGE=":     envVars["GO_BASE_IMAGE"],
	}

	for expectedPrefix, expectedValue := range expectedArgs {
		arg := findArg(args, expectedPrefix)
		if arg == "" {
			t.Fatalf("❌ Required argument '%s' is missing!", expectedPrefix)
		}

		actualValue := extractArgValue(arg, expectedPrefix)
		if actualValue != expectedValue {
			t.Fatalf("❌ Argument '%s' doesn't match environment variable! Expected '%s', got '%s'", expectedPrefix, expectedValue, actualValue)
		}

		t.Logf("✅ Argument '%s' = '%s' ✓ (matches env var)", expectedPrefix, actualValue)
	}

	t.Logf("✅ Environment variable simulation test PASSED!")
}

// 🧪 testCreateKanikoContainerEnvironmentVariables - "Test that Kaniko container gets correct environment variables"
func testCreateKanikoContainerEnvironmentVariables(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	buildRequest := createTestBuildRequest()
	destinationImageURI := "test-registry/test-image:latest"
	resourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2000m"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2000m"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
	}

	// Act
	container := jobManager.createKanikoContainer(buildRequest, destinationImageURI, resourceRequirements)

	// Log container details
	t.Logf("🔍 Kaniko Container Details:")
	t.Logf("  Name: %s", container.Name)
	t.Logf("  Image: %s", container.Image)
	t.Logf("  Args count: %d", len(container.Args))
	t.Logf("  Env vars count: %d", len(container.Env))

	// Log ALL environment variables
	t.Logf("🔍 ALL Environment Variables:")
	for i, envVar := range container.Env {
		t.Logf("  [%d] %s = '%s'", i, envVar.Name, envVar.Value)
	}

	// Assert - Check for required environment variables
	requiredEnvVars := []string{
		"AWS_REGION",
		"REGISTRY_MIRROR",
		"SKIP_TLS_VERIFY_REGISTRY",
	}

	for _, requiredEnvVar := range requiredEnvVars {
		found := false
		for _, envVar := range container.Env {
			if envVar.Name == requiredEnvVar {
				found = true
				if envVar.Value == "" {
					t.Fatalf("❌ Environment variable '%s' is EMPTY!", requiredEnvVar)
				}
				t.Logf("✅ Environment variable '%s' = '%s' ✓", requiredEnvVar, envVar.Value)
				break
			}
		}
		if !found {
			t.Fatalf("❌ Required environment variable '%s' is missing!", requiredEnvVar)
		}
	}

	// Assert - Check for registry-related environment variables specifically
	registryMirrorEnv := findEnvVar(container.Env, "REGISTRY_MIRROR")
	if registryMirrorEnv == "" {
		t.Fatal("❌ REGISTRY_MIRROR environment variable is missing or empty!")
	}

	skipTLSEnv := findEnvVar(container.Env, "SKIP_TLS_VERIFY_REGISTRY")
	if skipTLSEnv == "" {
		t.Fatal("❌ SKIP_TLS_VERIFY_REGISTRY environment variable is missing or empty!")
	}

	t.Logf("✅ Registry environment variables validation PASSED!")
	t.Logf("✅ REGISTRY_MIRROR = '%s'", registryMirrorEnv)
	t.Logf("✅ SKIP_TLS_VERIFY_REGISTRY = '%s'", skipTLSEnv)
}

// 🧪 testCreateKanikoContainerEnvironmentVariablesMissing - "Test the REAL scenario where env vars are missing"
func testCreateKanikoContainerEnvironmentVariablesMissing(t *testing.T) {
	// Arrange - Create JobManager WITHOUT the registry environment variables
	// This simulates the REAL problem where the environment variables are not being passed
	missingEnvAWSConfig := &config.AWSConfig{
		AWSRegion:             "us-west-2",
		AWSAccountID:          "339954290315",
		ECRRegistry:           "339954290315.dkr.ecr.us-west-2.amazonaws.com",
		ECRRepositoryName:     "knative-lambdas",
		S3SourceBucket:        "test-source-bucket",
		S3TempBucket:          "test-temp-bucket",
		RegistryMirror:        "", // 🔥 MISSING - This is the REAL problem!
		SkipTLSVerifyRegistry: "", // 🔥 MISSING - This is the REAL problem!
		NodeBaseImage:         "docker.io/library/node:22-alpine",
		PythonBaseImage:       "docker.io/library/python:3.11-alpine",
		GoBaseImage:           "docker.io/library/golang:1.21-alpine",
		UseEKSPodIdentity:     true,
		PodIdentityRole:       "test-pod-identity-role",
	}

	jobManager := &JobManagerImpl{
		awsConfig:     missingEnvAWSConfig,
		buildConfig:   createTestBuildConfig(),
		config:        createTestK8sConfig(),
		storageConfig: createTestStorageConfig(),
	}

	buildRequest := createTestBuildRequest()
	destinationImageURI := "test-registry/test-image:latest"
	resourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2000m"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2000m"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
	}

	// Act
	container := jobManager.createKanikoContainer(buildRequest, destinationImageURI, resourceRequirements)

	// Log container details
	t.Logf("🔍 Kaniko Container with MISSING environment variables:")
	t.Logf("  Name: %s", container.Name)
	t.Logf("  Image: %s", container.Image)
	t.Logf("  Args count: %d", len(container.Args))
	t.Logf("  Env vars count: %d", len(container.Env))

	// Log ALL environment variables
	t.Logf("🔍 ALL Environment Variables:")
	for i, envVar := range container.Env {
		t.Logf("  [%d] %s = '%s'", i, envVar.Name, envVar.Value)
	}

	// Log ALL arguments
	t.Logf("🔍 ALL Kaniko Arguments:")
	for i, arg := range container.Args {
		t.Logf("  [%d] %s", i, arg)
	}

	// Assert - Check for missing registry environment variables (verify that empty values ARE set when config is empty)
	registryMirrorEnv := findEnvVar(container.Env, "REGISTRY_MIRROR")
	if registryMirrorEnv != "" {
		t.Fatalf("❌ Expected empty REGISTRY_MIRROR environment variable, got: '%s'", registryMirrorEnv)
	}
	t.Logf("✅ REGISTRY_MIRROR environment variable is EMPTY as expected (config has empty RegistryMirror)")
	t.Logf("   This correctly reproduces: 'kaniko error building image: strict validation requires the registry to be explicitly defined'")
	t.Logf("   AWS Config RegistryMirror: '%s'", missingEnvAWSConfig.RegistryMirror)

	skipTLSEnv := findEnvVar(container.Env, "SKIP_TLS_VERIFY_REGISTRY")
	if skipTLSEnv != "" {
		t.Fatalf("❌ Expected empty SKIP_TLS_VERIFY_REGISTRY environment variable, got: '%s'", skipTLSEnv)
	}
	t.Logf("✅ SKIP_TLS_VERIFY_REGISTRY environment variable is EMPTY as expected (config has empty SkipTLSVerifyRegistry)")
	t.Logf("   This correctly reproduces: 'kaniko error building image: strict validation requires the registry to be explicitly defined'")
	t.Logf("   AWS Config SkipTLSVerifyRegistry: '%s'", missingEnvAWSConfig.SkipTLSVerifyRegistry)

	// Check the actual Kaniko arguments (verify that empty values ARE generated when config is empty)
	registryMirrorArg := findArg(container.Args, "--registry-mirror=")
	registryMirrorValue := extractArgValue(registryMirrorArg, "--registry-mirror=")
	if registryMirrorValue != "" {
		t.Fatalf("❌ Expected empty --registry-mirror argument value, got: '%s'", registryMirrorValue)
	}
	t.Logf("✅ --registry-mirror argument value is EMPTY as expected")
	t.Logf("   Generated argument: '%s'", registryMirrorArg)

	skipTLSArg := findArg(container.Args, "--skip-tls-verify-registry=")
	skipTLSValue := extractArgValue(skipTLSArg, "--skip-tls-verify-registry=")
	if skipTLSValue != "" {
		t.Fatalf("❌ Expected empty --skip-tls-verify-registry argument value, got: '%s'", skipTLSValue)
	}
	t.Logf("✅ --skip-tls-verify-registry argument value is EMPTY as expected")
	t.Logf("   Generated argument: '%s'", skipTLSArg)

	t.Logf("✅ Missing environment variables test PASSED!")
}

// 🧪 testCreateKanikoArgsSkipTLSVerifyRegistryNotEmpty - "Test that skip TLS verify registry argument is not empty"
func testCreateKanikoArgsSkipTLSVerifyRegistryNotEmpty(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	buildRequest := createTestBuildRequest()
	destinationImageURI := "test-registry/test-image:latest"

	// Log the AWS config to verify values
	t.Logf("🔍 AWS Config RegistryMirror: '%s'", jobManager.awsConfig.RegistryMirror)
	t.Logf("🔍 AWS Config SkipTLSVerifyRegistry: '%s'", jobManager.awsConfig.SkipTLSVerifyRegistry)

	// Act
	args := jobManager.createKanikoArgs(buildRequest, destinationImageURI)

	// Log ALL arguments for debugging
	t.Logf("🔍 ALL Kaniko arguments:")
	for i, arg := range args {
		t.Logf("  [%d] %s", i, arg)
	}

	// Assert - Skip TLS Verify Registry
	skipTLSArg := findArg(args, "--skip-tls-verify-registry=")
	if skipTLSArg == "" {
		t.Fatal("❌ Skip TLS verify registry argument is missing!")
	}

	// Extract the value after the equals sign
	skipTLSValue := extractArgValue(skipTLSArg, "--skip-tls-verify-registry=")
	if skipTLSValue == "" {
		t.Fatal("❌ Skip TLS verify registry value is empty! Expected 'docker.io'")
	}

	if skipTLSValue != "docker.io" {
		t.Fatalf("❌ Skip TLS verify registry value is incorrect! Expected 'docker.io', got '%s'", skipTLSValue)
	}

	t.Logf("✅ Skip TLS verify registry argument found: %s", skipTLSArg)
	t.Logf("✅ Skip TLS verify registry value: %s", skipTLSValue)
	t.Logf("✅ Skip TLS verify registry validation PASSED!")
}

// 🧪 testCreateKanikoArgsBaseImagesCorrect - "Test that base image arguments are correct"
func testCreateKanikoArgsBaseImagesCorrect(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	buildRequest := createTestBuildRequest()
	destinationImageURI := "test-registry/test-image:latest"

	// Act
	args := jobManager.createKanikoArgs(buildRequest, destinationImageURI)

	// Assert
	expectedNodeImage := "docker.io/library/node:22-alpine"
	expectedPythonImage := "docker.io/library/python:3.11-alpine"
	expectedGoImage := "docker.io/library/golang:1.21-alpine"

	// Check Node base image
	nodeArg := findArg(args, "--build-arg=NODE_BASE_IMAGE=")
	if nodeArg == "" {
		t.Fatal("❌ NODE_BASE_IMAGE argument is missing!")
	}
	nodeValue := extractArgValue(nodeArg, "--build-arg=NODE_BASE_IMAGE=")
	if nodeValue != expectedNodeImage {
		t.Fatalf("❌ NODE_BASE_IMAGE value is incorrect! Expected '%s', got '%s'", expectedNodeImage, nodeValue)
	}

	// Check Python base image
	pythonArg := findArg(args, "--build-arg=PYTHON_BASE_IMAGE=")
	if pythonArg == "" {
		t.Fatal("❌ PYTHON_BASE_IMAGE argument is missing!")
	}
	pythonValue := extractArgValue(pythonArg, "--build-arg=PYTHON_BASE_IMAGE=")
	if pythonValue != expectedPythonImage {
		t.Fatalf("❌ PYTHON_BASE_IMAGE value is incorrect! Expected '%s', got '%s'", expectedPythonImage, pythonValue)
	}

	// Check Go base image
	goArg := findArg(args, "--build-arg=GO_BASE_IMAGE=")
	if goArg == "" {
		t.Fatal("❌ GO_BASE_IMAGE argument is missing!")
	}
	goValue := extractArgValue(goArg, "--build-arg=GO_BASE_IMAGE=")
	if goValue != expectedGoImage {
		t.Fatalf("❌ GO_BASE_IMAGE value is incorrect! Expected '%s', got '%s'", expectedGoImage, goValue)
	}

	t.Logf("✅ All base image arguments are correct")
}

// 🧪 testCreateKanikoArgsAllArgumentsPresent - "Test that all required Kaniko arguments are present"
func testCreateKanikoArgsAllArgumentsPresent(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	buildRequest := createTestBuildRequest()
	destinationImageURI := "test-registry/test-image:latest"

	// Act
	args := jobManager.createKanikoArgs(buildRequest, destinationImageURI)

	// Assert
	requiredArgs := []string{
		"--context=",
		"--destination=",
		"--dockerfile=",
		"--registry-mirror=",
		"--skip-tls-verify-registry=",
		"--build-arg=NODE_BASE_IMAGE=",
		"--build-arg=PYTHON_BASE_IMAGE=",
		"--build-arg=GO_BASE_IMAGE=",
	}

	for _, requiredArg := range requiredArgs {
		if findArg(args, requiredArg) == "" {
			t.Fatalf("❌ Required argument '%s' is missing!", requiredArg)
		}
	}

	t.Logf("✅ All required Kaniko arguments are present")
	t.Logf("✅ Total arguments: %d", len(args))
}

// 🧪 testNewJobManagerValidConfigSuccess - "Test creating JobManager with valid config"
func testNewJobManagerValidConfigSuccess(t *testing.T) {
	// Arrange
	config := JobManagerConfig{
		K8sClient:       fake.NewSimpleClientset(),
		K8sConfig:       createTestK8sConfig(),
		BuildConfig:     createTestBuildConfig(),
		AWSConfig:       createTestAWSConfig(),
		RateLimitConfig: createTestRateLimitConfig(),
		Observability:   createTestObservability(),
		RateLimiter:     createTestRateLimiter(),
	}

	// Act
	jobManager, err := NewJobManager(config)

	// Assert
	if err != nil {
		t.Fatalf("❌ Expected no error, got: %v", err)
	}
	if jobManager == nil {
		t.Fatal("❌ Expected JobManager instance, got nil")
	}

	t.Logf("✅ JobManager created successfully")
}

// 🧪 testNewJobManagerNilK8sClientError - "Test creating JobManager with nil K8s client"
func testNewJobManagerNilK8sClientError(t *testing.T) {
	// Arrange
	config := JobManagerConfig{
		K8sClient:       nil, // This should cause an error
		K8sConfig:       createTestK8sConfig(),
		BuildConfig:     createTestBuildConfig(),
		AWSConfig:       createTestAWSConfig(),
		RateLimitConfig: createTestRateLimitConfig(),
		Observability:   createTestObservability(),
		RateLimiter:     createTestRateLimiter(),
	}

	// Act
	jobManager, err := NewJobManager(config)

	// Assert
	if err == nil {
		t.Fatal("❌ Expected error for nil K8s client, got nil")
	}
	if jobManager != nil {
		t.Fatal("❌ Expected nil JobManager, got instance")
	}

	t.Logf("✅ Correctly rejected nil K8s client: %v", err)
}

// 🧪 testNewJobManagerNilK8sConfigError - "Test creating JobManager with nil K8s config"
func testNewJobManagerNilK8sConfigError(t *testing.T) {
	// Arrange
	config := JobManagerConfig{
		K8sClient:       fake.NewSimpleClientset(),
		K8sConfig:       nil, // This should cause an error
		BuildConfig:     createTestBuildConfig(),
		AWSConfig:       createTestAWSConfig(),
		RateLimitConfig: createTestRateLimitConfig(),
		Observability:   createTestObservability(),
		RateLimiter:     createTestRateLimiter(),
	}

	// Act
	jobManager, err := NewJobManager(config)

	// Assert
	if err == nil {
		t.Fatal("❌ Expected error for nil K8s config, got nil")
	}
	if jobManager != nil {
		t.Fatal("❌ Expected nil JobManager, got instance")
	}

	t.Logf("✅ Correctly rejected nil K8s config: %v", err)
}

// 🧪 testNewJobManagerNilBuildConfigError - "Test creating JobManager with nil build config"
func testNewJobManagerNilBuildConfigError(t *testing.T) {
	// Arrange
	config := JobManagerConfig{
		K8sClient:       fake.NewSimpleClientset(),
		K8sConfig:       createTestK8sConfig(),
		BuildConfig:     nil, // This should cause an error
		AWSConfig:       createTestAWSConfig(),
		RateLimitConfig: createTestRateLimitConfig(),
		Observability:   createTestObservability(),
		RateLimiter:     createTestRateLimiter(),
	}

	// Act
	jobManager, err := NewJobManager(config)

	// Assert
	if err == nil {
		t.Fatal("❌ Expected error for nil build config, got nil")
	}
	if jobManager != nil {
		t.Fatal("❌ Expected nil JobManager, got instance")
	}

	t.Logf("✅ Correctly rejected nil build config: %v", err)
}

// 🧪 testNewJobManagerNilObservabilityError - "Test creating JobManager with nil observability"
func testNewJobManagerNilObservabilityError(t *testing.T) {
	// Arrange
	config := JobManagerConfig{
		K8sClient:       fake.NewSimpleClientset(),
		K8sConfig:       createTestK8sConfig(),
		BuildConfig:     createTestBuildConfig(),
		AWSConfig:       createTestAWSConfig(),
		RateLimitConfig: createTestRateLimitConfig(),
		Observability:   nil, // This should cause an error
		RateLimiter:     createTestRateLimiter(),
	}

	// Act
	jobManager, err := NewJobManager(config)

	// Assert
	if err == nil {
		t.Fatal("❌ Expected error for nil observability, got nil")
	}
	if jobManager != nil {
		t.Fatal("❌ Expected nil JobManager, got instance")
	}

	t.Logf("✅ Correctly rejected nil observability: %v", err)
}

// 🧪 testGenerateJobNameValidInputsSuccess - "Test generating job name with valid inputs"
func testGenerateJobNameValidInputsSuccess(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	thirdPartyID := "test-third-party"
	parserID := "test-parser"

	// Act
	jobName := jobManager.GenerateJobName(thirdPartyID, parserID)

	// Assert
	if jobName == "" {
		t.Fatal("❌ Generated job name is empty")
	}

	expectedPrefix := "knative-lambda-builder-"
	if len(jobName) <= len(expectedPrefix) {
		t.Fatalf("❌ Job name too short: %s", jobName)
	}

	t.Logf("✅ Generated job name: %s", jobName)
}

// 🧪 testIsJobRunningActiveJobTrue - "Test IsJobRunning with active job"
func testIsJobRunningActiveJobTrue(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	job := &batchv1.Job{
		Status: batchv1.JobStatus{
			Active: 1,
		},
	}

	// Act
	isRunning := jobManager.IsJobRunning(job)

	// Assert
	if !isRunning {
		t.Fatal("❌ Expected job to be running")
	}

	t.Logf("✅ Correctly identified running job")
}

// 🧪 testIsJobRunningNoActiveJobFalse - "Test IsJobRunning with no active job"
func testIsJobRunningNoActiveJobFalse(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	job := &batchv1.Job{
		Status: batchv1.JobStatus{
			Active: 0,
		},
	}

	// Act
	isRunning := jobManager.IsJobRunning(job)

	// Assert
	if isRunning {
		t.Fatal("❌ Expected job to not be running")
	}

	t.Logf("✅ Correctly identified non-running job")
}

// 🧪 testIsJobFailedFailedJobTrue - "Test IsJobFailed with failed job"
func testIsJobFailedFailedJobTrue(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	job := &batchv1.Job{
		Status: batchv1.JobStatus{
			Failed:    1,
			Succeeded: 0,
		},
	}

	// Act
	isFailed := jobManager.IsJobFailed(job)

	// Assert
	if !isFailed {
		t.Fatal("❌ Expected job to be failed")
	}

	t.Logf("✅ Correctly identified failed job")
}

// 🧪 testIsJobFailedSucceededJobFalse - "Test IsJobFailed with succeeded job"
func testIsJobFailedSucceededJobFalse(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	job := &batchv1.Job{
		Status: batchv1.JobStatus{
			Failed:    0,
			Succeeded: 1,
		},
	}

	// Act
	isFailed := jobManager.IsJobFailed(job)

	// Assert
	if isFailed {
		t.Fatal("❌ Expected job to not be failed")
	}

	t.Logf("✅ Correctly identified non-failed job")
}

// 🧪 testIsJobSucceededSucceededJobTrue - "Test IsJobSucceeded with succeeded job"
func testIsJobSucceededSucceededJobTrue(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	job := &batchv1.Job{
		Status: batchv1.JobStatus{
			Succeeded: 1,
		},
	}

	// Act
	isSucceeded := jobManager.IsJobSucceeded(job)

	// Assert
	if !isSucceeded {
		t.Fatal("❌ Expected job to be succeeded")
	}

	t.Logf("✅ Correctly identified succeeded job")
}

// 🧪 testIsJobSucceededFailedJobFalse - "Test IsJobSucceeded with failed job"
func testIsJobSucceededFailedJobFalse(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	job := &batchv1.Job{
		Status: batchv1.JobStatus{
			Succeeded: 0,
		},
	}

	// Act
	isSucceeded := jobManager.IsJobSucceeded(job)

	// Assert
	if isSucceeded {
		t.Fatal("❌ Expected job to not be succeeded")
	}

	t.Logf("✅ Correctly identified non-succeeded job")
}

// 🧪 testGenerateImageURISuccess - "Test generating image URI with valid inputs"
func testGenerateImageURISuccess(t *testing.T) {
	// Arrange
	jobManager := createTestJobManager(t)
	thirdPartyID := "test-third-party"
	parserID := "test-parser"
	contentHash := "abcdef1234567890abcdef1234567890"

	// Act
	imageURI := jobManager.generateImageURI(thirdPartyID, parserID, contentHash)

	// Assert
	if imageURI == "" {
		t.Fatal("❌ Generated image URI is empty")
	}

	expectedPrefix := "339954290315.dkr.ecr.us-west-2.amazonaws.com/knative-lambdas:"
	if len(imageURI) <= len(expectedPrefix) {
		t.Fatalf("❌ Image URI too short: %s", imageURI)
	}

	t.Logf("✅ Generated image URI: %s", imageURI)
}

// 🧪 testGenerateImageURINilAWSConfigFallback - "Test generating image URI with nil AWS config"
func testGenerateImageURINilAWSConfigFallback(t *testing.T) {
	// Arrange
	jobManager := &JobManagerImpl{
		awsConfig: nil, // Nil AWS config
	}
	thirdPartyID := "test-third-party"
	parserID := "test-parser"
	contentHash := "abcdef1234567890abcdef1234567890"

	// Act
	imageURI := jobManager.generateImageURI(thirdPartyID, parserID, contentHash)

	// Assert
	if imageURI == "" {
		t.Fatal("❌ Generated image URI is empty")
	}

	expectedPrefix := "unknown-registry/"
	if len(imageURI) <= len(expectedPrefix) {
		t.Fatalf("❌ Image URI too short: %s", imageURI)
	}

	t.Logf("✅ Generated fallback image URI: %s", imageURI)
}

// 🔧 Helper functions

// createTestJobManager creates a test JobManager instance
func createTestJobManager(t *testing.T) *JobManagerImpl {
	config := JobManagerConfig{
		K8sClient:       fake.NewSimpleClientset(),
		K8sConfig:       createTestK8sConfig(),
		BuildConfig:     createTestBuildConfig(),
		AWSConfig:       createTestAWSConfig(),
		StorageConfig:   createTestStorageConfig(),
		RateLimitConfig: createTestRateLimitConfig(),
		Observability:   createTestObservability(),
		RateLimiter:     createTestRateLimiter(),
	}

	jobManager, err := NewJobManager(config)
	if err != nil {
		t.Fatalf("Failed to create test JobManager: %v", err)
	}

	return jobManager.(*JobManagerImpl)
}

// createTestK8sConfig creates a test Kubernetes config
func createTestK8sConfig() *config.KubernetesConfig {
	return &config.KubernetesConfig{
		Namespace:                "test-namespace",
		ServiceAccount:           "test-service-account",
		JobTTLSeconds:            3600,
		RunAsUser:                1000,
		InCluster:                true,
		JobDeletionWaitTimeout:   30 * time.Second,
		JobDeletionCheckInterval: 5 * time.Second,
	}
}

// createTestBuildConfig creates a test build config
func createTestBuildConfig() *config.BuildConfig {
	return &config.BuildConfig{
		KanikoImage:   "gcr.io/kaniko-project/executor:v1.19.2",
		SidecarImage:  "test-sidecar:latest",
		BuildTimeout:  30 * time.Minute,
		CPULimit:      "2000m",
		MemoryLimit:   "2Gi",
		MaxParserSize: 104857600,
	}
}

// createTestAWSConfig creates a test AWS config
func createTestAWSConfig() *config.AWSConfig {
	return &config.AWSConfig{
		AWSRegion:             "us-west-2",
		AWSAccountID:          "339954290315",
		ECRRegistry:           "339954290315.dkr.ecr.us-west-2.amazonaws.com",
		ECRRepositoryName:     "knative-lambdas",
		S3SourceBucket:        "test-source-bucket",
		S3TempBucket:          "test-temp-bucket",
		RegistryMirror:        "docker.io", // 🔧 Set to docker.io
		SkipTLSVerifyRegistry: "docker.io", // 🔧 Set to docker.io
		NodeBaseImage:         "docker.io/library/node:22-alpine",
		PythonBaseImage:       "docker.io/library/python:3.11-alpine",
		GoBaseImage:           "docker.io/library/golang:1.21-alpine",
		UseEKSPodIdentity:     true,
		PodIdentityRole:       "test-pod-identity-role",
	}
}

// createTestRateLimitConfig creates a test rate limit config
func createTestRateLimitConfig() *config.RateLimitingConfig {
	return &config.RateLimitingConfig{
		Enabled:                    true,
		BuildContextRequestsPerMin: 5,
		BuildContextBurstSize:      2,
		K8sJobRequestsPerMin:       10,
		K8sJobBurstSize:            3,
		ClientRequestsPerMin:       5,
		ClientBurstSize:            2,
		S3UploadRequestsPerMin:     50,
		S3UploadBurstSize:          10,
		MaxMemoryUsagePercent:      80.0,
		MemoryCheckInterval:        30 * time.Second,
		CleanupInterval:            5 * time.Minute,
		ClientTTL:                  1 * time.Hour,
		MaxConcurrentBuilds:        10,
		MaxConcurrentJobs:          5,
		BuildTimeout:               30 * time.Minute,
		JobTimeout:                 1 * time.Hour,
		RequestTimeout:             5 * time.Minute,
	}
}

// createTestObservability creates a test observability instance
func createTestObservability() *observability.Observability {
	config := observability.Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		LogLevel:       "info",
		MetricsEnabled: false,
		TracingEnabled: false,
		OTLPEndpoint:   "",
		SampleRate:     1.0,
	}

	obs, err := observability.New(config)
	if err != nil {
		// Return a minimal observability instance for testing
		return &observability.Observability{}
	}

	return obs
}

// createTestRateLimiter creates a test rate limiter
func createTestRateLimiter() *resilience.MultiLevelRateLimiter {
	return &resilience.MultiLevelRateLimiter{}
}

// createTestStorageConfig creates a test storage config
func createTestStorageConfig() *config.StorageConfig {
	return &config.StorageConfig{
		Provider: "s3",
		S3: config.S3Config{
			SourceBucket: "test-source-bucket",
			TempBucket:   "test-temp-bucket",
		},
	}
}

// createTestBuildRequest creates a test build request
func createTestBuildRequest() *builds.BuildRequest {
	return &builds.BuildRequest{
		ThirdPartyID:  "test-third-party",
		ParserID:      "test-parser",
		ContentHash:   "abcdef1234567890abcdef1234567890",
		CorrelationID: "test-correlation-id",
		BuildType:     "nodejs",
		Runtime:       "nodejs22",
		SourceBucket:  "test-source-bucket",
		SourceKey:     "test/source/key",
	}
}

// findArg finds an argument that starts with the given prefix
func findArg(args []string, prefix string) string {
	for _, arg := range args {
		if len(arg) >= len(prefix) && arg[:len(prefix)] == prefix {
			return arg
		}
	}
	return ""
}

// extractArgValue extracts the value after the equals sign in an argument
func extractArgValue(arg, prefix string) string {
	if len(arg) <= len(prefix) {
		return ""
	}
	return arg[len(prefix):]
}

// findEnvVar finds an environment variable by name and returns its value
func findEnvVar(envVars []corev1.EnvVar, name string) string {
	for _, envVar := range envVars {
		if envVar.Name == name {
			return envVar.Value
		}
	}
	return ""
}
