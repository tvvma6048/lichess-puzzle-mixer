package db

import (
	"embed"
)

//go:embed schema.sql
var schemaFS embed.FS

func loadSchemaSQL() (string, error) {
	b, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
