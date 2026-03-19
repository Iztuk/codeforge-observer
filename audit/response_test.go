package audit

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAuditResponse_StatusNotDefined(t *testing.T) {
	doc := testContract()
	op := doc.Paths["/api/accounts"].GET

	resp := &http.Response{
		StatusCode: 500,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(`{"message":"boom"}`)),
	}

	errs := AuditResponse(resp, op, doc.Components)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Error(), "response status code not defined: 500") {
		t.Fatalf("unexpected error: %v", errs[0])
	}
}

func TestAuditResponse_Documented404WithoutContent(t *testing.T) {
	doc := testContract()
	op := doc.Paths["/api/accounts"].GET

	resp := &http.Response{
		StatusCode: 404,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(`<html>not found</html>`)),
	}
	resp.Header.Set("Content-Type", "text/html; charset=utf-8")

	errs := AuditResponse(resp, op, doc.Components)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestAuditResponse_MissingRequiredFields(t *testing.T) {
	doc := testContract()
	op := doc.Paths["/api/accounts"].POST

	resp := &http.Response{
		StatusCode: 201,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(`{"id":"123"}`)),
	}
	resp.Header.Set("Content-Type", "application/json")

	errs := AuditResponse(resp, op, doc.Components)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Error(), "missing required field: email") {
		t.Fatalf("unexpected error: %v", errs[0])
	}
}

func TestAuditResponse_Valid201(t *testing.T) {
	doc := testContract()
	op := doc.Paths["/api/accounts"].POST

	resp := &http.Response{
		StatusCode: 201,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(`{"id":"123","email":"test@example.com"}`)),
	}
	resp.Header.Set("Content-Type", "application/json")

	errs := AuditResponse(resp, op, doc.Components)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestAuditResponse_NilOperation(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(`{"ok": true}`)),
	}

	errs := AuditResponse(resp, nil, nil)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Error(), "operation is nil") {
		t.Fatalf("unexpected error: %v", errs[0])
	}
}

func TestAuditResponse_EmptyBody_WithDeclaredContent(t *testing.T) {
	doc := testContract()
	op := doc.Paths["/api/accounts"].POST

	resp := &http.Response{
		StatusCode: 201,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBuffer(nil)), // empty body
	}
	resp.Header.Set("Content-Type", "application/json")

	errs := AuditResponse(resp, op, doc.Components)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !strings.Contains(errs[0].Error(), "response body is missing") {
		t.Fatalf("unexpected error: %v", errs[0])
	}
}
