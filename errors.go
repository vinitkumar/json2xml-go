// Package json2xml provides utilities to convert JSON data to XML format.
package json2xml

import "errors"

// Custom errors for json2xml package
var (
	// ErrJSONRead is returned when there is an error reading JSON data.
	ErrJSONRead = errors.New("invalid JSON file")

	// ErrInvalidData is returned when the data is invalid.
	ErrInvalidData = errors.New("invalid data")

	// ErrURLRead is returned when there is an error reading from a URL.
	ErrURLRead = errors.New("URL is not returning correct response")

	// ErrStringRead is returned when there is an error reading from a string.
	ErrStringRead = errors.New("input is not a proper JSON string")
)
