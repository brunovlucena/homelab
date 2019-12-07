package repository

import (
	"net/url"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/data"
)

type Postgres struct {
}

func NewPostgres() *Postgres {
	return &Postgres{}
}

func (p *Postgres) Create(config data.Config) (data.Config, error) {
	return nil, nil
}

func (p *Postgres) Find(id string) (data.Config, error) {
	return nil, nil
}

func (p *Postgres) FindAll() ([]data.Config, error) {
	return nil, nil
}

func (p *Postgres) Update(config data.Config) (data.Config, error) {
	return nil, nil
}

func (p *Postgres) Remove(name string) (data.Config, error) {
	return nil, nil
}

func (p *Postgres) Search(params url.Values) ([]data.Config, error) {
	return nil, nil
}
