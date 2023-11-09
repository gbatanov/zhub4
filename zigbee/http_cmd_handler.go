/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2023 GSB, Georgii Batanov gbatanov @ yandex.ru
*/
package zigbee

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gbatanov/zhub4/zigbee/zdo"
	"github.com/gbatanov/zhub4/zigbee/zdo/zcl"
)

func (c *Controller) handleHttpQuery(cmdFromHttp string) string {
	switch cmdFromHttp {
	case "device_list":
		return c.create_device_list()
	case "command_list":
		return c.create_command_list()
	default:
		return "Unknown request"
	}

}
func (c *Controller) create_device_list() string {

	var result string = ""

	boardTemperature := c.get_board_temperature()
	if boardTemperature > -100.0 {
		bt := fmt.Sprintf("%d", boardTemperature)
		result += "<p>" + "<b>Температура платы управления: </b>"
		result += bt + "</p>"
	}

	//#ifdef WITH_SIM800
	//   result = result + "<p>" + zhub->show_sim800_battery() + "</p>";
	//#endif
	result += "<p>Старт программы: " + c.format_date_time(c.startTime) + "</p>"
	la := c.get_last_motion_sensor_activity()
	result += "<p>Время последнего срабатывания датчиков движения: "
	result += c.format_date_time(la) + "</p>"
	result += c.show_device_statuses()

	return result
}

func (c *Controller) get_board_temperature() int {
	if strings.ToLower(c.config.Os) == "darwin" {
		return -100
	}
	dat, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
	if err != nil {
		fmt.Println("get_board_temperature:: OpenFile error: ", err)
		return -200.0
	} else {
	}
	var temp_f int

	n, err := fmt.Sscanf(string(dat), "%d", &temp_f)
	if err != nil || n == 0 {

		return -200.0
	}
	return int(temp_f / 1000)

}

func (c *Controller) create_command_list() string {
	var params string = ""
	fmt.Println(params)
	mapParams, _ := url.ParseQuery(params)
	fmt.Println(mapParams)
	_, idExists := mapParams["id"]
	_, cmdExists := mapParams["cmd"]

	if idExists && cmdExists {
		fmt.Println(mapParams["id"][0])
		macAddress, err := strconv.ParseUint(mapParams["id"][0], 0, 64)
		if err == nil {
			fmt.Println(macAddress)
			cmnd, err := strconv.Atoi(mapParams["cmd"][0])
			if err == nil {
				c.switch_relay(macAddress, uint8(cmnd), 1)
			}
		} else {
			fmt.Println(err)
		}
	}

	var result string = "<p>Relay 6 <a href=\"/command?id=0x54ef441000609dcc&cmd=1\">On</a>&nbsp;<a href=\"/command?id=0x54ef441000609dcc&cmd=0\">Off</a></p>"
	return result
}

