package json2xml

import (
	"bytes"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("creates new converter with defaults", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		conv := New(data)

		if conv == nil {
			t.Fatal("expected converter, got nil")
		}
		if conv.wrapper != "all" {
			t.Errorf("expected wrapper 'all', got %s", conv.wrapper)
		}
		if !conv.root {
			t.Error("expected root to be true")
		}
		if !conv.pretty {
			t.Error("expected pretty to be true")
		}
		if !conv.attrType {
			t.Error("expected attrType to be true")
		}
		if !conv.itemWrap {
			t.Error("expected itemWrap to be true")
		}
	})
}

func TestWithMethods(t *testing.T) {
	t.Run("WithWrapper", func(t *testing.T) {
		conv := New(nil).WithWrapper("custom")
		if conv.wrapper != "custom" {
			t.Errorf("expected wrapper 'custom', got %s", conv.wrapper)
		}
	})

	t.Run("WithRoot", func(t *testing.T) {
		conv := New(nil).WithRoot(false)
		if conv.root {
			t.Error("expected root to be false")
		}
	})

	t.Run("WithPretty", func(t *testing.T) {
		conv := New(nil).WithPretty(false)
		if conv.pretty {
			t.Error("expected pretty to be false")
		}
	})

	t.Run("WithAttrType", func(t *testing.T) {
		conv := New(nil).WithAttrType(false)
		if conv.attrType {
			t.Error("expected attrType to be false")
		}
	})

	t.Run("WithItemWrap", func(t *testing.T) {
		conv := New(nil).WithItemWrap(false)
		if conv.itemWrap {
			t.Error("expected itemWrap to be false")
		}
	})

	t.Run("WithXPathFormat", func(t *testing.T) {
		conv := New(nil).WithXPathFormat(true)
		if !conv.xpathFormat {
			t.Error("expected xpathFormat to be true")
		}
	})

	t.Run("method chaining", func(t *testing.T) {
		conv := New(nil).
			WithWrapper("test").
			WithRoot(false).
			WithPretty(false).
			WithAttrType(false).
			WithItemWrap(false)

		if conv.wrapper != "test" || conv.root || conv.pretty || conv.attrType || conv.itemWrap {
			t.Error("method chaining didn't work correctly")
		}
	})
}

