package bus

type Handler func(interface{})

type SlotType int

const (
	SlotHttp SlotType = iota
	SlotRedis
	SlotKafka
)

type Slot interface {
	AddHandler(handler Handler, matches ...interface{})
	DelHandler(matches ...interface{})
	Type() SlotType
}
