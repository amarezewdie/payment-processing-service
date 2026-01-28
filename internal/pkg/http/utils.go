package http

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response with a specific status code.
func ResponseJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Bind decodes the request body into the provided interface.
func DecodeRequestJSON(r *http.Request, i interface{}) error {
	return json.NewDecoder(r.Body).Decode(i)
}

// ErrorResponse represents a simple error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// HTTPError writes a JSON error response.
func HTTPError(w http.ResponseWriter, status int, message string) {
	ResponseJSON(w, status, ErrorResponse{Error: message})
}
