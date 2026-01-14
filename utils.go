package json2xml

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// ReadFromJSON reads a JSON file and returns the parsed data.
func ReadFromJSON(filename string) (any, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJSONRead, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJSONRead, err)
	}

	var result any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJSONRead, err)
	}

	return result, nil
}

// ReadFromURL loads JSON data from a URL and returns the parsed data.
func ReadFromURL(url string, params map[string]string) (any, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrURLRead, err)
	}

	// Add query parameters if provided
	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrURLRead, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrURLRead
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrURLRead, err)
	}

	var result any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrURLRead, err)
	}

	return result, nil
}

// ReadFromString parses a JSON string and returns the data.
func ReadFromString(jsonData string) (any, error) {
	if jsonData == "" {
		return nil, ErrStringRead
	}

	var result any
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrStringRead, err)
	}

	return result, nil
}
