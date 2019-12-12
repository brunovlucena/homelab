package repository

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

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

	// number of open connections
	openConn := p.dbconn.Stats().OpenConnections

	// query
	row := p.dbconn.QueryRow(sqlStatement, dataMap)
	err := row.Scan(&id)
	if err != nil {
		logErr("Create", sqlStatement, openConn, err)
		return nil, err
	}

	// log
	logInfo("Create", sqlStatement, "Record created!", openConn)
	logrus.WithFields(logrus.Fields{
		"cmd": "Create",
	}).Infoln("New record ID is", id)

	return config, nil
}

// Find finds a Record in the database.
func (p *Postgres) Find(name string) (*data.Config, error) {
	sqlStatement := `SELECT data FROM configs WHERE data->>'name' = $1;`
	row := p.dbconn.QueryRow(sqlStatement, name)
	config := new(data.Config)

	// number of open connections
	openConn := p.dbconn.Stats().OpenConnections

	var err error
	switch err = row.Scan(&config.Data); err {
	case sql.ErrNoRows:
		logInfo("Find", sqlStatement, "No Record found!", openConn)
	case nil:
		logInfo("Find", sqlStatement, "Record found!", openConn)
	}

	return config, err
}

// Update updates a Record in the database.
func (p *Postgres) Update(config *data.Config) (*data.Config, error) {
	sqlStatement := `UPDATE configs SET data = $1 WHERE data->>'name' = $2;`
	dataMap := config.Data

	// number of open connections
	openConn := p.dbconn.Stats().OpenConnections

	// query
	_, err := p.dbconn.Exec(sqlStatement, dataMap, dataMap["name"])
	if err != nil {
		logErr("Update", sqlStatement, openConn, err)
		return nil, err
	}

	// log
	logInfo("Update", sqlStatement, "Record updated!", openConn)

	return config, nil
}

// Remove removes a Record from database.
func (p *Postgres) Remove(name string) (*data.Config, error) {
	sqlStatement := `DELETE FROM configs WHERE data->>'name' = $1;`
	config := &data.Config{}

	// number of open connections
	openConn := p.dbconn.Stats().OpenConnections

	// query
	row := p.dbconn.QueryRow(sqlStatement, name)
	if row != nil {
		err := row.Scan(&config.Data)
		if err != nil {
			if err.Error() != "sql: no rows in result set" {
				logErr("Remove", sqlStatement, openConn, err)
				return nil, err
			}
		}
	}

	// log
	logInfo("Remove", sqlStatement, "Record removed!", openConn)

	return config, nil
}

// FindAll returns all Records in the database.
func (p *Postgres) FindAll() ([]*data.Config, error) {
	// return value
	var configs []*data.Config

	// query
	sqlStatement := `SELECT data FROM configs;`
	rows, err := p.dbconn.Query(sqlStatement)
	if rows != nil {
		defer rows.Close()
	}

	// number of open connections
	openConn := p.dbconn.Stats().OpenConnections

	// check for errors
	if err != nil {
		logErr("FindAll", sqlStatement, openConn, err)
		return nil, err
	}

	// Loop through the data
	for rows.Next() {
		var m data.DataMap
		err := rows.Scan(&m)
		// check for errors
		if err != nil {
			logErr("FindAll", sqlStatement, openConn, err)
			return nil, err
		}
		// append results
		configs = append(configs, &data.Config{Data: m})
	}

	// log
	logInfo("FindAll", sqlStatement, "Record(s) found!", openConn)

	return configs, err
}

// Search searchs a records in the Database.
// Query: 	/search?metadata.{key}={value}
func (p *Postgres) Search(params url.Values) ([]*data.Config, error) {
	var configs []*data.Config
	// parse params: map[/search?metadata.limits.cpu.value:[120m]]
	// result: `SELECT FROM configs WHERE data->>'name' = $1;`
	var sqlStatement strings.Builder
	sqlStatement.WriteString(`SELECT data FROM configs WHERE data->`)

	// build statement
	var val string
	for k, v := range params {
		// k = /search?metadata.limits.cpu.value&&name->>''
		// v = 120m
		sk := strings.Split(string(k), "?")
		si := strings.Split(sk[1], ".")
		// [metadata, limits, cpu, value]
		for j, i := range si {
			if j == len(si)-1 {
				sqlStatement.WriteString("'" + i + "'" + "=")
			} else if j == len(si)-2 {
				sqlStatement.WriteString("'" + i + "'" + "->>")
			} else {
				sqlStatement.WriteString("'" + i + "'" + "->")
			}
		}
		val = strings.Join(v, "")
		sqlStatement.WriteString("'" + val + "'" + ";")
	}

	// log
	openConn := p.dbconn.Stats().OpenConnections
	logInfo("search", sqlStatement.String(), "checking database", openConn)

	// query database
	rows, err := p.dbconn.Query(sqlStatement.String())
	if rows != nil {
		defer rows.Close()
	}

	// check for errors
	if err != nil {
		logErr("search", sqlStatement.String(), openConn, err)
		return nil, err
	}

	// Loop through the data
	for rows.Next() {
		var m data.DataMap
		err := rows.Scan(&m)
		// check for errors
		if err != nil {
			logErr("search", "check rows", openConn, err)
			return nil, err
		}

		// append results
		configs = append(configs, &data.Config{Data: m})
	}

	return configs, nil
}

// helper logger
func logInfo(cmd, topic, message string, connections int) {
	logrus.WithFields(logrus.Fields{
		"cmd":              cmd,
		"msg":              topic,
		"open_connections": connections,
	}).Info(message)
}

// helper logger
func logErr(cmd, msg string, connections int, err error) {
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":              cmd,
			"msg":              msg,
			"open_connections": connections,
		}).Error(err.Error())
	}
}
