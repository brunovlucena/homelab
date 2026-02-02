package proxies

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type SolanaProxy struct {
	RpcUrl string
}

func NewSolanaProxy(rpcUrl string) *SolanaProxy {
	return &SolanaProxy{RpcUrl: rpcUrl}
}

func (s *SolanaProxy) GetAccountInfo(ctx context.Context, account string) (map[string]interface{}, error) {
	// Example: POST to Solana RPC endpoint for getAccountInfo
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getAccountInfo",
		"params":  []interface{}{account, map[string]interface{}{"encoding": "base64"}},
	}
	return s.DoRpcCall(ctx, payload)
}

func (s *SolanaProxy) GetBalance(ctx context.Context, account string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getBalance",
		"params":  []interface{}{account},
	}
	return s.DoRpcCall(ctx, payload)
}

func (s *SolanaProxy) GetSlot(ctx context.Context) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getSlot",
		"params":  []interface{}{},
	}
	return s.DoRpcCall(ctx, payload)
}

func (s *SolanaProxy) GetProgramAccounts(ctx context.Context, pubkey string, encoding string, withContext bool, filterObject interface{}) (map[string]interface{}, error) {
	params := []interface{}{pubkey}
	options := map[string]interface{}{}
	if encoding != "" {
		options["encoding"] = encoding
	}
	if withContext {
		options["withContext"] = withContext
	}
	if filterObject != nil {
		options["filters"] = filterObject
	}
	if len(options) > 0 {
		params = append(params, options)
	}
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getProgramAccounts",
		"params":  params,
	}
	return s.DoRpcCall(ctx, payload)
}

func (s *SolanaProxy) GetTokenAccountBalance(ctx context.Context, pubkey string) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTokenAccountBalance",
		"params":  []interface{}{pubkey},
	}
	return s.DoRpcCall(ctx, payload)
}

func (s *SolanaProxy) GetMultipleAccounts(ctx context.Context, pubkeys []string) (map[string]interface{}, error) {
	params := []interface{}{pubkeys}
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getMultipleAccounts",
		"params":  params,
	}
	return s.DoRpcCall(ctx, payload)
}

func (s *SolanaProxy) GetTokenAccountsByOwner(ctx context.Context, pubkey string, mint string, programId string) (map[string]interface{}, error) {
	params := []interface{}{pubkey}
	options := map[string]interface{}{}
	if mint != "" {
		options["mint"] = mint
	}
	if programId != "" {
		options["programId"] = programId
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
	return s.DoRpcCall(ctx, payload)
}

func (s *SolanaProxy) GetTransaction(ctx context.Context, signature string, encoding string) (map[string]interface{}, error) {
	params := []interface{}{signature}
	if encoding != "" {
		params = append(params, map[string]interface{}{"encoding": encoding})
	}
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTransaction",
		"params":  params,
	}
	return s.DoRpcCall(ctx, payload)
}

func (s *SolanaProxy) DoRpcCall(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, "POST", s.RpcUrl, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
