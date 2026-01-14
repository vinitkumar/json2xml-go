package json2xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

// XPath 3.1 json-to-xml namespace
const XPathFunctionsNS = "http://www.w3.org/2005/xpath-functions"

// ItemFunc is a function that generates element names for list items.
type ItemFunc func(parent string) string

// DefaultItemFunc returns "item" for any parent element.
func DefaultItemFunc(parent string) string {
	return "item"
}

// Options configures the XML conversion behavior.
type Options struct {
	// Root specifies whether to wrap output in an XML root element.
	Root bool
	// CustomRoot specifies the name of the root element.
	CustomRoot string
	// IDs specifies whether elements get unique IDs.
	IDs bool
	// AttrType specifies whether elements get a data type attribute.
	AttrType bool
	// ItemWrap specifies whether to wrap list items in <item> elements.
	ItemWrap bool
	// ItemFunc generates element names for list items.
	ItemFunc ItemFunc
	// CDATA specifies whether string values should be wrapped in CDATA sections.
	CDATA bool
	// XMLNamespaces is a map of namespace prefixes to URIs.
	XMLNamespaces map[string]any
	// ListHeaders specifies whether to repeat headers for each list item.
	ListHeaders bool
	// XPathFormat specifies whether to use XPath 3.1 json-to-xml format.
	XPathFormat bool
}

// DefaultOptions returns the default conversion options.
func DefaultOptions() Options {
	return Options{
		Root:        true,
		CustomRoot:  "root",
		IDs:         false,
		AttrType:    true,
		ItemWrap:    true,
		ItemFunc:    DefaultItemFunc,
		CDATA:       false,
		ListHeaders: false,
		XPathFormat: false,
	}
}

// MakeID generates a random ID for a given element.
func MakeID(element string, start, end int) string {
	if start == 0 {
		start = 100000
	}
	if end == 0 {
		end = 999999
	}
	return fmt.Sprintf("%s_%d", element, rand.Intn(end-start+1)+start)
}

// GetUniqueID generates a unique ID for a given element.
func GetUniqueID(element string) string {
	return MakeID(element, 100000, 999999)
}

// GetXMLType returns the XML type string for a given value.
func GetXMLType(val any) string {
	if val == nil {
		return "null"
	}

	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "int"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.String:
		return "str"
	case reflect.Map:
		return "dict"
	case reflect.Slice, reflect.Array:
		return "list"
	default:
		// Check for time.Time
		if _, ok := val.(time.Time); ok {
			return "str"
		}
		return v.Type().Name()
	}
}

// EscapeXML escapes special XML characters in a string.
func EscapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// MakeAttrString creates a string of XML attributes from a map.
func MakeAttrString(attrs map[string]any) string {
	if len(attrs) == 0 {
		return ""
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := attrs[k]
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, EscapeXML(fmt.Sprintf("%v", v))))
	}
	return " " + strings.Join(parts, " ")
}

// KeyIsValidXML checks if a key is a valid XML name.
func KeyIsValidXML(key string) bool {
	if key == "" {
		return false
	}
	// XML names can't start with numbers or punctuation (except underscore)
	// and can only contain letters, digits, hyphens, underscores, and periods
	testXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" ?><%s>foo</%s>`, key, key)
	decoder := xml.NewDecoder(strings.NewReader(testXML))
	for {
		tok, err := decoder.Token()
		if err != nil {
			// EOF means we successfully parsed all tokens
			if err.Error() == "EOF" {
				return true
			}
			return false
		}
		if tok == nil {
			break
		}
	}
	return true
}

// MakeValidXMLName tests an XML name and fixes it if invalid.
func MakeValidXMLName(key string, attrs map[string]any) (string, map[string]any) {
	key = EscapeXML(key)

	// Pass through if key is already valid
	if KeyIsValidXML(key) {
		return key, attrs
	}

	// Prepend 'n' if the key is numeric
	if isNumeric(key) {
		return "n" + key, attrs
	}

	// Replace spaces with underscores if that fixes the problem
	keyWithUnderscores := strings.ReplaceAll(key, " ", "_")
	if KeyIsValidXML(keyWithUnderscores) {
		return keyWithUnderscores, attrs
	}

	// Allow namespace prefixes + ignore @flat in key
	keyClean := strings.ReplaceAll(key, ":", "")
	keyClean = strings.ReplaceAll(keyClean, "@flat", "")
	if KeyIsValidXML(keyClean) {
		return key, attrs
	}

	// Key is still invalid - move it into a name attribute
	if attrs == nil {
		attrs = make(map[string]any)
	}
	attrs["name"] = key
	return "key", attrs
}

