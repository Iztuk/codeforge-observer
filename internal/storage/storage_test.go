package storage

import (
	"bytes"
	"codeforge-observer/audit"
	"database/sql"
	"net/http"
	"testing"
)

func TestCreateAndReadHosts(t *testing.T) {
	db := setupTestDB(t)

	tests := []struct {
		name     string
		input    HostInfo
		wantName string
		wantUp   string
		wantErr  bool
	}{
		{
			name: "valid host",
			input: HostInfo{
				Name:     "api.local",
				Upstream: "http://localhost:8081",
				Contract: "codeforge.contracts.json",
				Resource: "codeforge.resources.json",
			},
			wantName: "api.local",
			wantUp:   "http://localhost:8081",
			wantErr:  false,
		},
		{
			name: "second valid host",
			input: HostInfo{
				Name:     "auth.local",
				Upstream: "http://localhost:8082",
				Contract: "auth.contracts.json",
				Resource: "auth.resources.json",
			},
			wantName: "auth.local",
			wantUp:   "http://localhost:8082",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateHost(tt.input, db)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreateHost() error = %v, wantErr %v", err, tt.wantErr)
			}

			hosts, err := ReadHosts(db)
			if err != nil {
				t.Fatalf("ReadHosts() error = %v", err)
			}

			var found *HostInfo
			for i := range hosts {
				if hosts[i].Name == tt.wantName {
					found = &hosts[i]
					break
				}
			}

			if found == nil {
				t.Fatalf("expected host %q to be present", tt.wantName)
			}

			if found.Upstream != tt.wantUp {
				t.Fatalf("got upstream %q, want %q", found.Upstream, tt.wantUp)
			}
		})
	}
}

