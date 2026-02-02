package grpcServices

import (
	"context"
	"encoding/json"
	"notifinetwork/localfusion/proto/blockchain_manager"
	"notifinetwork/localfusion/services/proxies"
	"os"
)

type FusionSolanaRpcService struct {
	blockchain_manager.UnimplementedFusionSolanaRpcServer
	solanaProxy *proxies.SolanaProxy
}

func NewFusionSolanaRpcService() *FusionSolanaRpcService {
	rpcUrl := os.Getenv("SOLANA_RPC_URL")
	if rpcUrl == "" {
		rpcUrl = "https://api.mainnet-beta.solana.com" // fallback default
	}
	return &FusionSolanaRpcService{
		solanaProxy: proxies.NewSolanaProxy(rpcUrl),
	}
}

func (s *FusionSolanaRpcService) GetSolanaBalance(ctx context.Context, req *blockchain_manager.BcmGetSolanaBalanceRequest) (*blockchain_manager.BcmGetSolanaBalanceResponse, error) {
	result, err := s.solanaProxy.GetBalance(ctx, req.GetPubkey())
	if err != nil {
		return nil, err
	}
	// Marshal the result to JSON string for the response
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &blockchain_manager.BcmGetSolanaBalanceResponse{
		BalanceResponse: string(jsonBytes),
	}, nil
}

func (s *FusionSolanaRpcService) GetSolanaAccountInfo(ctx context.Context, req *blockchain_manager.BcmGetSolanaAccountInfoRequest) (*blockchain_manager.BcmGetSolanaAccountInfoResponse, error) {
	result, err := s.solanaProxy.GetAccountInfo(ctx, req.GetPubkey())
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &blockchain_manager.BcmGetSolanaAccountInfoResponse{
		AccountInfoResponse: string(jsonBytes),
	}, nil
}

func (s *FusionSolanaRpcService) GetSolanaSlot(ctx context.Context, req *blockchain_manager.BcmGetSolanaSlotRequest) (*blockchain_manager.BcmGetSolanaSlotResponse, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getSlot",
		"params":  []interface{}{},
	}
	result, err := s.solanaProxy.DoRpcCall(ctx, payload)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &blockchain_manager.BcmGetSolanaSlotResponse{
		CurrentSlotResponse: string(jsonBytes),
	}, nil
}

func (s *FusionSolanaRpcService) GetSolanaProgramAccounts(ctx context.Context, req *blockchain_manager.BcmGetSolanaProgramAccountsRequest) (*blockchain_manager.BcmGetSolanaProgramAccountsResponse, error) {
	params := []interface{}{req.GetPubkey()}
	if req.GetEncoding() != "" || req.GetFilterObject() != nil || req.GetWithContext() {
		options := map[string]interface{}{}
		if req.GetEncoding() != "" {
			options["encoding"] = req.GetEncoding()
		}
		if req.GetWithContext() {
			options["withContext"] = req.GetWithContext()
		}
		if req.GetFilterObject() != nil {
			options["filters"] = req.GetFilterObject() // This may need conversion to map
		}
		params = append(params, options)
	}
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getProgramAccounts",
		"params":  params,
	}
	result, err := s.solanaProxy.DoRpcCall(ctx, payload)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &blockchain_manager.BcmGetSolanaProgramAccountsResponse{
		ProgramAccountsResponse: string(jsonBytes),
	}, nil
}

func (s *FusionSolanaRpcService) GetSolanaTokenAccountBalance(ctx context.Context, req *blockchain_manager.BcmGetSolanaTokenAccountBalanceRequest) (*blockchain_manager.BcmGetSolanaTokenAccountBalanceResponse, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTokenAccountBalance",
		"params":  []interface{}{req.GetPubkey()},
	}
	result, err := s.solanaProxy.DoRpcCall(ctx, payload)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &blockchain_manager.BcmGetSolanaTokenAccountBalanceResponse{
		BalanceResponse: string(jsonBytes),
	}, nil
}

func (s *FusionSolanaRpcService) GetSolanaMultipleAccounts(ctx context.Context, req *blockchain_manager.BcmGetSolanaMultipleAccountsRequest) (*blockchain_manager.BcmGetSolanaMultipleAccountsResponse, error) {
	params := []interface{}{req.GetPubkeys()}
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getMultipleAccounts",
		"params":  params,
	}
	result, err := s.solanaProxy.DoRpcCall(ctx, payload)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &blockchain_manager.BcmGetSolanaMultipleAccountsResponse{
		AccountsResponse: string(jsonBytes),
	}, nil
}

func (s *FusionSolanaRpcService) GetSolanaTokenAccountsByOwner(ctx context.Context, req *blockchain_manager.BcmGetSolanaTokenAccountsByOwnerRequest) (*blockchain_manager.BcmGetSolanaTokenAccountsByOwnerResponse, error) {
	params := []interface{}{req.GetPubkey()}
	options := map[string]interface{}{}
	if req.GetMint() != "" {
		options["mint"] = req.GetMint()
	}
	if req.GetProgramId() != "" {
		options["programId"] = req.GetProgramId()
	}
	if len(options) > 0 {
		params = append(params, options)
	}
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTokenAccountsByOwner",
		"params":  params,
	}
	result, err := s.solanaProxy.DoRpcCall(ctx, payload)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &blockchain_manager.BcmGetSolanaTokenAccountsByOwnerResponse{
		TokenAccountsResponse: string(jsonBytes),
	}, nil
}

func (s *FusionSolanaRpcService) GetSolanaTransaction(ctx context.Context, req *blockchain_manager.BcmGetSolanaTransactionRequest) (*blockchain_manager.BcmGetSolanaTransactionResponse, error) {
	result, err := s.solanaProxy.GetTransaction(ctx, req.GetTransactionSignature(), req.GetEncoding())
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &blockchain_manager.BcmGetSolanaTransactionResponse{
		TransactionResponse: string(jsonBytes),
	}, nil
}
