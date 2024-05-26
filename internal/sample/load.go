package sample

import (
	"embed"
	"path/filepath"

	"github.com/efritz/gostgres/internal/execution/engine"
	"github.com/efritz/gostgres/internal/syntax/parsing"
)

func LoadPagilaSampleSchemaAndData(engine *engine.Engine) error {
	schema, err := readPagilaFile("schema.sql")
	if err != nil {
		return err
	}

	data, err := readPagilaFile("data.sql")
	if err != nil {
		return err
	}

	for _, statement := range append(schema, data...) {
		if _, err := engine.Query(statement, false); err != nil {
			return err
		}
	}

	return nil
}

//go:embed data/pagila
var data embed.FS

func readPagilaFile(path string) ([]string, error) {
	schemaContent, err := data.ReadFile(filepath.Join("data/pagila", path))
	if err != nil {
		return nil, err
	}

	return parsing.SplitStatements(string(schemaContent)), nil
}
