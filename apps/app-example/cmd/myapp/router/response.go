package router

import (
	"net/http"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/data"
	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/utils"
)

type ConfigResponse struct {
	Config   *data.Config `json:"config"`
	ServedBy string       `json:"served_by"`
}

func (rd *ConfigResponse) Render(w http.ResponseWriter, r *http.Request) error {
	rd.ServedBy = utils.GetIP() // Pod's IP
	return nil
}

// NewConfigResponse is the response payload for the Config data model.
func NewConfigResponse(config *data.Config) *ConfigResponse {
	return &ConfigResponse{Config: config}
}
