package bus

import (
	"ms_bff/server/web"

	"github.com/kataras/iris/v12"
)

type Bus struct {
	web *web.Web
}

func NewBus(web *web.Web) *Bus {
	return &Bus{
		web: web,
	}
}

func (bus *Bus) RegisterHttp(method string, url string, handler ContextHandler) {
	bus.web.Register(method, url, func(ctx iris.Context) {
		busCtx := &Context{
			ctx, url,
		}
		handler(busCtx)
	})
}
