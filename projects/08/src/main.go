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
	"github.com/nfukaaswa/nand2tetris/08/src/vm"
)

type opts struct {
	Inputs []string `short:"i" long:"in" required:"true" description:"input file or directory path"`
	Output string   `short:"o" long:"out" required:"true" description:"output file or path"`
	Debug  bool     `short:"d" long:"debug"  description:"enable debug mode"`
}

func main() {
	var opts opts
	if _, err := flags.Parse(&opts); err != nil {
		fmt.Println(err)
		return
	}

	var err error

	srcs, err := collectSourceFiles(opts.Inputs)
	if err != nil {
		fmt.Println(err)
		return
	}

	out, err := os.OpenFile(opts.Output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		out.Close()
		if err != nil {
			os.Remove(opts.Output)
		}
	}()

	trans, err := vm.NewTranslator(out)
	if err != nil {
		fmt.Println(err)
		return
	}
	trans.Debug = opts.Debug

	for _, src := range srcs {
		if err = translateSourceFile(src, trans); err != nil {
			fmt.Println(err)
			return
		}
	}
}

func translateSourceFile(src string, trans *vm.Translator) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	t := trans.File(strings.TrimSuffix(filepath.Base(src), ".vm"))
	parser := vm.NewParser(src, file)
	for {
		cmd, err := parser.NextCommand()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if err := t.Command(cmd); err != nil {
			return err
		}
	}
	return nil
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
	if len(srcs) == 0 {
		return nil, fmt.Errorf(".vm file not found in: %v", inputs)
	}
	return srcs, nil
}
