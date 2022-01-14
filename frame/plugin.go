package frame

import (
	"github.com/jumboframes/tigerbalm/frame/capal"
	"github.com/robertkrimen/otto"
)

type Plugin struct {
	Name    string
	Url     string
	Method  string
	Handler otto.Value
}

func (plugin *Plugin) Handle(req *capal.Request) (*capal.Response, error) {
	this, err := otto.ToValue(nil)
	if err != nil {
		return nil, err
	}
	value, err := plugin.Handler.Call(this, req)
	if err != nil {
		return nil, err
	}
	// status
	statusValue, err := value.Object().Get("status")
	if err != nil {
		return nil, err
	}
	status, err := statusValue.ToInteger()
	if err != nil {
		return nil, err
	}

	// body
	bodyValue, err := value.Object().Get("body")
	if err != nil {
		return nil, err
	}
	body, err := bodyValue.ToString()
	if err != nil {
		return nil, err
	}
	rsp := &capal.Response{
		Status: int(status),
		Body:   body,
	}
	return rsp, nil
}
