package controller

import (
	"hypercloud-operator-go/pkg/controller/registry"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, registry.Add)
}
