package repository

import (
	"net/url"

	"github.com/brunovlucena/mobimeo/apps/app-example/cmd/myapp/data"
)

type MongoDB struct {
}

func NewMongoDB() *MongoDB {
	return &MongoDB{}
}

func (m *MongoDB) Create(config data.Config) (data.Config, error) {
	return nil, nil
}

func (m *MongoDB) Find(id string) (data.Config, error) {
	return nil, nil
}

func (m *MongoDB) FindAll() ([]data.Config, error) {
	return nil, nil
}

func (m *MongoDB) Update(config data.Config) (data.Config, error) {
	return nil, nil
}

func (m *MongoDB) Remove(name string) (data.Config, error) {
	return nil, nil
}

func (m *MongoDB) Search(params url.Values) ([]data.Config, error) {
	return nil, nil
}
