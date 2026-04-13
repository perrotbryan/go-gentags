package main

import (
	"flag"
	"fmt"

	"github.com/perrotbryan/gentags"
)

func main() {
	dirFlag := flag.String("dir", ".", "directory to scan")
	flag.Parse()

	gen := gentags.Generator{}
	pkgs := gen.Analyze(*dirFlag)

	err := gen.Generate(pkgs)
	if err != nil {
		fmt.Println(fmt.Errorf("Generate error: %w", err))
	}
}
