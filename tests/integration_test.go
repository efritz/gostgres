package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/efritz/gostgres/internal/execution/engine"
	"github.com/efritz/gostgres/internal/execution/protocol"
	"github.com/efritz/gostgres/internal/execution/serialization"
	"github.com/efritz/gostgres/internal/sample"
	"github.com/efritz/gostgres/internal/syntax/parsing"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	engine := engine.NewDefaultEngine()
	require.NoError(t, sample.LoadPagilaSampleSchemaAndData(engine))

	testCases, err := gatherTestCases(t)
	require.NoError(t, err)

	for _, testCase := range testCases {
		t.Run(testCase.relPath, func(t *testing.T) {
			got, err := runTestQuery(engine, testCase.contents)
			require.NoError(t, err)
			autogold.ExpectFile(t, got, autogold.Dir("golden"))
		})
	}
}

type testCase struct {
	relPath  string
	contents string
}

const rootDir = "queries"

func gatherTestCases(t *testing.T) (cases []testCase, _ error) {
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".sql") {
			return fmt.Errorf("unexpected file %s", path)
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		rawContents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		contents := string(rawContents)

		if strings.HasPrefix(contents, "-- SKIP\n") {
			t.Logf("Skipping %s", relPath)
			return nil
		}

		cases = append(cases, testCase{
			relPath:  relPath,
			contents: string(contents),
		})

		return nil
	})

	return cases, err
}

func runTestQuery(engine *engine.Engine, input string) (string, error) {
	statements := parsing.SplitStatements(input)
	if len(statements) != 1 {
		return "", fmt.Errorf("expected exactly one statement")
	}

	planRows, err := engine.QueryRows(protocol.Request{
		Query: fmt.Sprintf("EXPLAIN %s", statements[0]),
		Debug: false,
	})
	if err != nil {
		return "", err
	}

	resultRows, err := engine.QueryRows(protocol.Request{
		Query: statements[0],
		Debug: false,
	})
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
