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

func NewPostgres(host, port, user, pass, dbname string) *Postgres {
	// connect to database
	dbconn := Connect(host, port, user, pass, dbname)
	// return
	return &Postgres{host, port, user, pass, dbname, dbconn}
}

func Connect(host, port, user, pass, dbname string) *sql.DB {
	// info: sslmode - disabled
	psqlConn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbname)
	logrus.Infof("Conn: %s", psqlConn)
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
	// opens connection
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"host":     host,
			"port":     port,
			"database": dbname,
		}).Warn("Connect: Cannot connect!")
		panic(err)
	} else {
		// SQL.Open only creates the DB object, but dies not open
		//any connections to the database.
		err = db.Ping()
		if err != nil {
			fmt.Println("db.Ping failed:", err)
			panic(err)
		}
	}
	// Success
	logrus.Println("Successfully connected to Postgres!")
	return db
}

func (p *Postgres) Create(config *data.Config) (*data.Config, error) {
	sqlStatement := `INSERT INTO configs (data) VALUES ($1) RETURNING id`
	var id int
	dataMap := config.Data
	row := p.dbconn.QueryRow(sqlStatement, dataMap)
	err := row.Scan(&id)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	// some warn
	if id == 1000 {
		logrus.WithFields(logrus.Fields{
			"omg": true,
			"id":  id,
		}).Warn("Create: The number of records number increased tremendously!")
	}
	fmt.Println("Create: New record ID is:", id)
	return config, nil
}

func (p *Postgres) Find(name string) (*data.Config, error) {
	sqlStatement := `SELECT data FROM configs WHERE data->>'name' = $1;`
	row := p.dbconn.QueryRow(sqlStatement, name)
	config := new(data.Config)
	var err error
	switch err = row.Scan(&config.Data); err {
	case sql.ErrNoRows:
		logrus.Println("No rows were returned!")
	case nil:
		logrus.Println("Find: Record found!")
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
		logrus.Error(err)
		return nil, err
	}
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
		logrus.Error(err)
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
			logrus.Error(err)
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
