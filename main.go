package main

import (
	"flag"
	"log"

	"telkom/netclient"
)

var (
	username = flag.String("username", "", "Telkom username")
	password = flag.String("password", "", "Telkom password")
)

func main() {
	flag.Parse()

	services, err := netclient.GetBundles(*username, *password)
	if err != nil {
		log.Fatalf(err.Error())
	}

	for _, s := range services {
		log.Println(s)
	}
}
