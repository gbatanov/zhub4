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

type ElectricalMeasurementCluster struct {
	Ed          *zdo.EndDevice
	ChargerChan chan MotionMsg
}

// SmartPlug, double relay
func (e ElectricalMeasurementCluster) HandlerAttributes(endpoint zcl.Endpoint, attributes []zcl.Attribute) {
	//	log.Printf("ElectricalMeasurementCluster:: %s, endpoint address: 0x%04x number = %d \n", e.Ed.GetHumanName(), endpoint.Address, endpoint.Number)

	for _, attribute := range attributes {
		// log.Printf("ElectricalMeasurementCluster:: Datatype: 0x%02x, Value: 0x%04x", attribute.Datatype, attribute.Value)
		if attribute.Value[0] == 0xff && attribute.Value[1] == 0xff {
			continue
		}
		switch zcl.ElectricalMeasurementAttribute(attribute.Id) {

		case zcl.ElectricalMeasurement_0505: // RMS Voltage V
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			e.Ed.SetMainsVoltage(float64(val))
			// log.Printf(" Voltage %0.2fV ", e.Ed.GetMainsVoltage())

		case zcl.ElectricalMeasurement_0508: // RMS Current mA
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			e.Ed.SetCurrent(float64(val) / 1000)
			if e.Ed.MacAddress == zdo.PLUG_1_CHARGER {
				e.checkCharger(val)
			}
			// log.Printf(" Current %0.3fA ", e.Ed.Get_current())

		case zcl.ElectricalMeasurement_050B: // Active Power
			val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
			e.Ed.Set_power(float64(val))
			//			fmt.Printf(" Active Power %0.3fW \n", float32(val))
			/*
				case zcl.ElectricalMeasurement_050F: // Apparent Power, not supported by coordinator
					val := zcl.UINT16_(attribute.Value[0], attribute.Value[1])
					//			e.Ed.SetCurrent(float32(val) / 1000)
					fmt.Printf("Device %s Apparent Power %0.3fA \n", e.Ed.GetHumanName(), float32(val))
			*/
		default:
			//			fmt.Printf("Cluster::ELECTRICAL_MEASUREMENTS::  attribute.Id = 0x%04x\n", attribute.Id)
		} //switch

	} //for
	// fmt.Println("")
}

// Current in millampers
func (e ElectricalMeasurementCluster) checkCharger(val uint16) {
	// Розетка должна быть включена
	// если chargerOn == false и ток больше 20 мА, значит зарядник включен. Ставим признак chargerOn = true
	// если chargerOn == true и ток больше 20 мА, продолжаем зарядку
	// если chargerOn == true и ток меньше 20 мА, выключаем розетку
	if e.Ed.GetCurrentState(1) == "On" {
		if !e.Ed.ChargerOn && val > 12 {
			msg := MotionMsg{Ed: e.Ed, Cmd: 1} // для отправки в телеграм
			e.ChargerChan <- msg
			log.Println("Заряд включен")
			e.Ed.ChargerOn = true
		} else if e.Ed.ChargerOn && val < 6 {
			log.Println("Заряд выключен")
			msg := MotionMsg{Ed: e.Ed, Cmd: 0}
			e.ChargerChan <- msg
			e.Ed.ChargerOn = false
		}
	}
	if e.Ed.GetCurrentState(1) == "Off" {
		e.Ed.ChargerOn = false
	}

}
