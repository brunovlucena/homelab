// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//
//	🔗 NOTIFI SERVICE CONFIGURATION - Notifi service addresses for lambda functions
//
//	🎯 Purpose: Configuration for Notifi service addresses that lambda functions need to connect to
//	💡 Features: Service addresses, gRPC configuration, connection settings
//
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package config

import (
	"strconv"

	"knative-lambda-new/internal/constants"
	"knative-lambda-new/internal/errors"
)

// 🔗 NotifiConfig - "Notifi service addresses configuration"
type NotifiConfig struct {
	// Service Addresses
	SubscriptionManagerAddress string `envconfig:"SUBSCRIPTION_MANAGER_ADDRESS" default:"notifi-subscription-manager.notifi.svc.cluster.local:4000"`
	EphemeralStorageAddress    string `envconfig:"EPHEMERAL_STORAGE_ADDRESS" default:"notifi-storage-manager.notifi.svc.cluster.local:4000"`
	PersistentStorageAddress   string `envconfig:"PERSISTENT_STORAGE_ADDRESS" default:"notifi-storage-manager.notifi.svc.cluster.local:4000"`
	FusionFetchProxyAddress    string `envconfig:"FUSION_FETCH_PROXY_ADDRESS" default:"notifi-fetch-proxy.notifi.svc.cluster.local:4000"`
	EvmRpcAddress              string `envconfig:"EVM_RPC_ADDRESS" default:"notifi-blockchain-manager.notifi.svc.cluster.local:4000"`
	SolanaRpcAddress           string `envconfig:"SOLANA_RPC_ADDRESS" default:"notifi-blockchain-manager.notifi.svc.cluster.local:4000"`
	SuiRpcAddress              string `envconfig:"SUI_RPC_ADDRESS" default:"notifi-blockchain-manager.notifi.svc.cluster.local:4000"`

	// gRPC Configuration
	GrpcInsecure bool `envconfig:"GRPC_INSECURE" default:"true"`
}

// 🔧 NewNotifiConfig - "Create Notifi configuration with defaults"
func NewNotifiConfig() *NotifiConfig {
	return &NotifiConfig{
		SubscriptionManagerAddress: constants.NotifiSubscriptionManagerAddressDefault,
		EphemeralStorageAddress:    constants.NotifiEphemeralStorageAddressDefault,
		PersistentStorageAddress:   constants.NotifiPersistentStorageAddressDefault,
		FusionFetchProxyAddress:    constants.NotifiFusionFetchProxyAddressDefault,
		EvmRpcAddress:              constants.NotifiEvmRpcAddressDefault,
		SolanaRpcAddress:           constants.NotifiSolanaRpcAddressDefault,
		SuiRpcAddress:              constants.NotifiSuiRpcAddressDefault,
		GrpcInsecure:               constants.NotifiGrpcInsecureDefault,
	}
}

// 🔧 Validate - "Validate Notifi configuration"
func (c *NotifiConfig) Validate() error {
	// Validate service addresses are not empty
	if c.SubscriptionManagerAddress == "" {
		return errors.NewValidationError("subscription_manager_address", c.SubscriptionManagerAddress, "subscription manager address is required")
	}

	if c.EphemeralStorageAddress == "" {
		return errors.NewValidationError("ephemeral_storage_address", c.EphemeralStorageAddress, "ephemeral storage address is required")
	}

	if c.PersistentStorageAddress == "" {
		return errors.NewValidationError("persistent_storage_address", c.PersistentStorageAddress, "persistent storage address is required")
	}

	if c.FusionFetchProxyAddress == "" {
		return errors.NewValidationError("fusion_fetch_proxy_address", c.FusionFetchProxyAddress, "fusion fetch proxy address is required")
	}

	if c.EvmRpcAddress == "" {
		return errors.NewValidationError("evm_rpc_address", c.EvmRpcAddress, "EVM RPC address is required")
	}

	if c.SolanaRpcAddress == "" {
		return errors.NewValidationError("solana_rpc_address", c.SolanaRpcAddress, "Solana RPC address is required")
	}

	if c.SuiRpcAddress == "" {
		return errors.NewValidationError("sui_rpc_address", c.SuiRpcAddress, "Sui RPC address is required")
	}

	return nil
}

// 🔧 GetSubscriptionManagerAddress - "Get subscription manager address"
func (c *NotifiConfig) GetSubscriptionManagerAddress() string {
	return c.SubscriptionManagerAddress
}

// 🔧 GetEphemeralStorageAddress - "Get ephemeral storage address"
func (c *NotifiConfig) GetEphemeralStorageAddress() string {
	return c.EphemeralStorageAddress
}

// 🔧 GetPersistentStorageAddress - "Get persistent storage address"
func (c *NotifiConfig) GetPersistentStorageAddress() string {
	return c.PersistentStorageAddress
}

// 🔧 GetFusionFetchProxyAddress - "Get fusion fetch proxy address"
func (c *NotifiConfig) GetFusionFetchProxyAddress() string {
	return c.FusionFetchProxyAddress
}

// 🔧 GetEvmRpcAddress - "Get EVM RPC address"
func (c *NotifiConfig) GetEvmRpcAddress() string {
	return c.EvmRpcAddress
}

// 🔧 GetSolanaRpcAddress - "Get Solana RPC address"
func (c *NotifiConfig) GetSolanaRpcAddress() string {
	return c.SolanaRpcAddress
}

// 🔧 GetSuiRpcAddress - "Get Sui RPC address"
func (c *NotifiConfig) GetSuiRpcAddress() string {
	return c.SuiRpcAddress
}

// 🔧 GetGrpcInsecure - "Get gRPC insecure setting"
func (c *NotifiConfig) GetGrpcInsecure() string {
	return strconv.FormatBool(c.GrpcInsecure)
}
