// Package xmldom provides XML DOM processing, and supports xpath queries
package xmldom

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	xmlPrefix   = "xml"
	xmlUrl      = "http://www.w3.org/XML/1998/namespace"
	xmlnsPrefix = "xmlns"
	xmlnsUrl    = "http://www.w3.org/2000/xmlns"
	xlinkPrefix = "xlink"
	xlinkUrl    = "http://www.w3.org/1999/xlink"
	xsiPrefix   = "xsi"
	xsiUrl      = "http://www.w3.org/2001/XMLSchema-instance"
)

// DOMParser parses XML sources, converting them into a DOM. It is configurable, allowing
// the user to control some features of the XML parse, such as whitespace preservation in
// text entities.
type DOMParser interface {
	ParseXML(s string) (*Document, error)
	ParseFile(filename string) (*Document, error)
	Parse(r io.Reader) (*Document, error)
	PreserveWhitespace(f bool) DOMParser
}

type domParserSettings struct {
	preserveWhitespace bool
}

func NewDOMParser() DOMParser {
	return &domParserSettings{}
}

func (s *domParserSettings) PreserveWhitespace(f bool) DOMParser {
	s.preserveWhitespace = f
	return s
}

// Must parse without error, else panic. Helpful when there is no other path to following
// if the XML source is invalid.
func Must(doc *Document, err error) *Document {
	if err != nil {
		panic(err)
	}
	return doc
}

// ParseXML text, using default parser settings. For backwards compatibility.
func ParseXML(s string) (*Document, error) {
	return Parse(strings.NewReader(s))
}

// ParseXML text, using the parser settings from the receiver.
func (s *domParserSettings) ParseXML(text string) (*Document, error) {
	return s.Parse(strings.NewReader(text))
}

// ParseFile XML text, using default parser settings. For backwards compatibility.
func ParseFile(filename string) (*Document, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	return Parse(file)
}

// ParseFile XML text, using the parser settings from the receiver.
func (s *domParserSettings) ParseFile(filename string) (*Document, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	return s.Parse(file)
}

// Parse the XML text from the given reader, using default parser settings. For backwards compatibility.
func Parse(r io.Reader) (*Document, error) {
	return NewDOMParser().Parse(r)
}

// Parse the XML text from the given reader, using the parser settings from the receiver.
func (s *domParserSettings) Parse(r io.Reader) (*Document, error) {
	p := xml.NewDecoder(r)
	t, err := p.Token()
	if err != nil {
		return nil, err
	}

	doc := new(Document)
	var e *Node
	for t != nil {
		switch token := t.(type) {
		case xml.StartElement:
			// a new node
			el := new(Node)
			el.Document = doc
			el.Parent = e
			el.Name = token.Name.Local
			for _, attr := range token.Attr {
				var name, ns string
				if attr.Name.Space != "" {
					ns = attr.Name.Space
					switch ns {
					case xmlnsUrl:
						name = fmt.Sprintf("%s:%s", xmlnsPrefix, attr.Name.Local)
					case xmlUrl:
						name = fmt.Sprintf("%s:%s", xmlPrefix, attr.Name.Local)
					case xlinkUrl:
						name = fmt.Sprintf("%s:%s", xlinkPrefix, attr.Name.Local)
					case xsiUrl:
						name = fmt.Sprintf("%s:%s", xsiPrefix, attr.Name.Local)
					default:
						name = fmt.Sprintf("%s:%s", attr.Name.Space, attr.Name.Local)
					}
				} else {
					name = attr.Name.Local
				}
				el.Attributes = append(el.Attributes, &Attribute{
					Name:  name,
					Value: attr.Value,
				})
			}
			if e != nil {
				e.Children = append(e.Children, el)
			}
			e = el

			if doc.Root == nil {
				doc.Root = e
			}
		case xml.EndElement:
			e = e.Parent
		case xml.CharData:
			// text node
			if e != nil {
				if s.preserveWhitespace {
					e.Text = string(token)
				} else {
					e.Text = string(bytes.TrimSpace(token))
				}
			}
		case xml.ProcInst:
			doc.ProcInst = stringifyProcInst(&token)
		case xml.Directive:
			doc.Directives = append(doc.Directives, stringifyDirective(&token))
		}

		// get the next token
		t, err = p.Token()
	}

	// Make sure that reading stopped on EOF
	if err != io.EOF {
		return nil, err
	}

	// All is good, return the document
	return doc, nil
}
