package tests

import (
	"fmt"
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

type TestCase struct {
	name  string
	query string
}

func runTests(t *testing.T, testCases []TestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Skip()
			got, err := runTestQuery(testCase.query)
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
