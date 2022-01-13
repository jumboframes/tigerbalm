package frame

import (
	"github.com/robertkrimen/otto"
)

type Plugin struct {
	Name    string
	Url     string
	Method  string
	Handler otto.Value
}

func (plugin *Plugin) Handle(req *Request) ([]byte, error) {
	this, err := otto.ToValue(nil)
	if err != nil {
		return nil, err
	}
	value, err := plugin.Handler.Call(this, req)
	if err != nil {
		return nil, err
	}
	bodyValue, err := value.Object().Get("body")
	if err != nil {
		return nil, err
	}
	body, err := bodyValue.ToString()
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}
