package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/nfukaaswa/nand2tetris/07/src/vm"
)

type opts struct {
	Inputs []string `short:"i" long:"in" required:"true" description:"input file or directory path"`
}

func main() {
	var opts opts
	flags.Parse(&opts)

	files, err := collectSourceFiles(opts.Inputs)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		parser, err := vm.NewCommandParser(file)
		if err != nil {
			panic(err)
		}
		for {
			_, err := parser.NextCommand()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				panic(err)
			}
		}
	}
}

func collectSourceFiles(inputs []string) ([]string, error) {
	var srcs []string
	for _, in := range inputs {
		info, err := os.Stat(in)
		if err != nil {
			return nil, fmt.Errorf("os stat error: %v", err)
		}
		if info.IsDir() {
			filepath.Walk(in, func(path string, info fs.FileInfo, err error) error {
				if strings.HasSuffix(info.Name(), ".vm") {
					srcs = append(srcs, path)
				}
				return err
			})
		}
		if strings.HasSuffix(info.Name(), ".vm") {
			srcs = append(srcs, in)
		}
	}
	return srcs, nil
}
