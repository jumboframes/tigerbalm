package tbkafka

import "github.com/robertkrimen/otto"

type Message struct {
	Topic         string
	ConsumerGroup string
	Partition     int32
	Offset        int64
	Payload       string
}

func CGMessage2TbMessage(cgmsg *ConsumerGroupMessage) (*Message, error) {
	return &Message{
		Topic:         cgmsg.Topic,
		ConsumerGroup: cgmsg.ConsumerGroup,
		Partition:     cgmsg.Partition,
		Offset:        cgmsg.Offset,
		Payload:       string(cgmsg.Payload),
	}, nil
}

func TbMessage2OttoValue(msg *Message) (otto.Value, error) {
	return otto.New().ToValue(msg)
}
