package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/jinliming2/secure-dns/client"
	"github.com/jinliming2/secure-dns/config"
)

// PROGRAM is the name of this software
const PROGRAM = "secure-dns"

// VERSION should be replaced at compile
const VERSION = "UNKNOWN"

// BUILDHASH should be replaced at compile
const BUILDHASH = ""

func main() {
	var (
		configFile = flag.String("config", "/etc/"+PROGRAM+"/config.toml", "Config file")
		version    = flag.Bool("version", false, "Show version")
		logLevel   = flag.String("logLevel", "info", "Set log level (error, warn, info, verbose)")
		verbose    = flag.Bool("verbose", false, "Set log level to verbose")
	)

	flag.Parse()

	if *version {
		fmt.Printf("%s %s\n", PROGRAM, VERSION)
		fmt.Printf("Build with Go %s\n", runtime.Version())
		return
	}

	if *verbose {
		setLogLevel("verbose")
	} else {
		setLogLevel(*logLevel)
	}

	logger.Infof("Reading configuration file: %s", *configFile)
	config, err := config.LoadConfig(*configFile)
	if err != nil {
		logger.Errorf("Error while handling configuration file: %s", err.Error())
		os.Exit(1)
	}
	if loggerConfig.Level.Level() < 0 {
		json, _ := json.MarshalIndent(*config, "", "  ")
		logger.Debugf("Configuration file: %s", json)
	}

	dnsClient := client.NewClient(logger, config)
	go func() {
		err := dnsClient.ListenAndServe(config.Config.Listen)
		if err != nil {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	<-sig
	logger.Info("Exiting")
	dnsClient.Shutdown()
}
