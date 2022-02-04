package kafka

import (
	"github.com/jumboframes/tigerbalm"
	"github.com/jumboframes/tigerbalm/bus"
	"github.com/jumboframes/tigerbalm/frame/capal/tbkafka"
	"github.com/jumboframes/tigerbalm/frame/capal/tblog"
)

type Consumer struct {
	*tbkafka.ConsumerGroup
	failedCh chan *tbkafka.ConsumerGroupMessage
}

func NewConsumer() (*Consumer, error) {
	failedCh := make(chan *tbkafka.ConsumerGroupMessage)

	tbkafka.NewConsumerGroup(tigerbalm.Conf.Kafka.Brokers,
		tbkafka.OptionConsumerGroupFailedCh(failedCh),
		tbkafka.OptionConsumerGroupHeartbeatInterval(
			tigerbalm.Conf.Kafka.Consumer.Group.Heartbeat.Interval),
		tbkafka.OptionConsumerGroupOffsetsInitial(
			tigerbalm.Conf.Kafka.Consumer.Offsets.Initial),
		tbkafka.OptionConsumerGroupSessionTimeout(
			tigerbalm.Conf.Kafka.Consumer.Group.Session.Timeout))
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
	err := consumer.Add(topic, group, func(msg *tbkafka.ConsumerGroupMessage) {
		handler(msg)
	})
	if err != nil {
		tblog.Errorf("consumer::addhandler | consumer add err: %s", err)
	}
}

func (consumer *Consumer) DelHandler(matches ...interface{}) {
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
	err := consumer.Del(topic, group)
	if err != nil {
		tblog.Errorf("consumer::delhandler | consumer del err: %s", err)
	}
}

func (consumer *Consumer) Type() bus.SlotType {
	return bus.SlotKafka
}
