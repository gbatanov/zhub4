/*
GSB, 2023
gbatanov@yandex.ru
*/
package zigbee

import (
	"encoding/binary"
	"log"
	"time"
	"zhub4/zigbee/zdo/zcl"
)

type OnOffCluster struct {
	ed      *EndDevice
	msgChan chan MotionMsg
}

type MotionMsg struct {
	ed  *EndDevice
	cmd uint8
}

func (o OnOffCluster) handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("OnOffCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)
	for _, attribute := range attributes {
		// log.Printf("OnOff attribute id =0x%04x \n", attribute.Id)
		switch zcl.OnOffAttribute(attribute.Id) {
		case zcl.OnOff_ON_OFF: // 0x0000
			b_val := false
			u_val := attribute.Value[0]
			if attribute.Value[0] == 1 {
				b_val = true
			}
			log.Printf("OnOffCluster::handler_attributes: Device 0x%04x %s endpoint %d value[0]= %d \n", endpoint.Address, o.ed.get_human_name(), endpoint.Number, u_val)
			macAddress := o.ed.get_mac_address()
			if macAddress == 0x00124b0014db2724 {
				// custom2 coridor
				if endpoint.Number == 2 { // loght sensor
					log.Printf("OnOffCluster::handler_attributes: Освещенность %d \n", u_val)
					o.ed.set_luminocity(int8(u_val))
				}
				if endpoint.Number == 6 { // motion sensor (1 - no motion, 0 - motion)
					log.Printf("Прихожая: Движение %d \n", 1-u_val)
					msg := MotionMsg{ed: o.ed, cmd: 1 - u_val}
					o.msgChan <- msg
				}
			} else if macAddress == 0x00124b0009451438 {
				// custom3 - kitchen
				if endpoint.Number == 2 { // presence sensor - kitchen
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
				if endpoint.Number == 2 { // light sensor(1 - high, 0 - low)
					log.Printf("Custom3: Освещенность %d \n", u_val)
					o.ed.set_luminocity(int8(u_val))
				}
				if endpoint.Number == 4 { // motion sensor (1 - no motion, 0 - motion)
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
				o.ed.set_current_state(newState, endpoint.Number)
			} else {
				ts := time.Now() // get time now
				newState := "Off"
				if b_val {
					newState = "On"
				}
				o.ed.set_last_action(ts)
				o.ed.set_current_state(newState, 1)
			}

		case zcl.OnOff_ON_TIME:
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			log.Printf("Device 0x%04x endpoint %d ON_TIME =  %d s \n", endpoint.Address, endpoint.Number, val)

		case zcl.OnOff_OFF_WAIT_TIME:
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			log.Printf("Device 0x%04x endpoint %d OFF_WAIT_TIME =  %d s \n", endpoint.Address, endpoint.Number, val)

		case zcl.OnOff_00F5: //from relay aqara T1
			//  every 30 second approximately
			//  0x03<short_addr>mm, by switch on or off
			val := binary.LittleEndian.Uint32(attribute.Value)
			log.Printf("OnOffCluster::handler_attributes: attribute Id 0x%04x in cluster ON_OFF Device 0x%04x val 0x%08x\n", attribute.Id, endpoint.Address, val)

		case zcl.OnOff_F000, // dualchannel relay, like relay in cluster 00F5
			zcl.OnOff_F500, // from relay aqara T1
			zcl.OnOff_F501: // from relay aqara T1
			val := binary.LittleEndian.Uint32(attribute.Value)
			log.Printf("OnOffCluster::handler_attributes: attribute Id 0x%04x in cluster ON_OFF Device 0x%04x val 0x%08x\n", attribute.Id, endpoint.Address, val)

		case zcl.OnOff_00F7: // ???
			val := string(attribute.Value)
			log.Printf("OnOffCluster::handler_attributes: attribute Id 0x%04x in cluster ON_OFF Device 0x%04x value: %s \n", attribute.Id, endpoint.Address, val)

		case zcl.OnOff_5000,
			zcl.OnOff_8000,
			zcl.OnOff_8001,
			zcl.OnOff_8002:
			// Valves
			log.Printf("OnOffCluster::handler_attributes: attribute Id 0x%04x in cluster ON_OFF device: 0x%04x\n", attribute.Id, endpoint.Address)
		default:
			log.Printf("OnOffCluster::handler_attributes: unused attribute Id 0x%04x in cluster ON_OFF device: 0x%04x\n", attribute.Id, endpoint.Address)
		} //switch
	} //for

}
