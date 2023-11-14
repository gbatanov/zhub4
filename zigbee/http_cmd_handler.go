/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2022-2023 GSB, Georgii Batanov gbatanov@yandex.ru
MIT License
*/
package zigbee

import (
	"fmt"
	"strings"
	"time"

	"github.com/gbatanov/zhub4/zigbee/zdo"
	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

func (c *Controller) formatDateTime(la time.Time) string {
	dt := fmt.Sprintf("%d", la.Year()) + "-" +
		fmt.Sprintf("%02d", la.Month()) + "-" +
		fmt.Sprintf("%02d", la.Day()) + "  " +
		fmt.Sprintf("%02d", la.Hour()) + ":" +
		fmt.Sprintf("%02d", la.Minute()) + ":" +
		fmt.Sprintf("%02d", la.Second())
	if strings.Contains(dt, "1-01-01") {
		dt = ""
	}
	return dt
}

// Output devices state on display
func (c *Controller) showDeviceStatuses() map[int]map[uint16]WebDeviceInfo {

	var result map[int]map[uint16]WebDeviceInfo = make(map[int]map[uint16]WebDeviceInfo)
	ClimatSensors := []uint64{0x00124b000b1bb401}
	MotionSensors := zdo.GetDevicesByType(uint8(2)) // sonoff sensors
	MotionSensors = append(MotionSensors, []uint64{0x00124b0007246963, 0x00124b0014db2724, 0x00124b0009451438, 0x0c4314fffe17d8a8}...)
	WaterSensors := zdo.GetDevicesByType(uint8(5))
	WaterValves := zdo.GetDevicesByType(uint8(6))
	DoorSensors := zdo.GetDevicesByType(uint8(3))
	Relays := zdo.GetDevicesByType(uint8(9))
	Relays = append(Relays, zdo.GetDevicesByType(uint8(11))...)
	SmartPlugs := zdo.GetDevicesByType(uint8(10))
	Buttons := zdo.GetDevicesByType(uint8(1))
	Buttons = append(Buttons, zdo.GetDevicesByType(uint8(7))...)
	allDevices := [][]uint64{ClimatSensors, MotionSensors, WaterSensors, DoorSensors, Relays, SmartPlugs, WaterValves, Buttons}

	for ind, di := range allDevices {
		var resultTmp map[uint16]WebDeviceInfo = make(map[uint16]WebDeviceInfo, 0)
		for _, addr := range di {
			ed := c.getDeviceByMac(addr)
			if ed.ShortAddress != 0 && ed.Di.Available == 1 {
				resultTmp[ed.ShortAddress] = c.showOneType(ed)
			}
		}
		result[ind] = resultTmp
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
	wdi.State = ed.GetCurrentState(1)
	if ed.GetDeviceType() == 11 {
		wdi.State += "/" + ed.GetCurrentState(2)
	}
	wdi.LQ = fmt.Sprintf("%d", ed.Get_linkquality())
	if ed.Get_temperature() > -90 {
		wdi.Tmp = fmt.Sprintf("%d", ed.Get_temperature())
	} else {
		wdi.Tmp = " "
	}
	var powerSrc string = ""
	if ed.GetPowerSource() == uint8(zcl.PowerSource_BATTERY) { // battery
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
	} else if ed.GetPowerSource() == uint8(zcl.PowerSource_SINGLE_PHASE) { // 220V
		voltage := ed.GetMainsVoltage()
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
