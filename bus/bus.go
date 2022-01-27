package bus

import (
	"sync"

	"github.com/jumboframes/tigerbalm"
)

type Bus interface {
	AddSlot(slot Slot)
	AddSlotHandler(slotType SlotType, handler Handler, matches ...interface{}) error
	DelSlotHandler(slotType SlotType, matches ...interface{}) error
}

type SlotBus struct {
	mu    sync.RWMutex
	slots map[SlotType]Slot
}

func NewSlotBus() *SlotBus {
	return &SlotBus{
		slots: make(map[SlotType]Slot),
	}
}

func (bus *SlotBus) AddSlot(slot Slot) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.slots[slot.Type()] = slot
}

func (bus *SlotBus) AddSlotHandler(slotType SlotType, handler Handler,
	matches ...interface{}) error {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	slot, ok := bus.slots[slotType]
	if !ok {
		return tigerbalm.ErrNoSuchSlot
	}
	slot.AddHandler(handler, matches...)
	return nil
}

func (bus *SlotBus) DelSlotHandler(slotType SlotType, matches ...interface{}) error {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	slot, ok := bus.slots[slotType]
	if !ok {
		return tigerbalm.ErrNoSuchSlot
	}
	slot.DelHandler(matches...)
	return nil
}
