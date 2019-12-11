package router

import (
	"net/http"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/data"
	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/utils"
	"github.com/go-chi/render"
)

// ConfigResponse represents a Response.
type ConfigResponse struct {
	Config   *data.Config `json:"config"`
	ServedBy string       `json:"served_by"`
}

// ConfigResponse implements Render Interface.
func (rd *ConfigResponse) Render(w http.ResponseWriter, r *http.Request) error {
	rd.ServedBy = utils.GetIP() // Pod's IP
	return nil
}

// NewConfigResponse is the response payload for the Config data model.
func NewConfigResponse(config *data.Config) *ConfigResponse {
	return &ConfigResponse{Config: config}
}

// NewConfigListResponse is the response payload with a list of Config data model.
func NewConfigListResponse(configs []*data.Config) []render.Renderer {
	list := []render.Renderer{}
	if len(configs) == 0 {
		errNotFound := &ErrResponse{HTTPStatusCode: 404, StatusText: "Resources not found."}
		list = append(list, errNotFound)
		return list
	}
	for _, config := range configs {
		list = append(list, NewConfigResponse(config))
	}
	return list
}
