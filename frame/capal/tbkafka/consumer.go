package tbkafka

import (
	"crypto/tls"
	"errors"
	"sync"

	"github.com/Shopify/sarama"
)

type ConsumerOption func(*Consumer) error

type ConsumerMessage struct {
	Headers   []*sarama.RecordHeader // only set if kafka is version 0.11+
	Topic     string
	Partition int32
	Offset    int64
	Payload   []byte
	Error     error
}

type Consumer struct {
	unshare bool

	failedCh chan<- *ConsumerMessage
	queue    uint64
	outputCh chan *ConsumerMessage

	tps *sync.Map

	addrs  []string
	config *sarama.Config
	csr    sarama.Consumer
}

type topicParamsConsumer struct {
	csr        sarama.Consumer
	topic      string
	partitions []int
	offset     int64
	prtCsrs    []sarama.PartitionConsumer

	ws []*workerC
}

func NewConsumer(addrs []string, options ...ConsumerOption) (*Consumer, error) {
	c := &Consumer{
		queue:  0,
		addrs:  addrs,
		config: sarama.NewConfig(),
		tps:    new(sync.Map),
	}
	for _, option := range options {
		option(c)
	}
	c.outputCh = make(chan *ConsumerMessage, c.queue)
	if !c.unshare {
		csr, err := sarama.NewConsumer(c.addrs, c.config)
		if err != nil {
			return nil, err
		}
		c.csr = csr
	}
	return c, nil
}

func (c *Consumer) Add(topic string, partitions []int, offset int64) error {
	_, ok := c.tps.Load(topic)
	if !ok {
		tp := &topicParamsConsumer{
			topic:      topic,
			partitions: partitions,
			offset:     offset,
		}
		_, loaded := c.tps.LoadOrStore(topic, tp)
		if !loaded {
			err := c.spawn(tp)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return errors.New("topic existed")
}

func OptionConsumerSasl(user, passwd string) ConsumerOption {
	return func(c *Consumer) error {
		c.config.Net.SASL.Enable = true
		c.config.Net.SASL.Version = 1
		c.config.Net.SASL.User = user
		c.config.Net.SASL.Password = passwd
		return nil
	}
}

func OptionConsumerFailedCh(ch chan<- *ConsumerMessage) ConsumerOption {
	return func(c *Consumer) error {
		c.config.Consumer.Return.Errors = true
		c.failedCh = ch
		return nil
	}
}

func (c *Consumer) Output() <-chan *ConsumerMessage {
	return c.outputCh
}

func (c *Consumer) Fini() {
	c.tps.Range(func(key, value interface{}) bool {
		tp := value.(*topicParamsConsumer)
		for index, prtCsr := range tp.prtCsrs {
			prtCsr.AsyncClose()
			tp.ws[index].fini()
		}
		if !c.unshare {
			tp.csr.Close()
		}
		return true
	})
	close(c.outputCh)
	return
}

func (c *Consumer) spawn(tp *topicParamsConsumer) error {
	var csr sarama.Consumer
	var err error
	if !c.unshare {
		csr = c.csr
	} else {
		csr, err = sarama.NewConsumer(c.addrs, c.config)
		if err != nil {
			return err
		}
	}

	prtCsrs := make([]sarama.PartitionConsumer, len(tp.partitions))
	workers := make([]*workerC, len(tp.partitions))
	for i := 0; i < len(tp.partitions); i++ {
		prtCsrs[i], err = csr.ConsumePartition(tp.topic, int32(i), tp.offset)
		if err != nil {
			return err
		}
		w := newworkerC()
		w.work(c.outputCh, c.failedCh, prtCsrs[i])
		workers[i] = w
	}
	tp.csr = csr
	tp.prtCsrs = prtCsrs
	tp.ws = workers
	return nil
}

type workerC struct {
	wg *sync.WaitGroup
}

func newworkerC() *workerC {
	w := &workerC{
		wg: new(sync.WaitGroup),
	}
	return w
}

func (w *workerC) work(outCh chan<- *ConsumerMessage,
	failedCh chan<- *ConsumerMessage, prtCsr sarama.PartitionConsumer) {

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for msg := range prtCsr.Messages() {
			outCh <- &ConsumerMessage{
				Headers:   msg.Headers,
				Topic:     msg.Topic,
				Partition: msg.Partition,
				Offset:    msg.Offset,
				Payload:   msg.Value,
			}
		}
	}()

	if failedCh != nil {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			for e := range prtCsr.Errors() {
				failedCh <- &ConsumerMessage{
					Topic:     e.Topic,
					Partition: e.Partition,
					Error:     e.Err,
				}
			}

		}()
	}
}

func (w *workerC) fini() {
	w.wg.Wait()
}

func OptionConsumerNetTLS(config *tls.Config) ConsumerOption {
	return func(c *Consumer) error {
		c.config.Net.TLS.Enable = true
		c.config.Net.TLS.Config = config
		return nil
	}
}
