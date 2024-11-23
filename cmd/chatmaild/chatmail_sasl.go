package main

import (
	"fmt"
	"log"
	"net"

	"github.com/emersion/go-sasl"
	"github.com/foxcpp/go-dovecot-sasl"
)

type sasl_server struct {
	server   *dovecotsasl.Server
	listener net.Listener
}

func new_sasl_server(listen_uri string) (sasl_server, error) {
	server := dovecotsasl.NewServer()
	server.AddMechanism("PLAIN", dovecotsasl.Mechanism{}, func(*dovecotsasl.AuthReq) sasl.Server {
		return sasl.NewPlainServer(authenticator)
	})

	ln, err := make_listener(listen_uri)
	if err != nil {
		return sasl_server{}, fmt.Errorf("failed to set up listener for SASL server: %q", err)
	}

	log.Printf("using %s as SASL server listen socket\n", listen_uri)
	return sasl_server{server, ln}, nil
}

func (ss *sasl_server) serve() error {
	return ss.server.Serve(ss.listener)
}

func (ss *sasl_server) stop() error {
	return ss.server.Close()
}

func authenticator(_, user, pass string) error {
	return fmt.Errorf("rejecting login from %s", user)
}
