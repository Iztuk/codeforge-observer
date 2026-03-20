package audit

import (
	"codeforge-observer/utils"
	"fmt"
	"net/http"
	"strings"
)

func comparePath(path string, contractPaths map[string]OpenApiPathItem) (*OpenApiPathItem, *Finding) {
	if val, ok := contractPaths[path]; ok {
		return &val, nil
	}

	for pattern, val := range contractPaths {
		if matchOpenApiPath(pattern, path) {
			return &val, nil
		}
	}

	return nil, &Finding{
		Source:   ApiContract,
		Stage:    RequestStage,
		Severity: SeverityError,
		Code:     CodePathNotFound,
		Message:  fmt.Sprintf("path %s not found in contract", path),
	}
}

func matchOpenApiPath(pattern, path string) bool {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			// Path param: accept any non-empty segment
			if pathParts[i] == "" {
				return false
			}
			continue
		}

		if pathParts[i] != part {
			return false
		}
	}

	return true
}

func compareMethod(method string, pathItem *OpenApiPathItem) (*OpenApiOperation, *Finding) {
	var op *OpenApiOperation

	switch method {
	case http.MethodGet:
		op = pathItem.GET
	case http.MethodPost:
		op = pathItem.POST
	case http.MethodPut:
		op = pathItem.PUT
	case http.MethodPatch:
		op = pathItem.PATCH
	case http.MethodDelete:
		op = pathItem.DELETE
	case http.MethodHead:
		op = pathItem.HEAD
	case http.MethodOptions:
		op = pathItem.OPTIONS
	default:
		return nil, &Finding{
			Source:   ApiContract,
			Stage:    RequestStage,
			Severity: SeverityError,
			Code:     CodeMethodNotDefined,
			Message:  fmt.Sprintf("unsupported method %s", method),
		}
	}

	if op == nil {
		return nil, &Finding{
			Source:   ApiContract,
			Stage:    RequestStage,
			Severity: SeverityError,
			Code:     CodeMethodNotDefined,
			Message:  fmt.Sprintf("method %s not defined for path", method),
		}
	}

	return op, nil
}

func fetchRequestBodySchema(contentType string, op *OpenApiOperation) (string, OpenApiSchemaRef, *Finding) {
	// Content type normalization
	ct := strings.Split(contentType, ";")[0]
	ct = strings.ToLower(strings.TrimSpace(ct))

	mt, ok := op.RequestBody.Content[ct]
	if !ok {
		return "", OpenApiSchemaRef{}, &Finding{
			Source:   ApiContract,
			Stage:    RequestStage,
			Severity: SeverityError,
			Code:     CodeRequestContentTypeUnsupported,
			Message:  fmt.Sprintf("content type not supported: %s", ct),
		}
	}

	if mt.Schema == nil {
		return "", OpenApiSchemaRef{}, &Finding{
			Source:   ApiContract,
			Stage:    RequestStage,
			Severity: SeverityError,
			Code:     CodeRequestSchemaMissing,
			Message:  fmt.Sprintf("no schema defined for content type: %s", ct),
		}
	}

	// Return the schema reference
	if mt.Schema.Ref != "" {
		return mt.Schema.Ref, OpenApiSchemaRef{}, nil
	}

	return "", *mt.Schema, nil
}

func fetchResponseBodySchema(contentType string, res *OpenApiResponse) (string, OpenApiSchemaRef, error) {
	if res == nil {
		return "", OpenApiSchemaRef{}, fmt.Errorf("response is nil")
	}
	// Content type normalization
	ct := strings.Split(contentType, ";")[0]
	ct = strings.ToLower(strings.TrimSpace(ct))

	mt, ok := res.Content[ct]
	if !ok {
		return "", OpenApiSchemaRef{}, fmt.Errorf("content type not supported: %s", ct)
	}

	if mt.Schema == nil {
		return "", OpenApiSchemaRef{}, fmt.Errorf("no schema defined for content type: %s", ct)
	}

	// Return the schema reference
	if mt.Schema.Ref != "" {
		return mt.Schema.Ref, OpenApiSchemaRef{}, nil
	}

	return "", *mt.Schema, nil
}

