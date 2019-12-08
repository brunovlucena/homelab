package repository

import (
	"encoding/json"
	"io/ioutil"
	"os"
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
	// Open our jsonFile
	jsonFile, err := os.Open("postgres_test.json")
	// if we os.Open returns an error then handle it
	utils.LogErr(err)
	utils.LogPrint("Successfully Opened postgres_test.json")
	// defer the closing of our jsonFile
	defer jsonFile.Close()
	// read our opened json
	byteValue, _ := ioutil.ReadAll(jsonFile)
	// we unmarshal our byteArray
	json.Unmarshal(byteValue, &configs)
	// initialize repository
	dType := "postgres"
	dHost := "0.0.0.0"
	dPort := "5432"
	dUser := "postgres"
	dPass := "postgres"
	dbName := "myapp"
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
		"name": "datacenter-2",
		"metadata": map[string]interface{}{
			"monitoring": map[string]interface{}{
				"enabled": "true",
			},
		},
	}
	// update datacenter-2
	_, err := rep.Update(&data.Config{Data: newData})
	utils.LogErr(err)
	// TestFind is gonna fail if TestUpdate fails.
}

func TestFind(t *testing.T) {
	// find datacenter-2
	config, err := rep.Find("datacenter-2")
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
	_, err := rep.Remove("datacenter-3")
	utils.LogErr(err)
	assert.Equal(t, "sql: no rows in result set", err.Error())
}
