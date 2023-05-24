package zigbee

import (
	"log"
	"strings"
	"zhub4/zigbee/zdo/zcl"
)

type AnalogInputCluster struct {
	ed *EndDevice
}

func (a AnalogInputCluster) handler_attributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	var value float32 = -100.0
	var unit string
	log.Printf("AnalogInputCluster::endpoint address: 0x%04x number = %d \n", endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {
		log.Printf("attribute id =0x%04x \n", attribute.Id)
		switch zcl.AnalogInputAttribute(attribute.Id) {
		case zcl.AnalogInput_0055: // value
			//  на реле показывает суммарный ток в 0,1 А (потребляемый нагрузкой и самим реле)
			// показывает сразу после изменения нагрузки в отличие от получаемого в репортинге
			value = float32(attribute.Value[0])
			log.Printf("Device 0x%04x endpoint %d Analog Input Value =  %f \n", endpoint.Address, endpoint.Number, value)

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
			a.ed.set_humidity(int8(value))
			a.ed.set_current_state("On", endpoint.Number)
		} else if unit == "C" {
			a.ed.set_temperature(int8(value))
		} else if unit == "V" {
			a.ed.set_battery_params(0, value)
		} else if unit == "Pa" {
			a.ed.set_pressure(value)
		} else {
			log.Printf("Device 0x%04x endpoint %d Analog Input Unit =  %s \n", endpoint.Address, endpoint.Number, unit)
		}
		value = -100.0
		unit = ""
	} else if (a.ed.get_device_type() == 11 || a.ed.get_device_type() == 9 || a.ed.get_device_type() == 10) && (value > -100.0) {
		a.ed.set_current(value / 100)
	}
}
