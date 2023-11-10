/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2022-2023 GSB, Georgii Batanov gbatanov@yandex.ru
MIT License
*/
package httpServer

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ActionHandler struct {
	answerChan chan interface{}
	queryChan  chan map[string]string
	os         string
	programDir string
}

func NewActionHandler(answerChan chan interface{}, queryChan chan map[string]string, os string, programDir string) *ActionHandler {
	ah := ActionHandler{answerChan, queryChan, os, programDir}
	return &ah
}

func (ah *ActionHandler) page404(c *gin.Context) {
	c.HTML(http.StatusNotFound, "page404.tmpl", gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
}

// Обработчик команд из web
func (ah *ActionHandler) cmdHandler(c *gin.Context) {
	id := c.Query("id")   //c.Params.ByName("id")
	cmd := c.Query("cmd") //c.Params.ByName("cmd")

	log.Printf("%s %s", id, cmd)

	// HTML ответ на основе шаблона
	c.HTML(http.StatusOK, "index.tmpl", gin.H{"title": "GSB website", "id": id, "cmd": cmd})
}

// Главная страница
func (ah *ActionHandler) otherHandler(c *gin.Context) {
	cmdMap := make(map[string]string)
	cmdMap["device_list"] = ""
	ah.queryChan <- cmdMap // отправляем запрос и ждем ответ TODO: синхронизация
	answer := <-ah.answerChan

	// HTML ответ на основе шаблона
	c.HTML(http.StatusOK, "index.tmpl", gin.H{"title": "GSB Smart Home", "deviceList": answer.(map[uint16]WebDeviceInfo)})

}
