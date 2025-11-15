package main

import "flag"

type config struct {
	url         string
	clientRef   string
	clientPort  string
	templateDir string
}

func loadConfig() config {
	var cfg config
	flag.StringVar(&cfg.url, "url", "", "URL of the SIRI endpoint")
	flag.StringVar(&cfg.clientRef, "clientref", "client", "Client Reference to use in requests")
	flag.StringVar(&cfg.clientPort, "port", ":8000", "Port where the client is listening for incoming requests")
	flag.StringVar(&cfg.templateDir, "templates", "templates", "Folder where SIRI request templates are stored")

	flag.Parse()

	return cfg
}