// isNumeric checks if a string represents a number.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// WrapCDATA wraps a string in CDATA sections.
func WrapCDATA(s string) string {
	s = strings.ReplaceAll(s, "]]>", "]]]]><![CDATA[>")
	return "<![CDATA[" + s + "]]>"
}

// GetXPath31TagName determines XPath 3.1 tag name by value type.
func GetXPath31TagName(val any) string {
	if val == nil {
		return "null"
	}

	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Map:
		return "map"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	case reflect.Slice, reflect.Array:
		// Check if it's []byte
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return "string"
		}
		return "array"
	default:
		return "string"
	}
}

// ConvertToXPath31 converts a value to XPath 3.1 json-to-xml format.
func ConvertToXPath31(obj any, parentKey string) string {
	keyAttr := ""
	if parentKey != "" {
		keyAttr = fmt.Sprintf(` key="%s"`, EscapeXML(parentKey))
	}
	tagName := GetXPath31TagName(obj)

	switch tagName {
	case "null":
		return fmt.Sprintf("<null%s/>", keyAttr)
	case "boolean":
		return fmt.Sprintf("<boolean%s>%s</boolean>", keyAttr, strings.ToLower(fmt.Sprintf("%v", obj)))
	case "number":
		return fmt.Sprintf("<number%s>%v</number>", keyAttr, obj)
	case "string":
		return fmt.Sprintf("<string%s>%s</string>", keyAttr, EscapeXML(fmt.Sprintf("%v", obj)))
	case "map":
		var children strings.Builder
		m := toMap(obj)
		// Sort keys for consistent output
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			children.WriteString(ConvertToXPath31(m[k], k))
		}
		return fmt.Sprintf("<map%s>%s</map>", keyAttr, children.String())
	case "array":
		var children strings.Builder
		items := toSlice(obj)
		for _, item := range items {
			children.WriteString(ConvertToXPath31(item, ""))
		}
		return fmt.Sprintf("<array%s>%s</array>", keyAttr, children.String())
	default:
		return fmt.Sprintf("<string%s>%s</string>", keyAttr, EscapeXML(fmt.Sprintf("%v", obj)))
	}
}

// toMap converts an interface to a map[string]any.
func toMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Map {
		result := make(map[string]any)
		for _, key := range rv.MapKeys() {
			result[fmt.Sprintf("%v", key.Interface())] = rv.MapIndex(key).Interface()
		}
		return result
	}
	return nil
}

// toSlice converts an interface to a []any.
func toSlice(v any) []any {
	if s, ok := v.([]any); ok {
		return s
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		result := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = rv.Index(i).Interface()
		}
		return result
	}
	return nil
}

// IsPrimitiveType checks if a value is a primitive type.
func IsPrimitiveType(val any) bool {
	t := GetXMLType(val)
	return t == "str" || t == "int" || t == "float" || t == "bool" || t == "null"
}

// Convert routes elements to the right function based on their data type.
func Convert(obj any, opts Options, parent string) string {
	itemName := opts.ItemFunc(parent)

	if obj == nil {
		return ConvertNone(itemName, opts.AttrType, nil, opts.CDATA)
	}

	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Bool:
		return ConvertBool(itemName, obj.(bool), opts.AttrType, nil, opts.CDATA)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return ConvertKV(itemName, obj, opts.AttrType, nil, opts.CDATA)
	case reflect.String:
		return ConvertKV(itemName, obj, opts.AttrType, nil, opts.CDATA)
	case reflect.Map:
		return ConvertDict(toMap(obj), opts, parent)
	case reflect.Slice, reflect.Array:
		return ConvertList(toSlice(obj), opts, parent)
	default:
		// Check for time.Time
		if t, ok := obj.(time.Time); ok {
			return ConvertKV(itemName, t.Format(time.RFC3339), opts.AttrType, nil, opts.CDATA)
		}
		return ConvertKV(itemName, fmt.Sprintf("%v", obj), opts.AttrType, nil, opts.CDATA)
	}
}

