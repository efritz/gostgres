package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/lexing"
	"github.com/efritz/gostgres/internal/syntax/parsing"
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

func runTestQuery(query string) (string, error) {
	tables, err := CreateStandardTestTables("")
	if err != nil {
		return "", err
	}

	node, err := parsing.Parse(lexing.Lex(query), tables)
	if err != nil {
		return "", fmt.Errorf("failed to parse node: %s", err)
	}
	node.Optimize()

	scanner, err := node.Scanner(scan.ScanContext{
		Tables: tables,
	})
	if err != nil {
		return "", err
	}
	rows, err := shared.NewRows(node.Fields())
	if err != nil {
		return "", err
	}
	rows, err = scan.ScanIntoRows(scanner, rows)
	if err != nil {
		return "", fmt.Errorf("failed to execute query: %s", err)
	}

	return fmt.Sprintf(
		"\nQuery:\n\n%v\n\nPlan:\n\n%v\nResults:\n\n%v",
		strings.TrimSpace(query),
		serialization.SerializePlanString(node),
		serialization.SerializeRowsString(rows),
	), nil
}
