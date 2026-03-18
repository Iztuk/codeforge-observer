package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

func AuditRequest(r *http.Request, contractsDoc OpenApiDoc) ([]error, *OpenApiOperation) {
	var findings []error

	pi, err := comparePath(r.URL.Path, contractsDoc.Paths)
	if err != nil {
		findings = append(findings, err)
		return findings, nil
	}

	op, err := compareMethod(r.Method, pi)
	if err != nil {
		findings = append(findings, err)
		return findings, nil
	}

	if op.RequestBody == nil {
		return findings, op
	}

	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, err = io.ReadAll(r.Body)
		if err != nil {
			findings = append(findings, err)
			return findings, op
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	if len(bodyBytes) == 0 {
		if op.RequestBody.Required {
			findings = append(findings, fmt.Errorf("required request body is missing"))
		}
		return findings, op
	}

	ct := r.Header.Get("Content-Type")
	ref, schema, err := fetchRequestBodySchema(ct, op)
	if err != nil {
		findings = append(findings, err)
		return findings, op
	}

	if ref != "" {
		if contractsDoc.Components == nil {
			findings = append(findings, fmt.Errorf("components are nil"))
			return findings, op
		}

		const prefix = "#/components/schemas/"
		name := strings.TrimPrefix(ref, prefix)

		resolved, ok := contractsDoc.Components.Schemas[name]
		if !ok {
			findings = append(findings, fmt.Errorf("schema ref not found: %s", ref))
			return findings, op
		}

		schema = resolved
	}

	var body any
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		findings = append(findings, err)
		return findings, op
	}

	obj, ok := body.(map[string]any)
	if !ok {
		findings = append(findings, fmt.Errorf("request body is not a JSON object"))
		return findings, op
	}

	findings = append(findings, compareBody(schema, obj)...)
	return findings, op
}

func AuditResponse(r *http.Response, op *OpenApiOperation, components *OpenApiComponents) []error {
	var findings []error
	if op == nil {
		findings = append(findings, fmt.Errorf("operation is nil"))
		return findings
	}

	res, ok := op.Responses[strconv.Itoa(r.StatusCode)]
	if !ok {
		findings = append(findings, fmt.Errorf("response status code not defined: %d", r.StatusCode))
		return findings
	}

	if len(res.Content) == 0 {
		return findings
	}

	var bodyBytes []byte
	if r.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(r.Body)
		if err != nil {
			findings = append(findings, err)
			return findings
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	if len(bodyBytes) == 0 && len(res.Content) > 0 {
		findings = append(findings, fmt.Errorf("response body is missing"))
		return findings
	}

	ct := r.Header.Get("Content-Type")
	ref, schema, err := fetchResponseBodySchema(ct, &res)
	if err != nil {
		findings = append(findings, err)
		return findings
	}

	if ref != "" {
		if components == nil {
			findings = append(findings, fmt.Errorf("components are nil"))
			return findings
		}

		const prefix = "#/components/schemas/"
		name := strings.TrimPrefix(ref, prefix)

		resolved, ok := components.Schemas[name]
		if !ok {
			findings = append(findings, fmt.Errorf("schema ref not found: %s", ref))
			return findings
		}

		schema = resolved
	}

	var body any
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		findings = append(findings, err)
		return findings
	}

	obj, ok := body.(map[string]any)
	if !ok {
		findings = append(findings, fmt.Errorf("response body is not a JSON object"))
		return findings
	}

	findings = append(findings, compareBody(schema, obj)...)

	return findings
}