// ConvertDict converts a map into an XML string.
func ConvertDict(obj map[string]any, opts Options, parent string) string {
	var output strings.Builder

	// Sort keys for consistent output
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		val := obj[key]
		attrs := make(map[string]any)

		if opts.IDs {
			attrs["id"] = GetUniqueID(parent)
		}

		key, attrs = MakeValidXMLName(key, attrs)

		switch v := val.(type) {
		case bool:
			output.WriteString(ConvertBool(key, v, opts.AttrType, attrs, opts.CDATA))
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			output.WriteString(ConvertKV(key, v, opts.AttrType, attrs, opts.CDATA))
		case string:
			output.WriteString(ConvertKV(key, v, opts.AttrType, attrs, opts.CDATA))
		case time.Time:
			output.WriteString(ConvertKV(key, v.Format(time.RFC3339), opts.AttrType, attrs, opts.CDATA))
		case map[string]any:
			output.WriteString(Dict2XMLStr(opts, attrs, v, key, false, parent))
		case []any:
			output.WriteString(List2XMLStr(opts, attrs, v, key))
		default:
			if val == nil {
				output.WriteString(ConvertNone(key, opts.AttrType, attrs, opts.CDATA))
			} else {
				rv := reflect.ValueOf(val)
				switch rv.Kind() {
				case reflect.Bool:
					output.WriteString(ConvertBool(key, rv.Bool(), opts.AttrType, attrs, opts.CDATA))
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					output.WriteString(ConvertKV(key, rv.Int(), opts.AttrType, attrs, opts.CDATA))
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					output.WriteString(ConvertKV(key, rv.Uint(), opts.AttrType, attrs, opts.CDATA))
				case reflect.Float32, reflect.Float64:
					output.WriteString(ConvertKV(key, rv.Float(), opts.AttrType, attrs, opts.CDATA))
				case reflect.String:
					output.WriteString(ConvertKV(key, rv.String(), opts.AttrType, attrs, opts.CDATA))
				case reflect.Map:
					output.WriteString(Dict2XMLStr(opts, attrs, toMap(val), key, false, parent))
				case reflect.Slice, reflect.Array:
					output.WriteString(List2XMLStr(opts, attrs, toSlice(val), key))
				default:
					output.WriteString(ConvertKV(key, fmt.Sprintf("%v", val), opts.AttrType, attrs, opts.CDATA))
				}
			}
		}
	}

	return output.String()
}

// Dict2XMLStr parses dict to XML string.
func Dict2XMLStr(opts Options, attrs map[string]any, item map[string]any, itemName string, parentIsList bool, parent string) string {
	if opts.AttrType {
		attrs["type"] = GetXMLType(item)
	}

	// Check for custom @attrs
	var valAttrs map[string]any
	if customAttrs, ok := item["@attrs"]; ok {
		if ca, ok := customAttrs.(map[string]any); ok {
			valAttrs = ca
			delete(item, "@attrs")
		}
	}
	if valAttrs == nil {
		valAttrs = attrs
	}

	// Check for @val
	var rawItem any = item
	if val, ok := item["@val"]; ok {
		rawItem = val
		delete(item, "@val")
	}

	// Check for @flat
	flat := false
	if f, ok := item["@flat"]; ok {
		if fb, ok := f.(bool); ok && fb {
			flat = true
		}
		delete(item, "@flat")
	}

	var subtree string
	if IsPrimitiveType(rawItem) {
		switch v := rawItem.(type) {
		case string:
			subtree = EscapeXML(v)
		default:
			subtree = EscapeXML(fmt.Sprintf("%v", rawItem))
		}
	} else {
		subtree = Convert(rawItem, opts, itemName)
	}

	if parentIsList && opts.ListHeaders {
		if len(valAttrs) > 0 && !opts.ItemWrap {
			attrString := MakeAttrString(valAttrs)
			return fmt.Sprintf("<%s%s>%s</%s>", parent, attrString, subtree, parent)
		}
		return fmt.Sprintf("<%s>%s</%s>", parent, subtree, parent)
	} else if flat || (parentIsList && !opts.ItemWrap) {
		return subtree
	}

	attrString := MakeAttrString(valAttrs)
	return fmt.Sprintf("<%s%s>%s</%s>", itemName, attrString, subtree, itemName)
}

// List2XMLStr converts a list to XML string.
func List2XMLStr(opts Options, attrs map[string]any, items []any, itemName string) string {
	if opts.AttrType {
		attrs["type"] = GetXMLType(items)
	}

	flat := false
	if strings.HasSuffix(itemName, "@flat") {
		itemName = itemName[:len(itemName)-5]
		flat = true
	}

	subtree := ConvertList(items, opts, itemName)

	if flat || (len(items) > 0 && IsPrimitiveType(items[0]) && !opts.ItemWrap) {
		return subtree
	} else if opts.ListHeaders {
		return subtree
	}

	attrString := MakeAttrString(attrs)
	return fmt.Sprintf("<%s%s>%s</%s>", itemName, attrString, subtree, itemName)
}

