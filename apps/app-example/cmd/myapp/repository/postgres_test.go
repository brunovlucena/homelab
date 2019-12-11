package repository

import (
	"testing"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/data"
	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/utils"
	"github.com/stretchr/testify/assert"
)

var (
	configs []map[string]interface{}
	rep     Repository
)

func init() {
	// load json
	utils.LoadJson("postgres_test.json", &configs)
	// initialize repository
	dType := "postgres"
	dHost := "0.0.0.0" // or postgres.storage
	dPort := "5432"
	dUser := "postgres"
	dPass := "postgres"
	dbName := "myapp"
	var err error
	rep, err = NewRepository(dType, dHost, dPort, dUser, dPass, dbName)
	utils.LogErr(err)
	utils.LogPrint("Successfully Loaded tests")
}

func TestCreate(t *testing.T) {
	for _, c := range configs {
		// create configs
		_, err := rep.Create(&data.Config{Data: c})
		utils.LogErr(err)
	}
}

func TestUpdate(t *testing.T) {
	// new changed metadata
	newData := map[string]interface{}{
		"name": "pod-2p",
		"metadata": map[string]interface{}{
			"monitoring": map[string]interface{}{
				"enabled": "true",
			},
		},
	}
	// update pod-2
	_, err := rep.Update(&data.Config{Data: newData})
	utils.LogErr(err)
	// TestFind is gonna fail if TestUpdate fails.
}

func TestFind(t *testing.T) {
	// find pod-2
	config, err := rep.Find("pod-2p")
	utils.LogErr(err)
	// compare metadata
	data := config.Data
	metadata := data["metadata"].(map[string]interface{})
	monitoring := metadata["monitoring"].(map[string]interface{})
	enabled := monitoring["enabled"].(string)
	// it becomes true because TestUpdate
	assert.Equal(t, "true", enabled)
}

func TestRemove(t *testing.T) {
	// remove pod-13-idonotexist
	_, err := rep.Remove("pod-13-idonotexit")
	utils.LogErr(err)
	if err != nil {
		assert.Equal(t, "sql: no rows in result set", err.Error())
	}
}

func TestFindAll(t *testing.T) {
	// find pod-11
	configs, err := rep.FindAll()
	utils.LogErr(err)
	// compare metadata
	config := configs[0]
	data := config.Data
	metadata := data["metadata"].(map[string]interface{})
	monitoring := metadata["monitoring"].(map[string]interface{})
	enabled := monitoring["enabled"].(bool)
	// it becomes true because TestUpdate
	assert.Equal(t, false, enabled)
}
