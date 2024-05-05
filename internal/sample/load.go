package sample

import (
	"embed"
	"path/filepath"
	"strings"

	"github.com/efritz/gostgres/internal/engine"
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
		if _, err := engine.Query(statement); err != nil {
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

	return parseStatements(string(schemaContent)), nil
}

func parseStatements(content string) []string {
	var normalized []string
	for _, line := range strings.Split(string(content), "\n") {
		normalized = append(normalized, strings.Split(line, "--")[0])
	}
	statements := strings.Split(strings.Join(normalized, "\n"), ";")

	var filtered []string
	for _, s := range statements {
		s = strings.TrimSpace(s)
		if s != "" {
			filtered = append(filtered, s)
		}
	}

	return filtered
}
