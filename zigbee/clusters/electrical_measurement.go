package clusters

import (
	"fmt"
	"log"

	"github.com/gbatanov/zhub4/zigbee/zdo"

	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

type ElectricalMeasurementCluster struct {
	Ed *zdo.EndDevice
}

// SmartPlug
func (e ElectricalMeasurementCluster) Handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	log.Printf("ElectricalMeasurementCluster:: %s, endpoint address: 0x%04x number = %d \n", e.Ed.Get_human_name(), endpoint.Address, endpoint.Number)
	fmt.Printf("Device %s ", e.Ed.Get_human_name())
	for _, attribute := range attributes {
		switch zcl.ElectricalMeasurementAttribute(attribute.Id) {

		case zcl.ElectricalMeasurement_0505: // RMS Voltage V
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			e.Ed.Set_mains_voltage(float32(val))
			fmt.Printf(" Voltage %0.2fV ", e.Ed.Get_mains_voltage())

		case zcl.ElectricalMeasurement_0508: // RMS Current mA
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			e.Ed.Set_current(float32(val) / 1000)
			fmt.Printf(" Current %0.3fA ", e.Ed.Get_current())

		case zcl.ElectricalMeasurement_050B: // Active Power
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			e.Ed.Set_power(float32(val))
			fmt.Printf(" Active Power %0.3fW \n", float32(val))
			/*
				case zcl.ElectricalMeasurement_050F: // Apparent Power, not supported by coordinator
					val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
					//			e.Ed.Set_current(float32(val) / 1000)
					fmt.Printf("Device %s Apparent Power %0.3fA \n", e.Ed.Get_human_name(), float32(val))
			*/
		default:
			fmt.Printf("Cluster::ELECTRICAL_MEASUREMENTS::  attribute.Id = 0x%04x\n", attribute.Id)
		} //switch

	} //for
	fmt.Println("")
}
