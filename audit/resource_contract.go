package audit

import (
	"encoding/json"
	"os"
)

type ResourceDoc struct{}

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
