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
	Id   CommandId // CommandId (uint16)
	Emit chan Command
}
type EventAsync struct {
	Id   byte
	Emit chan byte
}

type EventHandler struct {
	Events      map[CommandId]Event // События синхронных команд
	EventsAsync map[byte]EventAsync // События асинхронных команд
}

var evMtx sync.Mutex

// Добавление  события для синхронных команд
func (eh *EventHandler) AddEvent(id CommandId) {
	evMtx.Lock()
	eh.Events[id] = Event{Id: id, Emit: make(chan Command, 1)}
	evMtx.Unlock()
}

// Добавление  события для асинхронных команд
func (eh *EventHandler) AddEventAsync(id byte) {
	evMtx.Lock()
	eh.EventsAsync[id] = EventAsync{Id: id, Emit: make(chan byte, 1)}
	evMtx.Unlock()
}

func (eh *EventHandler) EventExists(id CommandId) bool {
	var key bool
	evMtx.Lock()
	_, key = eh.Events[id]
	evMtx.Unlock()
	return key
}
func (eh *EventHandler) EventAsyncExists(id byte) bool {
	var key bool
	evMtx.Lock()
	_, key = eh.EventsAsync[id]
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

func (eh *EventHandler) RemoveEventAsync(id byte) {
	evMtx.Lock()
	_, key := eh.EventsAsync[id]
	if !key {
		delete(eh.EventsAsync, id)
	}
	evMtx.Unlock()
}

func (eh *EventHandler) GetEvent(id CommandId) *Event {
	evMtx.Lock()
	val := eh.Events[id]
	evMtx.Unlock()
	return &val
}
func (eh *EventHandler) GetEventAsync(id byte) *EventAsync {
	evMtx.Lock()
	val := eh.EventsAsync[id]
	evMtx.Unlock()
	return &val
}

func (eh *EventHandler) emit(id CommandId, cmd Command) {
	if eh.EventExists(id) {
		event := eh.GetEvent(id)
		event.Emit <- cmd
	}
}

func (eh *EventHandler) emitAsync(id byte, status byte) {
	if eh.EventAsyncExists(id) {
		event := eh.GetEventAsync(id)
		event.Emit <- status
	}
}

// Waiting for a sync response
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

// Waiting for an async response
func (eh *EventHandler) waitAsync(id byte, timeout time.Duration) byte {
	if !eh.EventAsyncExists(id) {
		eh.AddEventAsync(id)
	}
	event := eh.GetEventAsync(id)
	ticker := time.NewTicker(timeout)
	var status byte
	select {
	case status = <-event.Emit:
	case <-ticker.C:
		log.Printf("Wait status for %d timeout", id)
		status = 1 // Error
	}
	eh.RemoveEventAsync(id)
	return status
}

func Create_event_handler() *EventHandler {
	var eh EventHandler
	eh.Events = make(map[CommandId]Event)
	eh.EventsAsync = make(map[byte]EventAsync)
	return &eh
}
