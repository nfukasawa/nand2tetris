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
	entries, err := assets.ReadDir("os")
	go func() {
		if err != nil {
			ch <- VMReader{ReadCloser: errorReadCloser(err)}
			close(ch)
			return
		}

		for _, e := range entries {
			f, err := assets.Open("os/" + e.Name())
			if err != nil {
				ch <- VMReader{Name: e.Name(), ReadCloser: errorReadCloser(err)}
				break
			}
			ch <- VMReader{Name: e.Name(), ReadCloser: f}
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

func errorReadCloser(err error) io.ReadCloser {
	return io.NopCloser(&errReader{err: err})
}

func (r *errReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
