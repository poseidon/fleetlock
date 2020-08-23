package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/poseidon/fleetlock"
)

var (
	// version provided by compile time -ldflags
	version = "was not built properly"
	// logger defaults to info logging
	log = logrus.New()
)

func main() {
	flags := struct {
		address  string
		logLevel string
		version  bool
		help     bool
	}{}

	flag.StringVar(&flags.address, "address", "0.0.0.0:8080", "HTTP listen address")
	// log levels https://github.com/sirupsen/logrus/blob/master/logrus.go#L36
	flag.StringVar(&flags.logLevel, "log-level", "info", "Set the logging level")
	// subcommands
	flag.BoolVar(&flags.version, "version", false, "Print version and exit")
	flag.BoolVar(&flags.help, "help", false, "Print usage and exit")

	// parse command line arguments
	flag.Parse()

	if flags.version {
		fmt.Println(version)
		return
	}

	if flags.help {
		flag.Usage()
		return
	}

	// logger
	lvl, err := logrus.ParseLevel(flags.logLevel)
	if err != nil {
		log.Fatalf("invalid log-level: %v", err)
	}
	log.Level = lvl

	// HTTP Server
	config := &fleetlock.Config{
		Logger: log,
	}
	server, err := fleetlock.NewServer(config)
	if err != nil {
		log.Fatalf("main: NewServer error %v", err)
	}

	log.Infof("main: starting fleetlock on %s", flags.address)
	err = http.ListenAndServe(flags.address, server)
	if err != nil {
		log.Fatalf("main: ListenAndServe error: %v", err)
	}
}
