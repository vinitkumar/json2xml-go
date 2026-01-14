# json2xml-go

A Go port of the Python [json2xml](https://github.com/vinitkumar/json2xml) library for converting JSON data to XML format.

[![Go Reference](https://pkg.go.dev/badge/github.com/vinitkumar/json2xml-go.svg)](https://pkg.go.dev/github.com/vinitkumar/json2xml-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/vinitkumar/json2xml-go)](https://goreportcard.com/report/github.com/vinitkumar/json2xml-go)

## Features

- Convert JSON/Go maps to XML
- Customizable root element names
- Optional type attributes on elements
- List item wrapping control
- CDATA sections support
- XML namespaces support
- XPath 3.1 json-to-xml format support
- Pretty printing with indentation
- **Command-line tool** for easy conversion

## Installation

### As a Library

```bash
go get github.com/vinitkumar/json2xml-go
```

### As a CLI Tool

**Quick install (binary only):**

```bash
go install github.com/vinitkumar/json2xml-go/cmd/json2xml-go@latest
```

**Full install with man page:**

```bash
# Clone the repository
git clone https://github.com/vinitkumar/json2xml-go.git
cd json2xml-go

# Install binary and man page (may need sudo for /usr/local)
sudo make install

# Or install to a custom location (no sudo needed)
make PREFIX=~/.local install
```

After installation, you can access the man page:

```bash
man json2xml-go
```

**Uninstall:**

```bash
sudo make uninstall
# Or if installed to custom location:
make PREFIX=~/.local uninstall
```

## CLI Usage

The `json2xml-go` command-line tool provides an easy way to convert JSON to XML from the terminal.

### Basic Examples

```bash
# Convert a JSON file to XML
json2xml-go data.json

# Convert with custom wrapper element
json2xml-go -w root data.json

# Read JSON from string
json2xml-go -s '{"name": "John", "age": 30}'

# Read from stdin
cat data.json | json2xml-go -

# Output to file
json2xml-go -o output.xml data.json

# Use XPath 3.1 format
json2xml-go -x data.json

# Disable pretty printing and type attributes
json2xml-go -p=false -t=false data.json

# Without item wrapping for lists
json2xml-go -i=false data.json
```

### CLI Options

```
Input Options:
  -u, --url string        Read JSON from URL
  -s, --string string     Read JSON from string
  [input-file]            Read JSON from file (use - for stdin)

Output Options:
  -o, --output string     Output file (default: stdout)

Conversion Options:
  -w, --wrapper string    Wrapper element name (default "all")
  -r, --root              Include root element (default true)
  -p, --pretty            Pretty print output (default true)
  -t, --type              Include type attributes (default true)
  -i, --item-wrap         Wrap list items in <item> elements (default true)
  -x, --xpath             Use XPath 3.1 json-to-xml format
  -c, --cdata             Wrap string values in CDATA sections
  -l, --list-headers      Repeat headers for each list item

Other Options:
  -v, --version           Show version information
  -h, --help              Show help message
```

## Library Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "log"

    json2xml "github.com/vinitkumar/json2xml-go"
)

func main() {
    data := map[string]any{
        "name": "John",
        "age":  30,
        "active": true,
    }

    converter := json2xml.New(data)
    xml, err := converter.ToXMLString()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(xml)
}
```

Output:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<all>
  <active type="bool">true</active>
  <age type="int">30</age>
  <name type="str">John</name>
</all>
```

### Reading from JSON File

```go
data, err := json2xml.ReadFromJSON("data.json")
if err != nil {
    log.Fatal(err)
}

xml, err := json2xml.New(data).ToXMLString()
```

### Reading from JSON String

```go
jsonStr := `{"login":"mojombo","id":1}`
data, err := json2xml.ReadFromString(jsonStr)
if err != nil {
    log.Fatal(err)
}

xml, err := json2xml.New(data).ToXMLString()
```

### Customization Options

```go
data := map[string]any{"key": "value"}

xml, err := json2xml.New(data).
    WithWrapper("root").       // Custom wrapper element name
    WithRoot(true).            // Include root element
    WithPretty(true).          // Pretty print output
    WithAttrType(true).        // Include type attributes
    WithItemWrap(true).        // Wrap list items in <item> elements
    WithXPathFormat(false).    // Use XPath 3.1 format
    ToXMLString()
```

### Without Item Wrapping

```go
data := map[string]any{"colors": []any{"red", "green", "blue"}}

// With ItemWrap = true (default):
// <colors><item>red</item><item>green</item><item>blue</item></colors>

// With ItemWrap = false:
// <colors>red</colors><colors>green</colors><colors>blue</colors>

xml, err := json2xml.New(data).
    WithItemWrap(false).
    WithAttrType(false).
    ToXMLString()
```

### XPath 3.1 Format

The library supports [XPath 3.1 json-to-xml](https://www.w3.org/TR/xpath-functions-31/#json-to-xml-mapping) format:

```go
data := map[string]any{"name": "John", "age": 30}

xml, err := json2xml.New(data).
    WithXPathFormat(true).
    WithPretty(false).
    ToXMLString()
```

Output:
```xml
<?xml version="1.0" encoding="UTF-8" ?><map xmlns="http://www.w3.org/2005/xpath-functions"><number key="age">30</number><string key="name">John</string></map>
```

### Using DictToXML Directly

For more control, you can use the `DictToXML` function directly:

```go
data := map[string]any{"key": "value"}
opts := json2xml.DefaultOptions()
opts.CustomRoot = "myroot"
opts.AttrType = false

xmlBytes := json2xml.DictToXML(data, opts)
```

### XML Namespaces

```go
data := map[string]any{"ns1:node1": "data in namespace 1"}
opts := json2xml.DefaultOptions()
opts.XMLNamespaces = map[string]any{
    "ns1": "https://example.com/ns1",
    "xmlns": "http://example.com/default",
}

xmlBytes := json2xml.DictToXML(data, opts)
```

## API Reference

### Types

#### Json2xml

Main converter struct with fluent API:

- `New(data any) *Json2xml` - Create new converter
- `WithWrapper(name string)` - Set wrapper element name (default: "all")
- `WithRoot(bool)` - Include root element (default: true)
- `WithPretty(bool)` - Pretty print output (default: true)
- `WithAttrType(bool)` - Include type attributes (default: true)
- `WithItemWrap(bool)` - Wrap list items (default: true)
- `WithXPathFormat(bool)` - Use XPath 3.1 format (default: false)
- `ToXML() (any, error)` - Convert to XML
- `ToXMLString() (string, error)` - Convert to XML string
- `ToXMLBytes() ([]byte, error)` - Convert to XML bytes

#### Options

Configuration struct for `DictToXML`:

```go
type Options struct {
    Root          bool              // Wrap in root element
    CustomRoot    string            // Root element name
    IDs           bool              // Add unique IDs
    AttrType      bool              // Add type attributes
    ItemWrap      bool              // Wrap list items
    ItemFunc      ItemFunc          // Custom item name function
    CDATA         bool              // Wrap strings in CDATA
    XMLNamespaces map[string]any    // XML namespaces
    ListHeaders   bool              // Repeat headers for list items
    XPathFormat   bool              // XPath 3.1 format
}
```

### Functions

- `ReadFromJSON(filename string) (any, error)` - Read JSON file
- `ReadFromString(jsonData string) (any, error)` - Parse JSON string
- `ReadFromURL(url string, params map[string]string) (any, error)` - Fetch JSON from URL
- `DictToXML(obj any, opts Options) []byte` - Convert to XML bytes
- `ConvertToXML(data any, opts *Options) ([]byte, error)` - Convenience function

### Errors

- `ErrJSONRead` - Error reading JSON file
- `ErrInvalidData` - Invalid data error
- `ErrURLRead` - Error reading from URL
- `ErrStringRead` - Error parsing JSON string

## Related Projects

### Python Version

This is a Go port of the original Python [json2xml](https://github.com/vinitkumar/json2xml) library. If you prefer Python, you can install it via pip:

```bash
pip install json2xml
```

The Python version includes the same features and a CLI tool (`json2xml-py`).

## License

MIT License - same as the original Python library.

## Author

Vinit Kumar <mail@vinitkumar.me>
