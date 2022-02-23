package compiler

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func Compile(inputs []string, outDir string) error {
	srcs, err := collectSourceFiles(inputs)
	if err != nil {
		return err
	}

	for _, src := range srcs {
		file, err := os.Open(src)
		if err != nil {
			return err
		}

		// tokenize
		tokens, err := Tokenize(file)
		if err != nil {
			file.Close()
			return err
		}
		file.Close()

		// analyze
		cls, err := Analyze(tokens)
		if err != nil {
			return err
		}

		// compile
		out := bytes.NewBuffer(nil)
		if err := CompileClass(NewJackVM(out), cls); err != nil {
			return err
		}

		// output
		name := strings.TrimSuffix(filepath.Base(src), ".jack") + ".vm"
		if err := writeFile(filepath.Join(outDir, name), out); err != nil {
			fmt.Println(err)
			return err
		}
	}

	// output os VMs
	for vm := range OSVMs() {
		if err := writeFile(filepath.Join(outDir, vm.Name), vm); err != nil {
			vm.Close()
			return err
		}
		vm.Close()
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
