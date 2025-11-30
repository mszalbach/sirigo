package main

import "flag"

type config struct {
	url             string
	clientRef       string
	clientPort      string
	templateDir     string
	autoresponseDir string
	logFile         string
}

func loadConfig() config {
	var cfg config
	flag.StringVar(&cfg.url, "url", "http://localhost:8080", "URL of the SIRI endpoint")
	flag.StringVar(&cfg.clientRef, "clientref", "client", "Client Reference to use in requests")
	flag.StringVar(&cfg.clientPort, "port", ":8000", "Port where the client is listening for incoming requests")
	flag.StringVar(
		&cfg.templateDir,
		"templates",
		"templates/siri/request",
		"Folder where SIRI request templates are stored",
	)
	flag.StringVar(
		&cfg.autoresponseDir,
		"autoresponse",
		"templates/siri/autoresponse",
		"Folder where SIRI autoresponse templates are stored",
	)
	flag.StringVar(&cfg.logFile, "log", "sirigo.log", "Location of the log file")

	flag.Parse()

	return cfg
}
