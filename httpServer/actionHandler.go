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
	os         string
	programDir string
}

func NewActionHandler(os string, programDir string) *ActionHandler {
	ah := ActionHandler{os, programDir}
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
	id := c.Query("id")   //c.Params.ByName("id")
	cmd := c.Query("cmd") //c.Params.ByName("cmd")

	log.Printf("%s %s", id, cmd)

	// HTML ответ на основе шаблона
	c.HTML(http.StatusOK, "index.tmpl", gin.H{"title": "GSB website", "id": id, "cmd": cmd})
}

/*
func (ah *ActionHandler) cssHandler(c *gin.Context) {

	var my_resp MyResponse

	cssPath := "/usr/local/etc/zhub4/web/"
	if ah.os == "windows" {
		cssPath = ah.programDir + "\\html\\"
	}

	css := c.Params.ByName("style")
	str, err := os.ReadFile(cssPath + css)
	if err != nil {
		log.Println("Error open css file")
		my_resp.body = ""
	} else {
		my_resp.body = string(str)
	}
	var headers map[string]string = make(map[string]string)
	web.sendAnswer(w, my_resp, 200, "text/css", headers)

}
*/