// ConvertList converts a slice into an XML string.
func ConvertList(items []any, opts Options, parent string) string {
	var output strings.Builder
	itemName := opts.ItemFunc(parent)

	if strings.HasSuffix(itemName, "@flat") {
		itemName = itemName[:len(itemName)-5]
	}

	for _, item := range items {
		attrs := make(map[string]any)

		switch v := item.(type) {
		case bool:
			output.WriteString(ConvertBool(itemName, v, opts.AttrType, attrs, opts.CDATA))
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			if opts.ItemWrap {
				output.WriteString(ConvertKV(itemName, v, opts.AttrType, attrs, opts.CDATA))
			} else {
				output.WriteString(ConvertKV(parent, v, opts.AttrType, attrs, opts.CDATA))
			}
		case string:
			if opts.ItemWrap {
				output.WriteString(ConvertKV(itemName, v, opts.AttrType, attrs, opts.CDATA))
			} else {
				output.WriteString(ConvertKV(parent, v, opts.AttrType, attrs, opts.CDATA))
			}
		case time.Time:
			output.WriteString(ConvertKV(itemName, v.Format(time.RFC3339), opts.AttrType, attrs, opts.CDATA))
		case map[string]any:
			output.WriteString(Dict2XMLStr(opts, attrs, v, itemName, true, parent))
		case []any:
			output.WriteString(List2XMLStr(opts, attrs, v, itemName))
		default:
			if item == nil {
				output.WriteString(ConvertNone(itemName, opts.AttrType, attrs, opts.CDATA))
			} else {
				rv := reflect.ValueOf(item)
				switch rv.Kind() {
				case reflect.Bool:
					output.WriteString(ConvertBool(itemName, rv.Bool(), opts.AttrType, attrs, opts.CDATA))
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if opts.ItemWrap {
						output.WriteString(ConvertKV(itemName, rv.Int(), opts.AttrType, attrs, opts.CDATA))
					} else {
						output.WriteString(ConvertKV(parent, rv.Int(), opts.AttrType, attrs, opts.CDATA))
					}
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if opts.ItemWrap {
						output.WriteString(ConvertKV(itemName, rv.Uint(), opts.AttrType, attrs, opts.CDATA))
					} else {
						output.WriteString(ConvertKV(parent, rv.Uint(), opts.AttrType, attrs, opts.CDATA))
					}
				case reflect.Float32, reflect.Float64:
					if opts.ItemWrap {
						output.WriteString(ConvertKV(itemName, rv.Float(), opts.AttrType, attrs, opts.CDATA))
					} else {
						output.WriteString(ConvertKV(parent, rv.Float(), opts.AttrType, attrs, opts.CDATA))
					}
				case reflect.String:
					if opts.ItemWrap {
						output.WriteString(ConvertKV(itemName, rv.String(), opts.AttrType, attrs, opts.CDATA))
					} else {
						output.WriteString(ConvertKV(parent, rv.String(), opts.AttrType, attrs, opts.CDATA))
					}
				case reflect.Map:
					output.WriteString(Dict2XMLStr(opts, attrs, toMap(item), itemName, true, parent))
				case reflect.Slice, reflect.Array:
					output.WriteString(List2XMLStr(opts, attrs, toSlice(item), itemName))
				default:
					output.WriteString(ConvertKV(itemName, fmt.Sprintf("%v", item), opts.AttrType, attrs, opts.CDATA))
				}
			}
		}
	}

	return output.String()
}

// ConvertKV converts a key-value pair into an XML element.
func ConvertKV(key string, val any, attrType bool, attrs map[string]any, cdata bool) string {
	if attrs == nil {
		attrs = make(map[string]any)
	}
	key, attrs = MakeValidXMLName(key, attrs)

	// Convert time.Time to ISO format
	if t, ok := val.(time.Time); ok {
		val = t.Format(time.RFC3339)
	}

	if attrType {
		attrs["type"] = GetXMLType(val)
	}

	attrString := MakeAttrString(attrs)
	valStr := fmt.Sprintf("%v", val)
	if cdata {
		valStr = WrapCDATA(valStr)
	} else {
		valStr = EscapeXML(valStr)
	}

	return fmt.Sprintf("<%s%s>%s</%s>", key, attrString, valStr, key)
}

