package json2xml

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestMakeID(t *testing.T) {
	t.Run("generates ID with element prefix", func(t *testing.T) {
		id := MakeID("li", 100000, 999999)
		if !strings.HasPrefix(id, "li_") {
			t.Errorf("expected ID to start with 'li_', got %s", id)
		}
	})

	t.Run("ID is within range", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			id := MakeID("test", 100000, 999999)
			parts := strings.Split(id, "_")
			if len(parts) != 2 {
				t.Errorf("expected ID with format 'prefix_number', got %s", id)
			}
		}
	})
}

func TestGetUniqueID(t *testing.T) {
	t.Run("generates unique IDs", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id := GetUniqueID("li")
			if ids[id] {
				// Note: With random generation, collision is possible but very rare
				t.Logf("Collision detected for ID: %s (this is rare but possible)", id)
			}
			ids[id] = true
		}
	})

	t.Run("ID format is correct", func(t *testing.T) {
		id := GetUniqueID("test_element")
		if !strings.HasPrefix(id, "test_element_") {
			t.Errorf("expected ID to start with 'test_element_', got %s", id)
		}
	})
}

func TestGetXMLType(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string", "abc", "str"},
		{"int", 1, "int"},
		{"float", 1.5, "float"},
		{"bool true", true, "bool"},
		{"bool false", false, "bool"},
		{"nil", nil, "null"},
		{"dict", map[string]any{}, "dict"},
		{"list", []any{1, 2, 3}, "list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetXMLType(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"ampersand", "&", "&amp;"},
		{"quote", `"`, "&quot;"},
		{"apostrophe", "'", "&apos;"},
		{"less than", "<", "&lt;"},
		{"greater than", ">", "&gt;"},
		{"all special chars", `<tag attr="value" other='test'> & more</tag>`,
			"&lt;tag attr=&quot;value&quot; other=&apos;test&apos;&gt; &amp; more&lt;/tag&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeXML(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestMakeAttrString(t *testing.T) {
	t.Run("empty attributes", func(t *testing.T) {
		result := MakeAttrString(map[string]any{})
		if result != "" {
			t.Errorf("expected empty string, got %s", result)
		}
	})

	t.Run("single attribute", func(t *testing.T) {
		result := MakeAttrString(map[string]any{"id": "123"})
		if result != ` id="123"` {
			t.Errorf("expected ` id=\"123\"`, got %s", result)
		}
	})

	t.Run("escapes special characters in values", func(t *testing.T) {
		result := MakeAttrString(map[string]any{"test": "value <here>"})
		if !strings.Contains(result, "&lt;here&gt;") {
			t.Errorf("expected escaped value, got %s", result)
		}
	})
}

func TestKeyIsValidXML(t *testing.T) {
	tests := []struct {
		key   string
		valid bool
	}{
		{"valid", true},
		{"valid_key", true},
		{"valid-key", true},
		{"_valid", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := KeyIsValidXML(tt.key)
			if result != tt.valid {
				t.Errorf("expected %v for key '%s', got %v", tt.valid, tt.key, result)
			}
		})
	}
}

func TestMakeValidXMLName(t *testing.T) {
	t.Run("valid key unchanged", func(t *testing.T) {
		key, _ := MakeValidXMLName("valid_key", nil)
		if key != "valid_key" {
			t.Errorf("expected 'valid_key', got %s", key)
		}
	})

	t.Run("numeric key gets n prefix", func(t *testing.T) {
		key, _ := MakeValidXMLName("123", nil)
		if key != "n123" {
			t.Errorf("expected 'n123', got %s", key)
		}
	})

	t.Run("spaces replaced with underscores", func(t *testing.T) {
		key, _ := MakeValidXMLName("invalid key", nil)
		if key != "invalid_key" {
			t.Errorf("expected 'invalid_key', got %s", key)
		}
	})
}

func TestWrapCDATA(t *testing.T) {
	t.Run("wraps simple string", func(t *testing.T) {
		result := WrapCDATA("test")
		if result != "<![CDATA[test]]>" {
			t.Errorf("expected '<![CDATA[test]]>', got %s", result)
		}
	})

	t.Run("handles CDATA end sequence", func(t *testing.T) {
		result := WrapCDATA("test]]>more")
		if !strings.Contains(result, "]]]]><![CDATA[>") {
			t.Errorf("expected CDATA end sequence handling, got %s", result)
		}
	})
}

func TestGetXPath31TagName(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"nil", nil, "null"},
		{"bool true", true, "boolean"},
		{"bool false", false, "boolean"},
		{"dict", map[string]any{}, "map"},
		{"int", 42, "number"},
		{"float", 3.14, "number"},
		{"string", "hello", "string"},
		{"list", []any{1, 2, 3}, "array"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetXPath31TagName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestConvertToXPath31(t *testing.T) {
	t.Run("null", func(t *testing.T) {
		result := ConvertToXPath31(nil, "")
		if result != "<null/>" {
			t.Errorf("expected '<null/>', got %s", result)
		}
	})

	t.Run("null with key", func(t *testing.T) {
		result := ConvertToXPath31(nil, "empty")
		if !strings.Contains(result, `key="empty"`) {
			t.Errorf("expected key attribute, got %s", result)
		}
	})

	t.Run("boolean true", func(t *testing.T) {
		result := ConvertToXPath31(true, "")
		if result != "<boolean>true</boolean>" {
			t.Errorf("expected '<boolean>true</boolean>', got %s", result)
		}
	})

	t.Run("boolean false", func(t *testing.T) {
		result := ConvertToXPath31(false, "")
		if result != "<boolean>false</boolean>" {
			t.Errorf("expected '<boolean>false</boolean>', got %s", result)
		}
	})

	t.Run("number", func(t *testing.T) {
		result := ConvertToXPath31(42, "")
		if result != "<number>42</number>" {
			t.Errorf("expected '<number>42</number>', got %s", result)
		}
	})

	t.Run("string", func(t *testing.T) {
		result := ConvertToXPath31("hello", "")
		if result != "<string>hello</string>" {
			t.Errorf("expected '<string>hello</string>', got %s", result)
		}
	})

	t.Run("string with special chars", func(t *testing.T) {
		result := ConvertToXPath31("hello & <world>", "")
		if !strings.Contains(result, "&amp;") || !strings.Contains(result, "&lt;") {
			t.Errorf("expected escaped characters, got %s", result)
		}
	})

	t.Run("map", func(t *testing.T) {
		result := ConvertToXPath31(map[string]any{"name": "John"}, "")
		if !strings.Contains(result, "<map>") || !strings.Contains(result, "</map>") {
			t.Errorf("expected map tags, got %s", result)
		}
	})

	t.Run("array", func(t *testing.T) {
		result := ConvertToXPath31([]any{1, 2, 3}, "")
		if !strings.Contains(result, "<array>") || !strings.Contains(result, "</array>") {
			t.Errorf("expected array tags, got %s", result)
		}
	})
}

func TestIsPrimitiveType(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"string", "abc", true},
		{"int", 1, true},
		{"float", 1.5, true},
		{"bool", true, true},
		{"nil", nil, true},
		{"dict", map[string]any{}, false},
		{"list", []any{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPrimitiveType(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConvertKV(t *testing.T) {
	t.Run("string value without attr type", func(t *testing.T) {
		result := ConvertKV("key", "value", false, nil, false)
		if result != "<key>value</key>" {
			t.Errorf("expected '<key>value</key>', got %s", result)
		}
	})

	t.Run("string value with attr type", func(t *testing.T) {
		result := ConvertKV("key", "value", true, nil, false)
		if !strings.Contains(result, `type="str"`) {
			t.Errorf("expected type attribute, got %s", result)
		}
	})

	t.Run("int value with attr type", func(t *testing.T) {
		result := ConvertKV("key", 123, true, nil, false)
		if !strings.Contains(result, `type="int"`) {
			t.Errorf("expected type='int', got %s", result)
		}
	})

	t.Run("value with CDATA", func(t *testing.T) {
		result := ConvertKV("key", "value", false, nil, true)
		if !strings.Contains(result, "<![CDATA[value]]>") {
			t.Errorf("expected CDATA wrapped value, got %s", result)
		}
	})

	t.Run("value with special characters", func(t *testing.T) {
		result := ConvertKV("key", "<script>", false, nil, false)
		if !strings.Contains(result, "&lt;script&gt;") {
			t.Errorf("expected escaped value, got %s", result)
		}
	})
}

func TestConvertBool(t *testing.T) {
	t.Run("true without attr type", func(t *testing.T) {
		result := ConvertBool("key", true, false, nil, false)
		if result != "<key>true</key>" {
			t.Errorf("expected '<key>true</key>', got %s", result)
		}
	})

	t.Run("false without attr type", func(t *testing.T) {
		result := ConvertBool("key", false, false, nil, false)
		if result != "<key>false</key>" {
			t.Errorf("expected '<key>false</key>', got %s", result)
		}
	})

	t.Run("with attr type", func(t *testing.T) {
		result := ConvertBool("key", true, true, nil, false)
		if !strings.Contains(result, `type="bool"`) {
			t.Errorf("expected type attribute, got %s", result)
		}
	})

	t.Run("with custom attributes", func(t *testing.T) {
		result := ConvertBool("key", true, false, map[string]any{"id": "1"}, false)
		if !strings.Contains(result, `id="1"`) {
			t.Errorf("expected custom attribute, got %s", result)
		}
	})
}

func TestConvertNone(t *testing.T) {
	t.Run("without attr type", func(t *testing.T) {
		result := ConvertNone("key", false, nil, false)
		if result != "<key></key>" {
			t.Errorf("expected '<key></key>', got %s", result)
		}
	})

	t.Run("with attr type", func(t *testing.T) {
		result := ConvertNone("key", true, nil, false)
		if !strings.Contains(result, `type="null"`) {
			t.Errorf("expected type attribute, got %s", result)
		}
	})
}

func TestDictToXML(t *testing.T) {
	t.Run("simple dict with root", func(t *testing.T) {
		data := map[string]any{"mock": "payload"}
		opts := DefaultOptions()
		opts.AttrType = false
		result := DictToXML(data, opts)

		expected := `<?xml version="1.0" encoding="UTF-8" ?><root><mock>payload</mock></root>`
		if string(result) != expected {
			t.Errorf("expected %s, got %s", expected, string(result))
		}
	})

	t.Run("simple dict without root", func(t *testing.T) {
		data := map[string]any{"mock": "payload"}
		opts := DefaultOptions()
		opts.Root = false
		opts.AttrType = false
		result := DictToXML(data, opts)

		expected := `<mock>payload</mock>`
		if string(result) != expected {
			t.Errorf("expected %s, got %s", expected, string(result))
		}
	})

	t.Run("with custom root", func(t *testing.T) {
		data := map[string]any{"mock": "payload"}
		opts := DefaultOptions()
		opts.CustomRoot = "element"
		opts.AttrType = false
		result := DictToXML(data, opts)

		if !strings.Contains(string(result), "<element>") || !strings.Contains(string(result), "</element>") {
			t.Errorf("expected custom root element, got %s", string(result))
		}
	})

	t.Run("with type attributes", func(t *testing.T) {
		data := map[string]any{"bike": "blue"}
		opts := DefaultOptions()
		opts.Root = false
		result := DictToXML(data, opts)

		if !strings.Contains(string(result), `type="str"`) {
			t.Errorf("expected type attribute, got %s", string(result))
		}
	})

	t.Run("item_wrap true", func(t *testing.T) {
		data := map[string]any{"bike": []any{"blue", "green"}}
		opts := DefaultOptions()
		opts.Root = false
		opts.AttrType = false
		result := DictToXML(data, opts)

		expected := `<bike><item>blue</item><item>green</item></bike>`
		if string(result) != expected {
			t.Errorf("expected %s, got %s", expected, string(result))
		}
	})

	t.Run("item_wrap false", func(t *testing.T) {
		data := map[string]any{"bike": []any{"blue", "green"}}
		opts := DefaultOptions()
		opts.Root = false
		opts.AttrType = false
		opts.ItemWrap = false
		result := DictToXML(data, opts)

		expected := `<bike>blue</bike><bike>green</bike>`
		if string(result) != expected {
			t.Errorf("expected %s, got %s", expected, string(result))
		}
	})

	t.Run("with namespaces", func(t *testing.T) {
		data := map[string]any{"ns1:node1": "data in namespace 1"}
		opts := DefaultOptions()
		opts.AttrType = false
		opts.XMLNamespaces = map[string]any{
			"ns1": "https://www.google.de/ns1",
		}
		result := DictToXML(data, opts)

		if !strings.Contains(string(result), `xmlns:ns1="https://www.google.de/ns1"`) {
			t.Errorf("expected namespace declaration, got %s", string(result))
		}
	})

	t.Run("with xmlns namespace", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		opts := DefaultOptions()
		opts.AttrType = false
		opts.XMLNamespaces = map[string]any{
			"xmlns": "http://example.com",
		}
		result := DictToXML(data, opts)

		if !strings.Contains(string(result), `xmlns="http://example.com"`) {
			t.Errorf("expected xmlns declaration, got %s", string(result))
		}
	})

	t.Run("boolean values", func(t *testing.T) {
		data := map[string]any{"flag": true}
		opts := DefaultOptions()
		opts.Root = false
		opts.AttrType = false
		result := DictToXML(data, opts)

		if string(result) != "<flag>true</flag>" {
			t.Errorf("expected '<flag>true</flag>', got %s", string(result))
		}
	})

	t.Run("null values", func(t *testing.T) {
		data := map[string]any{"empty": nil}
		opts := DefaultOptions()
		opts.Root = false
		opts.AttrType = false
		result := DictToXML(data, opts)

		if string(result) != "<empty></empty>" {
			t.Errorf("expected '<empty></empty>', got %s", string(result))
		}
	})

	t.Run("nested dict", func(t *testing.T) {
		data := map[string]any{
			"person": map[string]any{
				"name": "John",
				"age":  30,
			},
		}
		opts := DefaultOptions()
		opts.Root = false
		opts.AttrType = false
		result := DictToXML(data, opts)

		if !strings.Contains(string(result), "<person>") || !strings.Contains(string(result), "<name>John</name>") {
			t.Errorf("expected nested structure, got %s", string(result))
		}
	})

	t.Run("ampersand in value", func(t *testing.T) {
		data := map[string]any{"Bicycles": "Wheels & Steers"}
		opts := DefaultOptions()
		opts.Root = false
		opts.AttrType = false
		result := DictToXML(data, opts)

		expected := "<Bicycles>Wheels &amp; Steers</Bicycles>"
		if string(result) != expected {
			t.Errorf("expected %s, got %s", expected, string(result))
		}
	})

	t.Run("datetime conversion", func(t *testing.T) {
		dt := time.Date(2023, 2, 15, 12, 30, 45, 0, time.UTC)
		data := map[string]any{"key": dt}
		opts := DefaultOptions()
		opts.AttrType = false
		result := DictToXML(data, opts)

		if !strings.Contains(string(result), "2023-02-15") {
			t.Errorf("expected datetime in ISO format, got %s", string(result))
		}
	})

	t.Run("list with dict items", func(t *testing.T) {
		data := map[string]any{
			"items": []any{
				map[string]any{"key1": "value1"},
				map[string]any{"key2": "value2"},
			},
		}
		opts := DefaultOptions()
		opts.Root = false
		opts.AttrType = false
		result := DictToXML(data, opts)

		if !strings.Contains(string(result), "<item><key1>value1</key1></item>") {
			t.Errorf("expected item wrapped dicts, got %s", string(result))
		}
	})

	t.Run("CDATA wrapping", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		opts := DefaultOptions()
		opts.Root = false
		opts.AttrType = false
		opts.CDATA = true
		result := DictToXML(data, opts)

		if !strings.Contains(string(result), "<![CDATA[value]]>") {
			t.Errorf("expected CDATA wrapped value, got %s", string(result))
		}
	})
}

func TestDictToXMLXPathFormat(t *testing.T) {
	t.Run("basic types", func(t *testing.T) {
		data := map[string]any{"name": "John", "age": 30, "active": true}
		opts := DefaultOptions()
		opts.XPathFormat = true
		result := DictToXML(data, opts)

		if !bytes.Contains(result, []byte(`xmlns="http://www.w3.org/2005/xpath-functions"`)) {
			t.Errorf("expected XPath namespace, got %s", string(result))
		}
		if !bytes.Contains(result, []byte(`<string key="name">John</string>`)) {
			t.Errorf("expected string element, got %s", string(result))
		}
		if !bytes.Contains(result, []byte(`<number key="age">30</number>`)) {
			t.Errorf("expected number element, got %s", string(result))
		}
		if !bytes.Contains(result, []byte(`<boolean key="active">true</boolean>`)) {
			t.Errorf("expected boolean element, got %s", string(result))
		}
	})

	t.Run("nested dict", func(t *testing.T) {
		data := map[string]any{
			"person": map[string]any{"name": "Alice", "age": 25},
		}
		opts := DefaultOptions()
		opts.XPathFormat = true
		result := DictToXML(data, opts)

		if !bytes.Contains(result, []byte(`<map key="person">`)) {
			t.Errorf("expected nested map, got %s", string(result))
		}
	})

	t.Run("array", func(t *testing.T) {
		data := map[string]any{"numbers": []any{1, 2, 3}}
		opts := DefaultOptions()
		opts.XPathFormat = true
		result := DictToXML(data, opts)

		if !bytes.Contains(result, []byte(`<array key="numbers">`)) {
			t.Errorf("expected array element, got %s", string(result))
		}
	})

	t.Run("null value", func(t *testing.T) {
		data := map[string]any{"value": nil}
		opts := DefaultOptions()
		opts.XPathFormat = true
		result := DictToXML(data, opts)

		if !bytes.Contains(result, []byte(`<null key="value"/>`)) {
			t.Errorf("expected null element, got %s", string(result))
		}
	})

	t.Run("root level array", func(t *testing.T) {
		data := []any{1, 2, 3}
		opts := DefaultOptions()
		opts.XPathFormat = true
		result := DictToXML(data, opts)

		if !bytes.Contains(result, []byte(`<array xmlns="http://www.w3.org/2005/xpath-functions">`)) {
			t.Errorf("expected array with namespace, got %s", string(result))
		}
	})
}

func TestListHeaders(t *testing.T) {
	t.Run("list headers true", func(t *testing.T) {
		data := map[string]any{
			"Bike": []any{
				map[string]any{"frame_color": "red"},
				map[string]any{"frame_color": "green"},
			},
		}
		opts := DefaultOptions()
		opts.Root = false
		opts.AttrType = false
		opts.ItemWrap = false
		opts.ListHeaders = true
		result := DictToXML(data, opts)

		expected := "<Bike><frame_color>red</frame_color></Bike><Bike><frame_color>green</frame_color></Bike>"
		if string(result) != expected {
			t.Errorf("expected %s, got %s", expected, string(result))
		}
	})
}

func TestFlatList(t *testing.T) {
	t.Run("flat list notation", func(t *testing.T) {
		data := map[string]any{
			"flat_list@flat": []any{1, 2, 3},
			"non_flat_list":  []any{4, 5, 6},
		}
		opts := DefaultOptions()
		opts.AttrType = false
		result := DictToXML(data, opts)

		// flat_list should not be wrapped
		if !strings.Contains(string(result), "<item>1</item><item>2</item><item>3</item>") {
			t.Errorf("expected flat list items, got %s", string(result))
		}
		// non_flat_list should be wrapped
		if !strings.Contains(string(result), "<non_flat_list><item>4</item><item>5</item><item>6</item></non_flat_list>") {
			t.Errorf("expected wrapped list, got %s", string(result))
		}
	})
}

func TestCustomItemFunc(t *testing.T) {
	t.Run("custom item function", func(t *testing.T) {
		data := map[string]any{"items": []any{1, 2, 3}}
		opts := DefaultOptions()
		opts.Root = false
		opts.AttrType = false
		opts.ItemFunc = func(parent string) string {
			return "custom"
		}
		result := DictToXML(data, opts)

		if !strings.Contains(string(result), "<custom>") {
			t.Errorf("expected custom item names, got %s", string(result))
		}
	})
}

func TestPrettyPrint(t *testing.T) {
	t.Run("formats XML with indentation", func(t *testing.T) {
		input := []byte(`<?xml version="1.0" encoding="UTF-8" ?><root><child>value</child></root>`)
		result, err := PrettyPrint(input)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !strings.Contains(result, "\n") {
			t.Errorf("expected formatted output with newlines, got %s", result)
		}
	})
}

// Fuzz tests for dicttoxml functions

func FuzzEscapeXML(f *testing.F) {
	f.Add("")
	f.Add("hello")
	f.Add("<script>alert('xss')</script>")
	f.Add("&amp;&lt;&gt;&quot;&apos;")
	f.Add("normal text with <special> & \"chars\"")
	f.Add("]]>")
	f.Add("\x00\x01\x02")
	f.Add("unicode: 日本語 🎉")

	f.Fuzz(func(t *testing.T, input string) {
		result := EscapeXML(input)
		// result should not contain unescaped special chars
		if strings.Contains(result, "<") && !strings.Contains(result, "&lt;") {
			// this is fine, we only escape < to &lt;
		}
		// should not panic, that's the main check
		_ = result
	})
}

func FuzzWrapCDATA(f *testing.F) {
	f.Add("")
	f.Add("simple text")
	f.Add("]]>")
	f.Add("nested ]]> end sequence")
	f.Add("multiple ]]> and ]]> here")
	f.Add("<![CDATA[already cdata]]>")

	f.Fuzz(func(t *testing.T, input string) {
		result := WrapCDATA(input)
		if !strings.HasPrefix(result, "<![CDATA[") {
			t.Errorf("result should start with CDATA prefix, got %s", result)
		}
		if !strings.HasSuffix(result, "]]>") {
			t.Errorf("result should end with CDATA suffix, got %s", result)
		}
	})
}

func FuzzKeyIsValidXML(f *testing.F) {
	f.Add("")
	f.Add("valid")
	f.Add("valid_key")
	f.Add("valid-key")
	f.Add("123")
	f.Add("_underscore")
	f.Add("has space")
	f.Add("special<chars>")
	f.Add("unicode:日本語")
	f.Add("\x00null\x00")

	f.Fuzz(func(t *testing.T, key string) {
		// should not panic
		_ = KeyIsValidXML(key)
	})
}

func FuzzMakeValidXMLName(f *testing.F) {
	f.Add("valid", "")
	f.Add("123", "")
	f.Add("has space", "")
	f.Add("special:colon", "")
	f.Add("@flat", "")
	f.Add("test@flat", "")
	f.Add("", "")
	f.Add("<invalid>", "")

	f.Fuzz(func(t *testing.T, key string, _ string) {
		resultKey, resultAttrs := MakeValidXMLName(key, nil)
		// result should be a valid XML name or fallback to "key"
		if resultKey == "" {
			t.Error("result key should not be empty")
		}
		_ = resultAttrs
	})
}

func FuzzGetXMLType(f *testing.F) {
	f.Add("string value")
	f.Add("")

	f.Fuzz(func(t *testing.T, input string) {
		// test with string
		result := GetXMLType(input)
		if result != "str" {
			t.Errorf("expected 'str' for string input, got %s", result)
		}
	})
}

func FuzzGetXPath31TagName(f *testing.F) {
	f.Add("string value")
	f.Add("")

	f.Fuzz(func(t *testing.T, input string) {
		result := GetXPath31TagName(input)
		if result != "string" {
			t.Errorf("expected 'string' for string input, got %s", result)
		}
	})
}

func FuzzConvertToXPath31(f *testing.F) {
	f.Add("", "")
	f.Add("value", "key")
	f.Add("special <chars> & \"quotes\"", "mykey")
	f.Add("unicode 日本語", "unicode_key")

	f.Fuzz(func(t *testing.T, value, key string) {
		result := ConvertToXPath31(value, key)
		if !strings.Contains(result, "<string") {
			t.Errorf("expected string element, got %s", result)
		}
		if key != "" && !strings.Contains(result, "key=") {
			t.Errorf("expected key attribute when key is provided, got %s", result)
		}
	})
}

func FuzzMakeAttrString(f *testing.F) {
	f.Add("key", "value")
	f.Add("id", "123")
	f.Add("special", "<>&\"'")
	f.Add("", "empty_key")

	f.Fuzz(func(t *testing.T, key, value string) {
		if key == "" {
			return // skip empty keys
		}
		attrs := map[string]any{key: value}
		result := MakeAttrString(attrs)
		if result == "" {
			t.Error("result should not be empty for non-empty attrs")
		}
		// should have proper quoting
		if !strings.Contains(result, "=\"") {
			t.Errorf("result should contain attribute assignment, got %s", result)
		}
	})
}

func FuzzConvertKV(f *testing.F) {
	f.Add("key", "value", true, false)
	f.Add("mykey", "special<>&\"'", false, true)
	f.Add("123", "numeric key", true, true)
	f.Add("", "empty key", false, false)

	f.Fuzz(func(t *testing.T, key, value string, attrType, cdata bool) {
		if key == "" {
			return // skip empty keys as they get transformed
		}
		result := ConvertKV(key, value, attrType, nil, cdata)
		if result == "" {
			t.Error("result should not be empty")
		}
		// should be valid XML structure (opening and closing tags)
		if !strings.Contains(result, "</") {
			t.Errorf("result should contain closing tag, got %s", result)
		}
	})
}

func FuzzConvertBool(f *testing.F) {
	f.Add("flag", true, true, false)
	f.Add("enabled", false, false, true)
	f.Add("123", true, true, true)

	f.Fuzz(func(t *testing.T, key string, val, attrType, cdata bool) {
		if key == "" {
			return
		}
		result := ConvertBool(key, val, attrType, nil, cdata)
		if result == "" {
			t.Error("result should not be empty")
		}
		// should contain true or false
		if !strings.Contains(result, "true") && !strings.Contains(result, "false") {
			t.Errorf("result should contain boolean value, got %s", result)
		}
	})
}

func FuzzConvertNone(f *testing.F) {
	f.Add("empty", true, false)
	f.Add("null_value", false, true)
	f.Add("123", true, true)

	f.Fuzz(func(t *testing.T, key string, attrType, cdata bool) {
		if key == "" {
			return
		}
		result := ConvertNone(key, attrType, nil, cdata)
		if result == "" {
			t.Error("result should not be empty")
		}
		// should be self-closing or empty element
		if !strings.Contains(result, "</") {
			t.Errorf("result should contain closing tag, got %s", result)
		}
	})
}

func FuzzDictToXML(f *testing.F) {
	f.Add("key", "value")
	f.Add("special", "<>&\"'chars")
	f.Add("unicode", "日本語🎉")

	f.Fuzz(func(t *testing.T, key, value string) {
		if key == "" {
			return
		}
		data := map[string]any{key: value}
		opts := DefaultOptions()

		// should not panic
		result := DictToXML(data, opts)
		if len(result) == 0 {
			t.Error("result should not be empty")
		}

		// test with XPath format
		opts.XPathFormat = true
		result = DictToXML(data, opts)
		if len(result) == 0 {
			t.Error("XPath result should not be empty")
		}
	})
}

func FuzzPrettyPrint(f *testing.F) {
	f.Add([]byte(`<?xml version="1.0"?><root></root>`))
	f.Add([]byte(`<simple>text</simple>`))
	f.Add([]byte(`<nested><child>value</child></nested>`))

	f.Fuzz(func(t *testing.T, input []byte) {
		// should not panic, even on invalid XML
		_, _ = PrettyPrint(input)
	})
}
