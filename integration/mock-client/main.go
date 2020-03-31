package main

import (
	"fmt"
	"github.com/smartcontractkit/external-initiator/integration/mock-client/web"
)

func main() {
	fmt.Println("Starting mock blockchain client")

	web.RunWebserver()
}
