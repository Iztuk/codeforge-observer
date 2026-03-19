package audit

func testContract() OpenApiDoc {
	return OpenApiDoc{
		OpenAPI: "3.0.3",
		Info: OpenApiInfo{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: map[string]OpenApiPathItem{
			"/api/accounts": {
				GET: &OpenApiOperation{
					OperationID: "accounts.readMany",
					Responses: map[string]OpenApiResponse{
						"200": {
							Description: "OK",
							Content: map[string]OpenApiMediaType{
								"application/json": {
									Schema: &OpenApiSchemaRef{
										Ref: "#/components/schemas/AccountsListResponse",
									},
								},
							},
						},
						"404": {
							Description: "Not found",
						},
					},
				},
				POST: &OpenApiOperation{
					OperationID: "accounts.create",
					RequestBody: &OpenApiRequestBody{
						Required: true,
						Content: map[string]OpenApiMediaType{
							"application/json": {
								Schema: &OpenApiSchemaRef{
									Ref: "#/components/schemas/AccountsCreateRequest",
								},
							},
						},
					},
					Responses: map[string]OpenApiResponse{
						"201": {
							Description: "Created",
							Content: map[string]OpenApiMediaType{
								"application/json": {
									Schema: &OpenApiSchemaRef{
										Ref: "#/components/schemas/AccountResponse",
									},
								},
							},
						},
					},
				},
			},
		},
		Components: &OpenApiComponents{
			Schemas: map[string]OpenApiSchemaRef{
				"AccountsCreateRequest": {
					Type:     "object",
					Required: []string{"email", "password"},
					Properties: map[string]OpenApiSchemaRef{
						"email":    {Type: "string"},
						"password": {Type: "string"},
					},
				},
				"AccountResponse": {
					Type:     "object",
					Required: []string{"id", "email"},
					Properties: map[string]OpenApiSchemaRef{
						"id":    {Type: "string"},
						"email": {Type: "string"},
					},
				},
				"AccountsListResponse": {
					Type:     "object",
					Required: []string{"items"},
					Properties: map[string]OpenApiSchemaRef{
						"items": {
							Type: "array",
							Items: &OpenApiSchemaRef{
								Ref: "#/components/schemas/AccountResponse",
							},
						},
					},
				},
			},
		},
	}
}
