/*
GSB, 2023
gbatanov@yandex.ru
*/
package zigbee

import (
	"encoding/binary"
	"log"
	"time"
)

type OnOffCluster struct {
	ed      *EndDevice
	msgChan chan MotionMsg
}

type MotionMsg struct {
	ed  *EndDevice
	cmd uint8
}

func (o OnOffCluster) handler_attributes(endpoint Endpoint, attributes []Attribute) {
	log.Printf("OnOffCluster::endpoint address: 0x%04x number = %d \n", endpoint.address, endpoint.number)
	for _, attribute := range attributes {
		// log.Printf("OnOff attribute id =0x%04x \n", attribute.id)
		switch OnOffAttribute(attribute.id) {
		case OnOff_ON_OFF: // 0x0000
			b_val := false
			u_val := attribute.value[0]
			if attribute.value[0] == 1 {
				b_val = true
			}
			log.Printf("OnOffCluster::handler_attributes: Device 0x%04x %s endpoint %d value[0]= %d \n", endpoint.address, o.ed.get_human_name(), endpoint.number, u_val)
			macAddress := o.ed.get_mac_address()
			if macAddress == 0x00124b0014db2724 {
				// custom2 coridor
				if endpoint.number == 2 { // loght sensor
					log.Printf("OnOffCluster::handler_attributes: Освещенность %d \n", u_val)
					o.ed.set_luminocity(int8(u_val))
				}
				if endpoint.number == 6 { // motion sensor (1 - no motion, 0 - motion)
					log.Printf("Прихожая: Движение %d \n", 1-u_val)
					msg := MotionMsg{ed: o.ed, cmd: 1 - u_val}
					o.msgChan <- msg
				}
			} else if macAddress == 0x00124b0009451438 {
				// custom3 - kitchen
				if endpoint.number == 2 { // presence sensor - kitchen
					log.Printf("Кухня: Присутствие %d \n", 1-u_val)
					msg := MotionMsg{ed: o.ed, cmd: 1 - u_val}
					o.msgChan <- msg
				}
			} else if macAddress == 0x0c4314fffe17d8a8 {
				// motion sensor IKEA
				log.Printf("датчик движения IKEA %d \n", u_val)
				msg := MotionMsg{ed: o.ed, cmd: u_val}
				o.msgChan <- msg
			} else if macAddress == 0x00124b0007246963 {
				// Custom3
				if endpoint.number == 2 { // light sensor(1 - high, 0 - low)
					log.Printf("Custom3: Освещенность %d \n", u_val)
					o.ed.set_luminocity(int8(u_val))
				}
				if endpoint.number == 4 { // motion sensor (1 - no motion, 0 - motion)
					log.Printf("Custom3: Движение %d \n", 1-u_val)
					msg := MotionMsg{ed: o.ed, cmd: 1 - u_val}
					o.msgChan <- msg

				}
			} else if o.ed.get_device_type() == 10 { // SmartPlug
				currentState := o.ed.get_current_state(1)
				newState := "Off"
				if b_val {
					newState = "On"
				}
				if newState != currentState {
					ts := time.Now() // get time now
					o.ed.set_last_action(ts)
					o.ed.set_current_state(newState, 1)
				}
			} else if o.ed.get_device_type() == 11 { // duochannel relay has EP1 and EP2
				ts := time.Now() // get time now
				newState := "Off"
				if b_val {
					newState = "On"
				}
				o.ed.set_last_action(ts)
				o.ed.set_current_state(newState, endpoint.number)
			} else {
				ts := time.Now() // get time now
				newState := "Off"
				if b_val {
					newState = "On"
				}
				o.ed.set_last_action(ts)
				o.ed.set_current_state(newState, 1)
			}

		case OnOff_ON_TIME:
			val := UINT16_(attribute.value[0], attribute.value[1])
			log.Printf("Device 0x%04x endpoint %d ON_TIME =  %d s \n", endpoint.address, endpoint.number, val)

		case OnOff_OFF_WAIT_TIME:
			val := UINT16_(attribute.value[0], attribute.value[1])
			log.Printf("Device 0x%04x endpoint %d OFF_WAIT_TIME =  %d s \n", endpoint.address, endpoint.number, val)

		case OnOff_00F5: //from relay aqara T1
			//  every 30 second approximately
			//  0x03<short_addr>mm, by switch on or off
			val := binary.LittleEndian.Uint32(attribute.value)
			log.Printf("OnOffCluster::handler_attributes: attribute Id 0x%04x in cluster ON_OFF Device 0x%04x val 0x%08x\n", attribute.id, endpoint.address, val)

		case OnOff_F000, // dualchannel relay, like relay in cluster 00F5
			OnOff_F500, // from relay aqara T1
			OnOff_F501: // from relay aqara T1
			val := binary.LittleEndian.Uint32(attribute.value)
			log.Printf("OnOffCluster::handler_attributes: attribute Id 0x%04x in cluster ON_OFF Device 0x%04x val 0x%08x\n", attribute.id, endpoint.address, val)

		case OnOff_00F7: // ???
			val := string(attribute.value)
			log.Printf("OnOffCluster::handler_attributes: attribute Id 0x%04x in cluster ON_OFF Device 0x%04x value: %s \n", attribute.id, endpoint.address, val)

		case OnOff_5000,
			OnOff_8000,
			OnOff_8001,
			OnOff_8002:
			// Valves
			log.Printf("OnOffCluster::handler_attributes: attribute Id 0x%04x in cluster ON_OFF device: 0x%04x\n", attribute.id, endpoint.address)
		default:
			log.Printf("OnOffCluster::handler_attributes: unused attribute Id 0x%04x in cluster ON_OFF device: 0x%04x\n", attribute.id, endpoint.address)
		} //switch
	} //for

}
