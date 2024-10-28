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

type closeable_server interface {
    Serve(l net.Listener) error
    Close() error
}

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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		if err := milter_server.stop(); err != nil {
			log.Fatal("Failed to close milter: ", err)
		}
	}()

	// TODO: SQLite account database?
}
