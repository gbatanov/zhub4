package zigbee

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
func (eh *EventHandler) clear(id CommandId) {
	event := eh.GetEvent(id)
	select {
	case <-event.Emit:
	default:
	}
}
func (eh *EventHandler) emit(id CommandId, cmd Command) {
	event := eh.GetEvent(id)
	//	log.Printf("emit id = 0x%04x %s \n", uint16(event.Id), event.Id.String())
	//	 log.Printf("emit cmd = 0x%02x %02x\n", uint16(cmd.Id), cmd.Payload)
	event.Emit <- cmd
}

func (eh *EventHandler) wait(id CommandId, timeout time.Duration) Command {
	//	log.Printf("wait id = 0x%04x %s\n", uint16(id), id.String())
	event := eh.GetEvent(id)
	ticker := time.NewTicker(timeout)
	select {
	case cmd := <-event.Emit:
		//		 log.Printf("wait after emit cmd = 0x%02x %02x\n", uint16(cmd.Id), cmd.Payload)
		return cmd
	case <-ticker.C:
		log.Println("ticker action")
		return *NewCommand(0)
	}
}

func CreateEventHandler() *EventHandler {
	var eh EventHandler
	eh.Events = make(map[CommandId]Event)
	return &eh
}
