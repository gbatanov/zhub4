package zigbee

import (
	"log"
	"strings"
)

type AnalogInputCluster struct {
	ed *EndDevice
}

func (a AnalogInputCluster) handler_attributes(endpoint Endpoint, attributes []Attribute) {
	var value float32 = -100.0
	var unit string

	for _, attribute := range attributes {
		log.Printf("attribute id =0x%04x \n", attribute.id)
		switch AnalogInputAttribute(attribute.id) {
		case AnalogInput_0055: // value
			//  на реле показывает суммарный ток в 0,1 А (потребляемый нагрузкой и самим реле)
			// показывает сразу после изменения нагрузки в отличие от получаемого в репортинге
			value = float32(attribute.value[0])
			log.Printf("Device 0x%04x endpoint %d Analog Input Value =  %f \n", endpoint.address, endpoint.number, value)

		case AnalogInput_006f:
			{
			}

		case AnalogInput_001c: // unit
			// duochannel relay hasn't unit
			unit = string(attribute.value)
			if strings.Index(unit, ",") > 0 {
				unit = strings.Split(unit, ",")[0]
			}

		default:
			log.Printf("Coordinator::on_attribute_report: unknown attribute Id 0x%04x in cluster ANALOG_INPUT device: 0x%04x\n", attribute.id, endpoint.address)
		} //switch

	} //for
	if len(unit) > 0 && value > -100.0 {
		if unit == "%" {
			a.ed.set_humidity(int8(value))
			a.ed.set_current_state("On", endpoint.number)
		} else if unit == "C" {
			a.ed.set_temperature(int8(value))
		} else if unit == "V" {
			a.ed.set_battery_params(0, value)
		} else if unit == "Pa" {
			a.ed.set_pressure(value)
		} else {
			log.Printf("Device 0x%04x endpoint %d Analog Input Unit =  %s \n", endpoint.address, endpoint.number, unit)
		}
		value = -100.0
		unit = ""
	} else if (a.ed.get_device_type() == 11 || a.ed.get_device_type() == 9 || a.ed.get_device_type() == 10) && (value > -100.0) {
		a.ed.set_current(value / 100)
	}
}
