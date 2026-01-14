// Package json2xml provides utilities to convert JSON data to XML format.
//
// This package is a Go port of the Python json2xml library.
// It supports various conversion options including:
//   - Custom root element names
//   - Type attributes on elements
//   - List item wrapping
//   - CDATA sections
//   - XML namespaces
//   - XPath 3.1 json-to-xml format
//
// Example usage:
//
//	data := map[string]any{"name": "John", "age": 30}
//	converter := json2xml.New(data)
//	xml, err := converter.ToXML()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(xml)
package json2xml

import (
	"fmt"
)

// Version information
const (
	Version = "1.0.0"
	Author  = "Vinit Kumar"
	Email   = "mail@vinitkumar.me"
)

// Json2xml is the main converter struct.
type Json2xml struct {
	data        any
	wrapper     string
	root        bool
	pretty      bool
	attrType    bool
	itemWrap    bool
	xpathFormat bool
}

// New creates a new Json2xml converter with default options.
func New(data any) *Json2xml {
	return &Json2xml{
		data:        data,
		wrapper:     "all",
		root:        true,
		pretty:      true,
		attrType:    true,
		itemWrap:    true,
		xpathFormat: false,
	}
}

// WithWrapper sets a custom wrapper element name.
func (j *Json2xml) WithWrapper(wrapper string) *Json2xml {
	j.wrapper = wrapper
	return j
}

// WithRoot sets whether to include root element.
func (j *Json2xml) WithRoot(root bool) *Json2xml {
	j.root = root
	return j
}

// WithPretty sets whether to pretty-print the output.
func (j *Json2xml) WithPretty(pretty bool) *Json2xml {
	j.pretty = pretty
	return j
}

// WithAttrType sets whether to include type attributes.
func (j *Json2xml) WithAttrType(attrType bool) *Json2xml {
	j.attrType = attrType
	return j
}

// WithItemWrap sets whether to wrap list items in <item> elements.
func (j *Json2xml) WithItemWrap(itemWrap bool) *Json2xml {
	j.itemWrap = itemWrap
	return j
}

// WithXPathFormat sets whether to use XPath 3.1 json-to-xml format.
func (j *Json2xml) WithXPathFormat(xpathFormat bool) *Json2xml {
	j.xpathFormat = xpathFormat
	return j
}

// ToXML converts the data to XML.
// Returns the XML as a string when pretty=true, or as bytes when pretty=false.
// Returns nil if data is empty or nil.
func (j *Json2xml) ToXML() (any, error) {
	if j.data == nil {
		return nil, nil
	}

	// Check if data is an empty map or slice
	switch v := j.data.(type) {
	case map[string]any:
		if len(v) == 0 {
			return nil, nil
		}
	case []any:
		if len(v) == 0 {
			return nil, nil
		}
	}

	opts := Options{
		Root:        j.root,
		CustomRoot:  j.wrapper,
		AttrType:    j.attrType,
		ItemWrap:    j.itemWrap,
		ItemFunc:    DefaultItemFunc,
		XPathFormat: j.xpathFormat,
	}

	xmlData := DictToXML(j.data, opts)

	if j.pretty {
		prettyXML, err := PrettyPrint(xmlData)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidData, err)
		}
		return prettyXML, nil
	}

	return xmlData, nil
}

// ToXMLString converts the data to XML and returns it as a string.
func (j *Json2xml) ToXMLString() (string, error) {
	result, err := j.ToXML()
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", nil
	}

	switch v := result.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// ToXMLBytes converts the data to XML and returns it as bytes.
func (j *Json2xml) ToXMLBytes() ([]byte, error) {
	result, err := j.ToXML()
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	switch v := result.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return []byte(fmt.Sprintf("%v", v)), nil
	}
}

// ConvertToXML is a convenience function to convert JSON data to XML.
func ConvertToXML(data any, opts *Options) ([]byte, error) {
	if data == nil {
		return nil, nil
	}

	if opts == nil {
		defaultOpts := DefaultOptions()
		opts = &defaultOpts
	}

	if opts.ItemFunc == nil {
		opts.ItemFunc = DefaultItemFunc
	}

	return DictToXML(data, *opts), nil
}
