package audit

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAuditRequest_PathNotFound(t *testing.T) {
	doc := testContract()

	req, err := http.NewRequest(http.MethodGet, "/does-not-exist", nil)
	if err != nil {
		t.Fatal(err)
	}

	errs, op := AuditRequest(req, doc)
	if op != nil {
		t.Fatalf("expected nil operation")
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Error(), "path /does-not-exist not found in contract") {
		t.Fatalf("unexpected error: %v", errs[0])
	}
}

func TestAuditRequest_MethodNotDefined(t *testing.T) {
	doc := testContract()

	req, err := http.NewRequest(http.MethodDelete, "/api/accounts", nil)
	if err != nil {
		t.Fatal(err)
	}

	errs, _ := AuditRequest(req, doc)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Error(), "method DELETE not defined for path") {
		t.Fatalf("unexpected error: %v", errs[0])
	}
}

func TestAuditRequest_RequiredBodyMissing(t *testing.T) {
	doc := testContract()

	req, err := http.NewRequest(http.MethodPost, "/api/accounts", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	errs, _ := AuditRequest(req, doc)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Error(), "required request body is missing") {
		t.Fatalf("unexpected error: %v", errs[0])
	}
}

func TestAuditRequest_RequiredBodyMissing_NoContentType(t *testing.T) {
	doc := testContract()

	req, err := http.NewRequest(http.MethodPost, "/api/accounts", nil)
	if err != nil {
		t.Fatal(err)
	}

	// NOTE: intentionally NOT setting Content-Type

	errs, _ := AuditRequest(req, doc)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}

	if !strings.Contains(errs[0].Error(), "required request body is missing") {
		t.Fatalf("unexpected error: %v", errs[0])
	}
}

func TestAuditRequest_MissingRequiredFields(t *testing.T) {
	doc := testContract()

	body := []byte(`{}`)
	req, err := http.NewRequest(http.MethodPost, "/api/accounts", io.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	errs, _ := AuditRequest(req, doc)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(errs))
	}
}

func TestAuditRequest_ValidPost(t *testing.T) {
	doc := testContract()

	body := []byte(`{"email":"test@example.com","password":"secret"}`)
	req, err := http.NewRequest(http.MethodPost, "/api/accounts", io.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	errs, op := AuditRequest(req, doc)
	if op == nil {
		t.Fatalf("expected operation")
	}
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
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
				t.Errorf("matchOpenApiPath(%s, %s) = %t; expected %t", tc.pattern, tc.path, actual, tc.expected)
			}
		})
	}
}
