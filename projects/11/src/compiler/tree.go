package compiler

import (
	"bytes"
	"io"
	"strings"
)

type NodeIface interface {
	ToNode() *Node
}

type Node struct {
	Name      string
	Value     string
	Children  []Node
	SkipLayer bool
}

func (e *Node) AddChild(c NodeIface) {
	if e == nil || c == nil {
		return
	}

	node := c.ToNode()
	if node == nil {
		return
	}

	e.Children = append(e.Children, *node)
}

func (e *Node) ToNode() *Node {
	return e
}

func (e *Node) MarshalXML(w io.Writer) error {
	return e.marshalXML(w, "")
}

func (e *Node) marshalXML(w io.Writer, indent string) error {
	if e == nil {
		return nil
	}

	if e.SkipLayer {
		for i, c := range e.Children {
			if i != 0 {
				if err := writeString(w, "\n"); err != nil {
					return err
				}
			}
			if err := c.marshalXML(w, indent); err != nil {
				return err
			}
		}
		return nil
	}

	if err := writeString(w, indent+"<"+string(e.Name)+">"); err != nil {
		return err
	}

	if e.Value != "" {
		if err := writeString(w, " "+xmlEscaper.Replace(e.Value)+" "); err != nil {
			return err
		}
	}

	if e.Children != nil {
		childIndent := indent + "  "
		for _, c := range e.Children {
			if err := writeString(w, "\n"); err != nil {
				return err
			}
			if err := c.marshalXML(w, childIndent); err != nil {
				return err
			}
		}
		if err := writeString(w, "\n"+indent); err != nil {
			return err
		}
	}

	if err := writeString(w, "</"+string(e.Name)+">"); err != nil {
		return err
	}

	return nil
}

func writeString(w io.Writer, str string) error {
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
