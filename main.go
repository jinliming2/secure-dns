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
	"github.com/jinliming2/secure-dns/versions"
)

func main() {
	var (
		configFile = flag.String("config", "/etc/"+versions.PROGRAM+"/config.toml", "Config file")
		version    = flag.Bool("version", false, "Show version")
		logLevel   = flag.String("logLevel", "info", "Set log level (error, warn, info, verbose)")
		verbose    = flag.Bool("verbose", false, "Set log level to verbose")
	)

	flag.Parse()

	if *version {
		fmt.Printf("%s/%s %s %s/%s\n", versions.PROGRAM, versions.VERSION, versions.BUILDHASH, runtime.GOOS, runtime.GOARCH)
		fmt.Printf("Build with Go %s (%s)\n", runtime.Version(), runtime.Compiler)
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

	dnsClient, err := client.NewClient(logger, config)
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

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
