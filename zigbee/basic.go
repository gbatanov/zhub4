package zigbee

import (
	"log"
)

type BasicCluster struct {
	ed *EndDevice
}

func (b BasicCluster) handler_attributes(endpoint Endpoint, attributes []Attribute) {
	for _, attribute := range attributes {
		log.Printf("attribute id =0x%04x \n", attribute.id)

		switch BasicAttribute(attribute.id) {
		case Basic_MANUFACTURER_NAME:
			if attribute.size > 0 {
				identifier := string(attribute.value)
				b.ed.set_manufacturer(identifier)
				log.Printf("MANUFACTURER_NAME: 0x%02x %s \n\n", endpoint.address, identifier)
			}

		case Basic_MODEL_IDENTIFIER:
			if attribute.size > 0 {
				identifier := string(attribute.value)
				b.ed.set_model_identifier(identifier)
				log.Printf("MODEL_IDENTIFIER: 0x%02x %s \n\n", endpoint.address, identifier)
			}
		case Basic_PRODUCT_CODE:
			if attribute.size > 0 {
				identifier := string(attribute.value)
				b.ed.set_product_code(identifier)
				log.Printf("PRODUCT_CODE: 0x%02x %s \n\n", endpoint.address, identifier)
			}

		case Basic_PRODUCT_LABEL,
			Basic_ZCL_VERSION,
			Basic_GENERIC_DEVICE_TYPE,
			Basic_GENERIC_DEVICE_CLASS,
			Basic_PRODUCT_URL,
			Basic_APPLICATION_VERSION,
			Basic_SW_BUILD_ID:
			{
			}
		case Basic_POWER_SOURCE: // uint8

			val := attribute.value[0]
			log.Printf("Device 0x%04x POWER_SOURCE: %d \n", endpoint.address, val)
			if val > 0 && val < 0x8f {
				b.ed.set_power_source(val)
			}

		case Basic_FF01: // string
			// water leak sensor Xiaomi. duochannel relay Aqara.

			// датчик протечек
			// 0x01 21 d1 0b // battery 3.025
			// 0x03 28 1e // температура 29 град
			// 0x04 21 a8 43  // no description
			// 0x05 21 08 00  // RSSI 128 dB UINT16
			// 0x06 24 01 00 00 00 00  // ?? UINT40
			// 08 21 06 02 // no  description
			// 09 21 00 04 // no  description
			// 0a 21 00 00  // parent NWK - coordinator (0000)
			// 64 10 00    // false - OFF
			// двухканальное реле
			// 0x03   0x28   0x1e //int8
			// 0x05   0x21   0x05   0x00
			// 0x08   0x21   0x2f   0x13
			// 0x09   0x21   0x01   0x02
			// 0x64   0x10   0x00 // bool
			// 0x65   0x10   0x00
			// 0x6e   0x20   0x00
			// 0x6f   0x20   0x00
			// 0x94   0x20   0x02  // uint8
			// 0x95   0x39   0x0a   0xd7   0xa3   0x39   // float
			// 0x96   0x39   0x58   0x48   0x0a   0x45
			// 0x97   0x39   0x00   0x30   0x68   0x3b
			// 0x98   0x39   0x00   0x00   0x00   0x00
			// 0x9b   0x21   0x00   0x00
			// 0x9c   0x20   0x01
			// 0x0a   0x21   0x00   0x00 //uint16
			// 0x0c   0x28   0x00   0x00
			for i := 0; i < len(attribute.value); i++ {
				switch attribute.value[i] {
				case 0x01: // battery voltage
					bat := float32(UINT16_(attribute.value[i+2], attribute.value[i+3]))
					i = i + 3
					b.ed.set_battery_params(0, bat/1000)

				case 0x03: // temperature
					i = i + 2
					b.ed.set_temperature(int8(attribute.value[i]))

				case 0x04:
					i = i + 3

				case 0x05: // RSSI  val - 90
					rssi := int16(UINT16_(attribute.value[i+2], attribute.value[i+3]) - 90)
					log.Printf("device 0x%04x RSSI:  %d dBm \n", endpoint.address, rssi)
					i = i + 3

				case 0x06: // ?
					i = i + 6

				case 0x08,
					0x09,
					0x0a: // nwk
					i = i + 3

				case 0x0c:
					i = i + 2

				case 0x64: // device state, channel 1
					i = i + 2
					state := "Off"
					if attribute.value[i] == 1 {
						state = "On"
					}
					b.ed.set_current_state(state, 1)

				case 0x65: //  device state, channel 2
					i = i + 2
					state := "Off"
					if attribute.value[i] == 1 {
						state = "On"
					}
					b.ed.set_current_state(state, 2)

				case 0x6e, // uint8
					0x6f,
					0x94:
					i = i + 2

				case 0x95, // energy for period
					0x98: // instant power
					i = i + 5

				case 0x96: // voltage
					value := float32(uint32(attribute.value[i+2]) + uint32(attribute.value[i+3])<<8 + uint32(attribute.value[i+4])<<16 + uint32(attribute.value[i+5])<<24)
					b.ed.set_power_source(0x01)
					b.ed.set_mains_voltage(value / 10)
					log.Printf("Напряжение:  %0.2f\n", value/10)
					i = i + 5

				case 0x97: // current
					value := float32(uint32(attribute.value[i+2]) + uint32(attribute.value[i+3])<<8 + uint32(attribute.value[i+4])<<16 + uint32(attribute.value[i+5])<<24)
					b.ed.set_current(value)
					log.Printf("Ток: %0.3f\n", value)
					i = i + 5

				case 0x9b:
					i = i + 3

				case 0x9c:
					i = i + 2
				} // switch
				if i >= len(attributes) {
					break
				}
			} // for
		}
	}
}