func TestToXML(t *testing.T) {
	t.Run("nil data returns nil", func(t *testing.T) {
		result, err := New(nil).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("empty map returns nil", func(t *testing.T) {
		result, err := New(map[string]any{}).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("empty slice returns nil", func(t *testing.T) {
		result, err := New([]any{}).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("simple dict with pretty print", func(t *testing.T) {
		data := map[string]any{"login": "mojombo", "id": 1}
		result, err := New(data).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlStr, ok := result.(string)
		if !ok {
			t.Fatalf("expected string, got %T", result)
		}

		if !strings.Contains(xmlStr, `encoding="UTF-8"`) {
			t.Errorf("expected encoding declaration, got %s", xmlStr)
		}
	})

	t.Run("simple dict without pretty print", func(t *testing.T) {
		data := map[string]any{"login": "mojombo", "id": 1}
		result, err := New(data).WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlBytes, ok := result.([]byte)
		if !ok {
			t.Fatalf("expected []byte, got %T", result)
		}

		if !bytes.Contains(xmlBytes, []byte(`encoding="UTF-8"`)) {
			t.Errorf("expected encoding declaration, got %s", string(xmlBytes))
		}
	})

	t.Run("custom wrapper", func(t *testing.T) {
		data := map[string]any{"login": "mojombo"}
		result, err := New(data).WithWrapper("test").WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlBytes := result.([]byte)
		if !bytes.Contains(xmlBytes, []byte("<test>")) || !bytes.Contains(xmlBytes, []byte("</test>")) {
			t.Errorf("expected custom wrapper, got %s", string(xmlBytes))
		}
	})

	t.Run("no root wrapper", func(t *testing.T) {
		data := map[string]any{"login": "mojombo"}
		result, err := New(data).WithRoot(false).WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlBytes := result.([]byte)
		if bytes.Contains(xmlBytes, []byte("<all>")) {
			t.Errorf("expected no wrapper, got %s", string(xmlBytes))
		}
	})

	t.Run("item wrap true", func(t *testing.T) {
		data := map[string]any{
			"my_items":     []any{map[string]any{"id": 1}, map[string]any{"id": 2}},
			"my_str_items": []any{"a", "b"},
		}
		result, err := New(data).WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlBytes := result.([]byte)
		// Item elements should be present (with or without type attributes)
		if !bytes.Contains(xmlBytes, []byte("<item")) {
			t.Errorf("expected item elements, got %s", string(xmlBytes))
		}
	})

	t.Run("item wrap false", func(t *testing.T) {
		data := map[string]any{
			"my_items": []any{map[string]any{"id": 1}, map[string]any{"id": 2}},
		}
		result, err := New(data).WithPretty(false).WithItemWrap(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlBytes := result.([]byte)
		xmlStr := string(xmlBytes)
		// Without item wrap, the dict contents should be in my_items (with type attribute)
		if !strings.Contains(xmlStr, "<my_items") {
			t.Errorf("expected my_items element, got %s", xmlStr)
		}
		// Should have id elements directly (not wrapped in <item>)
		if !strings.Contains(xmlStr, "<id") {
			t.Errorf("expected id elements, got %s", xmlStr)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		data := map[string]any{"empty_list": []any{}}
		result, err := New(data).WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlBytes := result.([]byte)
		if !bytes.Contains(xmlBytes, []byte("<empty_list")) {
			t.Errorf("expected empty_list element, got %s", string(xmlBytes))
		}
	})

	t.Run("type attributes", func(t *testing.T) {
		data := map[string]any{
			"my_string": "a",
			"my_int":    1,
			"my_float":  1.1,
			"my_bool":   true,
			"my_null":   nil,
		}
		result, err := New(data).WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlStr := string(result.([]byte))
		if !strings.Contains(xmlStr, `type="str"`) {
			t.Errorf("expected str type, got %s", xmlStr)
		}
		if !strings.Contains(xmlStr, `type="int"`) {
			t.Errorf("expected int type, got %s", xmlStr)
		}
		if !strings.Contains(xmlStr, `type="float"`) {
			t.Errorf("expected float type, got %s", xmlStr)
		}
		if !strings.Contains(xmlStr, `type="bool"`) {
			t.Errorf("expected bool type, got %s", xmlStr)
		}
		if !strings.Contains(xmlStr, `type="null"`) {
			t.Errorf("expected null type, got %s", xmlStr)
		}
	})

	t.Run("boolean values", func(t *testing.T) {
		data := map[string]any{
			"boolean": true,
		}
		result, err := New(data).WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlStr := string(result.([]byte))
		// Boolean should be lowercase "true" not "True"
		if !strings.Contains(xmlStr, ">true<") {
			t.Errorf("expected lowercase 'true', got %s", xmlStr)
		}
	})

	t.Run("XPath format basic", func(t *testing.T) {
		data := map[string]any{"name": "John", "age": 30, "active": true}
		result, err := New(data).WithXPathFormat(true).WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlBytes := result.([]byte)
		if !bytes.Contains(xmlBytes, []byte(`xmlns="http://www.w3.org/2005/xpath-functions"`)) {
			t.Errorf("expected XPath namespace, got %s", string(xmlBytes))
		}
		if !bytes.Contains(xmlBytes, []byte(`<string key="name">John</string>`)) {
			t.Errorf("expected string element, got %s", string(xmlBytes))
		}
		if !bytes.Contains(xmlBytes, []byte(`<number key="age">30</number>`)) {
			t.Errorf("expected number element, got %s", string(xmlBytes))
		}
		if !bytes.Contains(xmlBytes, []byte(`<boolean key="active">true</boolean>`)) {
			t.Errorf("expected boolean element, got %s", string(xmlBytes))
		}
	})

	t.Run("XPath format nested dict", func(t *testing.T) {
		data := map[string]any{"person": map[string]any{"name": "Alice", "age": 25}}
		result, err := New(data).WithXPathFormat(true).WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlBytes := result.([]byte)
		if !bytes.Contains(xmlBytes, []byte(`<map key="person">`)) {
			t.Errorf("expected nested map, got %s", string(xmlBytes))
		}
	})

	t.Run("XPath format array", func(t *testing.T) {
		data := map[string]any{"numbers": []any{1, 2, 3}}
		result, err := New(data).WithXPathFormat(true).WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlBytes := result.([]byte)
		if !bytes.Contains(xmlBytes, []byte(`<array key="numbers">`)) {
			t.Errorf("expected array element, got %s", string(xmlBytes))
		}
	})

	t.Run("XPath format null", func(t *testing.T) {
		data := map[string]any{"value": nil}
		result, err := New(data).WithXPathFormat(true).WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlBytes := result.([]byte)
		if !bytes.Contains(xmlBytes, []byte(`<null key="value"/>`)) {
			t.Errorf("expected null element, got %s", string(xmlBytes))
		}
	})

	t.Run("XPath format escaping", func(t *testing.T) {
		data := map[string]any{"text": "<script>alert('xss')</script>"}
		result, err := New(data).WithXPathFormat(true).WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlBytes := result.([]byte)
		if !bytes.Contains(xmlBytes, []byte("&lt;script&gt;")) {
			t.Errorf("expected escaped characters, got %s", string(xmlBytes))
		}
	})

	t.Run("XPath format root array", func(t *testing.T) {
		data := []any{1, 2, 3}
		result, err := New(data).WithXPathFormat(true).WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		xmlBytes := result.([]byte)
		if !bytes.Contains(xmlBytes, []byte(`<array xmlns="http://www.w3.org/2005/xpath-functions">`)) {
			t.Errorf("expected array with namespace, got %s", string(xmlBytes))
		}
	})
}

func TestToXMLString(t *testing.T) {
	t.Run("nil data returns empty string", func(t *testing.T) {
		result, err := New(nil).ToXMLString()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != "" {
			t.Errorf("expected empty string, got %s", result)
		}
	})

	t.Run("valid data returns string", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		result, err := New(data).ToXMLString()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !strings.Contains(result, "<key") {
			t.Errorf("expected XML content, got %s", result)
		}
	})
}

func TestToXMLBytes(t *testing.T) {
	t.Run("nil data returns nil", func(t *testing.T) {
		result, err := New(nil).ToXMLBytes()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("valid data returns bytes", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		result, err := New(data).ToXMLBytes()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !bytes.Contains(result, []byte("<key")) {
			t.Errorf("expected XML content, got %s", string(result))
		}
	})
}

func TestConvertToXMLFunction(t *testing.T) {
	t.Run("nil data returns nil", func(t *testing.T) {
		result, err := ConvertToXML(nil, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("with default options", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		result, err := ConvertToXML(data, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !bytes.Contains(result, []byte("<root>")) {
			t.Errorf("expected root element, got %s", string(result))
		}
	})

	t.Run("with custom options", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		opts := DefaultOptions()
		opts.CustomRoot = "custom"
		result, err := ConvertToXML(data, &opts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !bytes.Contains(result, []byte("<custom>")) {
			t.Errorf("expected custom root, got %s", string(result))
		}
	})
}

func TestIntegration(t *testing.T) {
	t.Run("read JSON file and convert to XML", func(t *testing.T) {
		data, err := ReadFromJSON("testdata/booleanjson.json")
		if err != nil {
			t.Fatalf("failed to read JSON: %v", err)
		}

		result, err := New(data).ToXMLString()
		if err != nil {
			t.Fatalf("failed to convert to XML: %v", err)
		}

		if !strings.Contains(result, "<boolean") {
			t.Errorf("expected boolean element, got %s", result)
		}
	})

	t.Run("read JSON string and convert to XML", func(t *testing.T) {
		jsonStr := `{"login":"mojombo","id":1,"avatar_url":"https://avatars0.githubusercontent.com/u/1?v=4"}`
		data, err := ReadFromString(jsonStr)
		if err != nil {
			t.Fatalf("failed to read JSON string: %v", err)
		}

		result, err := New(data).ToXMLString()
		if err != nil {
			t.Fatalf("failed to convert to XML: %v", err)
		}

		if !strings.Contains(result, "<login") {
			t.Errorf("expected login element, got %s", result)
		}
	})

	t.Run("complex nested structure", func(t *testing.T) {
		data, err := ReadFromJSON("testdata/licht.json")
		if err != nil {
			t.Fatalf("failed to read JSON: %v", err)
		}

		result, err := New(data).WithPretty(false).ToXMLBytes()
		if err != nil {
			t.Fatalf("failed to convert to XML: %v", err)
		}

		if !bytes.Contains(result, []byte("<name")) {
			t.Errorf("expected name element, got %s", string(result))
		}
		if !bytes.Contains(result, []byte("<colors")) {
			t.Errorf("expected colors element, got %s", string(result))
		}
		if !bytes.Contains(result, []byte("<tokenColors")) {
			t.Errorf("expected tokenColors element, got %s", string(result))
		}
	})
}

func TestVersion(t *testing.T) {
	t.Run("version is set", func(t *testing.T) {
		if Version == "" {
			t.Error("version should not be empty")
		}
	})

	t.Run("author is set", func(t *testing.T) {
		if Author == "" {
			t.Error("author should not be empty")
		}
	})

	t.Run("email is set", func(t *testing.T) {
		if Email == "" {
			t.Error("email should not be empty")
		}
	})
}
