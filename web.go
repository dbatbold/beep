// web.go - web interface for beep
package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Web struct {
	req *http.Request
	res http.ResponseWriter
}

func startWebServer(address string) {
	var err error
	if len(address) == 0 {
		address = "localhost:4444"
	}
	web := &Web{}
	ip := "localhost"
	ipport := strings.Split(address, ":")
	if len(ipport[0]) > 0 {
		ip = ipport[0]
	}
	port, err := strconv.Atoi(ipport[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid port number: %s", ipport[1])
		os.Exit(1)
	}
	address = fmt.Sprintf("%s:%d", ip, port)
	fmt.Printf("Beep is listening on http://%s/\n", address)
	err = http.ListenAndServe(address, web)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start web server:", err)
		os.Exit(1)
	}
}

func (w *Web) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	w.req = req
	w.res = res
	w.serve()
}

func (w *Web) serve() {
	switch w.req.URL.Path {
	case "/":
		w.serveHome()
	default:
		fmt.Fprint(w.res, webFileMap["pageNotFound"])
	}
}

func (w *Web) serveHome() {
	fmt.Fprint(w.res, webFileMap["underconstruction"])
}

var webFileMap = map[string]string{
	"pageNotFound":      `<html><body>Page not found</body></html>`,
	"underconstruction": `<html><body>Under construction</body></html>`,
}
