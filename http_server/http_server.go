/*
GSB, 2023
gbatanov@yandex.ru
*/
package http_server

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
	srv        http.Server
	answerChan chan string
	queryChan  chan string
}

func Http_server_create(answerChan chan string, queryChan chan string) (*HttpServer, error) {
	var srv http.Server
	srv.Addr = "192.168.88.240:8180"
	httpServer := HttpServer{srv: srv, answerChan: answerChan, queryChan: queryChan}

	return &httpServer, nil
}

// register handlers
func (web *HttpServer) register_routing() {
	mux := http.NewServeMux()
	mux.HandleFunc("/css/", web.css_handler)
	mux.HandleFunc("/command", web.command_handler)
	mux.HandleFunc("/", web.other_handler)
	web.srv.Handler = mux
}

func (web *HttpServer) Start() error {
	//	log.Println("Web server Start()")
	web.register_routing()
	go func() {
		web.srv.ListenAndServe()
		log.Println("Web server stoped")
	}()

	return nil
}
func (web *HttpServer) Stop() {
	//	log.Println("Web server Stop()")
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
	web.send_answer(w, my_resp, 404, "text/html", headers)
}

func (web *HttpServer) other_handler(w http.ResponseWriter, r *http.Request) {
	//	log.Println("other_handler")
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	if len(r.URL.Path) > 0 && r.URL.Path != "/" {
		web.NotFound(w, r)
		return
	}
	web.main_page(w, r)

}

// TODO: path to config
func (web *HttpServer) css_handler(w http.ResponseWriter, r *http.Request) {

	var my_resp MyResponse

	css := r.URL.Path[5:]
	str, err := os.ReadFile("/usr/local/etc/zhub4/web/" + css)
	if err != nil {
		log.Println("Error open css file")
		my_resp.body = ""
	} else {
		my_resp.body = string(str)
	}
	var headers map[string]string = make(map[string]string)
	web.send_answer(w, my_resp, 200, "text/css", headers)

}
func (web *HttpServer) command_handler(w http.ResponseWriter, r *http.Request) {
	//	log.Println("command_handler")
	var my_resp MyResponse
	host := r.Host
	baseUrl, _ := url.Parse("http://" + host)

	my_resp.body = "<h3>Commands list</h3>"

	cmd := web.parse_command(r)
	web.queryChan <- cmd
	answer := <-web.answerChan
	my_resp.body += "<div>" + answer + "</div"

	my_resp.body += fmt.Sprintf("<p><a href=\"%s\">Home page</a></p>", baseUrl.String())
	var headers map[string]string = make(map[string]string)
	my_resp.head = "<title>Commands list</title>"
	my_resp.head += "<link href=\"/css/gsb_style.css\" rel=\"stylesheet\" type=\"text/css\">"
	web.send_answer(w, my_resp, 200, "text/html", headers)

}

func (web *HttpServer) parse_command(r *http.Request) string {
	return "command_list" //TODO: dummy
}

func (web *HttpServer) main_page(w http.ResponseWriter, r *http.Request) {
	//	log.Println("Http main page")
	var my_resp MyResponse
	host := r.Host
	baseUrl, _ := url.Parse("http://" + host)

	my_resp.body = ""

	web.queryChan <- "device_list"
	answer := <-web.answerChan
	my_resp.body += "<div>" + answer + "</div"
	my_resp.body += fmt.Sprintf("<br><a href=\"%s\">Commands</a>", baseUrl.String()+"/command")

	my_resp.head = "<title>Device list</title>"
	my_resp.head += "<link href=\"/css/gsb_style.css\" rel=\"stylesheet\" type=\"text/css\">"
	var headers map[string]string = make(map[string]string)
	web.send_answer(w, my_resp, 200, "text/html", headers)
}

// send answer to client
func (web *HttpServer) send_answer(w http.ResponseWriter, my_resp MyResponse, code int, mime string, headers map[string]string) {
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
	web.send_headers(w, code, headers)
	w.Write([]byte(result))
}

// get URL parameters
func (web *HttpServer) get_params(req *http.Request) (url.Values, error) {
	uri := req.RequestURI
	u, err := url.Parse(uri)
	if err != nil {
		return url.Values{}, err
	}
	m, _ := url.ParseQuery(u.RawQuery)
	return m, nil
}

// send header to client
func (web *HttpServer) send_headers(w http.ResponseWriter, code int, headers map[string]string) {

	for k, v := range headers {
		w.Header().Add(k, v)
	}
	w.WriteHeader(code)
}
