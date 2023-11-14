/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2022-2023 GSB, Georgii Batanov gbatanov@yandex.ru
MIT License
*/
package zigbee

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	srv *http.Server
	c   *Controller
}

// func NewHttpServer(addr string,
//
//	answerChan chan interface{},
//	queryChan chan map[string]string,
//	os string,
//	programDir string) (*HttpServer, error) {
func NewHttpServer(c *Controller) (*HttpServer, error) {
	httpserv := HttpServer{}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Используем Goview (Ginview - вариант для Gin)
	var vConfig goview.Config = goview.Config{
		Root:      "html/tpl",       //template root path
		Extension: ".tmpl",          //file extension
		Master:    "layouts/master", //master layout file
		//		Partials:  []string{"partials/head"}, //partial files
		DisableCache: true, //if disable cache, auto reload template file for debug.
		Delims:       goview.Delims{Left: "{{", Right: "}}"},
	}
	router.HTMLRender = ginview.New(vConfig)
	router.LoadHTMLGlob("/usr/local/etc/zhub4/web/tpl/*")

	actionHandler := NewActionHandler(c)

	router.GET("/metrics", actionHandler.metrics)
	router.GET("/join", actionHandler.join)
	router.GET("/command", actionHandler.cmdHandler)
	router.Static("/css", "/usr/local/etc/zhub4/web/css")
	router.GET("/", actionHandler.otherHandler)
	router.NoRoute(actionHandler.page404)

	httpserv.srv = &http.Server{
		Addr:    c.config.HttpAddress,
		Handler: router,
	}
	// Кастомный логгер
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s -  %s\"\n",
			param.TimeStamp.Format(time.RFC1123),
			param.ErrorMessage,
		)
	}))

	router.Use(gin.Recovery()) // Восстанавливает сервер после panic error

	return &httpserv, nil
}

// server start
func (h *HttpServer) Start() {
	go func() {
		if err := h.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("listen: %s\n", err.Error())
			errMap := make(map[string]string)
			errMap["error"] = err.Error()
			h.c.http.withHttp = false
		}
	}() // listen and serve
}

// Gracefull stop for http server
func (h *HttpServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP Server Shutdown: %s\n", err.Error())
		return
	}
	// catching ctx.Done(). timeout of 5 seconds.
	<-ctx.Done()
	log.Println("HTTP Server exiting")
}
