package data

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// The most convenient way to work with JSONB coming from a database would be in
// the form of a map[string]interface{}, not in the form of a JSON object and most
// certainly not as bytes.
// Luckely, the Go standard library has 2 built-in interfaces we can implement to
// create our own database compatible type: sql.Scanner & driver.Valuer

type Config struct {
	Data DataMap `db:"data"`
}

type DataMap map[string]interface{}

// To satisfy this interface, we must implement the Value method, which must
// transform our type to a database driver compatible type. In our case, we’ll
// marshall the map to JSONB data (= []byte):
func (p DataMap) Value() (driver.Value, error) {
	j, err := json.Marshal(p)
	return j, err
}

func (p *DataMap) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	var i interface{}
	err := json.Unmarshal(source, &i)
	if err != nil {
		return err
	}

	*p, ok = i.(map[string]interface{})
	if !ok {
		return errors.New("Type assertion .(map[string]interface{}) failed.")
	}

	return nil
}
