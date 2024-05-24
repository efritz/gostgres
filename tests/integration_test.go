package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/efritz/gostgres/internal/execution/engine"
	"github.com/efritz/gostgres/internal/sample"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/syntax/parsing"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

const rootDir = "queries"

func TestIntegration(t *testing.T) {
	engine := engine.NewDefaultEngine()
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
	statements := parsing.SplitStatements(input)
	if len(statements) != 1 {
		return "", fmt.Errorf("expected exactly one statement")
	}

	planRows, err := engine.Query(fmt.Sprintf("EXPLAIN %s", statements[0]))
	if err != nil {
		return "", err
	}

	resultRows, err := engine.Query(statements[0])
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"\nQuery:\n\n%v\n\nPlan:\n\n%v\nResults:\n\n%v",
		strings.TrimSpace(input),
		serialization.SerializeRows(planRows),
		serialization.SerializeRows(resultRows),
	), nil
}
