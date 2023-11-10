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

type PowerConfigurationCluster struct {
	Ed *zdo.EndDevice
}

func (p PowerConfigurationCluster) HandlerAttributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("PowerConfigurationCluster:: %s, endpoint address: 0x%04x number = %d \n", p.Ed.GetHumanName(), endpoint.Address, endpoint.Number)
	for _, attribute := range attributes {
		//		log.Printf("attribute id =0x%04x \n", attribute.Id)
		switch zcl.PowerConfigurationAttribute(attribute.Id) {
		case zcl.PowerConfiguration_MAINS_VOLTAGE:
			val := float32(attribute.Value[0])
			fmt.Printf("Mains voltage: %2.2fV \n", val/10)
			p.Ed.Set_mains_voltage(val)

		case zcl.PowerConfiguration_BATTERY_VOLTAGE:
			val := float32(attribute.Value[0])
			fmt.Printf("Battery voltage: %2.1fV ", val/10)
			p.Ed.Set_battery_params(0, val/10)

		case zcl.PowerConfiguration_BATTERY_REMAIN:
			val := attribute.Value[0] // 0x00-0x30 0x30-0x60 0x60-0x90 0x90-0xc8
			if val > 0xc8 {
				val = 0xc8
			}
			value := val / 2
			fmt.Printf(" remain: %d%% (0x%02x) \n\n", value, val)
			p.Ed.Set_battery_params(value, 0.0)

		default:
			fmt.Printf("unknown attribute 0x%04x \n", attribute.Id)
		} //switch
	} //for
}
