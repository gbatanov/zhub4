package zigbee

import (
	"log"
)

type PowerConfigurationCluster struct {
	ed *EndDevice
}

func (p PowerConfigurationCluster) handler_attributes(endpoint Endpoint, attributes []Attribute) {
	log.Printf("PowerConfigurationCluster::endpoint address: 0x%04x number = %d \n", endpoint.address, endpoint.number)
	for _, attribute := range attributes {
		log.Printf("attribute id =0x%04x \n", attribute.id)
		switch PowerConfigurationAttribute(attribute.id) {
		case PowerConfiguration_MAINS_VOLTAGE:
			val := float32(attribute.value[0])
			log.Printf("Device 0x%04x MAINS_VOLTAGE: %2.3f \n", endpoint.address, val/10)
			p.ed.set_mains_voltage(val)

		case PowerConfiguration_BATTERY_VOLTAGE:
			val := float32(attribute.value[0])
			log.Printf("Device 0x%04x BATTERY_VOLTAGE: %2.3f \n", endpoint.address, val/10)
			p.ed.set_battery_params(0, val/10)

		case PowerConfiguration_BATTERY_REMAIN:
			val := attribute.value[0] // 0x00-0x30 0x30-0x60 0x60-0x90 0x90-0xc8
			log.Printf("Device 0x%04x BATTERY_REMAIN: 0x%02x \n\n", endpoint.address, val)
			p.ed.set_battery_params(val, 0.0)

		default:
			log.Printf(" Cluster::POWER_CONFIGURATION: unused attribute 0x%04x \n", attribute.id)
		} //switch
	} //for
}
