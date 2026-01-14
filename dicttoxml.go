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

// XPathFunctionsNS is the XPath 3.1 json-to-xml namespace.
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
	testXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" ?><%s>foo</%s>`, key, key)
	decoder := xml.NewDecoder(strings.NewReader(testXML))
	for {
		tok, err := decoder.Token()
		if err != nil {
			return err.Error() == "EOF"
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

	if KeyIsValidXML(key) {
		return key, attrs
	}

	if isNumeric(key) {
		return "n" + key, attrs
	}

	keyWithUnderscores := strings.ReplaceAll(key, " ", "_")
	if KeyIsValidXML(keyWithUnderscores) {
		return keyWithUnderscores, attrs
	}

	keyClean := strings.ReplaceAll(key, ":", "")
	keyClean = strings.ReplaceAll(keyClean, "@flat", "")
	if KeyIsValidXML(keyClean) {
		return key, attrs
	}

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

	switch GetXPath31TagName(obj) {
	case "null":
		return fmt.Sprintf("<null%s/>", keyAttr)
	case "boolean":
		return fmt.Sprintf("<boolean%s>%s</boolean>", keyAttr, strings.ToLower(fmt.Sprintf("%v", obj)))
	case "number":
		return fmt.Sprintf("<number%s>%v</number>", keyAttr, obj)
	case "string":
		return fmt.Sprintf("<string%s>%s</string>", keyAttr, EscapeXML(fmt.Sprintf("%v", obj)))
	case "map":
		return convertXPathMap(obj, keyAttr)
	case "array":
		return convertXPathArray(obj, keyAttr)
	default:
		return fmt.Sprintf("<string%s>%s</string>", keyAttr, EscapeXML(fmt.Sprintf("%v", obj)))
	}
}

func convertXPathMap(obj any, keyAttr string) string {
	var children strings.Builder
	m := toMap(obj)
	keys := sortedKeys(m)
	for _, k := range keys {
		children.WriteString(ConvertToXPath31(m[k], k))
	}
	return fmt.Sprintf("<map%s>%s</map>", keyAttr, children.String())
}

func convertXPathArray(obj any, keyAttr string) string {
	var children strings.Builder
	for _, item := range toSlice(obj) {
		children.WriteString(ConvertToXPath31(item, ""))
	}
	return fmt.Sprintf("<array%s>%s</array>", keyAttr, children.String())
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

// sortedKeys returns sorted keys from a map.
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// IsPrimitiveType checks if a value is a primitive type.
func IsPrimitiveType(val any) bool {
	t := GetXMLType(val)
	return t == "str" || t == "int" || t == "float" || t == "bool" || t == "null"
}

// normalizeValue converts a value to a standard type using reflection.
// Returns the normalized value suitable for XML conversion.
func normalizeValue(val any) any {
	if val == nil {
		return nil
	}

	// Handle common concrete types directly (fast path)
	switch v := val.(type) {
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64,
		float32, float64, string, map[string]any, []any:
		return v
	case time.Time:
		return v.Format(time.RFC3339)
	}

	// Use reflection for other types
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint()
	case reflect.Float32, reflect.Float64:
		return rv.Float()
	case reflect.String:
		return rv.String()
	case reflect.Map:
		return toMap(val)
	case reflect.Slice, reflect.Array:
		return toSlice(val)
	default:
		return fmt.Sprintf("%v", val)
	}
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
		reflect.Float32, reflect.Float64, reflect.String:
		return ConvertKV(itemName, obj, opts.AttrType, nil, opts.CDATA)
	case reflect.Map:
		return ConvertDict(toMap(obj), opts, parent)
	case reflect.Slice, reflect.Array:
		return ConvertList(toSlice(obj), opts, parent)
	default:
		if t, ok := obj.(time.Time); ok {
			return ConvertKV(itemName, t.Format(time.RFC3339), opts.AttrType, nil, opts.CDATA)
		}
		return ConvertKV(itemName, fmt.Sprintf("%v", obj), opts.AttrType, nil, opts.CDATA)
	}
}

// ConvertDict converts a map into an XML string.
func ConvertDict(obj map[string]any, opts Options, parent string) string {
	var output strings.Builder

	for _, key := range sortedKeys(obj) {
		val := obj[key]
		attrs := make(map[string]any)

		if opts.IDs {
			attrs["id"] = GetUniqueID(parent)
		}

		key, attrs = MakeValidXMLName(key, attrs)
		output.WriteString(convertDictValue(key, val, attrs, opts, parent))
	}

	return output.String()
}

