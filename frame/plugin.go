package frame

import (
	"ms_bff/errors"
	"ms_bff/frame/capal"
	"sync"

	"github.com/robertkrimen/otto"
)

const (
	RegisterFunc = "register"
)

type Plugin struct {
	url    string
	method string

	pool           sync.Pool
	handlerFactory func() interface{}
}

func NewPlugin(content []byte) (*Plugin, error) {
	err := error(nil)
	url, method := "", ""

	handlerFactory := func() interface{} {
		vm := otto.New()
		_, err = vm.Run(content)
		if err != nil {
			return nil
		}
		var pluginRoute otto.Value
		pluginRoute, err = vm.Call(RegisterFunc, nil)
		if err != nil {
			return nil
		}
		if !pluginRoute.IsObject() {
			err = errors.ErrIllegalRegister
			return nil
		}

		var route *route
		route, err = getRoute(pluginRoute.Object())
		if err != nil {
			return nil
		}
		// set capability
		err = vm.Set("doRequest", capal.DoRequest)
		if err != nil {
			return nil
		}

		url = route.url
		method = route.method
		return route.handler
	}

	handler := handlerFactory()
	if err != nil {
		return nil, err
	}

	pool := sync.Pool{
		New: handlerFactory,
	}
	pool.Put(handler)

	return &Plugin{
		url:    url,
		method: method,
		pool:   pool,
	}, nil
}

func (plugin *Plugin) Url() string {
	return plugin.url
}

func (plugin *Plugin) Method() string {
	return plugin.method
}

func (plugin *Plugin) Handle(req *capal.Request) (*capal.Response, error) {
	handler := plugin.pool.Get()
	if handler == nil {
		return nil, errors.ErrNewInterpreter
	}
	this, err := otto.ToValue(nil)
	if err != nil {
		return nil, err
	}
	rsp, err := handler.(otto.Value).Call(this, req)
	if err != nil {
		return nil, err
	}
	plugin.pool.Put(handler)
	// status
	statusValue, err := rsp.Object().Get("status")
	if err != nil {
		return nil, err
	}
	status, err := statusValue.ToInteger()
	if err != nil {
		return nil, err
	}

	// header
	headerValue, err := rsp.Object().Get("header")
	if err != nil {
		return nil, err
	}
	header := map[string]string{}
	for _, key := range headerValue.Object().Keys() {
		v, err := headerValue.Object().Get(key)
		if err != nil {
			continue
		}
		value, err := v.ToString()
		if err != nil {
			continue
		}
		header[key] = value
	}

	// body
	bodyValue, err := rsp.Object().Get("body")
	if err != nil {
		return nil, err
	}
	body, err := bodyValue.ToString()
	if err != nil {
		return nil, err
	}
	response := &capal.Response{
		Status: int(status),
		Header: header,
		Body:   body,
	}
	return response, nil
}
