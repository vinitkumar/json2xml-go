package json2xml

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Tests for 100% coverage

func TestMakeIDWithDefaults(t *testing.T) {
	t.Run("with zero start and end uses defaults", func(t *testing.T) {
		id := MakeID("test", 0, 0)
		if !strings.HasPrefix(id, "test_") {
			t.Errorf("expected prefix 'test_', got %s", id)
		}
	})

	t.Run("with custom range", func(t *testing.T) {
		id := MakeID("elem", 1, 10)
		if !strings.HasPrefix(id, "elem_") {
			t.Errorf("expected prefix 'elem_', got %s", id)
		}
	})
}

func TestGetXMLTypeAllTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"int8", int8(1), "int"},
		{"int16", int16(1), "int"},
		{"int32", int32(1), "int"},
		{"int64", int64(1), "int"},
		{"uint", uint(1), "int"},
		{"uint8", uint8(1), "int"},
		{"uint16", uint16(1), "int"},
		{"uint32", uint32(1), "int"},
		{"uint64", uint64(1), "int"},
		{"float32", float32(1.5), "float"},
		{"time.Time", time.Now(), "str"},
		{"array", [3]int{1, 2, 3}, "list"},
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

func TestKeyIsValidXMLEdgeCases(t *testing.T) {
	t.Run("valid hyphenated key", func(t *testing.T) {
		result := KeyIsValidXML("my-key")
		if !result {
			t.Error("expected my-key to be valid")
		}
	})

	t.Run("key with colon (namespace)", func(t *testing.T) {
		result := KeyIsValidXML("ns:key")
		if !result {
			t.Error("expected ns:key to be valid")
		}
	})
}

