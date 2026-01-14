// json2xml-go is a command-line tool to convert JSON data to XML format.
//
// Usage:
//
//	json2xml-go [flags] [input-file]
//
// Flags:
//
//	-w, --wrapper string    Wrapper element name (default "all")
//	-r, --root              Include root element (default true)
//	-p, --pretty            Pretty print output (default true)
//	-t, --type              Include type attributes (default true)
//	-i, --item-wrap         Wrap list items in <item> elements (default true)
//	-x, --xpath             Use XPath 3.1 json-to-xml format
//	-o, --output string     Output file (default: stdout)
//	-u, --url string        Read JSON from URL
//	-s, --string string     Read JSON from string
//	-h, --help              Show help message
//	-v, --version           Show version information
//
// Examples:
//
//	# Convert a JSON file to XML
//	json2xml-go data.json
//
//	# Convert with custom wrapper
//	json2xml-go -w root data.json
//
//	# Read from URL
//	json2xml-go -u https://api.example.com/data.json
//
//	# Read from string
//	json2xml-go -s '{"name": "John", "age": 30}'
//
//	# Output to file
//	json2xml-go -o output.xml data.json
//
//	# Use XPath 3.1 format
//	json2xml-go -x data.json
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	json2xml "github.com/vinitkumar/json2xml-go"
)

const version = "1.0.0"

var (
	// Input options
	inputURL    string
	inputString string

	// Output options
	outputFile string

	// Conversion options
	wrapper     string
	root        bool
	pretty      bool
	attrType    bool
	itemWrap    bool
	xpathFormat bool
	cdata       bool
	listHeaders bool

	// Other options
	showVersion bool
	showHelp    bool
)

func init() {
	// Input options
	flag.StringVar(&inputURL, "u", "", "Read JSON from URL")
	flag.StringVar(&inputURL, "url", "", "Read JSON from URL")
	flag.StringVar(&inputString, "s", "", "Read JSON from string")
	flag.StringVar(&inputString, "string", "", "Read JSON from string")

	// Output options
	flag.StringVar(&outputFile, "o", "", "Output file (default: stdout)")
	flag.StringVar(&outputFile, "output", "", "Output file (default: stdout)")

	// Conversion options
	flag.StringVar(&wrapper, "w", "all", "Wrapper element name")
	flag.StringVar(&wrapper, "wrapper", "all", "Wrapper element name")
	flag.BoolVar(&root, "r", true, "Include root element")
	flag.BoolVar(&root, "root", true, "Include root element")
	flag.BoolVar(&pretty, "p", true, "Pretty print output")
	flag.BoolVar(&pretty, "pretty", true, "Pretty print output")
	flag.BoolVar(&attrType, "t", true, "Include type attributes")
	flag.BoolVar(&attrType, "type", true, "Include type attributes")
	flag.BoolVar(&itemWrap, "i", true, "Wrap list items in <item> elements")
	flag.BoolVar(&itemWrap, "item-wrap", true, "Wrap list items in <item> elements")
	flag.BoolVar(&xpathFormat, "x", false, "Use XPath 3.1 json-to-xml format")
	flag.BoolVar(&xpathFormat, "xpath", false, "Use XPath 3.1 json-to-xml format")
	flag.BoolVar(&cdata, "c", false, "Wrap string values in CDATA sections")
	flag.BoolVar(&cdata, "cdata", false, "Wrap string values in CDATA sections")
	flag.BoolVar(&listHeaders, "l", false, "Repeat headers for each list item")
	flag.BoolVar(&listHeaders, "list-headers", false, "Repeat headers for each list item")

	// Other options
	flag.BoolVar(&showVersion, "v", false, "Show version information")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.BoolVar(&showHelp, "help", false, "Show help message")

	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, `json2xml-go - Convert JSON to XML

Usage:
  json2xml-go [flags] [input-file]

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

Examples:
  # Convert a JSON file to XML
  json2xml-go data.json

  # Convert with custom wrapper
  json2xml-go -w root data.json

  # Read from URL
  json2xml-go -u https://api.example.com/data.json

  # Read from string
  json2xml-go -s '{"name": "John", "age": 30}'

  # Read from stdin
  cat data.json | json2xml-go -

  # Output to file
  json2xml-go -o output.xml data.json

  # Use XPath 3.1 format
  json2xml-go -x data.json

  # Disable pretty printing and type attributes
  json2xml-go -p=false -t=false data.json

`)
}

func main() {
	flag.Parse()

	if showHelp {
		usage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("json2xml-go version %s\n", version)
		fmt.Printf("Author: %s <%s>\n", json2xml.Author, json2xml.Email)
		os.Exit(0)
	}

	// Read input data
	data, err := readInput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Convert to XML
	converter := json2xml.New(data).
		WithWrapper(wrapper).
		WithRoot(root).
		WithPretty(pretty).
		WithAttrType(attrType).
		WithItemWrap(itemWrap).
		WithXPathFormat(xpathFormat)

	xmlOutput, err := converter.ToXMLString()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting to XML: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if err := writeOutput(xmlOutput); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}
}

func readInput() (any, error) {
	// Priority: URL > String > File > Stdin
	if inputURL != "" {
		return json2xml.ReadFromURL(inputURL, nil)
	}

	if inputString != "" {
		return json2xml.ReadFromString(inputString)
	}

	args := flag.Args()
	if len(args) > 0 {
		filename := args[0]
		if filename == "-" {
			// Read from stdin
			return readFromStdin()
		}
		return json2xml.ReadFromJSON(filename)
	}

	// Check if there's data on stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return readFromStdin()
	}

	return nil, fmt.Errorf("no input provided. Use -h for help")
}

func readFromStdin() (any, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}

	jsonStr := strings.TrimSpace(string(data))
	if jsonStr == "" {
		return nil, fmt.Errorf("empty input")
	}

	return json2xml.ReadFromString(jsonStr)
}

func writeOutput(output string) error {
	if outputFile != "" {
		return os.WriteFile(outputFile, []byte(output), 0644)
	}

	fmt.Println(output)
	return nil
}
