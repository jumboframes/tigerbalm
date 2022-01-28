package tbkafka

import (
	"os"
	"os/signal"
	"testing"
)

func TestConsumerGroup(t *testing.T) {
	cg, err := NewConsumerGroup([]string{"192.168.111.103:9092"})
	if err != nil {
		t.Error(err)
		return
	}
	err = cg.Add("topic_test", "group_test")
	if err != nil {
		t.Error(err)
		return
	}
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, os.Kill)
	for {
		select {
		case msg, ok := <-cg.Output():
			if !ok {
				goto QUIT
			}
			t.Log(string(msg.Payload))
		case <-s:
			cg.Fini()
			return
		}
	}
QUIT:
}
