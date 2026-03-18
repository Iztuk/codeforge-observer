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
	ct = strings.TrimSpace(ct)

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
	// Content type normalization
	ct := strings.Split(contentType, ";")[0]
	ct = strings.TrimSpace(ct)

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

// Compares the fields between the request body and the contract definitions
func compareBody(schema OpenApiSchemaRef, obj map[string]any) []error {
	var findings []error
	if schema.Type != "object" {
		findings = append(findings, fmt.Errorf("expected type object, got %s", schema.Type))
		return findings
	}

	for _, field := range schema.Required {
		if _, ok := obj[field]; !ok {
			findings = append(findings, fmt.Errorf("missing required field: %s", field))
		}
	}

	return findings
}
