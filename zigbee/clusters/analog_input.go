/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2023 GSB, Georgii Batanov gbatanov @ yandex.ru
*/

package clusters

import (
	"fmt"
	"log"
	"strings"

	"github.com/gbatanov/zhub4/zigbee/zdo"

	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

type AnalogInputCluster struct {
	Ed *zdo.EndDevice
}

func (a AnalogInputCluster) HandlerAttributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	var value float64 = -100.0
	var unit string
	log.Printf("AnalogInputCluster::%s, endpoint address: 0x%04x number = %d \n", a.Ed.GetHumanName(), endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {

		switch zcl.AnalogInputAttribute(attribute.Id) {
		case zcl.AnalogInput_0055: // value
			//
			value = float64(attribute.Value[0])
			if a.Ed.GetDeviceType() == 9 { // relay
				fmt.Printf("Analog Input Value =  %0.3f \n", value)
			} else {
				fmt.Printf("Analog Input Value =  %f \n", value)
			}

		case zcl.AnalogInput_006f:
			{
			}

		case zcl.AnalogInput_001c: // unit
			// duochannel relay hasn't unit
			unit = string(attribute.Value)
			if strings.Index(unit, ",") > 0 {
				unit = strings.Split(unit, ",")[0]
			}

		default:
			log.Printf("ANALOG_INPUT  unknown attribute Id 0x%04x,  device: 0x%04x\n", attribute.Id, endpoint.Address)
		} //switch

	} //for
	if len(unit) > 0 && value > -100.0 {
		if unit == "%" {
			a.Ed.Set_humidity(int8(value))
			a.Ed.SetCurrentState("On", endpoint.Number)
		} else if unit == "C" {
			a.Ed.Set_temperature(int8(value))
		} else if unit == "V" {
			a.Ed.Set_battery_params(0, value) // TODO ???
		} else if unit == "Pa" {
			a.Ed.Set_pressure(value)
		} else {
			log.Printf("Device 0x%04x endpoint %d Analog Input Unit =  %s \n", endpoint.Address, endpoint.Number, unit)
		}
		value = -100.0
		unit = ""
	} else if (a.Ed.GetDeviceType() == 11 || a.Ed.GetDeviceType() == 9 || a.Ed.GetDeviceType() == 10) && (value > -100.0) {
		a.Ed.Set_current(value / 100)
	}
	fmt.Println("")
}