// Compares the fields between the request/response body and the contract definitions
// NOTE:
// This currently performs shallow validation:
//   - Ensures correct top-level type (object/array)
//   - Validates required fields for objects
//   - Recursively validates array items
//
// TODO: (future enhancement)
//   - Validate object properties recursively (nested objects)
//   - Enforce property types (string, integer, boolean, etc.)
//   - Support additional schema constraints (enum, format, min/max, length, etc.)
//   - Detect unknown/extra fields not defined in schema
//   - Improve error reporting with full JSON path (e.g., "items[0].email")
//   - Support nullable and optional fields properly
func compareBody(schema OpenApiSchemaRef, value any, stage FindingStage, req *http.Request, res *http.Response) []Finding {
	var findings []Finding

	switch schema.Type {
	case "object":
		obj, ok := value.(map[string]any)
		if !ok {
			return []Finding{
				{
					Source:   ApiContract,
					Stage:    stage,
					Severity: SeverityError,
					Code:     codeForTypeMismatch(stage),
					Message:  "expected object",
					Metadata: bodyFindingMetadata(stage, req, res, ""),
				},
			}
		}

		for _, field := range schema.Required {
			if _, ok := obj[field]; !ok {
				findings = append(findings, Finding{
					Source:   ApiContract,
					Stage:    stage,
					Severity: SeverityError,
					Code:     codeForRequiredFieldMissing(stage),
					Message:  fmt.Sprintf("missing required field: %s", field),
					Metadata: bodyFindingMetadata(stage, req, res, field),
				})
			}
		}

	case "array":
		arr, ok := value.([]any)
		if !ok {
			return []Finding{
				{
					Source:   ApiContract,
					Stage:    stage,
					Severity: SeverityError,
					Code:     codeForTypeMismatch(stage),
					Message:  "expected array",
					Metadata: bodyFindingMetadata(stage, req, res, ""),
				},
			}
		}

		if schema.Items == nil {
			return []Finding{
				{
					Source:   ApiContract,
					Stage:    stage,
					Severity: SeverityError,
					Code:     CodeArrayItemsSchemaMissing,
					Message:  "array schema missing items definition",
					Metadata: bodyFindingMetadata(stage, req, res, ""),
				},
			}
		}

		for i, item := range arr {
			itemFindings := compareBody(*schema.Items, item, stage, req, res)
			for _, f := range itemFindings {
				f.Message = fmt.Sprintf("array index %d: %s", i, f.Message)
				findings = append(findings, f)
			}
		}
	}

	return findings
}

func bodyFindingMetadata(stage FindingStage, req *http.Request, res *http.Response, field string) *FindingMetadata {
	md := &FindingMetadata{
		Field: field,
	}

	if stage == RequestStage && req != nil {
		md.RequestID = utils.GetOrCreateRequestID(req)
		md.Host = req.Host
		md.Path = req.URL.Path
		md.Method = HttpMethod(req.Method)
		return md
	}

	if stage == ResponseStage && res != nil && res.Request != nil {
		md.RequestID = utils.GetOrCreateRequestID(res.Request)
		md.Host = res.Request.Host
		md.Path = res.Request.URL.Path
		md.Method = HttpMethod(res.Request.Method)
		return md
	}

	return md
}

func codeForRequiredFieldMissing(stage FindingStage) FindingCode {
	switch stage {
	case ResponseStage:
		return CodeResponseRequiredFieldMissing
	case RequestStage:
		return CodeRequestRequiredFieldMissing
	default:
		return CodeRequestRequiredFieldMissing
	}
}

func codeForTypeMismatch(stage FindingStage) FindingCode {
	switch stage {
	case ResponseStage:
		return CodeResponseTypeMismatch
	case RequestStage:
		return CodeRequestTypeMismatch
	default:
		return CodeRequestTypeMismatch
	}
}
