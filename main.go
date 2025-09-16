package main

import (
	"flag"
	"os"

	"github.com/harvester/storage-validator/pkg/validation"
	"github.com/sirupsen/logrus"
)

var (
	configFile string
	debug      bool
)

func main() {
	flag.StringVar(&configFile, "config", "config.yaml", "Path to config file")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	v := &validation.ValidationRun{
		ConfigFile: configFile,
	}

	// run validation
	if err := v.Execute(); err != nil {
		logrus.Errorf("error while running validation: %v", err)
		os.Exit(1)
	}
}
