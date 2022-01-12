package frame

import "github.com/robertkrimen/otto"

type Plugin struct {
	Type      string
	Namespace string
	Handler   otto.Value
}

func (plugin *Plugin) Handle(req *Request) {}
