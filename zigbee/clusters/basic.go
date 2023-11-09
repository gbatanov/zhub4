/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2023 GSB, Georgii Batanov gbatanov @ yandex.ru
*/
package clusters

import (
	"fmt"
	"log"

	"github.com/gbatanov/zhub4/zigbee/zdo"

	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

type BasicCluster struct {
	Ed *zdo.EndDevice
}

func (b BasicCluster) HandlerAttributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("BasicCluster:: %s, endpoint address: 0x%04x number = %d \n", b.Ed.GetHumanName(), endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {

		switch zcl.BasicAttribute(attribute.Id) {
		case zcl.Basic_MANUFACTURER_NAME: //0x0004
			if attribute.Size > 0 {
				identifier := string(attribute.Value)
				b.Ed.Set_manufacturer(identifier)
				fmt.Printf("MANUFACTURER_NAME: 0x%02x %s \n\n", endpoint.Address, identifier)
			}

		case zcl.Basic_MODEL_IDENTIFIER: //0x0005
			if attribute.Size > 0 {
				identifier := string(attribute.Value)
				b.Ed.Set_model_identifier(identifier)
				fmt.Printf("MODEL_IDENTIFIER: 0x%02x %s \n\n", endpoint.Address, identifier)
			}
		case zcl.Basic_PRODUCT_CODE: //0x000a
			if attribute.Size > 0 {
				identifier := string(attribute.Value)
				b.Ed.Set_product_code(identifier)
				fmt.Printf("PRODUCT_CODE: 0x%02x %s \n\n", endpoint.Address, identifier)
			}
		case zcl.Basic_APPLICATION_VERSION: //0x0001
			fmt.Printf("Basic_APPLICATION_VERSION: value: %d \n\n", attribute.Value[0])
		case zcl.Basic_PRODUCT_LABEL,
			zcl.Basic_ZCL_VERSION,
			zcl.Basic_GENERIC_DEVICE_TYPE,
			zcl.Basic_GENERIC_DEVICE_CLASS,
			zcl.Basic_PRODUCT_URL,
			zcl.Basic_SW_BUILD_ID:
			{
				if attribute.Size > 0 {
					identifier := string(attribute.Value)
					fmt.Printf("attribute: 0x%04x, ep: 0x%02x value: %s \n\n", attribute.Id, endpoint.Address, identifier)
				}
			}
		case zcl.Basic_POWER_SOURCE: // uint8

			val := attribute.Value[0]
			fmt.Printf("Device 0x%04x POWER_SOURCE: %d \n", endpoint.Address, val)
			if val > 0 && val < 0x8f {
				b.Ed.Set_power_source(val)
			}

		case zcl.Basic_FF01: // string
			// water leak sensor Xiaomi. duochannel relay Aqara.

			// датчик протечек
			// 0x01 21 d1 0b // battery 3.025
			// 0x03 28 1e // температура 29 град
			// 0x04 21 a8 43  // no description
			// 0x05 21 08 00  // RSSI 128 dB UINT16
			// 0x06 24 01 00 00 00 00  // ?? UINT40
			// 08 21 06 02 // no  description
			// 09 21 00 04 // no  description
			// 0a 21 00 00  // parent NWK - zhub (0000)
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
			for i := 0; i < len(attribute.Value); i++ {
				switch attribute.Value[i] {
				case 0x01: // battery voltage
					bat := float32(zcl.UINT16_(attribute.Value[i+2], attribute.Value[i+3]))
					i = i + 3
					b.Ed.Set_battery_params(0, bat/1000)

				case 0x03: // temperature
					i = i + 2
					b.Ed.Set_temperature(int8(attribute.Value[i]))

				case 0x04:
					i = i + 3

				case 0x05: // RSSI  val - 90
					rssi := int16(zcl.UINT16_(attribute.Value[i+2], attribute.Value[i+3]) - 90)
					fmt.Printf("device 0x%04x RSSI:  %d dBm \n", endpoint.Address, rssi)
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
					if attribute.Value[i] == 1 {
						state = "On"
					}
					b.Ed.SetCurrentState(state, 1)

				case 0x65: //  device state, channel 2
					i = i + 2
					state := "Off"
					if attribute.Value[i] == 1 {
						state = "On"
					}
					b.Ed.SetCurrentState(state, 2)

				case 0x6e, // uint8
					0x6f,
					0x94:
					i = i + 2

				case 0x95, // energy for period
					0x98: // instant power
					i = i + 5

				case 0x96: // voltage
					value := float32(uint32(attribute.Value[i+2]) + uint32(attribute.Value[i+3])<<8 + uint32(attribute.Value[i+4])<<16 + uint32(attribute.Value[i+5])<<24)
					b.Ed.Set_power_source(0x01)
					b.Ed.Set_mains_voltage(value / 10)
					fmt.Printf("Voltage:  %0.2fV\n", value/10)
					i = i + 5

				case 0x97: // current
					value := float32(uint32(attribute.Value[i+2]) + uint32(attribute.Value[i+3])<<8 + uint32(attribute.Value[i+4])<<16 + uint32(attribute.Value[i+5])<<24)
					b.Ed.Set_current(value)
					fmt.Printf("Current: %0.3fA\n", value)
					i = i + 5

				case 0x9b:
					i = i + 3

				case 0x9c:
					i = i + 2
				} // switch
				if i >= len(attributes) {
					break
				}
			} //
		default:
			fmt.Printf("attribute id =0x%04x value = %q \n", attribute.Id, attribute.Value)

		}
	}
}
