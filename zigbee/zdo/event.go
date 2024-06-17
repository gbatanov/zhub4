/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2022-2024 GSB, Georgii Batanov gbatanov@yandex.ru
MIT License
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
	eh.Events[id] = Event{Id: id & 0xffff, Emit: make(chan Command, 16)}
	evMtx.Unlock()
}

func (eh *EventHandler) RemoveEvent(id uint32) {
	id = id & 0xffff
	evMtx.Lock()
	_, key := eh.Events[id]
	if !key {
		delete(eh.Events, id)
	}
	evMtx.Unlock()
}

func (eh *EventHandler) GetEvent(id uint32) *Event {
	id = id & 0xffff
	evMtx.Lock()

	_, key := eh.Events[id]
	if !key {
		eh.Events[id] = Event{Id: id, Emit: make(chan Command, 16)}
	}
	val := eh.Events[id]
	evMtx.Unlock()
	return &val
}

func (eh *EventHandler) emit(id uint32, cmd Command) {
	event := eh.GetEvent(id & 0xffff)

	event.Emit <- cmd
}

// Waiting for a response
func (eh *EventHandler) wait(id uint32, timeout time.Duration) Command {
	event := eh.GetEvent(id)
	ticker := time.NewTicker(timeout)
	select {
	case cmd := <-event.Emit:
		eh.RemoveEvent(id)
		return cmd
	case <-ticker.C:
		log.Printf("Wait command 0x%08x timeout", id&0xffff)
		eh.RemoveEvent(id)
		return *NewCommand(0)
	}
}

func Create_event_handler() *EventHandler {
	var eh EventHandler
	eh.Events = make(map[uint32]Event)
	return &eh
}