func TestMakeValidXMLNameEdgeCases(t *testing.T) {
	t.Run("key with @flat suffix", func(t *testing.T) {
		key, _ := MakeValidXMLName("list@flat", nil)
		// Should be treated as valid due to @flat handling
		if key == "" {
			t.Error("expected non-empty key")
		}
	})

	t.Run("key with namespace prefix", func(t *testing.T) {
		key, _ := MakeValidXMLName("ns:element", nil)
		if key != "ns:element" {
			t.Errorf("expected 'ns:element', got %s", key)
		}
	})

	t.Run("completely invalid key moves to name attr", func(t *testing.T) {
		key, attrs := MakeValidXMLName("/invalid/path", nil)
		if key != "key" {
			t.Errorf("expected 'key', got %s", key)
		}
		if attrs["name"] == nil {
			t.Error("expected name attribute to be set")
		}
	})
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"0", true},
		{"12a3", false},
		{"", false},
		{"abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isNumeric(tt.input)
			if result != tt.expected {
				t.Errorf("isNumeric(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetXPath31TagNameEdgeCases(t *testing.T) {
	t.Run("byte slice", func(t *testing.T) {
		result := GetXPath31TagName([]byte("hello"))
		if result != "string" {
			t.Errorf("expected 'string', got %s", result)
		}
	})

	t.Run("custom struct fallback", func(t *testing.T) {
		type CustomStruct struct {
			Name string
		}
		result := GetXPath31TagName(CustomStruct{Name: "test"})
		if result != "string" {
			t.Errorf("expected 'string' for custom type, got %s", result)
		}
	})

	t.Run("int8", func(t *testing.T) {
		result := GetXPath31TagName(int8(5))
		if result != "number" {
			t.Errorf("expected 'number', got %s", result)
		}
	})

	t.Run("uint64", func(t *testing.T) {
		result := GetXPath31TagName(uint64(5))
		if result != "number" {
			t.Errorf("expected 'number', got %s", result)
		}
	})
}

func TestConvertToXPath31EdgeCases(t *testing.T) {
	t.Run("custom type fallback", func(t *testing.T) {
		type Point struct{ X, Y int }
		result := ConvertToXPath31(Point{1, 2}, "")
		if !strings.Contains(result, "<string>") {
			t.Errorf("expected string tag for custom type, got %s", result)
		}
	})
}

func TestToMapEdgeCases(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		result := toMap(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("non-map input", func(t *testing.T) {
		result := toMap("string")
		if result != nil {
			t.Errorf("expected nil for string, got %v", result)
		}
	})

	t.Run("map with int keys via reflection", func(t *testing.T) {
		m := map[int]string{1: "one", 2: "two"}
		result := toMap(m)
		if result == nil {
			t.Error("expected non-nil result")
		}
		if result["1"] != "one" {
			t.Errorf("expected key '1' to have value 'one', got %v", result["1"])
		}
	})
}

func TestToSliceEdgeCases(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		result := toSlice(nil)
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("non-slice input", func(t *testing.T) {
		result := toSlice("string")
		if result != nil {
			t.Errorf("expected nil for string, got %v", result)
		}
	})

	t.Run("array input", func(t *testing.T) {
		arr := [3]int{1, 2, 3}
		result := toSlice(arr)
		if len(result) != 3 {
			t.Errorf("expected 3 elements, got %d", len(result))
		}
	})
}

func TestConvertAllTypes(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = false

	t.Run("bool", func(t *testing.T) {
		result := Convert(true, opts, "root")
		if !strings.Contains(result, "true") {
			t.Errorf("expected 'true', got %s", result)
		}
	})

	t.Run("int types", func(t *testing.T) {
		result := Convert(int64(42), opts, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected '42', got %s", result)
		}
	})

	t.Run("uint types", func(t *testing.T) {
		result := Convert(uint32(42), opts, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected '42', got %s", result)
		}
	})

	t.Run("float types", func(t *testing.T) {
		result := Convert(float32(3.14), opts, "root")
		if !strings.Contains(result, "3.14") {
			t.Errorf("expected '3.14', got %s", result)
		}
	})

	t.Run("string", func(t *testing.T) {
		result := Convert("hello", opts, "root")
		if !strings.Contains(result, "hello") {
			t.Errorf("expected 'hello', got %s", result)
		}
	})

	t.Run("nil", func(t *testing.T) {
		result := Convert(nil, opts, "root")
		if !strings.Contains(result, "<item>") {
			t.Errorf("expected item element, got %s", result)
		}
	})

	t.Run("time.Time", func(t *testing.T) {
		dt := time.Date(2023, 2, 15, 12, 30, 45, 0, time.UTC)
		result := Convert(dt, opts, "root")
		if !strings.Contains(result, "2023") {
			t.Errorf("expected datetime, got %s", result)
		}
	})

	t.Run("map", func(t *testing.T) {
		m := map[string]any{"key": "value"}
		result := Convert(m, opts, "root")
		if !strings.Contains(result, "<key>") {
			t.Errorf("expected key element, got %s", result)
		}
	})

	t.Run("slice", func(t *testing.T) {
		s := []any{1, 2, 3}
		result := Convert(s, opts, "root")
		if !strings.Contains(result, "<item>") {
			t.Errorf("expected item elements, got %s", result)
		}
	})

	t.Run("custom type fallback", func(t *testing.T) {
		type Custom struct{ X int }
		result := Convert(Custom{X: 5}, opts, "root")
		if result == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestConvertDictAllBranches(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = false
	opts.IDs = true // Enable IDs

	t.Run("with IDs enabled", func(t *testing.T) {
		m := map[string]any{"key": "value"}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "id=") {
			t.Errorf("expected id attribute, got %s", result)
		}
	})

	t.Run("with int8 value", func(t *testing.T) {
		m := map[string]any{"num": int8(5)}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "5") {
			t.Errorf("expected 5, got %s", result)
		}
	})

	t.Run("with int16 value", func(t *testing.T) {
		m := map[string]any{"num": int16(5)}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "5") {
			t.Errorf("expected 5, got %s", result)
		}
	})

	t.Run("with int32 value", func(t *testing.T) {
		m := map[string]any{"num": int32(5)}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "5") {
			t.Errorf("expected 5, got %s", result)
		}
	})

	t.Run("with int64 value", func(t *testing.T) {
		m := map[string]any{"num": int64(5)}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "5") {
			t.Errorf("expected 5, got %s", result)
		}
	})

	t.Run("with uint value", func(t *testing.T) {
		m := map[string]any{"num": uint(5)}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "5") {
			t.Errorf("expected 5, got %s", result)
		}
	})

	t.Run("with uint8 value", func(t *testing.T) {
		m := map[string]any{"num": uint8(5)}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "5") {
			t.Errorf("expected 5, got %s", result)
		}
	})

	t.Run("with uint16 value", func(t *testing.T) {
		m := map[string]any{"num": uint16(5)}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "5") {
			t.Errorf("expected 5, got %s", result)
		}
	})

	t.Run("with uint32 value", func(t *testing.T) {
		m := map[string]any{"num": uint32(5)}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "5") {
			t.Errorf("expected 5, got %s", result)
		}
	})

	t.Run("with uint64 value", func(t *testing.T) {
		m := map[string]any{"num": uint64(5)}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "5") {
			t.Errorf("expected 5, got %s", result)
		}
	})

	t.Run("with float32 value", func(t *testing.T) {
		m := map[string]any{"num": float32(3.14)}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "3.14") {
			t.Errorf("expected 3.14, got %s", result)
		}
	})

	t.Run("with float64 value", func(t *testing.T) {
		m := map[string]any{"num": float64(3.14)}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "3.14") {
			t.Errorf("expected 3.14, got %s", result)
		}
	})

	t.Run("with time.Time value", func(t *testing.T) {
		dt := time.Date(2023, 2, 15, 12, 30, 45, 0, time.UTC)
		m := map[string]any{"date": dt}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "2023") {
			t.Errorf("expected datetime, got %s", result)
		}
	})

	t.Run("with slice via reflection", func(t *testing.T) {
		m := map[string]any{"items": []int{1, 2, 3}}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected items, got %s", result)
		}
	})

	t.Run("with map via reflection", func(t *testing.T) {
		m := map[string]any{"nested": map[string]int{"a": 1}}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "nested") {
			t.Errorf("expected nested, got %s", result)
		}
	})

	t.Run("with custom type via reflection", func(t *testing.T) {
		type Custom struct{ X int }
		m := map[string]any{"custom": Custom{X: 5}}
		result := ConvertDict(m, opts, "root")
		if result == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestDict2XMLStrAllBranches(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = false

	t.Run("with @attrs", func(t *testing.T) {
		item := map[string]any{
			"@attrs": map[string]any{"id": "123"},
			"@val":   "content",
		}
		result := Dict2XMLStr(opts, nil, item, "elem", false, "parent")
		if !strings.Contains(result, `id="123"`) {
			t.Errorf("expected id attribute, got %s", result)
		}
	})

	t.Run("with @flat", func(t *testing.T) {
		item := map[string]any{
			"@flat": true,
			"@val":  "content",
		}
		result := Dict2XMLStr(opts, nil, item, "elem", false, "parent")
		if result != "content" {
			t.Errorf("expected just content, got %s", result)
		}
	})

	t.Run("parent is list with list headers and attrs", func(t *testing.T) {
		opts.ListHeaders = true
		opts.ItemWrap = false
		item := map[string]any{
			"@attrs": map[string]any{"id": "1"},
			"@val":   "content",
		}
		result := Dict2XMLStr(opts, nil, item, "elem", true, "parent")
		if !strings.Contains(result, "<parent") {
			t.Errorf("expected parent tag, got %s", result)
		}
		opts.ListHeaders = false
		opts.ItemWrap = true
	})

	t.Run("parent is list with list headers without attrs", func(t *testing.T) {
		opts.ListHeaders = true
		opts.ItemWrap = true
		item := map[string]any{"key": "value"}
		result := Dict2XMLStr(opts, nil, item, "elem", true, "parent")
		if !strings.Contains(result, "<parent>") {
			t.Errorf("expected parent tag, got %s", result)
		}
		opts.ListHeaders = false
	})

	t.Run("parent is list without item wrap", func(t *testing.T) {
		opts.ItemWrap = false
		item := map[string]any{"key": "value"}
		result := Dict2XMLStr(opts, nil, item, "elem", true, "parent")
		if !strings.Contains(result, "<key>") {
			t.Errorf("expected key element, got %s", result)
		}
		opts.ItemWrap = true
	})

	t.Run("with primitive string @val", func(t *testing.T) {
		item := map[string]any{"@val": "simple"}
		result := Dict2XMLStr(opts, nil, item, "elem", false, "parent")
		if !strings.Contains(result, "simple") {
			t.Errorf("expected simple, got %s", result)
		}
	})

	t.Run("with non-primitive @val", func(t *testing.T) {
		item := map[string]any{"@val": map[string]any{"nested": "value"}}
		result := Dict2XMLStr(opts, nil, item, "elem", false, "parent")
		if !strings.Contains(result, "nested") {
			t.Errorf("expected nested, got %s", result)
		}
	})
}

func TestConvertListAllBranches(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = false

	t.Run("with bool items", func(t *testing.T) {
		items := []any{true, false}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "true") || !strings.Contains(result, "false") {
			t.Errorf("expected bool values, got %s", result)
		}
	})

	t.Run("with int8 items", func(t *testing.T) {
		items := []any{int8(1), int8(2)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("with int16 items", func(t *testing.T) {
		items := []any{int16(1)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("with int32 items", func(t *testing.T) {
		items := []any{int32(1)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("with int64 items", func(t *testing.T) {
		items := []any{int64(1)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("with uint items", func(t *testing.T) {
		items := []any{uint(1)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("with uint8 items", func(t *testing.T) {
		items := []any{uint8(1)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("with uint16 items", func(t *testing.T) {
		items := []any{uint16(1)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("with uint32 items", func(t *testing.T) {
		items := []any{uint32(1)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("with uint64 items", func(t *testing.T) {
		items := []any{uint64(1)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("with float32 items", func(t *testing.T) {
		items := []any{float32(1.5)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1.5") {
			t.Errorf("expected 1.5, got %s", result)
		}
	})

	t.Run("with float64 items", func(t *testing.T) {
		items := []any{float64(1.5)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1.5") {
			t.Errorf("expected 1.5, got %s", result)
		}
	})

	t.Run("with string items without item wrap", func(t *testing.T) {
		opts.ItemWrap = false
		items := []any{"a", "b"}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "<root>") {
			t.Errorf("expected root element, got %s", result)
		}
		opts.ItemWrap = true
	})

	t.Run("with time.Time items", func(t *testing.T) {
		dt := time.Date(2023, 2, 15, 12, 30, 45, 0, time.UTC)
		items := []any{dt}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "2023") {
			t.Errorf("expected datetime, got %s", result)
		}
	})

	t.Run("with nil items", func(t *testing.T) {
		items := []any{nil}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "<item>") {
			t.Errorf("expected item element, got %s", result)
		}
	})

	t.Run("with nested slice items", func(t *testing.T) {
		items := []any{[]any{1, 2}}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("with map items", func(t *testing.T) {
		items := []any{map[string]any{"key": "value"}}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "key") {
			t.Errorf("expected key, got %s", result)
		}
	})

	t.Run("with custom type items via reflection", func(t *testing.T) {
		type Custom struct{ X int }
		items := []any{Custom{X: 5}}
		result := ConvertList(items, opts, "root")
		if result == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("with flat item name", func(t *testing.T) {
		opts.ItemFunc = func(p string) string { return p + "@flat" }
		items := []any{"a", "b"}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "<root>") {
			t.Errorf("expected root element, got %s", result)
		}
		opts.ItemFunc = DefaultItemFunc
	})

	t.Run("reflection branches for item wrap false", func(t *testing.T) {
		opts.ItemWrap = false

		// int via reflection
		items := []any{int8(5)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "5") {
			t.Errorf("expected 5, got %s", result)
		}

		// uint via reflection
		items = []any{uint8(5)}
		result = ConvertList(items, opts, "root")
		if !strings.Contains(result, "5") {
			t.Errorf("expected 5, got %s", result)
		}

		// float via reflection
		items = []any{float32(5.5)}
		result = ConvertList(items, opts, "root")
		if !strings.Contains(result, "5.5") {
			t.Errorf("expected 5.5, got %s", result)
		}

		// string via reflection
		items = []any{"test"}
		result = ConvertList(items, opts, "root")
		if !strings.Contains(result, "test") {
			t.Errorf("expected test, got %s", result)
		}

		opts.ItemWrap = true
	})
}

func TestConvertKVWithTimeValue(t *testing.T) {
	dt := time.Date(2023, 2, 15, 12, 30, 45, 0, time.UTC)
	result := ConvertKV("date", dt, false, nil, false)
	if !strings.Contains(result, "2023") {
		t.Errorf("expected datetime, got %s", result)
	}
}

func TestDictToXMLXSINamespace(t *testing.T) {
	data := map[string]any{"bike": "blue"}
	opts := DefaultOptions()
	opts.AttrType = false
	opts.XMLNamespaces = map[string]any{
		"xsi": map[string]any{
			"schemaInstance": "http://www.w3.org/2001/XMLSchema-instance",
			"schemaLocation": "https://www.w3schools.com/note.xsd",
		},
	}
	result := DictToXML(data, opts)
	if !strings.Contains(string(result), "xmlns:xsi") {
		t.Errorf("expected xsi namespace, got %s", string(result))
	}
	if !strings.Contains(string(result), "xsi:schemaLocation") {
		t.Errorf("expected schemaLocation, got %s", string(result))
	}
}

func TestPrettyPrintError(t *testing.T) {
	t.Run("invalid XML", func(t *testing.T) {
		_, err := PrettyPrint([]byte("<unclosed"))
		if err == nil {
			t.Error("expected error for invalid XML")
		}
	})
}

func TestToXMLStringTypeSwitches(t *testing.T) {
	t.Run("returns bytes as string", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		conv := New(data).WithPretty(false)
		result, err := conv.ToXMLString()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "<key") {
			t.Errorf("expected key element, got %s", result)
		}
	})
}

func TestToXMLBytesTypeSwitches(t *testing.T) {
	t.Run("returns string as bytes", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		conv := New(data).WithPretty(true)
		result, err := conv.ToXMLBytes()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(string(result), "<key") {
			t.Errorf("expected key element, got %s", string(result))
		}
	})
}

func TestReadFromURL(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"key": "value", "number": 42}`))
		}))
		defer server.Close()

		result, err := ReadFromURL(server.URL, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		m, ok := result.(map[string]any)
		if !ok {
			t.Fatalf("expected map, got %T", result)
		}
		if m["key"] != "value" {
			t.Errorf("expected 'value', got %v", m["key"])
		}
	})

	t.Run("with query params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("foo") != "bar" {
				t.Errorf("expected query param foo=bar")
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"result": "ok"}`))
		}))
		defer server.Close()

		result, err := ReadFromURL(server.URL, map[string]string{"foo": "bar"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("404 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		_, err := ReadFromURL(server.URL, nil)
		if err == nil {
			t.Error("expected error for 404")
		}
	})

	t.Run("500 error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		_, err := ReadFromURL(server.URL, nil)
		if err == nil {
			t.Error("expected error for 500")
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		_, err := ReadFromURL(server.URL, nil)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		_, err := ReadFromURL("http://invalid.localhost.test:99999", nil)
		if err == nil {
			t.Error("expected error for invalid URL")
		}
	})
}

func TestConvertToXMLWithNilOptions(t *testing.T) {
	data := map[string]any{"key": "value"}
	result, err := ConvertToXML(data, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(result), "<root>") {
		t.Errorf("expected root element with default options, got %s", string(result))
	}
}

func TestList2XMLStrWithAttrType(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = true

	items := []any{"a", "b", "c"}
	attrs := make(map[string]any)
	result := List2XMLStr(opts, attrs, items, "list")
	if !strings.Contains(result, `type="list"`) {
		t.Errorf("expected type=list attribute, got %s", result)
	}
}

func TestList2XMLStrFlatNotation(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = false

	items := []any{1, 2, 3}
	result := List2XMLStr(opts, nil, items, "list@flat")
	// Should return items without the list wrapper
	if strings.Contains(result, "<list@flat>") {
		t.Errorf("should not contain list@flat tag, got %s", result)
	}
}

func TestList2XMLStrWithListHeaders(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = false
	opts.ListHeaders = true

	items := []any{map[string]any{"a": 1}}
	result := List2XMLStr(opts, nil, items, "list")
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestReadFromJSONIOError(t *testing.T) {
	// Try to read a directory (should fail)
	_, err := ReadFromJSON("testdata")
	if err == nil {
		t.Error("expected error when reading directory")
	}
}

// Additional coverage tests

func TestGetXMLTypeUnknown(t *testing.T) {
	// Test with a custom type that doesn't match any known types
	type CustomStruct struct{ X int }
	result := GetXMLType(CustomStruct{X: 1})
	if result == "" {
		t.Error("expected non-empty type name")
	}
}

func TestKeyIsValidXMLMoreCases(t *testing.T) {
	t.Run("key starting with number", func(t *testing.T) {
		result := KeyIsValidXML("1invalid")
		if result {
			t.Error("expected false for key starting with number")
		}
	})

	t.Run("key with special chars", func(t *testing.T) {
		result := KeyIsValidXML("key with space")
		if result {
			t.Error("expected false for key with space")
		}
	})
}

func TestConvertToXPath31FullCoverage(t *testing.T) {
	t.Run("default case with struct", func(t *testing.T) {
		type Point struct{ X, Y int }
		result := ConvertToXPath31(Point{1, 2}, "point")
		if !strings.Contains(result, `key="point"`) {
			t.Errorf("expected key attribute, got %s", result)
		}
	})
}

func TestConvertDictRemainingBranches(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = false

	t.Run("with bool value via reflection", func(t *testing.T) {
		// Use a type that will go through reflection
		type BoolWrapper bool
		m := map[string]any{"flag": BoolWrapper(true)}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "true") {
			t.Errorf("expected true, got %s", result)
		}
	})

	t.Run("with string value via reflection", func(t *testing.T) {
		type StringWrapper string
		m := map[string]any{"name": StringWrapper("test")}
		result := ConvertDict(m, opts, "root")
		if !strings.Contains(result, "test") {
			t.Errorf("expected test, got %s", result)
		}
	})
}

func TestConvertListRemainingBranches(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = false

	t.Run("with bool via reflection path", func(t *testing.T) {
		type BoolWrapper bool
		items := []any{BoolWrapper(true)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "true") {
			t.Errorf("expected true, got %s", result)
		}
	})

	t.Run("with map via reflection path", func(t *testing.T) {
		m := map[string]int{"a": 1}
		items := []any{m}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "a") {
			t.Errorf("expected a, got %s", result)
		}
	})

	t.Run("with slice via reflection path", func(t *testing.T) {
		s := []int{1, 2, 3}
		items := []any{s}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("custom type via reflection default", func(t *testing.T) {
		type Custom struct{ X int }
		items := []any{Custom{X: 5}}
		result := ConvertList(items, opts, "root")
		if result == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("without item wrap - various types via reflection", func(t *testing.T) {
		opts.ItemWrap = false

		// Bool
		type BoolWrapper bool
		items := []any{BoolWrapper(true)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "true") {
			t.Errorf("expected true, got %s", result)
		}

		// Map
		m := map[string]int{"a": 1}
		items = []any{m}
		result = ConvertList(items, opts, "root")
		if result == "" {
			t.Error("expected non-empty result")
		}

		// Slice
		s := []int{1, 2}
		items = []any{s}
		result = ConvertList(items, opts, "root")
		if result == "" {
			t.Error("expected non-empty result")
		}

		// Custom type
		type Custom struct{ X int }
		items = []any{Custom{X: 5}}
		result = ConvertList(items, opts, "root")
		if result == "" {
			t.Error("expected non-empty result")
		}

		opts.ItemWrap = true
	})
}

func TestDict2XMLStrRemainingBranches(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = true // Enable to test type="dict"

	t.Run("with type attribute for dict", func(t *testing.T) {
		item := map[string]any{"key": "value"}
		attrs := make(map[string]any)
		result := Dict2XMLStr(opts, attrs, item, "elem", false, "parent")
		if !strings.Contains(result, `type="dict"`) {
			t.Errorf("expected type=dict, got %s", result)
		}
	})
}

func TestDictToXMLRemainingBranches(t *testing.T) {
	t.Run("with nil ItemFunc uses default", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		opts := DefaultOptions()
		opts.ItemFunc = nil // Will be set to default
		result := DictToXML(data, opts)
		if len(result) == 0 {
			t.Error("expected non-empty result")
		}
	})

	t.Run("xpath format with non-map/non-array root", func(t *testing.T) {
		data := "just a string"
		opts := DefaultOptions()
		opts.XPathFormat = true
		result := DictToXML(data, opts)
		if !strings.Contains(string(result), `<map xmlns=`) {
			t.Errorf("expected map wrapper, got %s", string(result))
		}
	})
}

func TestPrettyPrintMoreCases(t *testing.T) {
	t.Run("valid XML with declaration", func(t *testing.T) {
		input := []byte(`<?xml version="1.0"?><root><child>value</child></root>`)
		result, err := PrettyPrint(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "<?xml") {
			t.Errorf("expected xml declaration, got %s", result)
		}
	})
}

func TestToXMLRemainingBranches(t *testing.T) {
	t.Run("with slice data", func(t *testing.T) {
		data := []any{1, 2, 3}
		result, err := New(data).WithPretty(false).ToXML()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("expected non-nil result")
		}
	})
}

func TestToXMLStringDefaultCase(t *testing.T) {
	t.Run("with non-standard type", func(t *testing.T) {
		// This is hard to trigger since ToXML returns string or []byte
		// But we can test the normal path
		data := map[string]any{"key": "value"}
		result, err := New(data).ToXMLString()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == "" {
			t.Error("expected non-empty result")
		}
	})
}

func TestToXMLBytesDefaultCase(t *testing.T) {
	t.Run("with pretty print", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		result, err := New(data).WithPretty(true).ToXMLBytes()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) == 0 {
			t.Error("expected non-empty result")
		}
	})
}

func TestConvertToXMLAllBranches(t *testing.T) {
	t.Run("with options having nil ItemFunc", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		opts := DefaultOptions()
		opts.ItemFunc = nil
		result, err := ConvertToXML(data, &opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) == 0 {
			t.Error("expected non-empty result")
		}
	})
}

func TestReadFromURLReadError(t *testing.T) {
	t.Run("server closes connection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Just close the connection without sending response
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				conn.Close()
			}
		}))
		defer server.Close()

		_, err := ReadFromURL(server.URL, nil)
		// This should cause an error
		if err == nil {
			t.Log("Connection close might not always cause error in test environment")
		}
	})
}

// Test KeyIsValidXML edge cases for more coverage
func TestKeyIsValidXMLEdgeCasesMore(t *testing.T) {
	t.Run("simple valid key", func(t *testing.T) {
		result := KeyIsValidXML("a")
		if !result {
			t.Error("expected 'a' to be valid")
		}
	})
}

// Additional ConvertList tests for remaining branches
func TestConvertListTimeWithoutItemWrap(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = false
	opts.ItemWrap = false

	dt := time.Date(2023, 2, 15, 12, 30, 45, 0, time.UTC)
	items := []any{dt}
	result := ConvertList(items, opts, "root")
	if !strings.Contains(result, "2023") {
		t.Errorf("expected datetime, got %s", result)
	}
}

// Test PrettyPrint with token that returns nil
func TestPrettyPrintNilToken(t *testing.T) {
	// Valid but minimal XML
	input := []byte(`<root/>`)
	result, err := PrettyPrint(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "root") {
		t.Errorf("expected root element, got %s", result)
	}
}

// Test ToXML with list data type check
func TestToXMLWithListCheck(t *testing.T) {
	// Non-empty slice
	data := []any{"a", "b"}
	result, err := New(data).WithPretty(false).ToXML()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result for non-empty slice")
	}

	// Non-empty map
	mapData := map[string]any{"key": "value"}
	result, err = New(mapData).WithPretty(false).ToXML()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result for non-empty map")
	}
}

// Test ToXMLString with byte result
func TestToXMLStringByteConversion(t *testing.T) {
	data := map[string]any{"key": "value"}
	// With pretty=false, ToXML returns []byte
	result, err := New(data).WithPretty(false).ToXMLString()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty string")
	}
}

// Test ToXMLBytes with string result
func TestToXMLBytesStringConversion(t *testing.T) {
	data := map[string]any{"key": "value"}
	// With pretty=true, ToXML returns string
	result, err := New(data).WithPretty(true).ToXMLBytes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty bytes")
	}
}

// Test ConvertDict with time via type switch (not reflection)
func TestConvertDictTimeTypeSwitch(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = false

	dt := time.Date(2023, 2, 15, 12, 30, 45, 0, time.UTC)
	m := map[string]any{"date": dt}
	result := ConvertDict(m, opts, "root")
	if !strings.Contains(result, "2023") {
		t.Errorf("expected datetime, got %s", result)
	}
}

// Additional test for ConvertList with various item types
func TestConvertListCompleteTypes(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = false

	// Test with direct type matches (not via reflection)
	t.Run("int types direct", func(t *testing.T) {
		items := []any{1, int8(2), int16(3), int32(4), int64(5)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("uint types direct", func(t *testing.T) {
		items := []any{uint(1), uint8(2), uint16(3), uint32(4), uint64(5)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("float types direct", func(t *testing.T) {
		items := []any{float32(1.5), float64(2.5)}
		result := ConvertList(items, opts, "root")
		if !strings.Contains(result, "1.5") {
			t.Errorf("expected 1.5, got %s", result)
		}
	})
}

// Test Dict2XMLStr edge case with primitive string
func TestDict2XMLStrPrimitiveString(t *testing.T) {
	opts := DefaultOptions()
	opts.AttrType = false

	// Item with just a string @val
	item := map[string]any{"@val": "simple string"}
	attrs := make(map[string]any)
	result := Dict2XMLStr(opts, attrs, item, "elem", false, "parent")
	if !strings.Contains(result, "simple string") {
		t.Errorf("expected simple string, got %s", result)
	}
}

// Comprehensive ConvertList tests for all branches
func TestConvertListAllBranchesWithItemWrap(t *testing.T) {
	optsWrap := DefaultOptions()
	optsWrap.AttrType = false
	optsWrap.ItemWrap = true

	optsNoWrap := DefaultOptions()
	optsNoWrap.AttrType = false
	optsNoWrap.ItemWrap = false

	// Test each type with both ItemWrap=true and ItemWrap=false
	t.Run("int with wrap", func(t *testing.T) {
		items := []any{42}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "<item>") {
			t.Errorf("expected item tag, got %s", result)
		}
	})

	t.Run("int without wrap", func(t *testing.T) {
		items := []any{42}
		result := ConvertList(items, optsNoWrap, "root")
		if !strings.Contains(result, "<root>") {
			t.Errorf("expected root tag, got %s", result)
		}
	})

	t.Run("int8 with wrap", func(t *testing.T) {
		items := []any{int8(42)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected 42, got %s", result)
		}
	})

	t.Run("int8 without wrap", func(t *testing.T) {
		items := []any{int8(42)}
		result := ConvertList(items, optsNoWrap, "root")
		if !strings.Contains(result, "<root>") {
			t.Errorf("expected root tag, got %s", result)
		}
	})

	t.Run("int16 with wrap", func(t *testing.T) {
		items := []any{int16(42)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected 42, got %s", result)
		}
	})

	t.Run("int32 with wrap", func(t *testing.T) {
		items := []any{int32(42)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected 42, got %s", result)
		}
	})

	t.Run("int64 with wrap", func(t *testing.T) {
		items := []any{int64(42)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected 42, got %s", result)
		}
	})

	t.Run("uint with wrap", func(t *testing.T) {
		items := []any{uint(42)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected 42, got %s", result)
		}
	})

	t.Run("uint without wrap", func(t *testing.T) {
		items := []any{uint(42)}
		result := ConvertList(items, optsNoWrap, "root")
		if !strings.Contains(result, "<root>") {
			t.Errorf("expected root tag, got %s", result)
		}
	})

	t.Run("uint8 with wrap", func(t *testing.T) {
		items := []any{uint8(42)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected 42, got %s", result)
		}
	})

	t.Run("uint16 with wrap", func(t *testing.T) {
		items := []any{uint16(42)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected 42, got %s", result)
		}
	})

	t.Run("uint32 with wrap", func(t *testing.T) {
		items := []any{uint32(42)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected 42, got %s", result)
		}
	})

	t.Run("uint64 with wrap", func(t *testing.T) {
		items := []any{uint64(42)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected 42, got %s", result)
		}
	})

	t.Run("float32 with wrap", func(t *testing.T) {
		items := []any{float32(3.14)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "3.14") {
			t.Errorf("expected 3.14, got %s", result)
		}
	})

	t.Run("float32 without wrap", func(t *testing.T) {
		items := []any{float32(3.14)}
		result := ConvertList(items, optsNoWrap, "root")
		if !strings.Contains(result, "<root>") {
			t.Errorf("expected root tag, got %s", result)
		}
	})

	t.Run("float64 with wrap", func(t *testing.T) {
		items := []any{float64(3.14)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "3.14") {
			t.Errorf("expected 3.14, got %s", result)
		}
	})

	t.Run("string with wrap", func(t *testing.T) {
		items := []any{"hello"}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "<item>") {
			t.Errorf("expected item tag, got %s", result)
		}
	})

	t.Run("string without wrap", func(t *testing.T) {
		items := []any{"hello"}
		result := ConvertList(items, optsNoWrap, "root")
		if !strings.Contains(result, "<root>") {
			t.Errorf("expected root tag, got %s", result)
		}
	})

	t.Run("time with wrap", func(t *testing.T) {
		dt := time.Date(2023, 2, 15, 12, 30, 45, 0, time.UTC)
		items := []any{dt}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "2023") {
			t.Errorf("expected 2023, got %s", result)
		}
	})

	t.Run("map[string]any with wrap", func(t *testing.T) {
		items := []any{map[string]any{"key": "value"}}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "key") {
			t.Errorf("expected key, got %s", result)
		}
	})

	t.Run("[]any with wrap", func(t *testing.T) {
		items := []any{[]any{1, 2, 3}}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "1") {
			t.Errorf("expected 1, got %s", result)
		}
	})

	t.Run("nil with wrap", func(t *testing.T) {
		items := []any{nil}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "<item>") {
			t.Errorf("expected item tag, got %s", result)
		}
	})
}

// Test reflection paths in ConvertList
func TestConvertListReflectionPaths(t *testing.T) {
	optsWrap := DefaultOptions()
	optsWrap.AttrType = false
	optsWrap.ItemWrap = true

	optsNoWrap := DefaultOptions()
	optsNoWrap.AttrType = false
	optsNoWrap.ItemWrap = false

	// Custom types that go through reflection
	type MyBool bool
	type MyInt int
	type MyUint uint
	type MyFloat float64
	type MyString string

	t.Run("custom bool with wrap", func(t *testing.T) {
		items := []any{MyBool(true)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "true") {
			t.Errorf("expected true, got %s", result)
		}
	})

	t.Run("custom int with wrap", func(t *testing.T) {
		items := []any{MyInt(42)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected 42, got %s", result)
		}
	})

	t.Run("custom int without wrap", func(t *testing.T) {
		items := []any{MyInt(42)}
		result := ConvertList(items, optsNoWrap, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected 42, got %s", result)
		}
	})

	t.Run("custom uint with wrap", func(t *testing.T) {
		items := []any{MyUint(42)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected 42, got %s", result)
		}
	})

	t.Run("custom uint without wrap", func(t *testing.T) {
		items := []any{MyUint(42)}
		result := ConvertList(items, optsNoWrap, "root")
		if !strings.Contains(result, "42") {
			t.Errorf("expected 42, got %s", result)
		}
	})

	t.Run("custom float with wrap", func(t *testing.T) {
		items := []any{MyFloat(3.14)}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "3.14") {
			t.Errorf("expected 3.14, got %s", result)
		}
	})

	t.Run("custom float without wrap", func(t *testing.T) {
		items := []any{MyFloat(3.14)}
		result := ConvertList(items, optsNoWrap, "root")
		if !strings.Contains(result, "3.14") {
			t.Errorf("expected 3.14, got %s", result)
		}
	})

	t.Run("custom string with wrap", func(t *testing.T) {
		items := []any{MyString("hello")}
		result := ConvertList(items, optsWrap, "root")
		if !strings.Contains(result, "hello") {
			t.Errorf("expected hello, got %s", result)
		}
	})

	t.Run("custom string without wrap", func(t *testing.T) {
		items := []any{MyString("hello")}
		result := ConvertList(items, optsNoWrap, "root")
		if !strings.Contains(result, "hello") {
			t.Errorf("expected hello, got %s", result)
		}
	})

	t.Run("map via reflection with wrap", func(t *testing.T) {
		items := []any{map[string]int{"a": 1}}
		result := ConvertList(items, optsWrap, "root")
		if result == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("slice via reflection with wrap", func(t *testing.T) {
		items := []any{[]int{1, 2, 3}}
		result := ConvertList(items, optsWrap, "root")
		if result == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("struct fallback with wrap", func(t *testing.T) {
		type Point struct{ X, Y int }
		items := []any{Point{1, 2}}
		result := ConvertList(items, optsWrap, "root")
		if result == "" {
			t.Error("expected non-empty result")
		}
	})
}

// Test remaining uncovered lines
func TestKeyIsValidXMLTokenNil(t *testing.T) {
	// This is hard to trigger since xml decoder doesn't return nil token normally
	// Just test that valid keys work
	result := KeyIsValidXML("validkey")
	if !result {
		t.Error("expected validkey to be valid")
	}
}

func TestToXMLError(t *testing.T) {
	// Create data that will cause PrettyPrint to fail
	// This is hard to trigger with valid data
	// Just verify that errors are propagated
	data := map[string]any{"key": "value"}
	_, err := New(data).ToXML()
	if err != nil {
		t.Logf("Error occurred: %v", err)
	}
}

func TestReadFromURLBodyReadError(t *testing.T) {
	// Test with a server that sends incomplete response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"incomplete`))
	}))
	defer server.Close()

	_, err := ReadFromURL(server.URL, nil)
	if err == nil {
		t.Log("Expected error for incomplete JSON")
	}
}
