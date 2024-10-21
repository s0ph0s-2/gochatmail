package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/emersion/go-milter"
)

func main() {
	listenURI := "unix:///tmp/mandatory-encryption-milter.sock"
	parts := strings.SplitN(listenURI, "://", 2)
	if len(parts) != 2 {
		log.Fatal("Invalid listen URI")
	}
	listenNetwork, listenAddr := parts[0], parts[1]

	server := milter.Server{
		NewMilter: func() milter.Milter {
			return &ChatmailMilter{}
		},
		Protocol: milter.OptNoConnect | milter.OptNoHelo,
	}
	ln, err := net.Listen(listenNetwork, listenAddr)
	if err != nil {
		log.Fatal("Failed to set up listener: ", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		if err := server.Close(); err != nil {
			log.Fatal("Failed to close server: ", err)
		}
	}()

	log.Println("Mandatory Encryption milter starting at", listenURI)
	if err = server.Serve(ln); err != nil && err != milter.ErrServerClosed {
		log.Fatal("Failed to start milter: ", err)
	}
	// TODO: Dovecot SASL socket?
	// TODO: SQLite account database?
}
