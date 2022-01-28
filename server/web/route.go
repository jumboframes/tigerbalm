package web

import (
	"github.com/jumboframes/tigerbalm/bus"
	"github.com/kataras/iris/v12"
)

func (web *Web) AddHandler(handler bus.Handler, matches ...interface{}) {
	if len(matches) != 2 {
		return
	}
	method, ok := matches[0].(string)
	if !ok {
		return
	}
	path, ok := matches[1].(string)
	if !ok {
		return
	}
	web.app.Handle(method, path, func(ctx iris.Context) {
		handler(bus.ContextHttp{ctx, path})
	})
	web.app.RefreshRouter()
}

func (web *Web) DelHandler(matches ...interface{}) {
	if len(matches) != 2 {
		return
	}
	method, ok := matches[0].(string)
	if !ok {
		return
	}
	path, ok := matches[1].(string)
	if !ok {
		return
	}
	web.app.Handle(method, path, func(ctx iris.Context) {
		ctx.NotFound()
	})
	web.app.RefreshRouter()
}

func (web *Web) Type() bus.SlotType {
	return bus.SlotHttp
}
