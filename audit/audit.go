package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

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

func AuditRequest(r *http.Request, contractsDoc OpenApiDoc) []error {
	var errors []error

	pi, err := comparePath(r.URL.Path, contractsDoc.Paths)
	if err != nil {
		errors = append(errors, err)
		return errors
	}

	op, err := compareMethod(r.Method, pi)
	if err != nil {
		errors = append(errors, err)
		return errors
	}

	if op.RequestBody == nil {
		return errors
	}

	ct := r.Header.Get("Content-Type")
	ref, schema, err := fetchSchemaRef(ct, op)
	if err != nil {
		errors = append(errors, err)
	}

	if ref != "" {
		if contractsDoc.Components == nil {
			errors = append(errors, fmt.Errorf("components are nil"))
			return errors
		}

		const prefix = "#/components/schemas/"
		name := strings.TrimPrefix(ref, prefix)

		resolved, ok := contractsDoc.Components.Schemas[name]
		if !ok {
			errors = append(errors, fmt.Errorf("schema ref not found: %s", ref))
			return errors
		}

		schema = resolved
	}

	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, err = io.ReadAll(r.Body)
		if err != nil {
			errors = append(errors, err)
			return errors
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	if len(bodyBytes) == 0 {
		if op.RequestBody.Required {
			errors = append(errors, fmt.Errorf("required request body is missing"))
		}
		return errors
	}

	var body any
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		errors = append(errors, err)
		return errors
	}

	obj, ok := body.(map[string]any)
	if !ok {
		errors = append(errors, fmt.Errorf("request body is not a JSON object"))
		return errors
	}

	errors = append(errors, compareRequestBody(schema, obj)...)
	return errors
}

func AuditResponse(r *http.Response) {

}
