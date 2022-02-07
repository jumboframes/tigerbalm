package web

import (
	"context"
	"net"

	"github.com/jumboframes/tigerbalm"
	"github.com/jumboframes/tigerbalm/frame/capal/tblog"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/recover"
)

type Web struct {
	app *iris.Application
	l   net.Listener
}

func NewWeb() (*Web, error) {
	l, err := net.Listen("tcp", tigerbalm.Conf.Web.Addr)
	if err != nil {
		return nil, err
	}
	app := iris.New()
	app.Use(recover.New())
	return &Web{app, l}, nil
}

func (web *Web) Serve(ctx context.Context) {
	go func() {
		<-ctx.Done()
		web.app.Shutdown(context.TODO())
	}()

	if err := web.app.Run(iris.Listener(web.l)); err != nil {
		if err == iris.ErrServerClosed {
			tblog.Info("web::server | app quit")
		} else {
			tblog.Errorf("web::server | app quit: %s", err)
		}
	}
}

func (web *Web) Fini() {
	web.l.Close()
}
