package controller

import (
	"github.com/brunovlucena/mobimeo/pkg/controller/myappexample"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, myappexample.Add)
}
