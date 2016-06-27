package main

import (
	//go builtin pkg
	"flag"

	//local pkg
	"github.com/Sirupsen/logrus"
)

var debugLogger *logrus.Logger = logrus.New()

func main() {
	var baseConfigFile *string
	var debugMode *bool
	var dispatcher *Dispatcher = NewDispatcher()

	//parse config file from cli argument
	baseConfigFile = flag.String("c", "/etc/ok_agent.json", "base config file path")
	debugMode = flag.Bool("d", false, "enable debug mode")
	flag.Parse()

	//create debug logger
	if *debugMode {
		debugLogger.Level = logrus.DebugLevel
	}

	dispatcher.Dispatch(*baseConfigFile)
}
