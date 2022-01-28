package web

import (
	"context"
	"net"

	"github.com/jumboframes/tigerbalm/frame/capal/tblog"
	"github.com/kataras/iris/v12"
)

type Web struct {
	app *iris.Application
	l   net.Listener
}

func (web *Web) Serve(ctx context.Context) {
	go func() {
		<-ctx.Done()
		web.app.Shutdown(context.TODO())
	}()

	if err := web.app.Run(iris.Listener(web.l)); err != nil {
		tblog.Errorf("web::server | app quit: %s", err)
	}
}

func (web *Web) Fini() {
	web.l.Close()
}
