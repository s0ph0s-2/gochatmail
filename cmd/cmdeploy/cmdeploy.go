package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type chatmail_config struct {
	MailFullyQualifiedDomainName string
}

func do_init(fqdn string) {
	config := chatmail_config{fqdn}
	output_txt, m_err := json.Marshal(config)
	if m_err != nil {
		panic(m_err)
	}
	f, err := os.Create("./chatmail.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	f.Write(output_txt)
}

func build_website() {
	// TODO: this
}

func serve_website() {
	// TODO: this
}

func main() {
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)

	webdevCmd := flag.NewFlagSet("webdev", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("expected 'init' or 'webdev' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])
		tail := initCmd.Args()
		if len(tail) < 1 {
			fmt.Println("you have to provide the fully qualified domain name of your new chat server")
			os.Exit(1)
		}
		fqdn := tail[0]
		do_init(fqdn)
	case "webdev":
		webdevCmd.Parse(os.Args[2:])
		build_website()
		serve_website()
	default:
		fmt.Println("expected 'init' or 'webdev' subcommands")
		os.Exit(1)
	}
}
