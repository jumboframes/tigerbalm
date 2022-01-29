package kafka

import (
	"github.com/jumboframes/tigerbalm/bus"
	"github.com/jumboframes/tigerbalm/frame/capal/tbkafka"
	"github.com/jumboframes/tigerbalm/frame/capal/tblog"
)

type Consumer struct {
	*tbkafka.ConsumerGroup
}

func (consumer *Consumer) AddHandler(handler bus.Handler, matches ...interface{}) {
	if len(matches) != 2 {
		return
	}
	topic, ok := matches[0].(string)
	if !ok {
		return
	}
	group, ok := matches[1].(string)
	if !ok {
		return
	}
	err := consumer.Add(topic, group)
	if err != nil {
		tblog.Errorf("consumer::addhandler | consumer add err: %s", err)
		return
	}
}
