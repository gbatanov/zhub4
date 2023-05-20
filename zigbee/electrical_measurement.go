package zigbee

import (
	"log"
)

type ElectricalMeasurementCluster struct {
	ed *EndDevice
}

// SmartPlug
func (e ElectricalMeasurementCluster) handler_attributes(endpoint Endpoint, attributes []Attribute) {
	for _, attribute := range attributes {
		log.Printf("ElectricalMeasurementCluster::  attribute id =0x%04x \n", attribute.id)
		switch ElectricalMeasurementAttribute(attribute.id) {

		case ElectricalMeasurement_0505: // RMS Voltage V
			val := UINT16_(attribute.value[0], attribute.value[1])
			e.ed.set_mains_voltage(float32(val))
			log.Printf("Device %s Voltage %0.3fV \n", e.ed.get_human_name(), e.ed.get_mains_voltage())

		case ElectricalMeasurement_0508: // RMS Current mA
			val := UINT16_(attribute.value[0], attribute.value[1])
			e.ed.set_current(float32(val) / 1000)
			log.Printf("Device %s Current %0.3fA \n", e.ed.get_human_name(), e.ed.get_current())

		default:
			log.Printf("Device 0x%04x Cluster::ELECTRICAL_MEASUREMENTS  attribute.id = 0x%04x\n", endpoint.address, attribute.id)
		} //switch
	} //for

}
