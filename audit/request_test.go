package audit

import (
	"bytes"
	"net/http"
	"testing"
)

func TestAuditRequest(t *testing.T) {
	tests := []struct {
		name                 string
		method               string
		path                 string
		body                 []byte
		contentType          string
		expectedOpNil        bool
		expectedFindingCodes []FindingCode
		expectedMessages     []string
	}{
		{
			name:                 "request path not found",
			method:               http.MethodGet,
			path:                 "/does-not-exist",
			expectedOpNil:        true,
			expectedFindingCodes: []FindingCode{CodePathNotFound},
			expectedMessages:     []string{"path /does-not-exist not found in contract"},
		},
		{
			name:                 "method not defined",
			method:               http.MethodDelete,
			path:                 "/api/accounts",
			expectedOpNil:        true,
			expectedFindingCodes: []FindingCode{CodeMethodNotDefined},
			expectedMessages:     []string{"method DELETE not defined for path"},
		},
		{
			name:                 "required body missing with content type",
			method:               http.MethodPost,
			path:                 "/api/accounts",
			contentType:          "application/json",
			expectedOpNil:        false,
			expectedFindingCodes: []FindingCode{CodeRequestBodyMissing},
			expectedMessages:     []string{"required request body is missing"},
		},
		{
			name:                 "required body missing without content type",
			method:               http.MethodPost,
			path:                 "/api/accounts",
			expectedOpNil:        false,
			expectedFindingCodes: []FindingCode{CodeRequestBodyMissing},
			expectedMessages:     []string{"required request body is missing"},
		},
		{
			name:          "missing required fields",
			method:        http.MethodPost,
			path:          "/api/accounts",
			body:          []byte(`{}`),
			contentType:   "application/json",
			expectedOpNil: false,
			expectedFindingCodes: []FindingCode{
				CodeRequestRequiredFieldMissing,
				CodeRequestRequiredFieldMissing,
			},
			expectedMessages: []string{
				"missing required field: email",
				"missing required field: password",
			},
		},
		{
			name:          "valid post",
			method:        http.MethodPost,
			path:          "/api/accounts",
			body:          []byte(`{"email":"test@example.com","password":"secret"}`),
			contentType:   "application/json",
			expectedOpNil: false,
		},
	}

	doc := testContract()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tc.body != nil {
				bodyReader = bytes.NewReader(tc.body)
			} else {
				bodyReader = bytes.NewReader(nil)
			}

			req, err := http.NewRequest(tc.method, tc.path, bodyReader)
			if err != nil {
				t.Fatal(err)
			}

			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			findings, op := AuditRequest(req, doc)

			if tc.expectedOpNil && op != nil {
				t.Fatalf("expected nil operation, got non-nil")
			}
			if !tc.expectedOpNil && op == nil {
				t.Fatalf("expected non-nil operation, got nil")
			}

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

func TestOpenApiPathPattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		path     string
		expected bool
	}{
		{
			name:     "exact match returns true",
			pattern:  "/api/accounts",
			path:     "/api/accounts",
			expected: true,
		},
		{
			name:     "path parameter match returns true",
			pattern:  "/api/accounts/{id}",
			path:     "/api/accounts/01",
			expected: true,
		},
		{
			name:     "different base path returns false",
			pattern:  "/api/accounts/{id}",
			path:     "/api/account/01",
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := matchOpenApiPath(tc.pattern, tc.path)
			if actual != tc.expected {
				t.Fatalf("matchOpenApiPath(%s, %s) = %t; expected %t", tc.pattern, tc.path, actual, tc.expected)
			}
		})
	}
}
