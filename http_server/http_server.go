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
)

type MyResponse struct {
	head string
	body string
}

type HttpServer struct {
	srv  http.Server
	Flag bool
}

func Http_server_create() (*HttpServer, error) {
	var srv http.Server
	srv.Addr = "192.168.88.240:8180"
	httpServer := HttpServer{srv: srv, Flag: true}

	return &httpServer, nil
}

// register handlers
func (web *HttpServer) register_routing() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", web.other_handler)
	web.srv.Handler = mux
}

func (web *HttpServer) Start() error {
	web.register_routing()
	go func() {
		web.srv.ListenAndServe()
		log.Println("Web server stoped")
	}()

	return nil
}
func (web *HttpServer) Stop() {
	web.srv.Shutdown(context.Background())
}

// page 404
func (web *HttpServer) NotFound(w http.ResponseWriter, r *http.Request) {

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
	web.send_answer(w, my_resp, 404)
}

func (web *HttpServer) other_handler(w http.ResponseWriter, r *http.Request) {

	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	if len(r.URL.Path) > 0 && r.URL.Path != "/" {
		web.NotFound(w, r)
		return
	}
	web.main_page(w, r)

}

func (web *HttpServer) main_page(w http.ResponseWriter, r *http.Request) {
	var my_resp MyResponse
	host := r.Host
	baseUrl, _ := url.Parse("http://" + host)

	my_resp.body = "<div>Device list</div>"
	//	my_resp.body += fmt.Sprintf("<a href=\"%s\">Home page</a>", baseUrl.String())
	my_resp.body += fmt.Sprintf("<br><a href=\"%s\">Commands</a>", baseUrl.String()+"/command")

	my_resp.head = "<title>Device list</title>"
	web.send_answer(w, my_resp, 200)
}

// send answer to client
func (web *HttpServer) send_answer(w http.ResponseWriter, my_resp MyResponse, code int) {

	var result string = "<html>"

	result += "<head>"
	result += my_resp.head
	result += "</head><body>"
	result += my_resp.body
	result += "</body></html>"
	send_headers(w, code)
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
func send_headers(w http.ResponseWriter, code int) {
	w.Header().Add("Content-Type", "text/html;charset=utf-8")
	w.WriteHeader(code)
}
