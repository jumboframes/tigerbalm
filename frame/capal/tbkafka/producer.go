package tbkafka

import (
	"crypto/tls"
	"errors"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

type RequireType int8

const (
	RequireNoResponse  RequireType = 0
	RequireOnlyLeader  RequireType = 1
	RequireAllReplicas RequireType = -1
)

type CompressionType int8

const (
	CompressionNone   CompressionType = 0
	CompressionGZIP   CompressionType = 1
	CompressionSnappy CompressionType = 2
)

const (
	defaultTopicQueue = 1024 * 1024
	defaultTopicConcu = 10
)

type ProducerOption func(*Producer) error

//用户消息
type ProducerMessage struct {
	//topic
	Topic string
	// The partitioning key for this message. Pre-existing Encoders include
	// StringEncoder and ByteEncoder.
	Key sarama.Encoder
	//Payload to produce.
	Payload []byte
	// The headers are key-value pairs that are transparently passed
	// by Kafka between producers and consumers.
	Headers []sarama.RecordHeader
	//Custom data which will be enqueued FailedQueue if failed,
	//or be enqueued SucceedQueue if succeed.
	Custom interface{}
	//Partition specified by user when OptionManualPartition
	//was set, or returned by broker.
	Partition int32
	//Offset returned by broker.
	Offset int64
	//Error will be set while showing up in FailedQueue.
	Error error
}

/*
*	Producer结束应该调用Fini，这样会等待本模块所有可能block
*	的函数结束才返回。
**/
type Producer struct {
	unshare bool
	queue   uint64
	concu   uint32

	failedCh  chan<- *ProducerMessage
	succeedCh chan<- *ProducerMessage
	inputCh   chan *ProducerMessage

	tps *sync.Map

	addrs  []string
	config *sarama.Config
	asyncP sarama.AsyncProducer

	wg *sync.WaitGroup
}

type topicParamsProducer struct {
	queue        uint64
	concu        uint32
	topicInputCh chan *ProducerMessage

	asyncP sarama.AsyncProducer
	w      *workerP
}

func NewProducer(addrs []string, options ...ProducerOption) (*Producer, error) {
	p := &Producer{
		queue:  0,
		concu:  1,
		addrs:  addrs,
		config: sarama.NewConfig(),
		tps:    new(sync.Map),
		wg:     new(sync.WaitGroup),
	}
	for _, option := range options {
		option(p)
	}
	p.inputCh = make(chan *ProducerMessage, p.queue)
	if !p.unshare {
		asyncP, err := sarama.NewAsyncProducer(addrs, p.config)
		if err != nil {
			return nil, err
		}
		p.asyncP = asyncP
	}

	p.tps.Range(func(key, value interface{}) bool {
		tp := value.(*topicParamsProducer)
		p.spawn(tp)
		return true
	})

	go p.dispatch()
	return p, nil
}

func (p *Producer) Input() chan<- *ProducerMessage {
	return p.inputCh
}

func (p *Producer) Fini() {
	if !p.unshare {
		p.asyncP.Close()
	}

	p.tps.Range(func(key, value interface{}) bool {
		tp := value.(*topicParamsProducer)
		if p.unshare {
			tp.asyncP.Close()
		}

		close(tp.topicInputCh)
		tp.w.fini()

		return true
	})

	close(p.inputCh)
	p.wg.Wait()
	return
}

//dispatch函数会在close(p.inputCh)后返回
func (p *Producer) dispatch() {
	p.wg.Add(int(p.concu))
	for i := 0; i < int(p.concu); i++ {
		go func() {
			defer p.wg.Done()

			for msg := range p.inputCh {
				tpIf, ok := p.tps.Load(msg.Topic)
				if !ok {
					var loaded bool
					tp := &topicParamsProducer{
						queue:        0,
						concu:        1,
						topicInputCh: make(chan *ProducerMessage, 0),
					}
					tpIf, loaded = p.tps.LoadOrStore(msg.Topic, tp)
					if !loaded {
						err := p.spawn(tp)
						if err != nil && p.failedCh != nil {
							msg.Error = err
							p.failedCh <- msg
						}
					}
				}
				tp := tpIf.(*topicParamsProducer)
				tp.topicInputCh <- msg
			}
		}()
	}
}

func (p *Producer) spawn(tp *topicParamsProducer) error {
	var asyncP sarama.AsyncProducer
	var err error
	if p.unshare {
		asyncP, err = sarama.NewAsyncProducer(p.addrs, p.config)
		if err != nil {
			return err
		}
		tp.asyncP = asyncP
	} else {
		tp.asyncP = p.asyncP
	}
	w := newworkerP()
	w.work(tp.topicInputCh, p.failedCh, p.succeedCh, tp.asyncP, tp.concu)
	tp.w = w
	return nil
}

//用户队列和分派并发
func OptionProducerConcu(queue uint64, concu uint32) ProducerOption {
	return func(p *Producer) error {
		p.concu = concu
		p.queue = queue
		return nil
	}
}

//每个topic使用独立的连接
func OptionProducerUnshare() ProducerOption {
	return func(p *Producer) error {
		p.unshare = true
		return nil
	}
}

//每个topic配置的读队列和并发
func OptionTopic(topic string, queue uint64, concu uint32) ProducerOption {
	return func(p *Producer) error {
		tpIf, ok := p.tps.Load(topic)
		if !ok {
			tpIf = &topicParamsProducer{}
			actual, loaded := p.tps.LoadOrStore(topic, tpIf)
			if loaded {
				tpIf = actual
			}
		}
		tp := tpIf.(*topicParamsProducer)
		tp.queue = queue
		tp.concu = concu
		if tp.topicInputCh != nil {
			close(tp.topicInputCh)
		}
		tp.topicInputCh = make(chan *ProducerMessage, queue)
		return nil
	}
}

//如果库消息发送成功，则把该消息回写ch
func OptionSucceedCh(ch chan<- *ProducerMessage) ProducerOption {
	return func(p *Producer) error {
		p.config.Producer.Return.Successes = true
		p.succeedCh = ch
		return nil
	}
}

func OptionFailedCh(ch chan<- *ProducerMessage) ProducerOption {
	return func(p *Producer) error {
		p.config.Producer.Return.Errors = true
		p.failedCh = ch
		return nil
	}
}

func OptionFlushBytes(size int) ProducerOption {
	return func(p *Producer) error {
		p.config.Producer.Flush.Bytes = size
		return nil
	}
}

func OptionFlushMessages(count int) ProducerOption {
	return func(p *Producer) error {
		p.config.Producer.Flush.Messages = count
		return nil
	}
}

func OptionSasl(user, passwd string) ProducerOption {
	return func(p *Producer) error {
		p.config.Net.SASL.Enable = true
		p.config.Net.SASL.Version = 1
		p.config.Net.SASL.User = user
		p.config.Net.SASL.Password = passwd
		return nil
	}
}

func OptionFlushFrequency(frequency time.Duration) ProducerOption {
	return func(p *Producer) error {
		p.config.Producer.Flush.Frequency = frequency
		return nil
	}
}

func OptionFlushMaxMessages(count int) ProducerOption {
	return func(p *Producer) error {
		p.config.Producer.Flush.MaxMessages = count
		return nil
	}
}

func OptionRetryMax(retry int) ProducerOption {
	return func(p *Producer) error {
		p.config.Producer.Retry.Max = retry
		return nil
	}
}

func OptionRetryBackoff(backoff time.Duration) ProducerOption {
	return func(p *Producer) error {
		p.config.Producer.Retry.Backoff = backoff
		return nil
	}
}

func OptionRequiredAcks(requireType RequireType) ProducerOption {
	return func(p *Producer) error {
		switch requireType {
		case RequireNoResponse:
			p.config.Producer.RequiredAcks = sarama.NoResponse
		case RequireOnlyLeader:
			p.config.Producer.RequiredAcks = sarama.WaitForLocal
		case RequireAllReplicas:
			p.config.Producer.RequiredAcks = sarama.WaitForAll
		default:
			return errors.New("no such type")
		}
		return nil
	}
}

func OptionTimeout(timeout time.Duration) ProducerOption {
	return func(p *Producer) error {
		p.config.Producer.Timeout = timeout
		return nil
	}
}

func OptionManualPartition() ProducerOption {
	return func(p *Producer) error {
		p.config.Producer.Partitioner = sarama.NewManualPartitioner
		return nil
	}
}

func OptionMaxMessageBytes(size int) ProducerOption {
	return func(p *Producer) error {
		p.config.Producer.MaxMessageBytes = size
		return nil
	}
}

func OptionCompression(compressionType CompressionType) ProducerOption {
	return func(p *Producer) error {
		switch compressionType {
		case CompressionNone:
			p.config.Producer.Compression = sarama.CompressionNone
		case CompressionGZIP:
			p.config.Producer.Compression = sarama.CompressionGZIP
		case CompressionSnappy:
			p.config.Producer.Compression = sarama.CompressionSnappy
		default:
			return errors.New("no such type")
		}
		return nil
	}

}

func OptionNetTLS(config *tls.Config) ProducerOption {
	return func(p *Producer) error {
		p.config.Net.TLS.Enable = true
		p.config.Net.TLS.Config = config
		return nil
	}
}

func OptionVersion(version sarama.KafkaVersion) ProducerOption {
	return func(p *Producer) error {
		p.config.Version = version
		return nil
	}
}

type workerP struct {
	wg *sync.WaitGroup
}

func newworkerP() *workerP {
	w := &workerP{
		wg: new(sync.WaitGroup),
	}
	return w
}

func (w *workerP) work(inCh <-chan *ProducerMessage, fCh, sCh chan<- *ProducerMessage, asyncP sarama.AsyncProducer, concu uint32) {
	w.wg.Add(int(concu))
	for i := 0; i < int(concu); i++ {
		go func() {
			defer w.wg.Done()
			for msg := range inCh {
				saramaMsg := &sarama.ProducerMessage{
					Topic:     msg.Topic,
					Key:       msg.Key,
					Value:     sarama.ByteEncoder(msg.Payload),
					Headers:   msg.Headers,
					Metadata:  msg.Custom,
					Partition: msg.Partition,
				}
				asyncP.Input() <- saramaMsg
			}
		}()
	}

	if fCh != nil {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			for pe := range asyncP.Errors() {
				payload, _ := pe.Msg.Value.Encode()
				msg := &ProducerMessage{
					Topic:     pe.Msg.Topic,
					Key:       pe.Msg.Key,
					Payload:   payload,
					Headers:   pe.Msg.Headers,
					Custom:    pe.Msg.Metadata,
					Partition: pe.Msg.Partition,
					Offset:    pe.Msg.Offset,
					Error:     pe.Err,
				}
				fCh <- msg
			}
		}()
	}

	if sCh != nil {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			for saramaMsg := range asyncP.Successes() {
				payload, _ := saramaMsg.Value.Encode()
				msg := &ProducerMessage{
					Topic:     saramaMsg.Topic,
					Key:       saramaMsg.Key,
					Payload:   payload,
					Headers:   saramaMsg.Headers,
					Custom:    saramaMsg.Metadata,
					Partition: saramaMsg.Partition,
					Offset:    saramaMsg.Offset,
				}
				sCh <- msg
			}
		}()
	}
}

func (w *workerP) fini() {
	w.wg.Wait()
}
