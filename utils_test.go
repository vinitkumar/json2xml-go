package json2xml

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFromJSON(t *testing.T) {
	t.Run("valid JSON file", func(t *testing.T) {
		data, err := ReadFromJSON("testdata/booleanjson.json")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if data == nil {
			t.Fatal("expected data, got nil")
		}

		m, ok := data.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", data)
		}

		if _, exists := m["boolean"]; !exists {
			t.Error("expected 'boolean' key in data")
		}
	})

	t.Run("valid JSON file with nested structure", func(t *testing.T) {
		data, err := ReadFromJSON("testdata/licht.json")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if data == nil {
			t.Fatal("expected data, got nil")
		}

		m, ok := data.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", data)
		}

		if name, exists := m["name"]; !exists || name != "Licht" {
			t.Error("expected 'name' key with value 'Licht'")
		}
	})

	t.Run("invalid JSON file (non-existent)", func(t *testing.T) {
		_, err := ReadFromJSON("nonexistent.json")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrJSONRead) {
			t.Errorf("expected ErrJSONRead, got %v", err)
		}
	})

	t.Run("invalid JSON content", func(t *testing.T) {
		_, err := ReadFromJSON("testdata/wrongjson.json")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrJSONRead) {
			t.Errorf("expected ErrJSONRead, got %v", err)
		}
	})
}

func TestReadFromString(t *testing.T) {
	t.Run("valid JSON string", func(t *testing.T) {
		jsonStr := `{"login":"mojombo","id":1,"avatar_url":"https://avatars0.githubusercontent.com/u/1?v=4"}`
		data, err := ReadFromString(jsonStr)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		m, ok := data.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", data)
		}

		if login, exists := m["login"]; !exists || login != "mojombo" {
			t.Error("expected 'login' key with value 'mojombo'")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := ReadFromString("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrStringRead) {
			t.Errorf("expected ErrStringRead, got %v", err)
		}
	})

	t.Run("invalid JSON string", func(t *testing.T) {
		_, err := ReadFromString(`{"login":"mojombo","id":1`)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrStringRead) {
			t.Errorf("expected ErrStringRead, got %v", err)
		}
	})

	t.Run("valid JSON array", func(t *testing.T) {
		jsonStr := `[1, 2, 3]`
		data, err := ReadFromString(jsonStr)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		arr, ok := data.([]any)
		if !ok {
			t.Fatalf("expected []any, got %T", data)
		}

		if len(arr) != 3 {
			t.Errorf("expected 3 elements, got %d", len(arr))
		}
	})

	t.Run("complex JSON object", func(t *testing.T) {
		jsonStr := `{"users": [{"name": "John", "age": 30}, {"name": "Jane", "age": 25}], "total": 2}`
		data, err := ReadFromString(jsonStr)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		m, ok := data.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", data)
		}

		users, exists := m["users"]
		if !exists {
			t.Fatal("expected 'users' key")
		}

		usersArr, ok := users.([]any)
		if !ok {
			t.Fatalf("expected []any for users, got %T", users)
		}

		if len(usersArr) != 2 {
			t.Errorf("expected 2 users, got %d", len(usersArr))
		}
	})
}

func TestReadFromJSONWithTempFile(t *testing.T) {
	t.Run("create and read temp JSON file", func(t *testing.T) {
		// Create temp file
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.json")
		content := []byte(`{"key": "value", "number": 42}`)
		if err := os.WriteFile(tmpFile, content, 0644); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}

		// Read the file
		data, err := ReadFromJSON(tmpFile)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		m, ok := data.(map[string]any)
		if !ok {
			t.Fatalf("expected map[string]any, got %T", data)
		}

		if key, exists := m["key"]; !exists || key != "value" {
			t.Error("expected 'key' with value 'value'")
		}

		if num, exists := m["number"]; !exists || num != float64(42) {
			t.Errorf("expected 'number' with value 42, got %v", m["number"])
		}
	})
}

func TestErrors(t *testing.T) {
	t.Run("ErrJSONRead", func(t *testing.T) {
		if ErrJSONRead.Error() != "invalid JSON file" {
			t.Errorf("unexpected error message: %s", ErrJSONRead.Error())
		}
	})

	t.Run("ErrInvalidData", func(t *testing.T) {
		if ErrInvalidData.Error() != "invalid data" {
			t.Errorf("unexpected error message: %s", ErrInvalidData.Error())
		}
	})

	t.Run("ErrURLRead", func(t *testing.T) {
		if ErrURLRead.Error() != "URL is not returning correct response" {
			t.Errorf("unexpected error message: %s", ErrURLRead.Error())
		}
	})

	t.Run("ErrStringRead", func(t *testing.T) {
		if ErrStringRead.Error() != "input is not a proper JSON string" {
			t.Errorf("unexpected error message: %s", ErrStringRead.Error())
		}
	})
}
