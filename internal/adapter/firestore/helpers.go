// Package firestore implements the inbound real-time push streaming adapter connecting Google Cloud Firestore snapshot streams to domain engines.
package firestore

// Import standard library packages for JSON serialization.
import (
	// encoding/json provides robust marshaling and unmarshaling between maps and structs.
	"encoding/json"
)

// domainToJSON converts an arbitrary data structure into standard JSON byte slice.
func domainToJSON(v any) ([]byte, error) {
	// Marshal input value to JSON bytes.
	return json.Marshal(v)
}

// domainFromJSON parses JSON bytes into a target destination struct pointer.
func domainFromJSON(data []byte, v any) error {
	// Unmarshal JSON bytes into destination pointer.
	return json.Unmarshal(data, v)
}
