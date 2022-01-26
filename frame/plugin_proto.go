package frame

import (
	"github.com/jumboframes/tigerbalm"

	"github.com/robertkrimen/otto"
)

const (
	MetaMatch   = "match"
	MetaPath    = "path"
	MetaMethod  = "method"
	MetaHandler = "handler"
)

const (
	FuncRegister = "register"
	FuncRequire  = "require"
	VarContext   = "context"
)

type context struct {
	Name string
}

type runtime struct {
	*route
	vm *otto.Otto
}

type route struct {
	path, method string
	handler      otto.Value
}

func getRoute(obj *otto.Object) (*route, error) {
	matchValue, err := obj.Get(MetaMatch)
	if err != nil {
		return nil, err
	}
	// path
	urlValue, err := matchValue.Object().Get(MetaPath)
	if err != nil {
		return nil, err
	}
	path, err := urlValue.ToString()
	if err != nil {
		return nil, err
	}
	// method
	methodValue, err := matchValue.Object().Get(MetaMethod)
	if err != nil {
		return nil, err
	}
	method, err := methodValue.ToString()
	if err != nil {
		return nil, err
	}
	// handler
	handler, err := obj.Get(MetaHandler)
	if err != nil {
		return nil, err
	}
	if !handler.IsFunction() {
		return nil, tigerbalm.ErrRegisterNotFunction
	}
	route := &route{
		path:    path,
		method:  method,
		handler: handler,
	}
	return route, nil
}
