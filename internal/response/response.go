package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type dataEnvelope struct {
	Data any `json:"data"`
}

type errorEnvelope struct {
	Error errBody `json:"error"`
}

type errBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// Json writes `{"data": <payload>}` with the given status.
func Json(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(dataEnvelope{Data: payload})
}

// Error writes a standard error envelope. The code slug is derived from the status.
func Error(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(errorEnvelope{
		Error: errBody{
			Code:    codeFor(statusCode),
			Message: message,
		},
	})
}

func codeFor(status int) string {
	text := http.StatusText(status)
	if text == "" {
		return fmt.Sprintf("status_%d", status)
	}
	return strings.ReplaceAll(strings.ToLower(text), " ", "_")
}
