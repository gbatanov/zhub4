/*
GSB, 2023
gbatanov@yandex.ru
*/
package zigbee

import (
	"fmt"
	"time"
	"zhub4/pi4"
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

	if pi4.Pi4Available {
		boardTemperature := c.get_board_temperature()
		if boardTemperature > -100.0 {
			bt := fmt.Sprintf("%d", boardTemperature)
			result += "<p>" + "<b>Температура платы управления: </b>"
			result += bt + "</p>"
		}
	}
	//#ifdef WITH_SIM800
	//   result = result + "<p>" + zhub->show_sim800_battery() + "</p>";
	//#endif
	la := c.get_last_motion_sensor_activity()
	result += "<p>Время последнего срабатывания датчиков движения: "
	result += c.formatDateTime(la) + "</p>"
	result += "<p>Старт программы: " + c.formatDateTime(c.startTime) + "</p>"
	/*
	   std::string list = zhub->show_device_statuses(true);
	   result = result + "<h3>Список устройств:</h3>";
	   if (list.empty())
	       result = result + "<p>(устройства отсутствуют)</p>";
	   else
	       result = result + list;

	   result = result + "<p><a href=\"/command\">Список команд</a></p>";
	*/
	return result
}

func (c *Controller) get_board_temperature() int {
	return -25
}
func (c *Controller) create_command_list() string {

	var result string = "Command list will be here"
	return result
}

func (c *Controller) formatDateTime(la time.Time) string {
	return fmt.Sprintf("%d", la.Year()) + "-" +
		fmt.Sprintf("%02d", la.Month()) + "-" +
		fmt.Sprintf("%02d", la.Day()) + "  " +
		fmt.Sprintf("%02d", la.Hour()) + ":" +
		fmt.Sprintf("%02d", la.Minute()) + ":" +
		fmt.Sprintf("%02d", la.Second())
}
