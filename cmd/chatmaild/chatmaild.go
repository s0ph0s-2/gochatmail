package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func make_listener(uri string) (net.Listener, error) {
	parts := strings.SplitN(uri, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid listen URI (missing '://' between protocol and details): %s", uri)
	}
	listenNetwork, listenAddr := parts[0], parts[1]
	return net.Listen(listenNetwork, listenAddr)
}

func main() {
	milter_listen_addr := "unix:///tmp/mandatory-encryption-milter.sock"
	sasl_listen_addr := "unix:///tmp/sasl.sock"

	milter_server, err := new_milter_server(milter_listen_addr)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := milter_server.serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	sasl_server, err := new_sasl_server(sasl_listen_addr)
	if err != nil {
		log.Fatal(err)
	}

	// Last listener must not be in a goroutine, otherwise nothing keeps the
	// program running.
	func() {
		err := sasl_server.serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		if err := milter_server.stop(); err != nil {
			log.Fatal("Failed to close milter: ", err)
		}
		if err := sasl_server.stop(); err != nil {
			log.Fatal("Failed to close SASL server: ", err)
		}
	}()

	// TODO: SQLite account database?
}
