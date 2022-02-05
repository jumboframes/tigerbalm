package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cihub/seelog"
	"github.com/jumboframes/tigerbalm"
	"github.com/jumboframes/tigerbalm/bus"
	"github.com/jumboframes/tigerbalm/frame"
	"github.com/jumboframes/tigerbalm/frame/capal/tblog"
	"github.com/jumboframes/tigerbalm/server/kafka"
	"github.com/jumboframes/tigerbalm/server/web"
)

func main() {
	err := tigerbalm.Init()
	if err != nil {
		tblog.Fatalf("main | init err: %s", err)
		return
	}
	ctx, cancel := context.WithCancel(context.TODO())

	// web
	web, err := web.NewWeb()
	if err != nil {
		tblog.Errorf("main | new web err: %s", err)
		return
	}
	defer web.Fini()
	go web.Serve(ctx)

	// kafka
	consumer, err := kafka.NewConsumer()
	if err != nil {
		tblog.Errorf("main | new consumer err: %s", err)
		return
	}
	defer consumer.Fini()

	// bus
	bus := bus.NewSlotBus()
	bus.AddSlot(web)
	bus.AddSlot(consumer)

	// frame
	frame, err := frame.NewFrame(bus, tigerbalm.Conf)
	if err != nil {
		tblog.Errorf("main | new frame err: %s", err)
		return
	}
	defer frame.Fini()

	// signal
}

func HandleSignal(ctx context.Context) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, os.Kill, syscall.SIGTERM)
	select {
	case sig := <-s:
		signal.Reset()
		seelog.Infof("main | got signal: %s", sig.String())

	case <-ctx.Done():
		seelog.Infof("main | sub routine exit")
	}
}
