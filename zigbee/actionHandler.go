/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2022-2023 GSB, Georgii Batanov gbatanov@yandex.ru
MIT License
*/
package zigbee

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gbatanov/zhub4/zigbee/zdo"
	"github.com/gin-gonic/gin"
)

type ActionHandler struct {
	// answerChan chan interface{}
	// queryChan  chan map[string]string
	// os         string
	// programDir string
	con *Controller
}

func NewActionHandler(con *Controller) *ActionHandler {
	ah := ActionHandler{con}
	return &ah
}

func (ah *ActionHandler) page404(c *gin.Context) {
	c.HTML(http.StatusNotFound, "page404.tmpl", gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
}

// Обработчик команд из web
func (ah *ActionHandler) cmdHandler(c *gin.Context) {
	cmnd := -1
	id := c.Query("off")
	if id == "" {
		id = c.Query("on")
		if id != "" {
			cmnd = 1
		}
	} else {
		cmnd = 0
	}
	eps := c.Query("ep") //c.Params.ByName("cmd")
	result := ""

	if len(eps) > 0 && cmnd > -1 {
		macAddress, err := strconv.ParseUint(id, 0, 64)
		if err == nil {
			ep, err := strconv.Atoi(eps)
			if err == nil {
				ah.con.switchRelay(macAddress, uint8(cmnd), uint8(ep))
				result += fmt.Sprintf("Cmd %d to device 0x%08x executed", cmnd, macAddress)
			}
		} else {
			result += fmt.Sprintf("%s", err.Error())
		}
	}
	// HTML ответ на основе шаблона
	c.HTML(http.StatusOK, "command.tmpl", gin.H{"Result": result})
}

// Главная страница
func (ah *ActionHandler) otherHandler(c *gin.Context) {

	answer := ah.con.showDeviceStatuses()
	sTime := ah.con.formatDateTime(ah.con.startTime)
	la := ah.con.getLastMotionSensorActivity()
	lMotion := ah.con.formatDateTime(la)

	weather := ah.con.showWeather()

	// HTML ответ на основе шаблона
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title":      "GSB Smart Home",
		"deviceList": answer,
		"StartTime":  sTime,
		"LMotion":    lMotion,
		"Weather":    weather})

}

func (ah *ActionHandler) metrics(c *gin.Context) {
	answer := ""
	// Получим давление
	di := ah.con.getDeviceByMac(0x00124b000b1bb401) // датчик климата в детской
	if di != nil {
		answer = answer + di.GetPromPressure()
	}
	for _, li := range zdo.PROM_MOTION_LIST {
		//		log.Printf("0x%08x \n", li)
		di = ah.con.getDeviceByMac(li)
		if di != nil {
			answer = answer + di.GetPromMotionString()
			//			log.Println(di.GetPromMotionString())
		}
	}
	for _, li := range zdo.PROM_RELAY_LIST {
		// для сдвоенного реле показываем по отдельности
		di = ah.con.getDeviceByMac(li)
		if di != nil {
			answer = answer + di.GetPromRelayString()
		}
	}
	for _, li := range zdo.PROM_DOOR_LIST {
		di = ah.con.getDeviceByMac(li)
		if di != nil {
			answer = answer + di.GetPromDoorString()
		}
	}
	c.String(http.StatusOK, "%s", answer)
}
func (ah *ActionHandler) join(c *gin.Context) {
	ah.con.GetZdo().PermitJoin(60 * time.Second)
	ah.otherHandler(c)
}