// convertDictValue handles conversion of a single dictionary value.
func convertDictValue(key string, val any, attrs map[string]any, opts Options, parent string) string {
	normalized := normalizeValue(val)

	switch v := normalized.(type) {
	case nil:
		return ConvertNone(key, opts.AttrType, attrs, opts.CDATA)
	case bool:
		return ConvertBool(key, v, opts.AttrType, attrs, opts.CDATA)
	case map[string]any:
		return Dict2XMLStr(opts, attrs, v, key, false, parent)
	case []any:
		return List2XMLStr(opts, attrs, v, key)
	default:
		return ConvertKV(key, v, opts.AttrType, attrs, opts.CDATA)
	}
}

// Dict2XMLStr parses dict to XML string.
func Dict2XMLStr(opts Options, attrs map[string]any, item map[string]any, itemName string, parentIsList bool, parent string) string {
	if opts.AttrType {
		attrs["type"] = GetXMLType(item)
	}

	valAttrs, rawItem, flat := extractSpecialAttrs(item, attrs)
	subtree := buildSubtree(rawItem, opts, itemName)

	return formatDictOutput(valAttrs, subtree, itemName, parent, parentIsList, flat, opts)
}

// extractSpecialAttrs extracts @attrs, @val, and @flat from an item.
func extractSpecialAttrs(item map[string]any, defaultAttrs map[string]any) (attrs map[string]any, rawItem any, flat bool) {
	attrs = defaultAttrs
	rawItem = item

	if customAttrs, ok := item["@attrs"]; ok {
		if ca, ok := customAttrs.(map[string]any); ok {
			attrs = ca
			delete(item, "@attrs")
		}
	}

	if val, ok := item["@val"]; ok {
		rawItem = val
		delete(item, "@val")
	}

	if f, ok := item["@flat"]; ok {
		if fb, ok := f.(bool); ok && fb {
			flat = true
		}
		delete(item, "@flat")
	}

	return attrs, rawItem, flat
}

// buildSubtree creates the XML subtree for a value.
func buildSubtree(rawItem any, opts Options, itemName string) string {
	if IsPrimitiveType(rawItem) {
		if v, ok := rawItem.(string); ok {
			return EscapeXML(v)
		}
		return EscapeXML(fmt.Sprintf("%v", rawItem))
	}
	return Convert(rawItem, opts, itemName)
}

// formatDictOutput formats the final dict XML output.
func formatDictOutput(valAttrs map[string]any, subtree, itemName, parent string, parentIsList, flat bool, opts Options) string {
	if parentIsList && opts.ListHeaders {
		if len(valAttrs) > 0 && !opts.ItemWrap {
			return fmt.Sprintf("<%s%s>%s</%s>", parent, MakeAttrString(valAttrs), subtree, parent)
		}
		return fmt.Sprintf("<%s>%s</%s>", parent, subtree, parent)
	}

	if flat || (parentIsList && !opts.ItemWrap) {
		return subtree
	}

	return fmt.Sprintf("<%s%s>%s</%s>", itemName, MakeAttrString(valAttrs), subtree, itemName)
}

// List2XMLStr converts a list to XML string.
func List2XMLStr(opts Options, attrs map[string]any, items []any, itemName string) string {
	if opts.AttrType {
		attrs["type"] = GetXMLType(items)
	}

	flat := strings.HasSuffix(itemName, "@flat")
	itemName = strings.TrimSuffix(itemName, "@flat")

	subtree := ConvertList(items, opts, itemName)

	if flat || (len(items) > 0 && IsPrimitiveType(items[0]) && !opts.ItemWrap) || opts.ListHeaders {
		return subtree
	}

	return fmt.Sprintf("<%s%s>%s</%s>", itemName, MakeAttrString(attrs), subtree, itemName)
}

// ConvertList converts a slice into an XML string.
func ConvertList(items []any, opts Options, parent string) string {
	var output strings.Builder
	itemName := strings.TrimSuffix(opts.ItemFunc(parent), "@flat")

	for _, item := range items {
		output.WriteString(convertListItem(item, itemName, parent, opts))
	}

	return output.String()
}

// convertListItem handles conversion of a single list item.
func convertListItem(item any, itemName, parent string, opts Options) string {
	attrs := make(map[string]any)
	normalized := normalizeValue(item)

	switch v := normalized.(type) {
	case nil:
		return ConvertNone(itemName, opts.AttrType, attrs, opts.CDATA)
	case bool:
		return ConvertBool(itemName, v, opts.AttrType, attrs, opts.CDATA)
	case map[string]any:
		return Dict2XMLStr(opts, attrs, v, itemName, true, parent)
	case []any:
		return List2XMLStr(opts, attrs, v, itemName)
	default:
		name := itemName
		if !opts.ItemWrap {
			name = parent
		}
		return ConvertKV(name, v, opts.AttrType, attrs, opts.CDATA)
	}
}

