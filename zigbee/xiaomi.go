/*
GSB, 2023
gbatanov@yandex.ru
*/
package zigbee

import (
	"log"
)

type XiaomiCluster struct {
	ed *EndDevice
}

func (x XiaomiCluster) handler_attributes(endpoint Endpoint, attributes []Attribute) {
	log.Printf("XiaomiCluster::endpoint address: 0x%04x number = %d \n", endpoint.address, endpoint.number)
	for _, attribute := range attributes {
		log.Printf("XiaomiCluster::attribute id =0x%04x \n", attribute.id)
		switch XiaomiAttribute(attribute.id) {
		case Xiaomi_0x00F7:
			// 03 28 1e          int8                 Device_temperature
			// 05 21 08 00       uint16                PowerOutages
			// 08 21 00 00       uint16
			// 09 21 00 00       uint16
			// 0b 28 00          uint8
			// 0с 20 00          uint8
			// 64 10 00          bool                    State    relay1
			// 65 10 00          bool                    State    relay2(for duochannel)
			// 95 39 00 00 00 00 float Single precision  // Energy for period
			// 96 39 00 00 00 00 float Single precision  Voltage
			// 97 39 00 00 00 00 float Single precision   Current
			// 98 39 00 00 00 00 float Single precision  Power    instant
			// 9a 28 00          uint8
			//
			for i := 0; i < len(attribute.value); i++ {

				attId := attribute.value[i]

				switch attId {

				case 0x03: // temperature
					i = i + 2
					log.Printf("Xiaomi temperature: %d \n", int8(attribute.value[i]))
					x.ed.set_temperature(int8(attribute.value[i]))

				case 0x05: // Power outages
					i = i + 3

				case 0x08, // uint16
					0x09: // uint16
					i = i + 3

				case 0x20: // bool
					i = i + 2

				case 0x64: // status
					i = i + 2
					state := "Off"
					if attribute.value[i] == 1 {
						state = "On"
					}
					x.ed.set_current_state(state, 1)
					log.Printf("State %s\n", state)

				case 0x65: // status2
					i = i + 2
					state := "Off"
					if attribute.value[i] == 1 {
						state = "On"
					}
					x.ed.set_current_state(state, 2)
					log.Printf("State2 %s\n", state)

				case 0x95: // energy
					i = i + 5

				case 0x96: // voltage
					value, err := x.ed.bytesToFloat32(attribute.value[i+2 : i+6])
					if err == nil {
						x.ed.set_power_source(0x01)
						x.ed.set_mains_voltage(value / 10)
						log.Printf("Voltage %0.2fV\n", value/10)
					}
					i = i + 5

				case 0x97: // current
					value, err := x.ed.bytesToFloat32(attribute.value[i+2 : i+6])
					if err == nil {
						val := value / 1000
						x.ed.set_current(val)
						log.Printf("Current %0.3fA\n", val)
					}
					i = i + 5

				case 0x98: // instant power
					value, err := x.ed.bytesToFloat32(attribute.value[i+2 : i+6])
					if err == nil {
						log.Printf("Текущая потребляемая мощность(0x98) %0.6f\n", value)
					}
					i = i + 5

				case 0x9a: // uint8
					i = i + 2

				case 0x0b: // uint8
					i = i + 2

				default:
					log.Printf("Необработанный тэг 0x%02x type 0x%02x \n ", attId, attribute.value[i+1])
					i = 1000 // big value for break
				} // switch
				if i >= len(attribute.value) {
					break
				}
			} // for

		case Xiaomi_0xFF01:
			for i := 0; i < int(attribute.size); i++ {

				switch attribute.value[i] {
				case 0x03: // device temperature
					i = i + 2
				case 0x05: // RSSI
					// rssi := int16(UINT16_(attribute.value[i+2], attribute.value[i+3]) - 90)
					i = i + 3
				}
			} //for
		default:
			log.Printf("Cluster::XIAOMI_SWITCH unknown attribute Id 0x%04x\n", attribute.id)
		} //switch
	} //for

}
