/*
GSB, 2023
gbatanov@yandex.ru
*/
package clusters

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"
	"zhub4/zigbee/zdo"
	"zhub4/zigbee/zdo/zcl"
)

type OnOffCluster struct {
	Ed      *zdo.EndDevice
	MsgChan chan MotionMsg
}

type MotionMsg struct {
	Ed  *zdo.EndDevice
	Cmd uint8
}

func (o OnOffCluster) Handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("OnOffCluster:: %s, endpoint address: 0x%04x number = %d \n", o.Ed.Get_human_name(), endpoint.Address, endpoint.Number)
	a0000 := false
	for _, attribute := range attributes {

		switch zcl.OnOffAttribute(attribute.Id) {
		case zcl.OnOff_ON_OFF: // 0x0000 multiple repetition
			if a0000 {
				break
			}
			a0000 = true
			b_val := false
			u_val := attribute.Value[0]
			if attribute.Value[0] == 1 {
				b_val = true
			}
			macAddress := o.Ed.Get_mac_address()
			if macAddress == 0x00124b0014db2724 {
				// custom2 coridor
				if endpoint.Number == 2 { // loght sensor
					fmt.Printf("Освещенность %d \n", u_val)
					o.Ed.Set_luminocity(int8(u_val))
				}
				if endpoint.Number == 6 { // motion sensor (1 - no motion, 0 - motion)
					fmt.Printf("Прихожая: Движение %d \n", 1-u_val)
					msg := MotionMsg{Ed: o.Ed, Cmd: 1 - u_val}
					o.MsgChan <- msg
				}
			} else if macAddress == 0x00124b0009451438 {
				// custom3 - kitchen
				if endpoint.Number == 2 { // presence sensor - kitchen
					fmt.Printf("Кухня: Присутствие %d \n", 1-u_val)
					msg := MotionMsg{Ed: o.Ed, Cmd: 1 - u_val}
					o.MsgChan <- msg
				}
			} else if macAddress == 0x0c4314fffe17d8a8 {
				// motion sensor IKEA
				fmt.Printf("Датчик движения IKEA %d \n", u_val)
				msg := MotionMsg{Ed: o.Ed, Cmd: u_val}
				o.MsgChan <- msg
			} else if macAddress == 0x00124b0007246963 {
				// Custom3
				if endpoint.Number == 2 { // light sensor(1 - high, 0 - low)
					fmt.Printf("Custom3: Освещенность %d \n", u_val)
					o.Ed.Set_luminocity(int8(u_val))
				}
				if endpoint.Number == 4 { // motion sensor (1 - no motion, 0 - motion)
					fmt.Printf("Custom3: Движение %d \n", 1-u_val)
					msg := MotionMsg{Ed: o.Ed, Cmd: 1 - u_val}
					o.MsgChan <- msg

				}
			} else if o.Ed.Get_device_type() == 10 { // SmartPlug
				currentState := o.Ed.Get_current_state(1)
				newState := "Off"
				if b_val {
					newState = "On"
				}
				fmt.Printf("SmartPlug %s \n", newState)
				if newState != currentState {
					ts := time.Now() // get time now
					o.Ed.Set_last_action(ts)
					o.Ed.Set_current_state(newState, 1)
				}
			} else if o.Ed.Get_device_type() == 11 { // duochannel relay has EP1 and EP2
				currentState := o.Ed.Get_current_state(endpoint.Number)

				newState := "Off"
				if b_val {
					newState = "On"
				}
				fmt.Printf("Duochannel relay %s channel %d\n", newState, endpoint.Number)
				if newState != currentState {
					ts := time.Now() // get time now
					o.Ed.Set_last_action(ts)
					o.Ed.Set_current_state(newState, endpoint.Number)
				}
			} else {
				currentState := o.Ed.Get_current_state(1)
				newState := "Off"
				if b_val {
					newState = "On"
				}
				if newState != currentState {
					ts := time.Now() // get time now
					o.Ed.Set_last_action(ts)
					o.Ed.Set_current_state(newState, 1)
				}
				fmt.Printf("Device 0x%04x %s endpoint %d state = %s \n", endpoint.Address, o.Ed.Get_human_name(), endpoint.Number, newState)
			}

		case zcl.OnOff_ON_TIME:
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			fmt.Printf("Device 0x%04x endpoint %d ON_TIME =  %d s \n", endpoint.Address, endpoint.Number, val)

		case zcl.OnOff_OFF_WAIT_TIME:
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			fmt.Printf("Device 0x%04x endpoint %d OFF_WAIT_TIME =  %d s \n", endpoint.Address, endpoint.Number, val)

		case zcl.OnOff_00F5: //from relay aqara T1
			//  every 30 second approximately
			//  0x03<short_addr>mm, by switch on or off
			val := binary.LittleEndian.Uint32(attribute.Value)
			fmt.Printf("Attribute Id 0x%04x in cluster ON_OFF Device 0x%04x val 0x%08x\n", attribute.Id, endpoint.Address, val)

		case zcl.OnOff_F000, // dualchannel relay, like relay in cluster 00F5
			zcl.OnOff_F500, // from relay aqara T1
			zcl.OnOff_F501: // from relay aqara T1
			val := binary.LittleEndian.Uint32(attribute.Value)
			fmt.Printf("Attribute Id 0x%04x in cluster ON_OFF Device 0x%04x val 0x%08x\n", attribute.Id, endpoint.Address, val)

		case zcl.OnOff_00F7: // ???
			val := string(attribute.Value)
			fmt.Printf("Attribute Id 0x%04x in cluster ON_OFF Device 0x%04x value: %s \n", attribute.Id, endpoint.Address, val)

		case zcl.OnOff_5000,
			zcl.OnOff_8000,
			zcl.OnOff_8001,
			zcl.OnOff_8002:
			// Valves
			fmt.Printf("Attribute Id 0x%04x in cluster ON_OFF device: 0x%04x\n", attribute.Id, endpoint.Address)
		default:
			fmt.Printf("Unused attribute Id 0x%04x in cluster ON_OFF device: 0x%04x\n", attribute.Id, endpoint.Address)
		} //switch
	} //for
	fmt.Println("")
}
