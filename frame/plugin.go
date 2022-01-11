package frame

import "github.com/robertkrimen/otto"

type Plugin struct {
	Type      string
	Namespace string
	entry     otto.Value
	exit      otto.Value
}

func (plugin *Plugin) Entry(req *Request) []*Request {
}

func (plugin *Plugin) Pipe(rsps []*Response) []*Request {
}

func (plugin *Plugin) Exit(rsps []*Response) *Response {
}
