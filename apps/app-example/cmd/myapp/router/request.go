package router

import (
	"encoding/json"
	"errors"
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
		logrus.WithFields(logrus.Fields{
			"cmd": "bind",
		}).Error(err.Error())
		return err
	}
	// validate json
	if data == nil { // edge case where json payload is null
		err := errors.New("Null is not valid!")
		logrus.WithFields(logrus.Fields{
			"cmd": "bind",
		}).Error(err.Error())
		return err
	}
	cr.Config.Data = data
	return nil
}
