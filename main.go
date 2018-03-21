package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"telkom"
)

const configFile = "telkombytes.conf"

type Config struct {
	MobileNumber string
}

func main() {
	// Read config
	cfg := Config{"0821234567"}
	configFormat, _ := json.Marshal(&cfg)
	cd, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Make sure configuration file '%s' exists in same "+
			"directory. Format: %s",
			configFile, configFormat)
	}
	err = json.Unmarshal(cd, &cfg)
	if err != nil || len(cfg.MobileNumber) != 10 || cfg.MobileNumber[0] != '0' {
		log.Fatalf("Config file format: %s", configFormat)
	}

	// Retrieve Telkom Mobile bundle information
	bundles, err := telkom.GetBundlesInfo(cfg.MobileNumber)
	if err != nil {
		log.Fatal(err)
	}
	for _, b := range bundles {
		log.Println(b)
	}
	log.Println(bundles.CapRemainingBytes())
}
