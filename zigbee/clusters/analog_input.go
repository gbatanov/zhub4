/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2023 GSB, Georgii Batanov gbatanov @ yandex.ru
*/

package clusters

import (
	"log"

	"github.com/gbatanov/zhub4/zigbee/zdo"

	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

type AnalogInputCluster struct {
	Ed *zdo.EndDevice
}

// Датчик климата
func (a AnalogInputCluster) HandlerAttributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	var value float32 = -100.0
	var val64 float64 = -100.0
	var unit []byte = make([]byte, 0)
	var err error
	//	log.Printf("AnalogInputCluster::%s, endpoint address: 0x%04x number = %d \n", a.Ed.GetHumanName(), endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {

		switch zcl.AnalogInputAttribute(attribute.Id) {
		case zcl.AnalogInput_0055: // value
			//
			val64, err = a.Ed.Bytes_to_float64(attribute.Value) // работает нормально
			if err == nil {
				value = float32(val64)
				/*
					if a.Ed.GetDeviceType() == 9 { // relay
						log.Printf("Relay Analog Input Value =  %0.3f \n", value) // current
					} else {
						log.Printf("Device Analog Input Value =  %f \n", value)
					}
				*/
			}
		case zcl.AnalogInput_006f:
			{
			}

		case zcl.AnalogInput_001c: // unit
			// duochannel relay hasn't unit
			//			log.Printf("ANALOG_INPUT unit %c", attribute.Value[0])
			unit = attribute.Value

		default:
			log.Printf("ANALOG_INPUT  unknown attribute Id 0x%04x,  device: 0x%04x\n", attribute.Id, endpoint.Address)
		} //switch

	} //for
	//	log.Printf("Analog Input Device 0x%04x endpoint %d  Unit =  %s  Value %f \n", endpoint.Address, endpoint.Number, unit, value)
	if len(unit) > 0 && value > -100.0 {
		//		log.Printf("Analog Input Device 0x%04x endpoint %d  Unit =  %s  Value %f \n", endpoint.Address, endpoint.Number, unit, value)

		if unit[0] == '%' {
			a.Ed.Set_humidity(int8(value))
			a.Ed.SetCurrentState("On", endpoint.Number)
		} else if unit[0] == 'C' {
			a.Ed.Set_temperature(int8(value))
		} else if unit[0] == 'V' {
			if a.Ed.Di.PowerSource == zcl.PowerSource_BATTERY {
				a.Ed.Set_battery_params(0, val64)
			} else {
				a.Ed.SetMainsVoltage(val64)
			}
		} else if unit[0] == 'P' {
			a.Ed.Set_pressure(float64(value))
		} else {
			//			log.Printf("Device 0x%04x endpoint %d Analog Input Unit =  %s \n", endpoint.Address, endpoint.Number, unit)
		}
		value = -100.0
		unit[0] = 0
	} else if (a.Ed.GetDeviceType() == 11 || a.Ed.GetDeviceType() == 9 || a.Ed.GetDeviceType() == 10) && (value > -100.0) {
		a.Ed.SetCurrent(float64(value / 100))
	}

}
