package web

import (
	"github.com/kataras/iris/v12/context"
)

func (web *Web) Register(method string, url string, handler context.Handler) {
	web.app.Handle(method, url, handler)
	web.app.RefreshRouter()
}
