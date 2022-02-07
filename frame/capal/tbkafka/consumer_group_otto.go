package tbkafka

import "github.com/robertkrimen/otto"

type CGMessage struct {
	Topic     string
	Group     string
	Partition int32
	Offset    int64
	Payload   string
}

func CGMessage2TbCGMessage(cgmsg *ConsumerGroupMessage) (*CGMessage, error) {
	return &CGMessage{
		Topic:     cgmsg.Topic,
		Group:     cgmsg.ConsumerGroup,
		Partition: cgmsg.Partition,
		Offset:    cgmsg.Offset,
		Payload:   string(cgmsg.Payload),
	}, nil
}

func TbMessage2OttoValue(msg *CGMessage) (otto.Value, error) {
	return otto.New().ToValue(msg)
}
