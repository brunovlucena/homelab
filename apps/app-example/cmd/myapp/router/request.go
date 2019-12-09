package router

import (
	"encoding/json"
	"net/http"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/data"
	"github.com/sirupsen/logrus"
)

type ConfigRequest struct {
	Config data.Config
}

func (cr *ConfigRequest) Bind(r *http.Request) error {
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		logrus.Error(err)
		return err
	}
	cr.Config.Data = data
	return nil
}
