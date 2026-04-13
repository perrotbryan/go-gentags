package main

import (
	"log"

	"github.com/perrotbryan/gentags"
)

func main() {
	gen := &gentags.Generator{}

	pkgs := gen.Analyze(".")

	if err := gen.Generate(pkgs); err != nil {
		log.Fatal(err)
	}
}
