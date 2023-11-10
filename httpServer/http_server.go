/*
zhub4 - Система домашней автоматизации на Go
Copyright (c) 2022-2023 GSB, Georgii Batanov gbatanov@yandex.ru
MIT License
*/
package httpServer

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type MyResponse struct {
	head string
	body string
}

type HttpServer struct {
	srv        *http.Server
	answerChan chan string
	queryChan  chan map[string]string
	os         string
	programDir string
}

func HttpServerCreate(address string, answerChan chan string, queryChan chan map[string]string, os string, programDir string) (*HttpServer, error) {
	var srv http.Server
	srv.Addr = address
	httpServer := HttpServer{&srv, answerChan, queryChan, os, programDir}

	return &httpServer, nil
}

// register handlers
func (web *HttpServer) registerRouting() {
	mux := http.NewServeMux()
	mux.HandleFunc("/css/", web.cssHandler)
	mux.HandleFunc("/command", web.commandHandler)
	mux.HandleFunc("/", web.otherHandler)
	web.srv.Handler = mux
}

func (web *HttpServer) Start() error {
	log.Println("Web server Start()")
	web.registerRouting()
	go func() {
		web.srv.ListenAndServe()
		log.Println("Web server stoped")
	}()

	return nil
}
func (web *HttpServer) Stop() {
	log.Println("Web server Stop()")
	web.srv.Shutdown(context.Background())
}

// page 404
func (web *HttpServer) NotFound(w http.ResponseWriter, r *http.Request) {
	//	log.Println("NotFound")

	host := r.Host
	var my_resp MyResponse
	var protocol string
	_, proto_redirect := r.Header["X-Forwarded-Proto"]
	if proto_redirect && r.Header["X-Forwarded-Proto"][0] == "https" {
		protocol = "https://"
	} else {
		protocol = "http://"
	}

	baseUrl, _ := url.Parse(protocol + host)
	my_resp.body = "<div>Wrong URL</div>"
	my_resp.body += fmt.Sprintf("<a href=\"%s\">Home page</a>", baseUrl.String())

	my_resp.head = "<title>Page not found</title>"
	my_resp.head += "<link href=\"/css/gsb_style.css\" rel=\"stylesheet\" type=\"text/css\">"
	var headers map[string]string = make(map[string]string)
	web.sendAnswer(w, my_resp, 404, "text/html", headers)
}

func (web *HttpServer) otherHandler(w http.ResponseWriter, r *http.Request) {
	//	log.Println("otherHandler")
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	if len(r.URL.Path) > 0 && r.URL.Path != "/" {
		web.NotFound(w, r)
		return
	}
	web.mainPage(w, r)

}

// TODO: path to config
func (web *HttpServer) cssHandler(w http.ResponseWriter, r *http.Request) {

	var my_resp MyResponse

	cssPath := "/usr/local/etc/zhub4/web/"
	if web.os == "windows" {
		cssPath = web.programDir + "\\httpServer\\"
	}

	css := r.URL.Path[5:]
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

// command handler
func (web *HttpServer) commandHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("commandHandler")

	cmdMap := make(map[string]string)
	var my_resp MyResponse
	host := r.Host
	baseUrl, _ := url.Parse("http://" + host)

	my_resp.body = "<h3>Commands list</h3>"

	cmd := r.RequestURI
	cmdMap["command_list"] = cmd
	web.queryChan <- cmdMap
	answer := <-web.answerChan
	my_resp.body += "<div>" + answer + "</div"

	my_resp.body += fmt.Sprintf("<p><a href=\"%s\">Home page</a></p>", baseUrl.String())
	var headers map[string]string = make(map[string]string)
	my_resp.head = "<title>Commands list</title>"
	my_resp.head += "<link href=\"/css/gsb_style.css\" rel=\"stylesheet\" type=\"text/css\">"
	web.sendAnswer(w, my_resp, 200, "text/html", headers)

}

func (web *HttpServer) mainPage(w http.ResponseWriter, r *http.Request) {
	//	log.Println("Http main page")
	var my_resp MyResponse
	host := r.Host
	baseUrl, _ := url.Parse("http://" + host)
	cmdMap := make(map[string]string)
	my_resp.body = ""
	cmdMap["device_list"] = ""
	web.queryChan <- cmdMap
	answer := <-web.answerChan
	my_resp.body += "<div>" + answer + "</div"
	my_resp.body += fmt.Sprintf("<br><a href=\"%s\">Commands</a>", baseUrl.String()+"/command")

	my_resp.head = "<title>Device list</title>"
	my_resp.head += "<link href=\"/css/gsb_style.css\" rel=\"stylesheet\" type=\"text/css\">"
	var headers map[string]string = make(map[string]string)
	web.sendAnswer(w, my_resp, 200, "text/html", headers)
}

// send answer to client
func (web *HttpServer) sendAnswer(w http.ResponseWriter, my_resp MyResponse, code int, mime string, headers map[string]string) {
	var result string = ""
	if mime == "text/html" {
		result = "<html>"

		result += "<head>"
		result += my_resp.head
		result += "</head><body>"
		result += my_resp.body
		result += "</body></html>"
	} else if mime == "text/css" {
		mime = "text/plain"
		result += my_resp.body
	} else {
		mime = "text/plain"
		result += my_resp.body
	}
	headers["Content-Type"] = mime + ";charset=utf-8"
	headers["Content-Length"] = strconv.Itoa(len(result))

	//	fmt.Printf("%q \n", headers)
	web.sendHeaders(w, code, headers)
	w.Write([]byte(result))
}

// send header to client
func (web *HttpServer) sendHeaders(w http.ResponseWriter, code int, headers map[string]string) {

	for k, v := range headers {
		w.Header().Add(k, v)
	}
	w.WriteHeader(code)
}
