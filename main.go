package main

import (
	"net"
	"net/http"
	"os"

	"example.com/router"
)

func handleError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	var err error
	var listener net.Listener
	var port string = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	listener, err = net.Listen("tcp4", "0.0.0.0:"+port)
	handleError(err)
	err = http.Serve(listener, router.Mux)
	handleError(err)
}
