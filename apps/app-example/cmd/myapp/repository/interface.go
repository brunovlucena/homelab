package repository

import (
	"errors"
	"net/url"
	"strings"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/data"
)

//Repository repository interface
type Repository interface {
	Create(config data.Config) (int, error)
	Find(name string) (data.Config, error)
	FindAll() ([]data.Config, error)
	Update(config data.Config) (data.Config, error)
	Remove(name string) (data.Config, error)
	Search(params url.Values) ([]data.Config, error)
}

func NewRepository(name string) (Repository, error) {
	switch strings.ToLower(name) {
	case "inmemdb":
		return NewInMemDB(), nil
	case "postgres":
		return NewPostgres(), nil
	}
	return nil, errors.New("Invalid base given")
}
