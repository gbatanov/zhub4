/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2022-2023 GSB, Georgii Batanov gbatanov@yandex.ru
MIT License
*/
package zigbee

import (
	"context"
	"log"
	"net/http"
	"time"

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
	router := gin.Default()
	router.LoadHTMLGlob("/usr/local/etc/zhub4/web/*")

	actionHandler := NewActionHandler(c)

	router.GET("/metrics", actionHandler.metrics)
	router.GET("/join", actionHandler.join)
	router.GET("/command", actionHandler.cmdHandler)
	router.Static("/css", "/usr/local/etc/zhub4/web")
	router.GET("/", actionHandler.otherHandler)
	router.NoRoute()
	httpserv.srv = &http.Server{
		Addr:    c.config.HttpAddress,
		Handler: router,
	}

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
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("HTTP Server exiting")
}