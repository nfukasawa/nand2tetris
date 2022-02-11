package compiler

import (
	"bytes"
	"io"
	"strings"
)

type XMLizer interface {
	ToXML() *XMLElm
}

type XMLElm struct {
	Name      string
	Value     string
	Children  []XMLElm
	SkipLayer bool
}

func (e *XMLElm) AddChild(c XMLizer) {
	if e == nil || c == nil {
		return
	}

	xml := c.ToXML()
	if xml == nil {
		return
	}

	if xml.SkipLayer {
		e.Children = append(e.Children, xml.Children...)
		return
	}

	e.Children = append(e.Children, *xml)
}

func (e XMLElm) ToXML() *XMLElm {
	return &e
}

func (e *XMLElm) Marshal(w io.Writer) error {
	return e.marshal(w, "")
}

func (e *XMLElm) marshal(w io.Writer, indent string) error {
	if e == nil {
		return nil
	}

	if err := e.write(w, indent+"<"+string(e.Name)+">"); err != nil {
		return err
	}

	if e.Value != "" {
		if err := e.write(w, " "+xmlEscaper.Replace(e.Value)+" "); err != nil {
			return err
		}
	}

	if e.Children != nil {
		childIndent := indent + "  "
		for _, c := range e.Children {
			if err := e.write(w, "\n"); err != nil {
				return err
			}
			if err := c.marshal(w, childIndent); err != nil {
				return err
			}
		}
		if err := e.write(w, "\n"+indent); err != nil {
			return err
		}
	}

	if err := e.write(w, "</"+string(e.Name)+">"); err != nil {
		return err
	}

	return nil
}

func (e *XMLElm) write(w io.Writer, str string) error {
	if _, err := io.Copy(w, bytes.NewBufferString(str)); err != nil {
		return err
	}
	return nil
}

var xmlEscaper = strings.NewReplacer(
	"<", "&lt;",
	">", "&gt;",
	"&", "&amp;",
	"'", "&apos;",
	"\"", "&quot;",
)
