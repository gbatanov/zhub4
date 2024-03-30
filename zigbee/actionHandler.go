/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2022-2023 GSB, Georgii Batanov gbatanov@yandex.ru
MIT License
*/
package zigbee

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/foolin/goview/supports/ginview"
	"github.com/gbatanov/zhub4/zigbee/zdo"
	"github.com/gin-gonic/gin"
)

type ActionHandler struct {
	con *Controller
}

func NewActionHandler(con *Controller) *ActionHandler {
	ah := ActionHandler{con}
	return &ah
}

func (ah *ActionHandler) page404(ctx *gin.Context) {
	ginview.HTML(ctx, http.StatusNotFound, "page404.tmpl", gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
}

// Обработчик команд из web
func (ah *ActionHandler) cmdHandler(ctx *gin.Context) {
	cmnd := -1
	id := ctx.Query("off")
	if id == "" {
		id = ctx.Query("on")
		if id != "" {
			cmnd = 1
		}
	} else {
		cmnd = 0
	}
	eps := ctx.Query("ep")
	result := ""

	if len(eps) > 0 && cmnd > -1 {
		macAddress, err := strconv.ParseUint(id, 0, 64)
		if err == nil {
			ep, err := strconv.Atoi(eps)
			if err == nil {
				if macAddress == zdo.RELAY_2_WASH {
					cmnd = 1 - cmnd
				}
				ah.con.switchRelay(macAddress, uint8(cmnd), uint8(ep))
				result += fmt.Sprintf("Cmd %d to device 0x%08x endpoint %d executed", cmnd, macAddress, ep)
			}
		} else {
			result += err.Error()
		}
	}
	// HTML ответ на основе шаблона
	ginview.HTML(ctx, http.StatusOK, "command.tmpl", gin.H{"Result": result})
}

// Главная страница
func (ah *ActionHandler) otherHandler(ctx *gin.Context) {

	answer := ah.con.showDeviceStatuses()
	sTime := ah.con.formatDateTime(ah.con.startTime)
	la := ah.con.getLastMotionSensorActivity()
	lMotion := ah.con.formatDateTime(la)

	weather := ah.con.showWeather()

	// HTML ответ на основе шаблона с использованием layouts
	ginview.HTML(ctx, http.StatusOK, "index", gin.H{
		"title":      "GSB Smart Home",
		"deviceList": answer,
		"StartTime":  sTime,
		"LMotion":    lMotion,
		"Weather":    weather})

}

func (ah *ActionHandler) metrics(ctx *gin.Context) {
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
	ctx.String(http.StatusOK, "%s", answer)
}
func (ah *ActionHandler) join(ctx *gin.Context) {
	ah.con.GetZdo().PermitJoin(60 * time.Second)
	ctx.Redirect(http.StatusPermanentRedirect, "/")
}

// Обработчик control
func (ah *ActionHandler) controlHandler(ctx *gin.Context) {
	cmd := ctx.Query("cmd")
	log.Println(cmd)

	// HTML ответ на основе шаблона

	kitchen_light := "unkn"
	kitchen_vent := "unkn"
	coridor_light := "unkn"

	ginview.HTML(ctx, http.StatusOK, "control.tmpl",
		gin.H{"Kitchen_light": kitchen_light,
			"Kitchen_vent":  kitchen_vent,
			"Coridor_light": coridor_light,
		})

}
