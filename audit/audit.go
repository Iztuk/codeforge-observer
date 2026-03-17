package audit

import "time"

type Observation struct {
	Timestamp time.Time `json:"timestamp"`
	Event     string    `json:"event"`

	RequestID string `json:"request_id"`

	Host     string `json:"host"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Query    string `json:"query"`
	Upstream string `json:"upstream"`

	Status     int   `json:"status"`
	DurationMs int64 `json:"duration_ms"`

	Error string `json:"error,omitempty"`

	RequestHeaders  map[string][]string `json:"request_headers,omitempty"`
	ResponseHeaders map[string][]string `json:"response_headers,omitempty"`
}
