package aws

import (
	"context"
	"testing"

	"knative-lambda-new/internal/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		region  string
		wantErr bool
	}{
		{
			name:    "valid region",
			region:  "us-west-2",
			wantErr: false,
		},
		{
			name:    "empty region",
			region:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cfg := ClientConfig{
				Region:        tt.region,
				Observability: &observability.Observability{},
			}

			client, err := NewClient(ctx, cfg)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, client)
			assert.Equal(t, tt.region, client.region)
		})
	}
}

func TestClient_GetImageURI(t *testing.T) {
	ctx := context.Background()
	cfg := ClientConfig{
		Region:            "us-west-2",
		ECRRegistry:       "123456789012.dkr.ecr.us-west-2.amazonaws.com",
		ECRRepositoryName: "test-repo",
		Observability:     &observability.Observability{},
	}

	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)

	imageURI := client.GetImageURI("test-third-party", "test-parser")
	expected := "123456789012.dkr.ecr.us-west-2.amazonaws.com/test-repo:test-third-party-test-parser"
	assert.Equal(t, expected, imageURI)
}
