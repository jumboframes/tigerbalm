package tbkafka

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

var (
	OffsetNewest = sarama.OffsetNewest
	OffsetOldest = sarama.OffsetOldest
)

type ConsumerGroupOption func(*ConsumerGroup) error

func OptionConsumerGroupFailedCh(ch chan<- *ConsumerGroupMessage) ConsumerGroupOption {
	return func(cg *ConsumerGroup) error {
		if ch == nil {
			return nil
		}

		cg.config.Consumer.Return.Errors = true
		cg.failedCh = ch
		return nil
	}
}

func OptionConsumerGroupOffsetsInitial(initial int64) ConsumerGroupOption {
	return func(cg *ConsumerGroup) error {
		cg.config.Consumer.Offsets.Initial = initial
		return nil
	}
}

// 0.10.2.1
// 2.8.0.0
func OptionConsumerGroupVersion(version string) ConsumerGroupOption {
	return func(cg *ConsumerGroup) error {
		kVersion, err := sarama.ParseKafkaVersion(version)
		if err != nil {
			return err
		}
		cg.config.Version = kVersion
		return nil
	}
}

func OptionConsumerGroupSessionTimeout(d time.Duration) ConsumerGroupOption {
	return func(cg *ConsumerGroup) error {
		cg.config.Consumer.Group.Session.Timeout = d
		return nil
	}
}

func OptionConsumerGroupHeartbeatInterval(d time.Duration) ConsumerGroupOption {
	return func(cg *ConsumerGroup) error {
		cg.config.Consumer.Group.Heartbeat.Interval = d
		return nil
	}
}

type ConsumerGroupMessage struct {
	Topic         string
	ConsumerGroup string
	Partition     int32
	Offset        int64
	Payload       []byte
	Error         error
}

type ConsumerGroup struct {
	failedCh chan<- *ConsumerGroupMessage
	queue    uint64
	outputCh chan *ConsumerGroupMessage

	tps *sync.Map

	addrs  []string
	config *sarama.Config
	quit   bool
}

func NewConsumerGroup(addrs []string, options ...ConsumerGroupOption) (*ConsumerGroup, error) {
	cg := &ConsumerGroup{
		queue:  0,
		addrs:  addrs,
		config: sarama.NewConfig(),
		tps:    new(sync.Map),
		quit:   false,
	}
	cg.config.Version = sarama.V0_10_2_1
	for _, option := range options {
		err := option(cg)
		if err != nil {
			return nil, err
		}
	}
	cg.outputCh = make(chan *ConsumerGroupMessage, cg.queue)
	return cg, nil
}

func (cg *ConsumerGroup) Add(topic string, group string,
	handlers ...func(*ConsumerGroupMessage)) error {

	key := topic + group
	_, ok := cg.tps.Load(key)
	if !ok {
		w := newWorkerCG(cg, topic, group, handlers...)
		_, loaded := cg.tps.LoadOrStore(key, w)
		if !loaded {
			err := w.spawn()
			if err != nil {
				return err
			}
		}
		return nil
	}
	return errors.New("topic existed")
}

func (cg *ConsumerGroup) Del(topic string, group string) error {
	key := topic + group
	v, ok := cg.tps.LoadAndDelete(key)
	if !ok {
		return errors.New("topic not exist")
	}
	v.(*workerCG).fini()
	return nil
}

func (cg *ConsumerGroup) Output() <-chan *ConsumerGroupMessage {
	return cg.outputCh
}

func (cg *ConsumerGroup) Fini() {
	cg.quit = true
	cg.tps.Range(func(key, value interface{}) bool {
		w := value.(*workerCG)
		w.fini()
		return true
	})
	close(cg.outputCh)
	return
}

type workerCG struct {
	quit     bool
	cg       *ConsumerGroup
	csr      sarama.ConsumerGroup
	topic    string
	group    string
	handlers []func(*ConsumerGroupMessage) // optional
}

func newWorkerCG(cg *ConsumerGroup, topic string, group string,
	handlers ...func(*ConsumerGroupMessage)) *workerCG {
	return &workerCG{
		cg:       cg,
		topic:    topic,
		group:    group,
		handlers: handlers,
	}
}

func (w *workerCG) spawn() error {
	csr, err := sarama.NewConsumerGroup(w.cg.addrs, w.group, w.cg.config)
	if err != nil {
		return err
	}
	w.csr = csr
	w.quit = false
	go w.work()
	go w.consume()
	return nil
}

func (w *workerCG) consume() {
	for !w.cg.quit && !w.quit {
		err := w.csr.Consume(context.TODO(), []string{w.topic}, w)
		if err != nil && w.cg.failedCh != nil {
			w.cg.failedCh <- &ConsumerGroupMessage{
				Topic:         w.topic,
				ConsumerGroup: w.group,
				Error:         err,
			}
		}
	}
}

func (w *workerCG) work() {
	for !w.cg.quit && !w.quit && w.cg.failedCh != nil {
		for e := range w.csr.Errors() {
			w.cg.failedCh <- &ConsumerGroupMessage{
				Topic:         w.topic,
				ConsumerGroup: w.group,
				Error:         e,
			}
		}
	}
}

func (w *workerCG) Setup(sess sarama.ConsumerGroupSession) error {
	return nil
}

func (w *workerCG) Cleanup(sess sarama.ConsumerGroupSession) error {
	return nil
}

func (w *workerCG) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			cgmsg := &ConsumerGroupMessage{
				Topic:         msg.Topic,
				Partition:     msg.Partition,
				Offset:        msg.Offset,
				Payload:       msg.Value,
				ConsumerGroup: w.group,
			}
			if w.handlers != nil {
				for _, handler := range w.handlers {
					handler(cgmsg)
				}
			} else {
				w.cg.outputCh <- cgmsg
			}
			sess.MarkMessage(msg, "synced")

		case <-sess.Context().Done():
			return sess.Context().Err()
		}
	}
	return nil
}

func (w *workerCG) fini() {
	w.quit = true
	if w.csr != nil {
		w.csr.Close()
	}
}
