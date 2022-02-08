package main

import (
	"context"
	"syscall"

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

	tblog.Infof(`
==================================================
                 TigerBalm Starts
==================================================`)

	// bus, io总线
	bus := bus.NewSlotBus()

	// web
	web, err := web.NewWeb()
	if err != nil {
		tblog.Errorf("main | new web err: %s", err)
		return
	}
	defer web.Fini()
	go web.Serve(ctx)
	bus.AddSlot(web)

	// kafka
	if tigerbalm.Conf.Kafka.Enable {
		consumer, err := kafka.NewConsumer()
		if err != nil {
			tblog.Errorf("main | new consumer err: %s", err)
			return
		}
		defer consumer.Fini()
		bus.AddSlot(consumer)
	}

	// frame
	frame, err := frame.NewFrame(bus)
	if err != nil {
		tblog.Errorf("main | new frame err: %s", err)
		return
	}
	defer frame.Fini()

	// signal
	sig := tigerbalm.NewSignal(tigerbalm.OptionSignalCancel(cancel))
	sig.Add(syscall.SIGHUP, frame)
	sig.Wait(ctx)
}
