package main

import (
	"flag"
	"fmt"
    "github.com/cloudmqtt/mosquitto-go-auth/common"
)

func main() {

	var cost = flag.Int("c", 10, "bcrypt cost (default 10)")
	var password = flag.String("p", "", "password")

	flag.Parse()

	pwHash, err := common.Hash(*password, *cost)
	if err != nil {
		fmt.Errorf("error: %s\n", err)
	} else {
		fmt.Println(pwHash)
	}

}
