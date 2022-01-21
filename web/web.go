package web

import (
	"context"
	"net"

	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
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
		logrus.Errorf("Web::Server | app quit: %s", err)
	}
}

func (web *Web) Fini() {
	web.l.Close()
}
