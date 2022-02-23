package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/nfukaaswa/nand2tetris/10/src/compiler"
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

	var err error

	srcs, err := collectSourceFiles(opts.Inputs)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, src := range srcs {
		file, err := os.Open(src)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		srcBase := strings.TrimSuffix(filepath.Base(src), ".jack")

		// tokenize
		tokens, err := compiler.Tokenize(file)
		if err != nil {
			fmt.Println(err)
			return
		}

		// write tokens xml
		if err := writeXML(filepath.Join(opts.Output, srcBase+"T.xml"), tokens.ToNode()); err != nil {
			fmt.Println(err)
			return
		}

		// analyze
		cls, err := compiler.Analyze(tokens)
		if err != nil {
			fmt.Println(err)
			return
		}

		// write tree xml
		if err := writeXML(filepath.Join(opts.Output, srcBase+".xml"), cls.ToNode()); err != nil {
			fmt.Println(err)
			return
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
				if strings.HasSuffix(info.Name(), ".jack") {
					srcs = append(srcs, path)
				}
				return err
			})
		}
		if strings.HasSuffix(info.Name(), ".jack") {
			srcs = append(srcs, in)
		}
	}
	if len(srcs) == 0 {
		return nil, fmt.Errorf(".jack file not found in: %v", inputs)
	}
	return srcs, nil
}

func writeXML(path string, node *compiler.Node) error {
	buf := bytes.NewBuffer(nil)
	if err := node.MarshalXML(buf); err != nil {
		return err
	}
	if err := writeFile(path, buf); err != nil {
		return err
	}
	return nil
}

func writeFile(path string, buf io.Reader) error {
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open file error: %s %v", path, err)
	}
	defer func() {
		out.Close()
		if err != nil {
			os.Remove(path)
		}
	}()
	_, err = io.Copy(out, buf)
	if err != nil {
		return fmt.Errorf("write file error: %s %v", path, err)
	}
	fmt.Println("out: " + path)
	return nil
}
