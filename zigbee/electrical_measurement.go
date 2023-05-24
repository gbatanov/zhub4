package zigbee

import (
	"fmt"
	"log"
	"zhub4/zigbee/zcl"
)

type ElectricalMeasurementCluster struct {
	ed *EndDevice
}

// SmartPlug
func (e ElectricalMeasurementCluster) handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("ElectricalMeasurementCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {
		log.Printf("ElectricalMeasurementCluster::  attribute id =0x%04x \n", attribute.Id)
		switch zcl.ElectricalMeasurementAttribute(attribute.Id) {

		case zcl.ElectricalMeasurement_0505: // RMS Voltage V
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			e.ed.set_mains_voltage(float32(val))
			log.Printf("Device %s Voltage %0.3fV \n", e.ed.get_human_name(), e.ed.get_mains_voltage())

		case zcl.ElectricalMeasurement_0508: // RMS Current mA
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			e.ed.set_current(float32(val) / 1000)
			log.Printf("Device %s Current %0.3fA \n", e.ed.get_human_name(), e.ed.get_current())

		default:
			log.Printf("Device 0x%04x Cluster::ELECTRICAL_MEASUREMENTS  attribute.Id = 0x%04x\n", endpoint.Address, attribute.Id)
		} //switch

	} //for
	fmt.Println("")
}
