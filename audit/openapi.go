package audit

import (
	"encoding/json"
	"os"
)

type OpenApiDoc struct {
	OpenAPI    string                     `json:"openapi"`
	Info       OpenApiInfo                `json:"info"`
	Paths      map[string]OpenApiPathItem `json:"paths"`
	Components *OpenApiComponents         `json:"components,omitempty"`
}

type OpenApiInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

type OpenApiPathItem struct {
	GET     *OpenApiOperation `json:"get,omitempty"`
	POST    *OpenApiOperation `json:"post,omitempty"`
	PUT     *OpenApiOperation `json:"put,omitempty"`
	PATCH   *OpenApiOperation `json:"patch,omitempty"`
	DELETE  *OpenApiOperation `json:"delete,omitempty"`
	HEAD    *OpenApiOperation `json:"head,omitempty"`
	OPTIONS *OpenApiOperation `json:"options,omitempty"`
}

type OpenApiOperation struct {
	Tags        []string `json:"tags,omitempty"`
	OperationID string   `json:"operationId,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Description string   `json:"description,omitempty"`

	XResource *RouteResourceBinding `json:"x-resource,omitempty"`

	Parameters  []OpenApiParameter         `json:"parameters,omitempty"`
	RequestBody *OpenApiRequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]OpenApiResponse `json:"responses,omitempty"`
}

type OpenApiParameter struct {
	Name        string            `json:"name"`
	In          string            `json:"in"` // path, query, header, cookie
	Description string            `json:"description,omitempty"`
	Required    bool              `json:"required,omitempty"`
	Schema      *OpenApiSchemaRef `json:"schema,omitempty"`
}

type OpenApiRequestBody struct {
	Description string                      `json:"description,omitempty"`
	Required    bool                        `json:"required,omitempty"`
	Content     map[string]OpenApiMediaType `json:"content,omitempty"`
}

type OpenApiResponse struct {
	Description string                      `json:"description"`
	Content     map[string]OpenApiMediaType `json:"content,omitempty"`
}

type OpenApiMediaType struct {
	Schema *OpenApiSchemaRef `json:"schema,omitempty"`
}

type OpenApiComponents struct {
	Schemas map[string]OpenApiSchemaRef `json:"schemas,omitempty"`
}

type OpenApiSchemaRef struct {
	Ref string `json:"$ref,omitempty"`

	Type       string                      `json:"type,omitempty"`
	Properties map[string]OpenApiSchemaRef `json:"properties,omitempty"`
	Required   []string                    `json:"required,omitempty"`
	Items      *OpenApiSchemaRef           `json:"items,omitempty"`

	Enum      []json.RawMessage `json:"enum,omitempty"`
	Format    string            `json:"format,omitempty"`
	Nullable  bool              `json:"nullable,omitempty"`
	MinLength *int              `json:"minLength,omitempty"`
	MaxLength *int              `json:"maxLength,omitempty"`
}

type RouteResourceBinding struct {
	ResourceName string `json:"resource_name"`
	Table        string `json:"table"` // e.g. "tableName"
}

type ResourceOp string

const (
	ResourceReadOne       ResourceOp = "readOne"
	ResourceReadMany      ResourceOp = "readMany"
	ResourceCreate        ResourceOp = "create"
	ResourceUpdate        ResourceOp = "update"
	ResourcePartialUpdate ResourceOp = "partialUpdate"
	ResourceDelete        ResourceOp = "delete"
)

func ReadOpenApiDoc(docPath string) (OpenApiDoc, error) {
	var doc OpenApiDoc

	if docPath == "" {
		return OpenApiDoc{}, nil
	}

	d, err := os.ReadFile(docPath)
	if err != nil {
		return OpenApiDoc{}, err
	}

	if err = json.Unmarshal(d, &doc); err != nil {
		return OpenApiDoc{}, err
	}

	return doc, nil
}
