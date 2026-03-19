package audit

import (
	"fmt"
	"net/http"
	"strings"
)

func comparePath(path string, contractPaths map[string]OpenApiPathItem) (*OpenApiPathItem, error) {
	if val, ok := contractPaths[path]; ok {
		return &val, nil
	}

	for pattern, val := range contractPaths {
		if matchOpenApiPath(pattern, path) {
			return &val, nil
		}
	}

	return nil, fmt.Errorf("path %s not found in contract", path)
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

func compareMethod(method string, pathItem *OpenApiPathItem) (*OpenApiOperation, error) {
	if pathItem == nil {
		return nil, fmt.Errorf("path item is nil")
	}

	switch method {
	case http.MethodGet:
		if pathItem.GET == nil {
			return nil, fmt.Errorf("method %s not defined for path", method)
		}
		return pathItem.GET, nil
	case http.MethodPost:
		if pathItem.POST == nil {
			return nil, fmt.Errorf("method %s not defined for path", method)
		}
		return pathItem.POST, nil
	case http.MethodPut:
		if pathItem.PUT == nil {
			return nil, fmt.Errorf("method %s not defined for path", method)
		}
		return pathItem.PUT, nil
	case http.MethodPatch:
		if pathItem.PATCH == nil {
			return nil, fmt.Errorf("method %s not defined for path", method)
		}
		return pathItem.PATCH, nil
	case http.MethodDelete:
		if pathItem.DELETE == nil {
			return nil, fmt.Errorf("method %s not defined for path", method)
		}
		return pathItem.DELETE, nil
	case http.MethodHead:
		if pathItem.HEAD == nil {
			return nil, fmt.Errorf("method %s not defined for path", method)
		}
		return pathItem.HEAD, nil
	case http.MethodOptions:
		if pathItem.OPTIONS == nil {
			return nil, fmt.Errorf("method %s not defined for path", method)
		}
		return pathItem.OPTIONS, nil
	default:
		return nil, fmt.Errorf("unsupported method %s", method)
	}
}

func fetchRequestBodySchema(contentType string, op *OpenApiOperation) (string, OpenApiSchemaRef, error) {
	if op == nil {
		return "", OpenApiSchemaRef{}, fmt.Errorf("operation is nil")
	}

	if op.RequestBody == nil {
		return "", OpenApiSchemaRef{}, fmt.Errorf("request body is not defined in contract")
	}

	// Content type normalization
	ct := strings.Split(contentType, ";")[0]
	ct = strings.ToLower(strings.TrimSpace(ct))

	mt, ok := op.RequestBody.Content[ct]
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
func compareBody(schema OpenApiSchemaRef, value any) []error {
	var findings []error

	switch schema.Type {
	case "object":
		obj, ok := value.(map[string]any)
		if !ok {
			return []error{fmt.Errorf("expected object")}
		}

		for _, field := range schema.Required {
			if _, ok := obj[field]; !ok {
				findings = append(findings, fmt.Errorf("missing required field: %s", field))
			}
		}

	case "array":
		arr, ok := value.([]any)
		if !ok {
			return []error{fmt.Errorf("expected array")}
		}

		if schema.Items == nil {
			return []error{fmt.Errorf("array schema missing items definition")}
		}

		for i, item := range arr {
			itemFindings := compareBody(*schema.Items, item)
			for _, err := range itemFindings {
				findings = append(findings, fmt.Errorf("array index %d: %w", i, err))
			}
		}
	}

	return findings
}
