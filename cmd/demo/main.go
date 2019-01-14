package main

import rest "github.com/oneiro-ndev/standard-webservice"

func main() {
	cf := rest.DefaultConfig()
	// add additional config items here if desired
	// or set new default values
	cf.SetDefault("port", 8008)
	cs := &countService{}
	rest.StandardMain(cf, cs)
}
