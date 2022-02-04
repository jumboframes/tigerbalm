package frame

import (
	"github.com/jumboframes/tigerbalm"

	"github.com/robertkrimen/otto"
)

const (
	MetaRoute   = "route"
	MetaConsume = "consume"
	MetaMatch   = "match"
	MetaPath    = "path"
	MetaMethod  = "method"
	MetaTopic   = "topic"
	MetaGroup   = "group"
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
	*registration
	vm *otto.Otto
}

type registration struct {
	route   *route
	consume *consume
}

type consume struct {
	topic, group string
	handler      otto.Value
}

type route struct {
	path, method string
	handler      otto.Value
}

func getRegistration(obj *otto.Object) (*registration, error) {
	registration := &registration{}

	routeValue, err := obj.Get(MetaRoute)
	if err != nil {
		return nil, err
	}
	if routeValue.IsDefined() {
		route, err := getRoute(routeValue.Object())
		if err != nil {
			return nil, err
		}
		registration.route = route
	}
	consumeValue, err := obj.Get(MetaConsume)
	if err != nil {
		return nil, err
	}
	if consumeValue.IsDefined() {
		consume, err := getConsume(consumeValue.Object())
		if err != nil {
			return nil, err
		}
		registration.consume = consume
	}
	return registration, nil
}

func getConsume(obj *otto.Object) (*consume, error) {
	matchValue, err := obj.Get(MetaMatch)
	if err != nil {
		return nil, err
	}
	// topic
	topicValue, err := matchValue.Object().Get(MetaTopic)
	if err != nil {
		return nil, err
	}
	topic, err := topicValue.ToString()
	if err != nil {
		return nil, err
	}
	// group
	groupValue, err := matchValue.Object().Get(MetaGroup)
	if err != nil {
		return nil, err
	}
	group, err := groupValue.ToString()
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
	consume := &consume{
		topic:   topic,
		group:   group,
		handler: handler,
	}
	return consume, nil
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
