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
	return &Postgres{host, port, user, pass, dbname, nil}
}

func (p *Postgres) Close() {
	p.dbconn.Close()
}

func (p *Postgres) Connect() {
	// info: sslmode - disabled
	psqlConn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", p.user, p.password, p.host, p.port, p.dbname)
	logrus.Infof("Conn: %s", psqlConn)
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", p.host, p.port, p.user, p.password, p.dbname)
	// opens connection
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"host":     p.host,
			"port":     p.port,
			"database": p.dbname,
		}).Warn("Connect: Cannot connect!")
		panic(err)
	} else {
		// SQL.Open only creates the DB object, but dies not open
		//any connections to the database.
		err = db.Ping()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"host":     p.host,
				"port":     p.port,
				"database": p.dbname,
			}).Error("Connect: failed to ping!")
			panic(err)
		}
	}
	// Success
	logrus.Println("Successfully connected to Postgres!")
	p.dbconn = db
}

func (p *Postgres) Create(config *data.Config) (*data.Config, error) {
	p.Connect()
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
	p.Close()
	return config, nil
}

func (p *Postgres) Find(name string) (*data.Config, error) {
	p.Connect()
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
	p.Close()
	return config, err
}

func (p *Postgres) FindAll() ([]*data.Config, error) {
	p.Connect()
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
	p.Close()
	return configs, err
}

func (p *Postgres) Update(config *data.Config) (*data.Config, error) {
	p.Connect()
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
	p.Close()
	return config, nil
}

func (p *Postgres) Remove(name string) (*data.Config, error) {
	p.Connect()
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
	p.Close()
	return config, nil
}

func (p *Postgres) Search(params url.Values) ([]*data.Config, error) {
	p.Connect()
	p.Close()
	return make([]*data.Config, 0, 1), nil
}
