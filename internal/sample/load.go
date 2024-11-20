package sample

import (
	"embed"
	"path/filepath"

	"github.com/efritz/gostgres/internal/execution/engine"
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/syntax/parsing"
)

func LoadPagilaSampleSchema(engine *engine.Engine) error {
	statements, err := readPagilaFile("schema.sql")
	if err != nil {
		return err
	}

	return loadPagilaSample(engine, statements)
}

func LoadPagilaSampleData(engine *engine.Engine) error {
	statements, err := readPagilaFile("data.sql")
	if err != nil {
		return err
	}

	return loadPagilaSample(engine, statements)
}

func LoadPagilaSampleSchemaAndData(engine *engine.Engine) error {
	if err := LoadPagilaSampleSchema(engine); err != nil {
		return err
	}

	if err := LoadPagilaSampleData(engine); err != nil {
		return err
	}

	return nil
}

func loadPagilaSample(engine *engine.Engine, statements []string) error {
	for _, statement := range statements {
		if err := engine.QueryError(protocol.Request{
			Query: statement,
			Debug: false,
		}); err != nil {
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
