package proxies

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"notifinetwork/localfusion/config"
	commonProto "notifinetwork/localfusion/proto/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EvmProxy struct {
	clients map[commonProto.BlockchainType]*ethclient.Client
}

type ClientNotFoundError struct {
	blockchainType commonProto.BlockchainType
}

func (e *ClientNotFoundError) Error() string {
	return "Client not found for blockchain type " + e.blockchainType.String()
}

func parseCustomChainConfigs() []config.ChainConfig {
	// Assuming CustomChainConfigs is a JSON array of ChainConfig
	var chainConfigs []config.ChainConfig
	err := json.Unmarshal([]byte(config.Config.CUSTOM_CHAIN_CONFIGS), &chainConfigs)
	if err != nil {
		log.Fatalf("Error parsing CUSTOM_CHAIN_CONFIGS json: %v", err)
	}

	// Now chainConfigs contains your configurations
	for _, config := range chainConfigs {
		fmt.Printf("Loaded custom config for Blockchain Type: %v, RPC URL: %s\n", config.BlockchainType, config.RpcUrl)
	}

	return chainConfigs
}

func hexStringToBigInt(hexString string) *big.Int {
	bigInt := big.Int{}
	// Remove the 0x prefix
	withoutPrefix := hexString[2:]
	bigInt.SetString(withoutPrefix, 16)
	return &bigInt
}

func (s *EvmProxy) SetupClients() {
	s.clients = make(map[commonProto.BlockchainType]*ethclient.Client)
	clientConfigurations := []config.ChainConfig{
		{
			BlockchainType: commonProto.BlockchainType_BLOCKCHAIN_TYPE_ETHEREUM,
			RpcUrl:         config.Config.ETHEREUM_RPC_URL,
		},
		{
			BlockchainType: commonProto.BlockchainType_BLOCKCHAIN_TYPE_POLYGON,
			RpcUrl:         config.Config.POLYGON_RPC_URL,
		},
		{
			BlockchainType: commonProto.BlockchainType_BLOCKCHAIN_TYPE_BINANCE,
			RpcUrl:         config.Config.BNB_RPC_URL,
		},
		{
			BlockchainType: commonProto.BlockchainType_BLOCKCHAIN_TYPE_AVALANCHE,
			RpcUrl:         config.Config.AVALANCHE_RPC_URL,
		},
		{
			BlockchainType: commonProto.BlockchainType_BLOCKCHAIN_TYPE_OPTIMISM,
			RpcUrl:         config.Config.OPTIMISM_RPC_URL,
		},
		{
			BlockchainType: commonProto.BlockchainType_BLOCKCHAIN_TYPE_ARBITRUM,
			RpcUrl:         config.Config.ARBITRUM_RPC_URL,
		},
	}

	// Load custom chain configs
	for _, cfg := range parseCustomChainConfigs() {
		clientConfigurations = append(clientConfigurations, config.ChainConfig{
			BlockchainType: cfg.BlockchainType,
			RpcUrl:         cfg.RpcUrl,
		})
	}

	for _, cfg := range clientConfigurations {
		client, err := ethclient.Dial(cfg.RpcUrl)
		if err != nil {
			log.Println("Failed to connect to blockchain, if the RPC is down please try configuring a different URL", "blockchainType", cfg.BlockchainType, "rpcUrl", cfg.RpcUrl, "error", err)
			continue
		}
		s.clients[cfg.BlockchainType] = client
	}
}

func (s *EvmProxy) GetAccountBalance(blockchainType commonProto.BlockchainType, accountAddress string, blockNumberHex string) (string, error) {
	var blockNumber *big.Int = nil
	if blockNumberHex != "" && blockNumberHex != "latest" {
		blockNumber = hexStringToBigInt(blockNumberHex)
	}

	log.Println("Getting account balance", "blockchainType", blockchainType, "accountAddress", accountAddress, "blockNumber", blockNumber)
	
	// Mock mode - return fake data to avoid rate limiting
	if config.Config.MOCK_MODE {
		log.Println("MOCK_MODE enabled - returning fake balance", "blockchainType", blockchainType, "accountAddress", accountAddress)
		// Return a consistent fake balance: 1000000000000000000 (1 ETH in wei)
		return "1000000000000000000", nil
	}
	
	client, ok := s.clients[blockchainType]
	if !ok {
		log.Println("Failed to retrieve client for blockchain", "blockchainType", blockchainType)
		return "", &ClientNotFoundError{}
	}

	balance, err := client.BalanceAt(context.Background(), common.HexToAddress(accountAddress), blockNumber)

	if err != nil {
		log.Println("Failed to retrieve account balance", "blockchainType", blockchainType, "accountAddress", accountAddress, "error", err)
		return "", err
	}

	return balance.String(), nil
}

func (s *EvmProxy) MakeEthCall(blockchainType commonProto.BlockchainType, fromAddress string, toAddress string, data *string, blockNumberHex string) (string, error) {
	var blockNumber *big.Int = nil
	if blockNumberHex != "" && blockNumberHex != "latest" {
		blockNumber = hexStringToBigInt(blockNumberHex)
	}
	log.Println("Making eth call", "blockchainType", blockchainType, "fromAddress", fromAddress, "toAddress", toAddress, "data", data, "blockNumber", blockNumber)
	
	// Mock mode - return fake data to avoid rate limiting
	if config.Config.MOCK_MODE {
		log.Println("MOCK_MODE enabled - returning fake eth call result", "blockchainType", blockchainType, "toAddress", toAddress)
		// Return a fake hex result
		return "0x0000000000000000000000000000000000000000000000000000000000000001", nil
	}
	
	client, ok := s.clients[blockchainType]

	if !ok {
		log.Println("Failed to retrieve client for blockchain", "blockchainType", blockchainType)
		return "", &ClientNotFoundError{}
	}

	toAddressHex := common.HexToAddress(toAddress)

	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		From:     common.HexToAddress(fromAddress),
		To:       &toAddressHex,
		Gas:      0,
		GasPrice: nil,
		Value:    nil,
		Data:     common.FromHex(*data),
	}, blockNumber)
	if err != nil {
		log.Println("Failed to make eth call", "blockchainType", blockchainType, "fromAddress", fromAddress, "toAddress", toAddress, "data", data, "blockNumber", blockNumber, "error", err)
		return "", err
	}

	resultHex := common.Bytes2Hex(result)

	return resultHex, nil
}

func NewEvmProxy() *EvmProxy {
	return &EvmProxy{
		clients: make(map[commonProto.BlockchainType]*ethclient.Client),
	}
}