// ConvertBool converts a boolean into an XML element.
func ConvertBool(key string, val bool, attrType bool, attrs map[string]any, cdata bool) string {
	if attrs == nil {
		attrs = make(map[string]any)
	}
	key, attrs = MakeValidXMLName(key, attrs)

	if attrType {
		attrs["type"] = GetXMLType(val)
	}

	attrString := MakeAttrString(attrs)
	valStr := strings.ToLower(fmt.Sprintf("%v", val))
	return fmt.Sprintf("<%s%s>%s</%s>", key, attrString, valStr, key)
}

// ConvertNone converts a null value into an XML element.
func ConvertNone(key string, attrType bool, attrs map[string]any, cdata bool) string {
	if attrs == nil {
		attrs = make(map[string]any)
	}
	key, attrs = MakeValidXMLName(key, attrs)

	if attrType {
		attrs["type"] = GetXMLType(nil)
	}

	attrString := MakeAttrString(attrs)
	return fmt.Sprintf("<%s%s></%s>", key, attrString, key)
}

// DictToXML converts a Go value into XML bytes.
func DictToXML(obj any, opts Options) []byte {
	if opts.ItemFunc == nil {
		opts.ItemFunc = DefaultItemFunc
	}

	// Handle XPath format
	if opts.XPathFormat {
		xmlContent := ConvertToXPath31(obj, "")
		var output bytes.Buffer
		output.WriteString(`<?xml version="1.0" encoding="UTF-8" ?>`)

		if strings.HasPrefix(xmlContent, "<map") {
			xmlContent = strings.Replace(xmlContent, "<map", fmt.Sprintf(`<map xmlns="%s"`, XPathFunctionsNS), 1)
		} else if strings.HasPrefix(xmlContent, "<array") {
			xmlContent = strings.Replace(xmlContent, "<array", fmt.Sprintf(`<array xmlns="%s"`, XPathFunctionsNS), 1)
		} else {
			xmlContent = fmt.Sprintf(`<map xmlns="%s">%s</map>`, XPathFunctionsNS, xmlContent)
		}
		output.WriteString(xmlContent)
		return output.Bytes()
	}

	// Build namespace string
	var namespaceStr strings.Builder
	if opts.XMLNamespaces != nil {
		for prefix, ns := range opts.XMLNamespaces {
			switch prefix {
			case "xsi":
				if xsiMap, ok := ns.(map[string]any); ok {
					for schemaAttr, schemaVal := range xsiMap {
						if schemaAttr == "schemaInstance" {
							namespaceStr.WriteString(fmt.Sprintf(` xmlns:%s="%s"`, prefix, schemaVal))
						} else if schemaAttr == "schemaLocation" {
							namespaceStr.WriteString(fmt.Sprintf(` xsi:%s="%s"`, schemaAttr, schemaVal))
						}
					}
				}
			case "xmlns":
				namespaceStr.WriteString(fmt.Sprintf(` xmlns="%s"`, ns))
			default:
				namespaceStr.WriteString(fmt.Sprintf(` xmlns:%s="%s"`, prefix, ns))
			}
		}
	}

	var output bytes.Buffer
	if opts.Root {
		output.WriteString(`<?xml version="1.0" encoding="UTF-8" ?>`)
		outputElem := Convert(obj, opts, opts.CustomRoot)
		output.WriteString(fmt.Sprintf("<%s%s>%s</%s>", opts.CustomRoot, namespaceStr.String(), outputElem, opts.CustomRoot))
	} else {
		output.WriteString(Convert(obj, opts, ""))
	}

	return output.Bytes()
}

// PrettyPrint formats XML with indentation.
func PrettyPrint(xmlBytes []byte) (string, error) {
	var buf bytes.Buffer
	decoder := xml.NewDecoder(bytes.NewReader(xmlBytes))
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")

	for {
		token, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", err
		}
		if token == nil {
			break
		}
		if err := encoder.EncodeToken(token); err != nil {
			return "", err
		}
	}
	if err := encoder.Flush(); err != nil {
		return "", err
	}

	// Add XML declaration
	result := buf.String()
	if !strings.HasPrefix(result, "<?xml") {
		result = `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + result
	}

	// Ensure proper formatting with newlines after declaration
	re := regexp.MustCompile(`<\?xml[^?]*\?>`)
	result = re.ReplaceAllStringFunc(result, func(s string) string {
		return s + "\n"
	})

	return result, nil
}
