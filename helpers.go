package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type JSON map[string]any

// Send encodes v as JSON and writes it to the response.
// It returns an error if encoding fails.
//
//	return httpx.Send(w, user)
func Send(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

// SendStatus encodes v as JSON and writes it to the response with the given status code.
// It returns an error if encoding fails.
//
//	return httpx.SendStatus(w, http.StatusCreated, user)
func SendStatus(w http.ResponseWriter, statusCode int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(v)
}

// Read decodes the JSON request body into T.
// Returns an error if the body is malformed JSON, or has wrong Content-Type.
// If Content-Type header is present, it must be "application/json" (or with charset).
//
//	user, err := httpx.Read[User](r)
//	if err != nil {
//	    return httperrors.New(err.Error(), http.StatusBadRequest)
//	}
func Read[T any](r *http.Request) (data T, err error) {
	// Check Content-Type header if present
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
		return data, fmt.Errorf("invalid content-type %s", contentType)
	}

	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return data, fmt.Errorf("invalid JSON: %w", err)
	}

	return data, nil
}
