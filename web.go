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
	ip := "localhost" // serve locally by default
	port := 4444

	parts := strings.Split(address, ":")
	if len(parts) == 2 {
		ip = parts[0]
		if len(parts[1]) > 0 {
			port, err = strconv.Atoi(parts[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid port number: %v", parts[1])
				os.Exit(1)
			}
		}
	}
	address = fmt.Sprintf("%s:%d", ip, port)
	fmt.Printf("Listening on http://%s/\n", address)

	web := &Web{}
	err = http.ListenAndServe(address, web)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to start web server:", err)
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
