package main

import (
	"fmt"
	"github.com/brunovlucena/guestbook-go/cmd/utils"
)

func init() {
	utils.LogrusSetup()
	utils.ViperSetup()
}

func main() {
	fmt.Println("Hello World!")
}