func (c *Controller) format_date_time(la time.Time) string {
	return fmt.Sprintf("%d", la.Year()) + "-" +
		fmt.Sprintf("%02d", la.Month()) + "-" +
		fmt.Sprintf("%02d", la.Day()) + "  " +
		fmt.Sprintf("%02d", la.Hour()) + ":" +
		fmt.Sprintf("%02d", la.Minute()) + ":" +
		fmt.Sprintf("%02d", la.Second())
}
func (c *Controller) show_device_statuses() string {
	var result string = ""
	ClimatSensors := []uint64{0x00124b000b1bb401}
	WaterSensors := []uint64{0x00158d0006e469a4, 0x00158d0006f8fc61, 0x00158d0006b86b79, 0x00158d0006ea99db}
	WaterValves := []uint64{0xa4c138d9758e1dcd, 0xa4c138373e89d731}
	MotionSensors := []uint64{0x00124b0007246963, 0x00124b0014db2724, 0x00124b0025137475, 0x00124b0024455048, 0x00124b002444d159, 0x00124b0009451438, 0x0c4314fffe17d8a8}
	DoorSensors := []uint64{0x00124b0025485ee6, 0x00124b002512a60b, 0x00124b00250bba63}
	Relays := []uint64{0x54ef44100019335b, 0x54ef441000193352, 0x54ef4410001933d3, 0x54ef44100018b523, 0x54ef4410005b2639, 0x54ef441000609dcc, 0x00158d0009414d7e}
	SmartPlugs := []uint64{0x70b3d52b6001b4a4, 0x70b3d52b6001b5d9, 0x70b3d52b60022ac9, 0x70b3d52b60022cfd}
	Buttons := []uint64{0x00124b0028928e8a, 0x00124b00253ba75f, 0x8cf681fffe0656ef}
	allDevices := [][]uint64{ClimatSensors, MotionSensors, WaterSensors, DoorSensors, Relays, SmartPlugs, WaterValves, Buttons}
	result += "<table>"
	result += "<tr><th>Адрес</th><th>Название</th><th>Статус</th><th>LQ</th><th>Температура<th>Питание</th><th>Last seen/action</th></tr>"

	for _, di := range allDevices {
		result += "<tr class='empty'><td colspan='8'><hr></td></tr>"
		for _, addr := range di {
			ed := c.get_device_by_mac(addr)
			if ed.ShortAddress != 0 && ed.Di.Test == 1 {
				result += c.show_one_type(ed)
			}
		}
	}
	result += "<tr class='empty'><td colspan='8'><hr></td></tr>"
	result += "</table>"
	addResult := "<table>"
	addResult += "<th>Комната</th><th>Температура</th><th>Влажность</th><th>Давление</th>"
	tmpResult := ""
	ed := c.get_device_by_mac(0x00124b000b1bb401) // Climat device on balconen, custom ptvo firmware
	if ed.ShortAddress != 0 && ed.Di.Test == 1 {

		tmpResult += "<tr><td>Балкон</td>"
		if ed.Get_temperature() > -90 {
			tmpResult += "<td>" + fmt.Sprintf("%d", ed.Get_temperature()) + "</td>"
		} else {
			tmpResult += "<td>&nbsp;</td>"
		}
		if ed.Get_humidity() > -90 {
			tmpResult += "<td>" + fmt.Sprintf("%d", ed.Get_humidity()) + "</td>"
		} else {
			tmpResult += "<td>&nbsp;</td>"
		}
		if ed.Get_pressure() > -90 {
			tmpResult += "<td>" + fmt.Sprintf("%d", uint64(ed.Get_pressure())) + "</td>"
		} else {
			tmpResult += "<td>&nbsp;</td></tr>"
		}
	}
	ed = c.get_device_by_mac(0x00124b0007246963) // Climat device on children room, custom ptvo firmware
	if ed.ShortAddress != 0 && ed.Di.Test == 1 {

		tmpResult += "<tr><td>Детская</td>"
		if ed.Get_temperature() > -90 {
			tmpResult += "<td>" + fmt.Sprintf("%d", ed.Get_temperature()) + "</td>"
		} else {
			tmpResult += "<td>&nbsp;</td>"
		}
		if ed.Get_humidity() > -90 {
			tmpResult += "<td>" + fmt.Sprintf("%d", ed.Get_humidity()) + "</td>"
		} else {
			tmpResult += "<td>&nbsp;</td>"
		}
		if ed.Get_pressure() > -90 {
			tmpResult += "<td>" + fmt.Sprintf("%d", uint64(ed.Get_pressure())) + "</td>"
		} else {
			tmpResult += "<td>&nbsp;</td></tr>"
		}
	}
	if len(tmpResult) > 0 {
		result += addResult + tmpResult + "</table>"
	}
	return result
}

func (c *Controller) show_one_type(ed *zdo.EndDevice) string {
	var result string = "<tr>"
	result += "<td class='addr'>" + fmt.Sprintf("0x%04x", ed.ShortAddress) +
		"</td><td>" + ed.Get_human_name() + "</td><td>"
	result += ed.Get_current_state(1)
	if ed.Get_device_type() == 11 {
		result += "/" + ed.Get_current_state(2)
	}

	result += "</td><td class='lq'>" + fmt.Sprintf("%d", ed.Get_linkquality()) + "</td>"
	if ed.Get_temperature() > -90 {
		result += "<td class='temp'>" + fmt.Sprintf("%d", ed.Get_temperature()) + "</td>"
	} else {
		result += "<td>&nbsp;</td>"
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

	result += "<td>&nbsp;"
	if len(powerSrc) > 0 {
		result += powerSrc
	}
	result += "</td>"

	lastSeen := ed.Get_last_seen()
	lastAction := ed.Get_last_action()

	result += "<td>&nbsp;" + c.format_date_time(lastSeen) + " / " + c.format_date_time(lastAction) + "</td></tr>"
	return result
}
