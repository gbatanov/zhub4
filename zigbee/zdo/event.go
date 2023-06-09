/*
GSB, 2023
gbatanov@yandex.ru
*/
package zdo

import (
	"log"
	"time"
)

type Event struct {
	Id   CommandId
	Emit chan Command
}

type EventHandler struct {
	Events map[CommandId]Event
}

func (eh *EventHandler) AddEvent(id CommandId) {
	eh.Events[id] = Event{Id: id, Emit: make(chan Command)}
}

func (eh *EventHandler) GetEvent(id CommandId) *Event {
	_, key := eh.Events[id]
	if !key {
		eh.AddEvent(id)
	}
	val := eh.Events[id]
	return &val
}

func (eh *EventHandler) emit(id CommandId, cmd Command) {
	event := eh.GetEvent(id)

	event.Emit <- cmd
}

func (eh *EventHandler) wait(id CommandId, timeout time.Duration) Command {
	event := eh.GetEvent(id)
	ticker := time.NewTicker(timeout)
	select {
	case cmd := <-event.Emit:
		return cmd
	case <-ticker.C:
		log.Println("ticker action")
		return *NewCommand(0)
	}
}

func Create_event_handler() *EventHandler {
	var eh EventHandler
	eh.Events = make(map[CommandId]Event)
	return &eh
}
