package audit

import (
	"fmt"
	"net/http"
)

// TODO: Get the request path and find it from API contracts
func comparePath(path string, contractPaths map[string]OpenApiPathItem) (*OpenApiPathItem, error) {
	val, ok := contractPaths[path]
	if !ok {
		return nil, fmt.Errorf("path %s not found in contract", path)
	}

	return &val, nil
}

// TODO: Get the request method and compare it to the OpenApiPathItem
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
