/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2023 GSB, Georgii Batanov gbatanov @ yandex.ru
*/

package zdo

import (
	"log"
	"sync"
	"time"
)

type Event struct {
	Id   uint32 // address << 16 + CommandId
	Emit chan Command
}

type EventHandler struct {
	Events map[uint32]Event
}

var evMtx sync.Mutex

func (eh *EventHandler) AddEvent(id uint32) {
	evMtx.Lock()
	eh.Events[id] = Event{Id: id, Emit: make(chan Command)}
	evMtx.Unlock()
}

func (eh *EventHandler) GetEvent(id uint32) *Event {
	evMtx.Lock()
	_, key := eh.Events[id]
	if !key {
		eh.Events[id] = Event{Id: id, Emit: make(chan Command)}
	}
	val := eh.Events[id]
	evMtx.Unlock()
	return &val
}

func (eh *EventHandler) emit(id uint32, cmd Command) {
	event := eh.GetEvent(id)

	event.Emit <- cmd
}

// Waiting for a response
func (eh *EventHandler) wait(id uint32, timeout time.Duration) Command {
	event := eh.GetEvent(id)
	ticker := time.NewTicker(timeout)
	select {
	case cmd := <-event.Emit:
		return cmd
	case <-ticker.C:
		log.Printf("Wait command 0x%08x timeout", id)
		return *NewCommand(0)
	}
}

func Create_event_handler() *EventHandler {
	var eh EventHandler
	eh.Events = make(map[uint32]Event)
	return &eh
}
