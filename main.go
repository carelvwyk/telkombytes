package main

import (
	"flag"
	"fmt"
	"os"

	"telkom"
)

var mobileNum = flag.String("mobilenumber", "",
	"The mobile number of your current Telkom connection e.g. 0812134567")

func main() {
	flag.Parse()

	if *mobileNum == "" {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Retrieve Telkom Mobile bundle information
	bundles, err := telkom.GetBundlesInfo(*mobileNum)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
	fmt.Print(bundles.CapRemainingBytes())
}