// ConvertKV converts a key-value pair into an XML element.
func ConvertKV(key string, val any, attrType bool, attrs map[string]any, cdata bool) string {
	if attrs == nil {
		attrs = make(map[string]any)
	}
	key, attrs = MakeValidXMLName(key, attrs)

	if t, ok := val.(time.Time); ok {
		val = t.Format(time.RFC3339)
	}

	if attrType {
		attrs["type"] = GetXMLType(val)
	}

	valStr := fmt.Sprintf("%v", val)
	if cdata {
		valStr = WrapCDATA(valStr)
	} else {
		valStr = EscapeXML(valStr)
	}

	return fmt.Sprintf("<%s%s>%s</%s>", key, MakeAttrString(attrs), valStr, key)
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

	return fmt.Sprintf("<%s%s>%s</%s>", key, MakeAttrString(attrs), strings.ToLower(fmt.Sprintf("%v", val)), key)
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

	return fmt.Sprintf("<%s%s></%s>", key, MakeAttrString(attrs), key)
}

// DictToXML converts a Go value into XML bytes.
func DictToXML(obj any, opts Options) []byte {
	if opts.ItemFunc == nil {
		opts.ItemFunc = DefaultItemFunc
	}

	if opts.XPathFormat {
		return buildXPathXML(obj)
	}

	return buildStandardXML(obj, opts)
}

// buildXPathXML creates XML in XPath 3.1 format.
func buildXPathXML(obj any) []byte {
	xmlContent := ConvertToXPath31(obj, "")
	var output bytes.Buffer
	output.WriteString(`<?xml version="1.0" encoding="UTF-8" ?>`)

	switch {
	case strings.HasPrefix(xmlContent, "<map"):
		xmlContent = strings.Replace(xmlContent, "<map", fmt.Sprintf(`<map xmlns="%s"`, XPathFunctionsNS), 1)
	case strings.HasPrefix(xmlContent, "<array"):
		xmlContent = strings.Replace(xmlContent, "<array", fmt.Sprintf(`<array xmlns="%s"`, XPathFunctionsNS), 1)
	default:
		xmlContent = fmt.Sprintf(`<map xmlns="%s">%s</map>`, XPathFunctionsNS, xmlContent)
	}
	output.WriteString(xmlContent)
	return output.Bytes()
}

// buildStandardXML creates XML in standard format.
func buildStandardXML(obj any, opts Options) []byte {
	var output bytes.Buffer
	if opts.Root {
		output.WriteString(`<?xml version="1.0" encoding="UTF-8" ?>`)
		outputElem := Convert(obj, opts, opts.CustomRoot)
		namespaceStr := buildNamespaceString(opts.XMLNamespaces)
		output.WriteString(fmt.Sprintf("<%s%s>%s</%s>", opts.CustomRoot, namespaceStr, outputElem, opts.CustomRoot))
	} else {
		output.WriteString(Convert(obj, opts, ""))
	}
	return output.Bytes()
}

// buildNamespaceString creates the namespace attribute string.
func buildNamespaceString(namespaces map[string]any) string {
	if namespaces == nil {
		return ""
	}

	var ns strings.Builder
	for prefix, value := range namespaces {
		switch prefix {
		case "xsi":
			ns.WriteString(buildXSINamespace(prefix, value))
		case "xmlns":
			ns.WriteString(fmt.Sprintf(` xmlns="%s"`, value))
		default:
			ns.WriteString(fmt.Sprintf(` xmlns:%s="%s"`, prefix, value))
		}
	}
	return ns.String()
}

// buildXSINamespace handles the special xsi namespace.
func buildXSINamespace(prefix string, value any) string {
	xsiMap, ok := value.(map[string]any)
	if !ok {
		return ""
	}

	var ns strings.Builder
	for schemaAttr, schemaVal := range xsiMap {
		switch schemaAttr {
		case "schemaInstance":
			ns.WriteString(fmt.Sprintf(` xmlns:%s="%s"`, prefix, schemaVal))
		case "schemaLocation":
			ns.WriteString(fmt.Sprintf(` xsi:%s="%s"`, schemaAttr, schemaVal))
		}
	}
	return ns.String()
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

	result := buf.String()
	if !strings.HasPrefix(result, "<?xml") {
		result = `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + result
	}

	re := regexp.MustCompile(`<\?xml[^?]*\?>`)
	result = re.ReplaceAllStringFunc(result, func(s string) string {
		return s + "\n"
	})

	return result, nil
}