func TestCreateHost_UniqueName(t *testing.T) {
	db := setupTestDB(t)

	first := HostInfo{
		Name:     "api.local",
		Upstream: "http://localhost:8081",
		Contract: "codeforge.contracts.json",
		Resource: "codeforge.resources.json",
	}

	if err := CreateHost(first, db); err != nil {
		t.Fatalf("initial CreateHost() error = %v", err)
	}

	tests := []struct {
		name    string
		input   HostInfo
		wantErr bool
	}{
		{
			name: "duplicate host name",
			input: HostInfo{
				Name:     "api.local",
				Upstream: "http://localhost:9999",
				Contract: "other.contracts.json",
				Resource: "other.resources.json",
			},
			wantErr: true,
		},
		{
			name: "different host name",
			input: HostInfo{
				Name:     "auth.local",
				Upstream: "http://localhost:8082",
				Contract: "auth.contracts.json",
				Resource: "auth.resources.json",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateHost(tt.input, db)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreateHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInsertFindings_AuditRequest_PersistsToDB(t *testing.T) {
	db := setupTestDB(t)

	doc := audit.TestContract()

	req, err := http.NewRequest(
		http.MethodPost,
		"/api/accounts",
		bytes.NewReader([]byte(`{}`)),
	)
	if err != nil {
		t.Fatal(err)
	}

	req.Host = "api.local"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Id", "req-123")

	findings, op := audit.AuditRequest(req, doc)

	if op == nil {
		t.Fatalf("expected non-nil operation, got nil")
	}

	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	for i := range findings {
		if findings[i].Metadata == nil {
			findings[i].Metadata = &audit.FindingMetadata{}
		}

		findings[i].Metadata.RequestID = "req-123"
		findings[i].Metadata.Host = "api.local"
		findings[i].Metadata.Path = "/api/accounts"
		findings[i].Metadata.Method = audit.HttpPost
		findings[i].Metadata.Body = `{}`
		findings[i].Metadata.OperationID = op.OperationID
	}

	if err := InsertFindings(findings, db); err != nil {
		t.Fatalf("InsertFindings() error = %v", err)
	}

	rows, err := db.Query(`
		SELECT
			source,
			stage,
			severity,
			code,
			message,
			request_id,
			host,
			path,
			method,
			body,
			operation_id,
			field,
			resource
		FROM findings
		ORDER BY message
	`)
	if err != nil {
		t.Fatalf("query persisted findings: %v", err)
	}
	defer rows.Close()

	type storedFinding struct {
		Source      string
		Stage       string
		Severity    string
		Code        string
		Message     string
		RequestID   sql.NullString
		Host        sql.NullString
		Path        sql.NullString
		Method      sql.NullString
		Body        sql.NullString
		OperationID sql.NullString
		Field       sql.NullString
		Resource    sql.NullString
	}

	var stored []storedFinding
	for rows.Next() {
		var f storedFinding
		if err := rows.Scan(
			&f.Source,
			&f.Stage,
			&f.Severity,
			&f.Code,
			&f.Message,
			&f.RequestID,
			&f.Host,
			&f.Path,
			&f.Method,
			&f.Body,
			&f.OperationID,
			&f.Field,
			&f.Resource,
		); err != nil {
			t.Fatalf("scan persisted finding: %v", err)
		}
		stored = append(stored, f)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("rows error: %v", err)
	}

	if len(stored) != 2 {
		t.Fatalf("expected 2 persisted findings, got %d", len(stored))
	}

	expectedMessages := map[string]bool{
		"missing required field: email":    false,
		"missing required field: password": false,
	}

	for _, f := range stored {
		if f.Source != string(audit.ApiContract) {
			t.Fatalf("expected source %q, got %q", audit.ApiContract, f.Source)
		}
		if f.Stage != string(audit.RequestStage) {
			t.Fatalf("expected stage %q, got %q", audit.RequestStage, f.Stage)
		}
		if f.Code != string(audit.CodeRequestRequiredFieldMissing) {
			t.Fatalf("expected code %q, got %q", audit.CodeRequestRequiredFieldMissing, f.Code)
		}
		if !f.RequestID.Valid || f.RequestID.String != "req-123" {
			t.Fatalf("expected request_id req-123, got %+v", f.RequestID)
		}
		if !f.Host.Valid || f.Host.String != "api.local" {
			t.Fatalf("expected host api.local, got %+v", f.Host)
		}
		if !f.Path.Valid || f.Path.String != "/api/accounts" {
			t.Fatalf("expected path /api/accounts, got %+v", f.Path)
		}
		if !f.Method.Valid || f.Method.String != string(audit.HttpPost) {
			t.Fatalf("expected method %q, got %+v", audit.HttpPost, f.Method)
		}
		if !f.Body.Valid || f.Body.String != `{}` {
			t.Fatalf("expected body {}, got %+v", f.Body)
		}
		if !f.OperationID.Valid || f.OperationID.String == "" {
			t.Fatalf("expected non-empty operation_id")
		}

		seen, ok := expectedMessages[f.Message]
		if !ok {
			t.Fatalf("unexpected message %q", f.Message)
		}
		if seen {
			t.Fatalf("duplicate persisted finding message %q", f.Message)
		}
		expectedMessages[f.Message] = true
	}

	for msg, seen := range expectedMessages {
		if !seen {
			t.Fatalf("expected persisted finding with message %q", msg)
		}
	}
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() err = %v", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS hosts (
		name TEXT PRIMARY KEY,
		upstream TEXT NOT NULL UNIQUE,
		api_contract_file TEXT,
		resource_contract_file TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS findings (
		id TEXT PRIMARY KEY,
		source TEXT NOT NULL,
		stage TEXT NOT NULL,
		severity TEXT NOT NULL,
		code TEXT NOT NULL,
		message TEXT NOT NULL,
		request_id TEXT,
		host TEXT,
		path TEXT,
		method TEXT,
		body TEXT,
		operation_id TEXT,
		field TEXT,
		resource TEXT
	);
	`

	if _, err := db.Exec(query); err != nil {
		t.Fatalf("creating schema failed: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
