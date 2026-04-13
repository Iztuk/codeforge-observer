package contracts

import (
	"encoding/json"
	"os"
)

type ResourceDoc struct {
	Resources map[string]Resource `json:"resources"`
}

type Resource struct {
	Name string       `json:"name"`
	Type ResourceType `json:"type"`

	DB *DB `json:"database,omitzero"`
}

type ResourceType string

const (
	DatabaseResource ResourceType = "database"
)

type DB struct {
	Dialect string             `json:"dialect"`
	Tables  map[string]DBTable `json:"tables"`
}

type DBTable struct {
	PrimaryKey []string             `json:"primary_key"`
	Fields     map[string]FieldSpec `json:"fields"`

	UniqueKeys [][]string `json:"unique_keys"`
}

type FieldSpec struct {
	ColumnName string  `json:"column_name"`
	Type       string  `json:"type"`
	Nullable   bool    `json:"nullable"`
	Default    *string `json:"default,omitzero"`

	Read    bool `json:"read"`
	Write   bool `json:"write"`
	Mutable bool `json:"mutable"`
}

func ReadResourceDoc(docPath string) (ResourceDoc, error) {
	var doc ResourceDoc

	if docPath == "" {
		return ResourceDoc{}, nil
	}

	d, err := os.ReadFile(docPath)
	if err != nil {
		return ResourceDoc{}, err
	}

	if err = json.Unmarshal(d, &doc); err != nil {
		return ResourceDoc{}, err
	}

	return doc, nil
}
