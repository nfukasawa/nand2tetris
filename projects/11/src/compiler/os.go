package compiler

import (
	"embed"
	"io"
)

//go:embed os/*
var assets embed.FS

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

func OSVMs() <-chan VMReader {
	ch := make(chan VMReader)
	go func() {
		for _, lib := range osLibs {
			f, err := assets.Open("os/" + lib + ".vm")
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
