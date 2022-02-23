package main

import (
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/nfukaaswa/nand2tetris/11/src/compiler"
)

type opts struct {
	Inputs []string `short:"i" long:"in" required:"true" description:"input file or directory path"`
	Output string   `short:"o" long:"out" required:"true" description:"output directory path"`
}

func main() {
	var opts opts
	if _, err := flags.Parse(&opts); err != nil {
		return
	}

	if err := compiler.Compile(opts.Inputs, opts.Output); err != nil {
		fmt.Println(err)
		return
	}
}
