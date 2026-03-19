package api

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, ErrorResponse{
		Error: ErrorDetail{Code: code, Message: message},
	})
}

// WriteList writes a JSON response with a top-level key wrapping the list.
// If data is nil, an empty JSON array is returned to avoid null in responses.
func WriteList(w http.ResponseWriter, status int, key string, data any) {
	if data == nil {
		WriteJSON(w, status, map[string]any{key: []any{}})
		return
	}
	WriteJSON(w, status, map[string]any{key: data})
}
