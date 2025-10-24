// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🏭 STORAGE FACTORY - Factory for creating storage clients
//
//	🎯 Purpose: Create storage clients based on provider configuration
//	💡 Features: Dynamic provider selection, validation, configuration
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package storage

import (
	"context"
	"fmt"

	"knative-lambda-new/internal/config"
	"knative-lambda-new/internal/errors"
	"knative-lambda-new/internal/observability"
)

// 🏭 StorageFactoryImpl - "Factory implementation for storage clients"
type StorageFactoryImpl struct {
	storageConfig *config.StorageConfig
	obs           *observability.Observability
}

// 🏗️ NewStorageFactory - "Create new storage factory"
func NewStorageFactory(storageConfig *config.StorageConfig, obs *observability.Observability) (*StorageFactoryImpl, error) {
	if storageConfig == nil {
		return nil, errors.NewConfigurationError("storage_factory", "storage_config", "storage config cannot be nil")
	}

	if obs == nil {
		return nil, errors.NewConfigurationError("storage_factory", "observability", "observability cannot be nil")
	}

	return &StorageFactoryImpl{
		storageConfig: storageConfig,
		obs:           obs,
	}, nil
}

// 🏭 CreateStorage - "Create storage client based on provider"
func (f *StorageFactoryImpl) CreateStorage(ctx context.Context, provider StorageProvider) (ObjectStorage, error) {
	ctx, span := f.obs.StartSpan(ctx, "create_storage")
	defer span.End()

	f.obs.Info(ctx, "Creating storage client",
		"provider", string(provider))

	switch provider {
	case ProviderS3:
		return f.createS3Storage(ctx)
	case ProviderMinIO:
		return f.createMinIOStorage(ctx)
	default:
		return nil, errors.NewConfigurationError("storage_factory", "provider", fmt.Sprintf("unsupported storage provider: %s", provider))
	}
}

// ☁️ createS3Storage - "Create S3 storage client"
func (f *StorageFactoryImpl) createS3Storage(ctx context.Context) (*S3Storage, error) {
	s3Config := S3StorageConfig{
		Region:        f.storageConfig.S3.Region,
		Endpoint:      f.storageConfig.S3.Endpoint,
		Observability: f.obs,
	}

	return NewS3Storage(ctx, s3Config)
}

// 🏠 createMinIOStorage - "Create MinIO storage client"
func (f *StorageFactoryImpl) createMinIOStorage(ctx context.Context) (*MinIOStorage, error) {
	minioConfig := MinIOStorageConfig{
		Endpoint:      f.storageConfig.MinIO.Endpoint,
		AccessKey:     f.storageConfig.MinIO.AccessKey,
		SecretKey:     f.storageConfig.MinIO.SecretKey,
		UseSSL:        f.storageConfig.MinIO.UseSSL,
		Region:        f.storageConfig.MinIO.Region,
		Observability: f.obs,
	}

	return NewMinIOStorage(ctx, minioConfig)
}

// 🔧 GetDefaultStorage - "Get storage client for the configured default provider"
func (f *StorageFactoryImpl) GetDefaultStorage(ctx context.Context) (ObjectStorage, error) {
	provider := StorageProvider(f.storageConfig.Provider)
	return f.CreateStorage(ctx, provider)
}
