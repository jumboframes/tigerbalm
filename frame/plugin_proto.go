package frame

import (
	"ms_bff/errors"

	"github.com/robertkrimen/otto"
)

const (
	MetaMatch   = "match"
	MetaUrl     = "url"
	MetaMethod  = "method"
	MetaHandler = "handler"
)

type route struct {
	url, method string
	handler     otto.Value
}

func getRoute(obj *otto.Object) (*route, error) {
	matchValue, err := obj.Get(MetaMatch)
	if err != nil {
		return nil, err
	}
	// url
	urlValue, err := matchValue.Object().Get(MetaUrl)
	if err != nil {
		return nil, err
	}
	url, err := urlValue.ToString()
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
		return nil, errors.ErrRegisterNotFunction
	}
	route := &route{
		url:     url,
		method:  method,
		handler: handler,
	}
	return route, nil
}
