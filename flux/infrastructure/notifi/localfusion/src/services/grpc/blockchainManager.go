package grpcServices

import (
	"context"
	"encoding/json"
	"log"
	proto "notifinetwork/localfusion/proto/blockchain_manager"
	"notifinetwork/localfusion/services/proxies"
)

type BlockchainManagerService struct {
	proto.UnimplementedBlockchainManagerServer
	evmProxy *proxies.EvmProxy
}

func NewBlockchainManagerService() *BlockchainManagerService {

	evmProxy := &proxies.EvmProxy{}
	evmProxy.SetupClients()

	return &BlockchainManagerService{
		evmProxy: evmProxy,
	}
}

func (s *BlockchainManagerService) GetAccountBalance(ctx context.Context, req *proto.GetAccountBalanceRequest) (*proto.GetAccountBalanceResponse, error) {
	balanceHex, err := s.evmProxy.GetAccountBalance(req.BlockchainType, req.AccountAddress, req.BlockId)
	if err != nil {
		log.Printf("Error getting account balance: %v", err)
		return nil, err
	}

	return &proto.GetAccountBalanceResponse{
		Value: "0x" + balanceHex,
	}, nil
}

func (s *BlockchainManagerService) RunEthCall(ctx context.Context, req *proto.EthCallRequest) (*proto.EthCallResponse, error) {
	log.Println("Received RunEthCall", "req", req)

	if req.FromAddress == nil {
		req.FromAddress = &req.ToAddress
	}

	result, err := s.evmProxy.MakeEthCall(req.BlockchainType, *req.FromAddress, req.ToAddress, req.Data, req.BlockNumberOrTag)

	if err != nil {
		log.Printf("Error making eth call: %v", err)
		return nil, err
	}

	response := &EthCallJson{
		Result:  "0x" + result,
		Id:      1,
		JsonRpc: "2.0",
	}

	log.Println("Returning RunEthCall", "response", response)

	//Marshal the response into JSON string and then to bytes
	responseBytes, err := json.Marshal(response)

	return &proto.EthCallResponse{
		Data: responseBytes,
	}, nil
}
