package clusters

import (
	"log"
	"zhub4/zigbee/zdo"
	"zhub4/zigbee/zdo/zcl"
)

type PowerConfigurationCluster struct {
	Ed *zdo.EndDevice
}

func (p PowerConfigurationCluster) Handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("PowerConfigurationCluster:: %s, endpoint address: 0x%04x number = %d \n", p.Ed.Get_human_name(), endpoint.Address, endpoint.Number)
	for _, attribute := range attributes {
		log.Printf("attribute id =0x%04x \n", attribute.Id)
		switch zcl.PowerConfigurationAttribute(attribute.Id) {
		case zcl.PowerConfiguration_MAINS_VOLTAGE:
			val := float32(attribute.Value[0])
			log.Printf("Device 0x%04x MAINS_VOLTAGE: %2.3f \n", endpoint.Address, val/10)
			p.Ed.Set_mains_voltage(val)

		case zcl.PowerConfiguration_BATTERY_VOLTAGE:
			val := float32(attribute.Value[0])
			log.Printf("Device 0x%04x BATTERY_VOLTAGE: %2.3f \n", endpoint.Address, val/10)
			p.Ed.Set_battery_params(0, val/10)

		case zcl.PowerConfiguration_BATTERY_REMAIN:
			val := attribute.Value[0] // 0x00-0x30 0x30-0x60 0x60-0x90 0x90-0xc8
			log.Printf("Device 0x%04x BATTERY_REMAIN: 0x%02x \n\n", endpoint.Address, val)
			p.Ed.Set_battery_params(val, 0.0)

		default:
			log.Printf(" Cluster::POWER_CONFIGURATION: unused attribute 0x%04x \n", attribute.Id)
		} //switch
	} //for
}
