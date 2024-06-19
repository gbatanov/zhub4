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
	Id   CommandId // CommandId
	Emit chan Command
}

type EventHandler struct {
	Events map[CommandId]Event
}

var evMtx sync.Mutex

func (eh *EventHandler) AddEvent(id CommandId) {
	evMtx.Lock()
	eh.Events[id] = Event{Id: id, Emit: make(chan Command, 1)} // unbuffered?
	evMtx.Unlock()
}

func (eh *EventHandler) EventExists(id CommandId) bool {
	var key bool
	evMtx.Lock()
	_, key = eh.Events[id]
	evMtx.Unlock()
	return key
}

func (eh *EventHandler) RemoveEvent(id CommandId) {

	evMtx.Lock()
	_, key := eh.Events[id]
	if !key {
		delete(eh.Events, id)
	}
	evMtx.Unlock()
}

func (eh *EventHandler) GetEvent(id CommandId) *Event {
	evMtx.Lock()
	val := eh.Events[id]
	evMtx.Unlock()
	return &val
}

func (eh *EventHandler) emit(id CommandId, cmd Command) {
	if eh.EventExists(id) {
		event := eh.GetEvent(id)
		event.Emit <- cmd
	}
}

// Waiting for a response
func (eh *EventHandler) wait(id CommandId, timeout time.Duration) Command {
	if !eh.EventExists(id) {
		eh.AddEvent(id)
	}
	event := eh.GetEvent(id)
	ticker := time.NewTicker(timeout)
	var cmd Command
	select {
	case cmd = <-event.Emit:
	case <-ticker.C:
		log.Printf("Wait command 0x%08x timeout", id)
		cmd = *NewCommand(0)
	}
	eh.RemoveEvent(id)
	return cmd
}

func Create_event_handler() *EventHandler {
	var eh EventHandler
	eh.Events = make(map[CommandId]Event)
	return &eh
}
