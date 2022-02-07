package kafka

import (
	"time"

	"github.com/jumboframes/tigerbalm"
	"github.com/jumboframes/tigerbalm/bus"
	"github.com/jumboframes/tigerbalm/frame/capal/tbkafka"
	"github.com/jumboframes/tigerbalm/frame/capal/tblog"
)

type Consumer struct {
	cg       *tbkafka.ConsumerGroup
	failedCh chan *tbkafka.ConsumerGroupMessage
}

func NewConsumer() (*Consumer, error) {
	failedCh := make(chan *tbkafka.ConsumerGroupMessage)
	cg, err := tbkafka.NewConsumerGroup(tigerbalm.Conf.Kafka.Brokers,
		tbkafka.OptionConsumerGroupFailedCh(failedCh),
		tbkafka.OptionConsumerGroupHeartbeatInterval(
			tigerbalm.Conf.Kafka.Consumer.Group.Heartbeat.Interval*time.Second),
		tbkafka.OptionConsumerGroupOffsetsInitial(
			tigerbalm.Conf.Kafka.Consumer.Offsets.Initial),
		tbkafka.OptionConsumerGroupSessionTimeout(
			tigerbalm.Conf.Kafka.Consumer.Group.Session.Timeout*time.Second))
	if err != nil {
		tblog.Errorf("newconsumer | new consumer group err: %s", err)
		return nil, err
	}
	consumer := &Consumer{cg, failedCh}
	go consumer.handleFailed()
	return consumer, nil
}

func (consumer *Consumer) Fini() {
	consumer.cg.Fini()
	close(consumer.failedCh)
}

func (consumer *Consumer) handleFailed() {
	for msg := range consumer.failedCh {
		tblog.Errorf("consumer::handlefailed | err: %s", msg.Error)
	}
}

func (consumer *Consumer) AddHandler(handler bus.Handler, matches ...interface{}) {
	if len(matches) != 2 {
		return
	}
	topic, ok := matches[0].(string)
	if !ok {
		tblog.Error("consumer::addhandler | matches 0 not string")
		return
	}
	group, ok := matches[1].(string)
	if !ok {
		tblog.Error("consumer::addhandler | matches 1 not string")
		return
	}
	err := consumer.cg.Add(topic, group, func(msg *tbkafka.ConsumerGroupMessage) {
		handler(msg)
	})
	if err != nil {
		tblog.Errorf("consumer::addhandler | add err: %s", err)
		return
	}
	tblog.Debugf("consumer::addhandler | add success, topic: %s, group: %s",
		topic, group)
}

func (consumer *Consumer) DelHandler(matches ...interface{}) {
	if len(matches) != 2 {
		return
	}
	topic, ok := matches[0].(string)
	if !ok {
		tblog.Error("consumer::delhandler | matches[0] not string")
		return
	}
	group, ok := matches[1].(string)
	if !ok {
		tblog.Error("consumer::delhandler | matches[1] not string")
		return
	}
	err := consumer.cg.Del(topic, group)
	if err != nil {
		tblog.Errorf("consumer::delhandler | del err: %s", err)
		return
	}
	tblog.Debugf("consumer::delhandler | del success, topic: %s, group: %s",
		topic, group)
}

func (consumer *Consumer) Type() bus.SlotType {
	return bus.SlotKafka
}
