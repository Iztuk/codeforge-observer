package audit

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func makeTestResponseRequest(t *testing.T, method, path string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "api.local"
	return req
}

func TestAuditResponse(t *testing.T) {
	doc := testContract()

	tests := []struct {
		name                 string
		op                   *OpenApiOperation
		method               string
		path                 string
		statusCode           int
		body                 []byte
		contentType          string
		expectedFindingCodes []FindingCode
		expectedMessages     []string
	}{
		{
			name:                 "status not defined",
			op:                   doc.Paths["/api/accounts"].GET,
			method:               http.MethodGet,
			path:                 "/api/accounts",
			statusCode:           500,
			body:                 []byte(`{"message":"boom"}`),
			contentType:          "application/json",
			expectedFindingCodes: []FindingCode{CodeResponseStatusNotDefined},
			expectedMessages:     []string{"response status code not defined: 500"},
		},
		{
			name:                 "documented 404 without content",
			op:                   doc.Paths["/api/accounts"].GET,
			method:               http.MethodGet,
			path:                 "/api/accounts",
			statusCode:           404,
			body:                 []byte(`<html>not found</html>`),
			contentType:          "text/html; charset=utf-8",
			expectedFindingCodes: nil,
			expectedMessages:     nil,
		},
		{
			name:                 "missing required response fields",
			op:                   doc.Paths["/api/accounts"].POST,
			method:               http.MethodPost,
			path:                 "/api/accounts",
			statusCode:           201,
			body:                 []byte(`{"id":"123"}`),
			contentType:          "application/json",
			expectedFindingCodes: []FindingCode{CodeResponseRequiredFieldMissing},
			expectedMessages:     []string{"missing required field: email"},
		},
		{
			name:                 "valid 201 response",
			op:                   doc.Paths["/api/accounts"].POST,
			method:               http.MethodPost,
			path:                 "/api/accounts",
			statusCode:           201,
			body:                 []byte(`{"id":"123","email":"test@example.com"}`),
			contentType:          "application/json",
			expectedFindingCodes: nil,
			expectedMessages:     nil,
		},
		{
			name:                 "nil operation",
			op:                   nil,
			method:               http.MethodGet,
			path:                 "/api/accounts",
			statusCode:           200,
			body:                 []byte(`{"ok": true}`),
			contentType:          "application/json",
			expectedFindingCodes: []FindingCode{CodeResponseOperationMissing},
			expectedMessages:     []string{"operation is nil"},
		},
		{
			name:                 "empty body with declared content",
			op:                   doc.Paths["/api/accounts"].POST,
			method:               http.MethodPost,
			path:                 "/api/accounts",
			statusCode:           201,
			body:                 nil,
			contentType:          "application/json",
			expectedFindingCodes: []FindingCode{CodeResponseBodyMissing},
			expectedMessages:     []string{"response body is missing"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var rc io.ReadCloser
			if tc.body == nil {
				rc = io.NopCloser(bytes.NewBuffer(nil))
			} else {
				rc = io.NopCloser(bytes.NewBuffer(tc.body))
			}

			resp := &http.Response{
				StatusCode: tc.statusCode,
				Header:     make(http.Header),
				Body:       rc,
				Request:    makeTestResponseRequest(t, tc.method, tc.path),
			}

			if tc.contentType != "" {
				resp.Header.Set("Content-Type", tc.contentType)
			}

			findings := AuditResponse(resp, tc.op, doc.Components)

			if len(findings) != len(tc.expectedFindingCodes) {
				t.Fatalf("expected %d findings, got %d: %+v", len(tc.expectedFindingCodes), len(findings), findings)
			}

			for i, expectedCode := range tc.expectedFindingCodes {
				if findings[i].Code != expectedCode {
					t.Fatalf("finding[%d]: expected code %q, got %q", i, expectedCode, findings[i].Code)
				}
			}

			for i, expectedMsg := range tc.expectedMessages {
				if findings[i].Message != expectedMsg {
					t.Fatalf("finding[%d]: expected message %q, got %q", i, expectedMsg, findings[i].Message)
				}
			}
		})
	}
}
