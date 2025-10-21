package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAWSConfig_GetRegion(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("NAMESPACE", "test-namespace")
	os.Setenv("AWS_REGION", "us-west-2")
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	os.Setenv("ECR_REGISTRY", "123456789012.dkr.ecr.us-west-2.amazonaws.com")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_ACCOUNT_ID")
		os.Unsetenv("ECR_REGISTRY")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "us-west-2", config.AWS.GetRegion())
}

func TestAWSConfig_GetS3SourceBucket(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("NAMESPACE", "test-namespace")
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	os.Setenv("ECR_REGISTRY", "123456789012.dkr.ecr.us-west-2.amazonaws.com")
	os.Setenv("S3_SOURCE_BUCKET", "test-source-bucket")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
		os.Unsetenv("AWS_ACCOUNT_ID")
		os.Unsetenv("ECR_REGISTRY")
		os.Unsetenv("S3_SOURCE_BUCKET")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "test-source-bucket", config.AWS.GetS3SourceBucket())
}

func TestAWSConfig_GetS3TempBucket(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("NAMESPACE", "test-namespace")
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	os.Setenv("ECR_REGISTRY", "123456789012.dkr.ecr.us-west-2.amazonaws.com")
	os.Setenv("S3_TMP_BUCKET", "test-tmp-bucket")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
		os.Unsetenv("AWS_ACCOUNT_ID")
		os.Unsetenv("ECR_REGISTRY")
		os.Unsetenv("S3_TMP_BUCKET")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "test-tmp-bucket", config.AWS.GetS3TempBucket())
}

func TestAWSConfig_GetECRRegistry(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("NAMESPACE", "test-namespace")
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	os.Setenv("ECR_REGISTRY", "123456789012.dkr.ecr.us-west-2.amazonaws.com")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
		os.Unsetenv("AWS_ACCOUNT_ID")
		os.Unsetenv("ECR_REGISTRY")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "123456789012.dkr.ecr.us-west-2.amazonaws.com", config.AWS.GetECRRegistry())
}

func TestAWSConfig_GetECRRepositoryName(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("NAMESPACE", "test-namespace")
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	os.Setenv("ECR_REGISTRY", "123456789012.dkr.ecr.us-west-2.amazonaws.com")
	os.Setenv("ECR_REPOSITORY", "test-repo")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
		os.Unsetenv("AWS_ACCOUNT_ID")
		os.Unsetenv("ECR_REGISTRY")
		os.Unsetenv("ECR_REPOSITORY")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "test-repo", config.AWS.GetECRRepositoryName())
}

func TestAWSConfig_GetAccountID(t *testing.T) {
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("NAMESPACE", "test-namespace")
	os.Setenv("AWS_ACCOUNT_ID", "123456789012")
	os.Setenv("ECR_REGISTRY", "123456789012.dkr.ecr.us-west-2.amazonaws.com")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("NAMESPACE")
		os.Unsetenv("AWS_ACCOUNT_ID")
		os.Unsetenv("ECR_REGISTRY")
	}()

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "123456789012", config.AWS.GetAccountID())
}
