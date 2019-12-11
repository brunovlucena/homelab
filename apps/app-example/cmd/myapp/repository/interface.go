package repository

import (
	"errors"
	"net/url"
	"strings"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/data"
)

//Repository repository interface
type Repository interface {
	Create(config *data.Config) (*data.Config, error)
	Find(name string) (*data.Config, error)
	FindAll() ([]*data.Config, error)
	Update(config *data.Config) (*data.Config, error)
	Remove(name string) (*data.Config, error)
	Search(params url.Values) ([]*data.Config, error)
}

// NewRepository returns a selected repository.
func NewRepository(name, host, port, user, pass, dbname string) (Repository, error) {
	switch strings.ToLower(name) {
	case "postgres":
		return NewPostgres(host, port, user, pass, dbname), nil
	}
	return nil, errors.New("Invalid base given")
}
