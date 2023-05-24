package clusters

import (
	"fmt"
	"log"
	"zhub4/zigbee/zdo"
	"zhub4/zigbee/zdo/zcl"
)

type ElectricalMeasurementCluster struct {
	Ed *zdo.EndDevice
}

// SmartPlug
func (e ElectricalMeasurementCluster) Handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("ElectricalMeasurementCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {
		log.Printf("ElectricalMeasurementCluster::  attribute id =0x%04x \n", attribute.Id)
		switch zcl.ElectricalMeasurementAttribute(attribute.Id) {

		case zcl.ElectricalMeasurement_0505: // RMS Voltage V
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			e.Ed.Set_mains_voltage(float32(val))
			log.Printf("Device %s Voltage %0.3fV \n", e.Ed.Get_human_name(), e.Ed.Get_mains_voltage())

		case zcl.ElectricalMeasurement_0508: // RMS Current mA
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			e.Ed.Set_current(float32(val) / 1000)
			log.Printf("Device %s Current %0.3fA \n", e.Ed.Get_human_name(), e.Ed.Get_current())

		default:
			log.Printf("Device 0x%04x Cluster::ELECTRICAL_MEASUREMENTS  attribute.Id = 0x%04x\n", endpoint.Address, attribute.Id)
		} //switch

	} //for
	fmt.Println("")
}
