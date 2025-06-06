package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"

	"example.com/entity"
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
	var srv http.Server = http.Server{
		Handler: router.Mux,
	}

	if port == "" {
		port = "8080"
	}
	listener, err = net.Listen("tcp4", "0.0.0.0:"+port)
	handleError(err)

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		err := srv.Shutdown(context.Background())
		handleError(err)
		entity.GracefulShutdown()
		//os.Exit(0)
		close(idleConnsClosed)
	}()

	err = srv.Serve(listener)
	if err != http.ErrServerClosed {
		handleError(err)
	}

	<-idleConnsClosed
}
