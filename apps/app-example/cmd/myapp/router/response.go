package router

import (
	"net/http"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/data"
)

type ConfigResponse struct {
	Config *data.Config
}

func (rd *ConfigResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// NewConfigResponse is the response payload for the Config data model.
func NewConfigResponse(config *data.Config) *ConfigResponse {
	return &ConfigResponse{Config: config}
}
