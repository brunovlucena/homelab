package vault

import (
	"context"
	"fmt"
	"os"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

type Client struct {
	client *vaultapi.Client
}

// NewClient creates a new Vault client
func NewClient(addr, token string) (*Client, error) {
	config := vaultapi.DefaultConfig()
	config.Address = addr
	config.Timeout = 30 * time.Second

	client, err := vaultapi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	if token != "" {
		client.SetToken(token)
	} else {
		// Try Kubernetes auth if no token provided
		if err := authenticateK8s(client); err != nil {
			return nil, fmt.Errorf("failed to authenticate with Vault: %w", err)
		}
	}

	return &Client{client: client}, nil
}

// authenticateK8s attempts Kubernetes authentication
func authenticateK8s(client *vaultapi.Client) error {
	// Read service account token
	tokenPath := "/var/run/secrets/kubernetes.io/serviceaccount/token"
	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		return fmt.Errorf("failed to read service account token: %w", err)
	}

	// Kubernetes auth
	authPath := "auth/kubernetes/login"
	data := map[string]interface{}{
		"role": "vaultwarden",
		"jwt":  string(tokenBytes),
	}

	resp, err := client.Logical().Write(authPath, data)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	if resp.Auth == nil || resp.Auth.ClientToken == "" {
		return fmt.Errorf("no token returned from Vault")
	}

	client.SetToken(resp.Auth.ClientToken)
	return nil
}

// StoreCipher stores an encrypted cipher (password entry) in Vault
func (c *Client) StoreCipher(ctx context.Context, userID, cipherID string, cipherData map[string]interface{}) error {
	path := fmt.Sprintf("secret/data/vaultwarden/users/%s/ciphers/%s", userID, cipherID)

	_, err := c.client.Logical().WriteWithContext(ctx, path, map[string]interface{}{
		"data": cipherData,
	})

	return err
}

// GetCipher retrieves an encrypted cipher from Vault
func (c *Client) GetCipher(ctx context.Context, userID, cipherID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("secret/data/vaultwarden/users/%s/ciphers/%s", userID, cipherID)

	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, err
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("cipher not found")
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid cipher data format")
	}

	return data, nil
}

// ListCiphers lists all ciphers for a user
func (c *Client) ListCiphers(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	path := fmt.Sprintf("secret/metadata/vaultwarden/users/%s/ciphers", userID)

	secret, err := c.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, err
	}

	if secret == nil || secret.Data == nil {
		return []map[string]interface{}{}, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []map[string]interface{}{}, nil
	}

	var ciphers []map[string]interface{}
	for _, key := range keys {
		cipherID := key.(string)
		cipher, err := c.GetCipher(ctx, userID, cipherID)
		if err != nil {
			continue
		}
		cipher["Id"] = cipherID
		ciphers = append(ciphers, cipher)
	}

	return ciphers, nil
}

// DeleteCipher deletes a cipher from Vault
func (c *Client) DeleteCipher(ctx context.Context, userID, cipherID string) error {
	path := fmt.Sprintf("secret/data/vaultwarden/users/%s/ciphers/%s", userID, cipherID)

	_, err := c.client.Logical().DeleteWithContext(ctx, path)
	return err
}

// StoreUser stores user metadata in Vault
func (c *Client) StoreUser(ctx context.Context, userID string, userData map[string]interface{}) error {
	path := fmt.Sprintf("secret/data/vaultwarden/users/%s/profile", userID)

	_, err := c.client.Logical().WriteWithContext(ctx, path, map[string]interface{}{
		"data": userData,
	})

	return err
}

// GetUser retrieves user metadata from Vault
func (c *Client) GetUser(ctx context.Context, userID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("secret/data/vaultwarden/users/%s/profile", userID)

	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, err
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("user not found")
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid user data format")
	}

	return data, nil
}

// StoreUserAuth stores user authentication data (password hash)
func (c *Client) StoreUserAuth(ctx context.Context, email string, authData map[string]interface{}) error {
	path := fmt.Sprintf("secret/data/vaultwarden/auth/%s", email)

	_, err := c.client.Logical().WriteWithContext(ctx, path, map[string]interface{}{
		"data": authData,
	})

	return err
}

// GetUserAuth retrieves user authentication data
func (c *Client) GetUserAuth(ctx context.Context, email string) (map[string]interface{}, error) {
	path := fmt.Sprintf("secret/data/vaultwarden/auth/%s", email)

	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, err
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("user not found")
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid auth data format")
	}

	return data, nil
}
