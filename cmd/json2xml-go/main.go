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

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// cliOptions holds all parsed CLI flags.
type cliOptions struct {
	inputURL    string
	inputString string
	outputFile  string
	wrapper     string
	root        bool
	pretty      bool
	attrType    bool
	itemWrap    bool
	xpathFormat bool
	cdata       bool
	listHeaders bool
	showVersion bool
	showHelp    bool
	args        []string
}

func parseFlags(arguments []string) (cliOptions, error) {
	var opts cliOptions
	fs := flag.NewFlagSet("json2xml-go", flag.ContinueOnError)

	fs.StringVar(&opts.inputURL, "u", "", "Read JSON from URL")
	fs.StringVar(&opts.inputURL, "url", "", "Read JSON from URL")
	fs.StringVar(&opts.inputString, "s", "", "Read JSON from string")
	fs.StringVar(&opts.inputString, "string", "", "Read JSON from string")
	fs.StringVar(&opts.outputFile, "o", "", "Output file (default: stdout)")
	fs.StringVar(&opts.outputFile, "output", "", "Output file (default: stdout)")
	fs.StringVar(&opts.wrapper, "w", "all", "Wrapper element name")
	fs.StringVar(&opts.wrapper, "wrapper", "all", "Wrapper element name")
	fs.BoolVar(&opts.root, "r", true, "Include root element")
	fs.BoolVar(&opts.root, "root", true, "Include root element")
	fs.BoolVar(&opts.pretty, "p", true, "Pretty print output")
	fs.BoolVar(&opts.pretty, "pretty", true, "Pretty print output")
	fs.BoolVar(&opts.attrType, "t", true, "Include type attributes")
	fs.BoolVar(&opts.attrType, "type", true, "Include type attributes")
	fs.BoolVar(&opts.itemWrap, "i", true, "Wrap list items in <item> elements")
	fs.BoolVar(&opts.itemWrap, "item-wrap", true, "Wrap list items in <item> elements")
	fs.BoolVar(&opts.xpathFormat, "x", false, "Use XPath 3.1 json-to-xml format")
	fs.BoolVar(&opts.xpathFormat, "xpath", false, "Use XPath 3.1 json-to-xml format")
	fs.BoolVar(&opts.cdata, "c", false, "Wrap string values in CDATA sections")
	fs.BoolVar(&opts.cdata, "cdata", false, "Wrap string values in CDATA sections")
	fs.BoolVar(&opts.listHeaders, "l", false, "Repeat headers for each list item")
	fs.BoolVar(&opts.listHeaders, "list-headers", false, "Repeat headers for each list item")
	fs.BoolVar(&opts.showVersion, "v", false, "Show version information")
	fs.BoolVar(&opts.showVersion, "version", false, "Show version information")
	fs.BoolVar(&opts.showHelp, "h", false, "Show help message")
	fs.BoolVar(&opts.showHelp, "help", false, "Show help message")

	fs.Usage = func() { printUsage(os.Stderr) }

	if err := fs.Parse(arguments); err != nil {
		return opts, err
	}
	opts.args = fs.Args()
	return opts, nil
}

func run(arguments []string, stdin io.Reader, stdout, stderr io.Writer) int {
	opts, err := parseFlags(arguments)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if opts.showHelp {
		printUsage(stderr)
		return 0
	}

	if opts.showVersion {
		fmt.Fprintf(stdout, "json2xml-go version %s\n", json2xml.Version)
		fmt.Fprintf(stdout, "Author: %s <%s>\n", json2xml.Author, json2xml.Email)
		return 0
	}

	data, err := readInput(opts, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "Error reading input: %v\n", err)
		return 1
	}

	converter := json2xml.New(data).
		WithWrapper(opts.wrapper).
		WithRoot(opts.root).
		WithPretty(opts.pretty).
		WithAttrType(opts.attrType).
		WithItemWrap(opts.itemWrap).
		WithXPathFormat(opts.xpathFormat)

	xmlOutput, err := converter.ToXMLString()
	if err != nil {
		fmt.Fprintf(stderr, "Error converting to XML: %v\n", err)
		return 1
	}

	if err := writeOutput(xmlOutput, opts.outputFile, stdout); err != nil {
		fmt.Fprintf(stderr, "Error writing output: %v\n", err)
		return 1
	}

	return 0
}

func readInput(opts cliOptions, stdin io.Reader) (any, error) {
	if opts.inputURL != "" {
		return json2xml.ReadFromURL(opts.inputURL, nil)
	}

	if opts.inputString != "" {
		return json2xml.ReadFromString(opts.inputString)
	}

	if len(opts.args) > 0 {
		filename := opts.args[0]
		if filename == "-" {
			return readFromReader(stdin)
		}
		return json2xml.ReadFromJSON(filename)
	}

	// Check if stdin has data (only when stdin is *os.File)
	if f, ok := stdin.(*os.File); ok {
		stat, _ := f.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			return readFromReader(stdin)
		}
	}

	return nil, fmt.Errorf("no input provided. Use -h for help")
}

func readFromReader(r io.Reader) (any, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}

	jsonStr := strings.TrimSpace(string(data))
	if jsonStr == "" {
		return nil, fmt.Errorf("empty input")
	}

	return json2xml.ReadFromString(jsonStr)
}

func writeOutput(output, outputFile string, stdout io.Writer) error {
	if outputFile != "" {
		return os.WriteFile(outputFile, []byte(output), 0644)
	}

	fmt.Fprintln(stdout, output)
	return nil
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `json2xml-go - Convert JSON to XML

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
