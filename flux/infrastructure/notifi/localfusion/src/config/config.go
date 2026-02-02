package config

import (
	commonProto "notifinetwork/localfusion/proto/types"

	"github.com/kelseyhightower/envconfig"
)

var Config AppConfig

type ChainConfig struct {
	BlockchainType commonProto.BlockchainType `json:"blockchainType"`
	RpcUrl         string                     `json:"rpcUrl"`
}
type AppConfig struct {
	EPHEMERAL_STORAGE_PORT    int `default:"50052" envconfig:"EPHEMERAL_STORAGE_PORT"`
	PERSISTENT_STORAGE_PORT   int `default:"50053" envconfig:"PERSISTENT_STORAGE_PORT"`
	EVM_RPC_PORT              int `default:"50054" envconfig:"EVM_RPC_PORT"`
	SOLANA_RPC_PORT           int `default:"50055" envconfig:"SOLANA_RPC_PORT"`
	SUBSCRIPTION_MANAGER_PORT int `default:"50056" envconfig:"SUBSCRIPTION_MANAGER_PORT"`
	SUI_RPC_PORT              int `default:"50055" envconfig:"SUI_RPC_PORT"`

	// FusionSuiPort int
	// FusionCosmosPort int
	// FusionSolPort int

	ETHEREUM_RPC_URL  string `default:"https://ethereum.publicnode.com" envconfig:"ETHEREUM_RPC_URL"`
	POLYGON_RPC_URL   string `default:"https://polygon-bor.publicnode.com" envconfig:"POLYGON_RPC_URL"`
	OPTIMISM_RPC_URL  string `default:"https://optimism.publicnode.com" envconfig:"OPTIMISM_RPC_URL"`
	ARBITRUM_RPC_URL  string `default:"https://arbitrum.publicnode.com" envconfig:"ARBITRUM_RPC_URL"`
	BASE_RPC_URL      string `default:"https://base.publicnode.com" envconfig:"BASE_RPC_URL"`
	AVALANCHE_RPC_URL string `default:"https://avalanche.publicnode.com" envconfig:"AVALANCHE_RPC_URL"`
	BNB_RPC_URL       string `default:"https://bsc.publicnode.com" envconfig:"BNB_RPC_URL"`

	CUSTOM_CHAIN_CONFIGS string `default:"[]" envconfig:"CUSTOM_CHAIN_CONFIGS"`
	
	// Mock mode - if true, returns fake data instead of calling real RPC endpoints
	MOCK_MODE bool `default:"false" envconfig:"MOCK_MODE"`
}

func LoadFromEnv() {
	err := envconfig.Process("", &Config)
	if err != nil {
		panic(err)
	}
}
