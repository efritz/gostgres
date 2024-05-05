package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/efritz/gostgres/internal/engine"
	"github.com/efritz/gostgres/internal/functions"
	"github.com/efritz/gostgres/internal/sample"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/table"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

const rootDir = "queries"

func TestIntegration(t *testing.T) {
	entries, err := os.ReadDir(rootDir)
	require.NoError(t, err)

	for _, entry := range entries {
		name := entry.Name()

		t.Run(name, func(t *testing.T) {
			query, err := os.ReadFile(filepath.Join(rootDir, name))
			require.NoError(t, err)

			got, err := runTestQuery(string(query))
			require.NoError(t, err)
			autogold.ExpectFile(t, got, autogold.Dir("golden"))
		})
	}
}

func runTestQuery(input string) (string, error) {
	tables := table.NewTablespace()
	functions := functions.NewFunctionspace()
	functions.SetFunction("now", func(args []any) (any, error) { return time.Now(), nil })
	engine := engine.NewEngine(tables, functions)

	if err := sample.LoadPagilaSampleSchemaAndData(engine); err != nil {
		return "", err
	}

	planRows, err := engine.Query(fmt.Sprintf("EXPLAIN %s", input))
	if err != nil {
		return "", err
	}

	resultRows, err := engine.Query(input)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"\nQuery:\n\n%v\n\nPlan:\n\n%v\nResults:\n\n%v",
		strings.TrimSpace(input),
		serialization.SerializeRowsString(planRows),
		serialization.SerializeRowsString(resultRows),
	), nil
}
