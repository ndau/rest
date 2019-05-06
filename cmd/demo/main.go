package main

import (
	"log"

	"github.com/oneiro-ndev/rest"
)

func main() {
	cf := rest.DefaultConfig()
	// add additional config items here if desired
	// or set new default values
	cf.AddString("passthrough", "http://localhost:9998")
	cf.SetDefault("port", 9999)
	// After this the configuration is available
	cf.Load()

	cs := &countService{
		PassthroughURL: cf.GetString("passthrough"),
	}
	server := rest.StandardSetup(cf, cs)
	if server != nil {
		// if your service requires cleanup before exiting, or can be restarted or
		// reloaded, set up the appropriate functions and pass them here.
		rest.WatchSignals(nil, rest.FatalFunc(cs, "SIGINT"), rest.FatalFunc(cs, "SIGTERM"))
		log.Fatal(server.ListenAndServe())
	}
}
