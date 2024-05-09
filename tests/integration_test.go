package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/efritz/gostgres/internal/engine"
	"github.com/efritz/gostgres/internal/functions"
	"github.com/efritz/gostgres/internal/sample"
	"github.com/efritz/gostgres/internal/sequence"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/table"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

const rootDir = "queries"

func TestIntegration(t *testing.T) {
	tables := table.NewTablespace()
	sequences := sequence.NewSequencespace()
	functions := functions.NewDefaultFunctionspace()

	engine := engine.NewEngine(tables, sequences, functions)
	require.NoError(t, sample.LoadPagilaSampleSchemaAndData(engine))

	entries, err := os.ReadDir(rootDir)
	require.NoError(t, err)

	for _, entry := range entries {
		t.Run(entry.Name(), func(t *testing.T) {
			query, err := os.ReadFile(filepath.Join(rootDir, entry.Name()))
			require.NoError(t, err)

			got, err := runTestQuery(engine, string(query))
			require.NoError(t, err)
			autogold.ExpectFile(t, got, autogold.Dir("golden"))
		})
	}
}

func runTestQuery(engine *engine.Engine, input string) (string, error) {
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
