package tbkafka

import (
	"github.com/jumboframes/tigerbalm"
	"github.com/robertkrimen/otto"
)

type TbProducer struct {
	p *Producer
}

func NewTbProducer() (*TbProducer, error) {
	producer, err := NewProducer(tigerbalm.Conf.Kafka.Brokers)
	if err != nil {
		return nil, err
	}
	return &TbProducer{producer}, nil
}

func (tbproducer *TbProducer) Produce(call otto.FunctionCall) otto.Value {
	argc := len(call.ArgumentList)
	if argc != 1 {
		return otto.FalseValue()
	}
	msg, err := OttoValue2PMessage(call.ArgumentList[0])
	if err != nil {
		return otto.FalseValue()
	}

	tbproducer.p.Input() <- msg
	return otto.TrueValue()
}

/*
{
	"Topic": "foo",
	"Payload": "bar"
}
*/
func OttoValue2PMessage(msg otto.Value) (*ProducerMessage, error) {
	// topic
	value, err := msg.Object().Get("Topic")
	if err != nil {
		return nil, err
	}
	topic, err := value.ToString()
	if err != nil {
		return nil, err
	}
	// payload
	value, err = msg.Object().Get("Payload")
	if err != nil {
		return nil, err
	}
	payload, err := value.ToString()
	if err != nil {
		return nil, err
	}
	return &ProducerMessage{
		Topic:   topic,
		Payload: []byte(payload),
	}, nil
}
