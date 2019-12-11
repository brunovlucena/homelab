package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/sirupsen/logrus"
)

// The most convenient way to work with JSONB coming from a database would be in
// the form of a map[string]interface{}, not in the form of a JSON object and most
// certainly not as bytes.
// Luckely, the Go standard library has 2 built-in interfaces we can implement to
// create our own database compatible type: sql.Scanner & driver.Valuer
type Config struct {
	Data DataMap `db:"data" json:"data"`
}

// DataMap represents the dynamic payload.
type DataMap map[string]interface{}

// To satisfy this interface, we must implement the Value method, which must
// transform our type to a database driver compatible type. In our case, we’ll
// marshall the map to JSONB data (= []byte):
func (p DataMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd": "value",
		}).Error(err.Error())
		return nil, err
	}
	return j, nil
}

// To use the interfacec interface, sql.Scanner, we need to implement Scan method.
// This method must take the raw data that comes from the database
// and transform it to our new type. In our case, the database will return JSONB
// ([]byte) that we must transform to our type (the reverse of what we did with
// driver.Valuer)
func (p *DataMap) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	var i interface{}
	err := json.Unmarshal(source, &i)
	if err != nil {
		return err
	}

	*p, ok = i.(map[string]interface{})
	if !ok {
		return errors.New("type assertion .(map[string]interface{}) failed")
	}

	return nil
}
