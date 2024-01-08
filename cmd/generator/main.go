package main

import (
	"github.com/kanryu/validagen/generator"
)

func main() {
	vp, err := generator.ParseToml("address.toml")
	if err != nil {
		panic(err)
	}
	err = vp.Generate()
	if err != nil {
		panic(err)
	}
}
