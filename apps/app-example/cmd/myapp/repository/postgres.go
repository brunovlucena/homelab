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

type Postgres struct {
	host     string
	port     string
	user     string
	password string
	dbname   string
	dbconn   *sql.DB
}

type Item struct {
	ID     int
	Config data.Config
}

func NewPostgres(host, port, user, password, dbname string) *Postgres {
	dbconn := connect(host, port, user, password, dbname)
	return &Postgres{host, port, user, password, dbname, dbconn}
}

//func (p *Postgres) Close() {
//logrus.Infoln("Disconnected from Postgres!")
//p.dbconn.Close()
//}

func connect(host, port, user, password, dbname string) *sql.DB {
	// info: sslmode - disabled
	psqlConn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
	logrus.Infof("Conn: %s", psqlConn)
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	// opens connection
	db, err := sql.Open("postgres", psqlInfo)
	// setup for highter performance
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(20)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"host":     host,
			"port":     port,
			"database": dbname,
		}).Error("Connect: Cannot connect!")
		//panic(err)
	} else {
		// SQL.Open only creates the DB object, but dies not open
		//any connections to the database.
		err = db.Ping()
		if err != nil {
			stats := db.Stats()
			logrus.WithFields(logrus.Fields{
				"host":             host,
				"port":             port,
				"database":         dbname,
				"max_connections":  stats.MaxOpenConnections, // to check stats
				"open_connections": stats.OpenConnections,    // to check stats
			}).Error("Connect: failed to ping!")
			//panic(err)
		}
	}
	// Success
	logrus.Infoln("Successfully connected to Postgres!")
	return db
}

func (p *Postgres) Create(config *data.Config) (*data.Config, error) {
	sqlStatement := `INSERT INTO configs (data) VALUES ($1) RETURNING id`
	var id int
	dataMap := config.Data
	row := p.dbconn.QueryRow(sqlStatement, dataMap)
	err := row.Scan(&id)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd:":        "Create",
			"config_name": config.Data["name"],
			"database":    p.dbname,
		}).Error(err.Error())
		return nil, err
	}
	// some crazy unescessary warn
	if id == 1000 {
		logrus.WithFields(logrus.Fields{
			"cmd:":     "Create",
			"id":       id,
			"database": p.dbname,
		}).Warn("Create: The number of records number increased tremendously!")
	}
	logrus.Infoln("Create: New record ID is", id)
	return config, nil
}

func (p *Postgres) Find(name string) (*data.Config, error) {
	sqlStatement := `SELECT data FROM configs WHERE data->>'name' = $1;`
	row := p.dbconn.QueryRow(sqlStatement, name)
	config := new(data.Config)
	var err error
	switch err = row.Scan(&config.Data); err {
	case sql.ErrNoRows:
		logrus.WithFields(logrus.Fields{
			"cmd:":        "Find",
			"config_name": name,
			"database":    p.dbname,
		}).Warn("No rows were returned!")
	case nil:
		logrus.Info("Find: Record found!")
	}
	return config, err
}

func (p *Postgres) FindAll() ([]*data.Config, error) {
	// return value
	var configs []*data.Config
	// query
	sqlStatement := `SELECT data FROM configs;`
	rows, err := p.dbconn.Query(sqlStatement)
	// check for errors
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":      "FindAll",
			"database": p.dbname,
		}).Error(err.Error())
		return nil, err
	}
	defer rows.Close()
	// Loop through the data
	for rows.Next() {
		var m data.DataMap
		err := rows.Scan(&m)
		// check for errors
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		// append results
		configs = append(configs, &data.Config{Data: m})
	}
	return configs, err
}

func (p *Postgres) Update(config *data.Config) (*data.Config, error) {
	sqlStatement := `UPDATE configs SET data = $1 WHERE data->>'name' = $2;`
	dataMap := config.Data
	_, err := p.dbconn.Exec(sqlStatement, dataMap, dataMap["name"])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"cmd":      "Update",
			"database": p.dbname,
		}).Error(err.Error())
		return nil, err
	}
	fmt.Println("Update: Record Updated!")
	return config, nil
}

func (p *Postgres) Remove(name string) (*data.Config, error) {
	sqlStatement := `DELETE FROM configs WHERE data->>'name' = $1;`
	config := &data.Config{}
	row := p.dbconn.QueryRow(sqlStatement, name)
	err := row.Scan(&config.Data)
	if err != nil {
		if err.Error() != "sql: no rows in result set" {
			logrus.WithFields(logrus.Fields{
				"cmd":      "Removee",
				"database": p.dbname,
			}).Error(err.Error())
			return nil, err
		}
	}
	// return removed record
	fmt.Println("Remove: Record removed!")
	return config, nil
}

func (p *Postgres) Search(params url.Values) ([]*data.Config, error) {
	return make([]*data.Config, 0, 1), nil
}
