/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2022-2023 GSB, Georgii Batanov gbatanov@yandex.ru
MIT License
*/
package zigbee

import (
	"fmt"
	"time"

	"github.com/gbatanov/zhub4/zigbee/zdo"
	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

func (c *Controller) formatDateTime(la time.Time) string {
	return fmt.Sprintf("%d", la.Year()) + "-" +
		fmt.Sprintf("%02d", la.Month()) + "-" +
		fmt.Sprintf("%02d", la.Day()) + "  " +
		fmt.Sprintf("%02d", la.Hour()) + ":" +
		fmt.Sprintf("%02d", la.Minute()) + ":" +
		fmt.Sprintf("%02d", la.Second())
}

func (c *Controller) showDeviceStatuses() map[uint16]WebDeviceInfo {
	//	 var result string = ""
	var result map[uint16]WebDeviceInfo = make(map[uint16]WebDeviceInfo)
	ClimatSensors := []uint64{0x00124b000b1bb401}
	WaterSensors := []uint64{0x00158d0006e469a4, 0x00158d0006f8fc61, 0x00158d0006b86b79, 0x00158d0006ea99db}
	WaterValves := []uint64{0xa4c138d9758e1dcd, 0xa4c138373e89d731}
	MotionSensors := []uint64{0x00124b0007246963, 0x00124b0014db2724, 0x00124b0025137475, 0x00124b0024455048, 0x00124b002444d159, 0x00124b0009451438, 0x0c4314fffe17d8a8}
	DoorSensors := []uint64{0x00124b0025485ee6, 0x00124b002512a60b, 0x00124b00250bba63}
	Relays := []uint64{0x54ef44100019335b, 0x54ef441000193352, 0x54ef4410001933d3, 0x54ef44100018b523, 0x54ef4410005b2639, 0x54ef441000609dcc, 0x00158d0009414d7e}
	SmartPlugs := []uint64{0x70b3d52b6001b4a4, 0x70b3d52b6001b5d9, 0x70b3d52b60022ac9, 0x70b3d52b60022cfd}
	Buttons := []uint64{0x00124b0028928e8a, 0x00124b00253ba75f, 0x8cf681fffe0656ef}
	allDevices := [][]uint64{ClimatSensors, MotionSensors, WaterSensors, DoorSensors, Relays, SmartPlugs, WaterValves, Buttons}

	for _, di := range allDevices {
		for _, addr := range di {
			ed := c.getDeviceByMac(addr)
			if ed.ShortAddress != 0 && ed.Di.Test == 1 {
				result[ed.ShortAddress] = c.showOneType(ed)
			}
		}
	}
	return result
}
func (c *Controller) showWeather() map[uint8]WebWeatherInfo {
	result := make(map[uint8]WebWeatherInfo, 1)
	wp := []uint64{0x00124b000b1bb401, 0x00124b0007246963}
	i := uint8(0)
	for _, addr := range wp {
		ed := c.getDeviceByMac(addr) // Climat device on balconen, custom ptvo firmware
		if ed.ShortAddress != 0 {
			var wi WebWeatherInfo = WebWeatherInfo{}
			if addr == 0x00124b000b1bb401 {
				wi.Room = "Балкон"
			} else {
				wi.Room = "Детская"
			}
			if ed.Get_temperature() > -90 {
				wi.Temp = fmt.Sprintf("%d", ed.Get_temperature())
			} else {
				wi.Temp = " "
			}
			if ed.Get_humidity() > -90 {
				wi.Humidity = fmt.Sprintf("%d", ed.Get_humidity())
			} else {
				wi.Humidity = " "
			}
			if ed.Get_pressure() > -90 {
				wi.Pressure = fmt.Sprintf("%d", uint64(ed.Get_pressure()))
			} else {
				wi.Pressure = " "
			}
			result[i] = wi
			i++
		}
	}
	return result
}

func (c *Controller) showOneType(ed *zdo.EndDevice) WebDeviceInfo {
	wdi := WebDeviceInfo{}
	wdi.ShortAddr = fmt.Sprintf("0x%04x", ed.ShortAddress)
	wdi.Name = ed.GetHumanName()
	wdi.State = ed.Get_current_state(1)
	if ed.GetDeviceType() == 11 {
		wdi.State += "/" + ed.Get_current_state(2)
	}
	wdi.LQ = fmt.Sprintf("%d", ed.Get_linkquality())
	if ed.Get_temperature() > -90 {
		wdi.Tmp = fmt.Sprintf("%d", ed.Get_temperature())
	} else {
		wdi.Tmp = " "
	}
	var powerSrc string = ""
	if ed.Get_power_source() == uint8(zcl.PowerSource_BATTERY) { // battery
		batL := ed.Get_battery_level()
		batV := ed.Get_battery_voltage()
		if batV > 0 {
			powerSrc += fmt.Sprintf("%0.1fV", batV)
		}
		if batL > 0 {
			powerSrc += fmt.Sprintf(" / %d%%", batL)
		}
		if len(powerSrc) == 0 {
			powerSrc = "Battery"
		}
	} else if ed.Get_power_source() == uint8(zcl.PowerSource_SINGLE_PHASE) { // 220V
		voltage := ed.Get_mains_voltage()
		current := ed.Get_current()
		if voltage > 0 {
			powerSrc += fmt.Sprintf("%0.1fV", voltage)
		}
		if current > -1 {
			powerSrc += fmt.Sprintf(" / %0.3fA", current)
		}
		if len(powerSrc) == 0 {
			powerSrc = "Single phase"
		}
	}
	wdi.Pwr = powerSrc
	lastSeen := ed.Get_last_seen()
	lastAction := ed.Get_last_action()

	wdi.LSeen = c.formatDateTime(lastSeen) + " / " + c.formatDateTime(lastAction)

	return wdi

}
