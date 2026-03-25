package storage

import (
	"codeforge-observer/audit"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

func InsertFindings(findings []audit.Finding, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO findings (
			id,
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
		)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, f := range findings {
		fId := uuid.New().String()
		_, err := stmt.Exec(
			fId,
			f.Source,
			f.Stage,
			f.Severity,
			f.Code,
			f.Message,
			f.Metadata.RequestID,
			f.Metadata.Host,
			f.Metadata.Path,
			f.Metadata.Method,
			f.Metadata.Body,
			f.Metadata.OperationID,
			f.Metadata.Field,
			f.Metadata.Resource)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
