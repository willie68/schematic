package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
)

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")

	// Buffer the JSON response to set Content-Length header
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// parseInt64Param parses a query parameter as int64, returning the default value on error or absence.
func parseInt64Param(r *http.Request, name string, defaultVal int64) int64 {
	s := r.URL.Query().Get(name)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}
