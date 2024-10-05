package main

import (
	"fmt"
	"net/http"
	"os"
)

func index(w http.ResponseWriter, req *http.Request) {
	// TODO: serve from disk?
	fmt.Fprintf(w, "hello, world!")
}

func main() {
	var port_str = os.Getenv("CM_WEB_PORT")
	if len(port_str) < 1 {
		port_str = "80"
	}
	http.HandleFunc("/", index)
	var listen_spec = fmt.Sprintf(":%s", port_str)
	http.ListenAndServe(listen_spec, nil)
}
