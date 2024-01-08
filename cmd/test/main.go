package main

import (
	"fmt"

	"github.com/kanryu/validagen/example"
)

func main() {

	a := example.Address{
		Street: "1235",
		City:   "Unknown",
		State:  "Viia",
		Zip:    "12345",
	}

	err := a.Validate()
	fmt.Println(err)
}
