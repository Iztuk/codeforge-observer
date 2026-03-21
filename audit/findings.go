package audit

import (
	"codeforge-observer/utils"
	"net/http"
)

type Finding struct {
	Source   FindingSource    `json:"source"`
	Stage    FindingStage     `json:"stage"`
	Severity FindingSeverity  `json:"severity"`
	Code     FindingCode      `json:"code"`
	Message  string           `json:"message"`
	Metadata *FindingMetadata `json:"metadata,omitempty"`
}

type FindingSource string

const (
	ApiContract      FindingSource = "api"
	ResourceContract FindingSource = "resource"
)

type FindingStage string

const (
	RequestStage  FindingStage = "request"
	ResponseStage FindingStage = "response"
)

type FindingSeverity string

const (
	SeverityWarning FindingSeverity = "warning"
	SeverityError   FindingSeverity = "error"
)

type FindingCode string

const (
	// ---- API contract: request ----

	CodePathNotFound                  FindingCode = "path_not_found"
	CodeMethodNotDefined              FindingCode = "method_not_defined"
	CodeRequestBodyMissing            FindingCode = "request_body_missing"
	CodeRequestBodyInvalidJSON        FindingCode = "request_body_invalid_json"
	CodeRequestBodyNotObject          FindingCode = "request_body_not_object"
	CodeRequestBodyNotArray           FindingCode = "request_body_not_array"
	CodeRequestBodyReadFailed         FindingCode = "request_body_read_failed"
	CodeRequestContentTypeUnsupported FindingCode = "request_content_type_unsupported"
	CodeRequestSchemaMissing          FindingCode = "request_schema_missing"
	CodeRequestSchemaRefNotFound      FindingCode = "request_schema_ref_not_found"
	CodeRequestRequiredFieldMissing   FindingCode = "request_required_field_missing"
	CodeRequestTypeMismatch           FindingCode = "request_type_mismatch"
	CodeArrayItemsSchemaMissing       FindingCode = "array_items_schema_missing"

	// ---- API contract: response ----

	CodeResponseStatusNotDefined       FindingCode = "response_status_not_defined"
	CodeResponseBodyMissing            FindingCode = "response_body_missing"
	CodeResponseBodyReadFailed         FindingCode = "response_body_read_failed"
	CodeResponseBodyInvalidJSON        FindingCode = "response_body_invalid_json"
	CodeResponseBodyNotObject          FindingCode = "response_body_not_object"
	CodeResponseBodyNotArray           FindingCode = "response_body_not_array"
	CodeResponseContentTypeUnsupported FindingCode = "response_content_type_unsupported"
	CodeResponseSchemaMissing          FindingCode = "response_schema_missing"
	CodeResponseSchemaRefNotFound      FindingCode = "response_schema_ref_not_found"
	CodeResponseRequiredFieldMissing   FindingCode = "response_required_field_missing"
	CodeResponseTypeMismatch           FindingCode = "response_type_mismatch"
	CodeResponseOperationMissing       FindingCode = "response_operation_missing"

	// ---- Resource contract: request ----

	CodeReadOnlyFieldWritable     FindingCode = "read_only_field_writable"
	CodeUnknownWritableField      FindingCode = "unknown_writable_field"
	CodeRestrictedFieldWritable   FindingCode = "restricted_field_writable"
	CodeImmutableFieldWritable    FindingCode = "immutable_field_writable"
	CodeOperationResourceMismatch FindingCode = "operation_resource_mismatch"

	// ---- Resource contract: response ----

	CodeReadRestrictedFieldExposed FindingCode = "read_restricted_field_exposed"
	CodeUnknownReadableField       FindingCode = "unknown_readable_field"
	CodeSensitiveFieldExposed      FindingCode = "sensitive_field_exposed"
)

type FindingMetadata struct {
	RequestID   string     `json:"request_id,omitempty"`
	Host        string     `json:"host,omitempty"`
	Path        string     `json:"path,omitempty"`
	Method      HttpMethod `json:"method,omitempty"`
	Body        string     `json:"body,omitempty"`
	OperationID string     `json:"operation_id,omitempty"`
	Field       string     `json:"field,omitempty"`
	Resource    string     `json:"resource,omitempty"`
}

type HttpMethod string

const (
	HttpGet     HttpMethod = "GET"
	HttpPost    HttpMethod = "POST"
	HttpPut     HttpMethod = "PUT"
	HttpPatch   HttpMethod = "PATCH"
	HttpDelete  HttpMethod = "DELETE"
	HttpHead    HttpMethod = "HEAD"
	HttpOptions HttpMethod = "OPTIONS"
)

func requestFindingMetadata(r *http.Request) *FindingMetadata {
	return &FindingMetadata{
		RequestID: utils.GetOrCreateRequestID(r),
		Host:      r.Host,
		Path:      r.URL.Path,
		Method:    HttpMethod(r.Method),
	}
}

func responseFindingMetadata(r *http.Response) *FindingMetadata {
	return &FindingMetadata{
		RequestID: utils.GetOrCreateRequestID(r.Request),
		Host:      r.Request.Host,
		Path:      r.Request.URL.Path,
		Method:    HttpMethod(r.Request.Method),
	}
}
