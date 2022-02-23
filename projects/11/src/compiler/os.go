package compiler

import (
	"embed"
	"io"
)

//go:embed os/*
var libs embed.FS

var osLibs = []string{
	"Array",
	"Keyboard",
	"Math",
	"Memory",
	"Output",
	"Screen",
	"String",
	"Sys",
}

var usedLibs = map[string]struct{}{}

func ClassMethodCalled(name string) {
	for _, osLib := range osLibs {
		if name == osLib {
			usedLibs[name] = struct{}{}
		}
	}
}

func OSVMs() <-chan VMReader {
	ch := make(chan VMReader)
	go func() {
		for lib := range usedLibs {
			f, err := libs.Open("os/" + lib + ".vm")
			if err != nil {
				ch <- VMReader{Name: lib, ReadCloser: io.NopCloser(&errReader{err: err})}
				break
			}
			ch <- VMReader{Name: lib, ReadCloser: f}
		}
		close(ch)
	}()
	return ch

}

type VMReader struct {
	Name string
	io.ReadCloser
}

type errReader struct {
	err error
}

func (r *errReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
