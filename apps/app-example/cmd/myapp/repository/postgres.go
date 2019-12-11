package repository

import (
	"database/sql"
	"fmt"
	"net/url"

	// import driver for database/sql
	_ "github.com/lib/pq"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/data"
	"github.com/sirupsen/logrus"
)

// Represents a Connection.
type Postgres struct {
	host     string
	port     string
	user     string
	password string
	dbname   string
	dbconn   *sql.DB
}

// Represents a Item in the Database.
type Item struct {
	ID     int
	Config data.Config
}

// NewPostgres creates a connection with the database.
func NewPostgres(host, port, user, password, dbname string) *Postgres {
	dbconn := connect(host, port, user, password, dbname)
	return &Postgres{host, port, user, password, dbname, dbconn}
}

// Connect stablishes a connection with the database
func connect(host, port, user, password, dbname string) *sql.DB {
	// info: sslmode - disabled
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// opens connection
	db, err := sql.Open("postgres", psqlInfo)

	// setup for highter performance
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(20)

	// error checking
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":      "connect",
			"host":     host,
			"port":     port,
			"database": dbname,
		}).Error("Cannot connect!")
	} else {
		// SQL.Open only creates the DB object, but dies not open
		//any connections to the database.
		err = db.Ping()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"cmd":      "connect",
				"host":     host,
				"port":     port,
				"database": dbname,
			}).Error("Failed to ping database!")
		}
	}

	// Success
	logrus.WithFields(logrus.Fields{
		"cmd":             "connect",
		"host":            host,
		"port":            port,
		"database":        dbname,
		"max_connections": db.Stats().MaxOpenConnections,
	}).Infoln("Successfully connected to Postgres!")

	return db
}

// Create creates a Record in the database.
func (p *Postgres) Create(config *data.Config) (*data.Config, error) {
	sqlStatement := `INSERT INTO configs (data) VALUES ($1) RETURNING id`
	var id int
	dataMap := config.Data

	// query
	row := p.dbconn.QueryRow(sqlStatement, dataMap)
	err := row.Scan(&id)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd:":             "Create",
			"config_name":      config.Data["name"],
			"open_connections": p.dbconn.Stats().OpenConnections,
			"database":         p.dbname,
		}).Error(err.Error())
		return nil, err
	}

	// log
	logrus.WithFields(logrus.Fields{
		"cmd":      "Create",
		"database": p.dbname,
	}).Infoln("New record ID is", id)

	return config, nil
}

// Find finds a Record in the database.
func (p *Postgres) Find(name string) (*data.Config, error) {
	sqlStatement := `SELECT data FROM configs WHERE data->>'name' = $1;`
	row := p.dbconn.QueryRow(sqlStatement, name)
	config := new(data.Config)

	var err error
	switch err = row.Scan(&config.Data); err {
	case sql.ErrNoRows:
		logrus.WithFields(logrus.Fields{
			"cmd:":             "Find",
			"config_name":      name,
			"database":         p.dbname,
			"open_connections": p.dbconn.Stats().OpenConnections,
		}).Warn("No rows were returned!")
	case nil:
		logrus.WithFields(logrus.Fields{
			"cmd":      "Find",
			"database": p.dbname,
		}).Infoln("Record founded!")
	}

	return config, err
}

// FindAll returns all Records in the database.
func (p *Postgres) FindAll() ([]*data.Config, error) {
	// return value
	var configs []*data.Config

	// query
	sqlStatement := `SELECT data FROM configs;`
	rows, err := p.dbconn.Query(sqlStatement)
	defer rows.Close()

	// check for errors
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":              "FindAll",
			"database":         p.dbname,
			"open_connections": p.dbconn.Stats().OpenConnections,
		}).Error(err.Error())
		return nil, err
	}

	// Loop through the data
	for rows.Next() {
		var m data.DataMap
		err := rows.Scan(&m)
		// check for errors
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"cmd":              "FindAll",
				"database":         p.dbname,
				"open_connections": p.dbconn.Stats().OpenConnections,
			}).Error(err.Error())
			return nil, err
		}
		// append results
		configs = append(configs, &data.Config{Data: m})
	}

	// log
	logrus.WithFields(logrus.Fields{
		"cmd":      "FindAll",
		"database": p.dbname,
	}).Infoln("Record(s) founded!")

	return configs, err
}

// Update updates a Record in the database.
func (p *Postgres) Update(config *data.Config) (*data.Config, error) {
	sqlStatement := `UPDATE configs SET data = $1 WHERE data->>'name' = $2;`
	dataMap := config.Data

	// query
	_, err := p.dbconn.Exec(sqlStatement, dataMap, dataMap["name"])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":              "Update",
			"database":         p.dbname,
			"open_connections": p.dbconn.Stats().OpenConnections,
		}).Error(err.Error())
		return nil, err
	}

	// log
	logrus.WithFields(logrus.Fields{
		"cmd":      "Update",
		"database": p.dbname,
	}).Infoln("Record updated!")

	return config, nil
}

// Remove removes a Record from database.
func (p *Postgres) Remove(name string) (*data.Config, error) {
	sqlStatement := `DELETE FROM configs WHERE data->>'name' = $1;`
	config := &data.Config{}

	// query
	row := p.dbconn.QueryRow(sqlStatement, name)
	err := row.Scan(&config.Data)
	if err != nil {
		if err.Error() != "sql: no rows in result set" {
			logrus.WithFields(logrus.Fields{
				"cmd":              "Remove",
				"database":         p.dbname,
				"open_connections": p.dbconn.Stats().OpenConnections,
			}).Error(err.Error())
			return nil, err
		}
	}

	// log
	logrus.WithFields(logrus.Fields{
		"cmd":      "Remove",
		"database": p.dbname,
	}).Infoln("Record removed!")

	return config, nil
}

// Search searchs a records in the Database.
func (p *Postgres) Search(params url.Values) ([]*data.Config, error) {
	return make([]*data.Config, 0, 1), nil
}
