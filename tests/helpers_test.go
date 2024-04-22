package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/efritz/gostgres/internal/engine"
	"github.com/efritz/gostgres/internal/sample"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

type TestCase struct {
	name  string
	query string
}

func runTests(t *testing.T, testCases []TestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := runTestQuery(testCase.query)
			require.NoError(t, err)
			autogold.ExpectFile(t, got, autogold.Dir("golden"))
		})
	}
}

func runTestQuery(input string) (string, error) {
	tables, err := sample.CreateSampleTables("")
	if err != nil {
		return "", err
	}

	engine := engine.NewEngine(tables)

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
