package tests

import (
	"fmt"

	"github.com/efritz/gostgres/internal/nodes"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/syntax/lexing"
	"github.com/efritz/gostgres/internal/syntax/parsing"
)

func testQuery(query string) (string, error) {
	tables, err := CreateStandardTestTables("")
	if err != nil {
		return "", err
	}

	node, err := parsing.Parse(lexing.Lex(query), tables)
	if err != nil {
		return "", fmt.Errorf("failed to parse node: %s", err)
	}
	node.Optimize()

	rows, err := ScanRows(node, scan.ScanContext{})
	if err != nil {
		return "", fmt.Errorf("failed to execute query: %s", err)
	}

	return fmt.Sprintf(
		"\nPlan:\n\n%v\nResults:\n\n%v",
		serialization.SerializePlanString(node),
		serialization.SerializeRowsString(rows),
	), nil
}

// TODO - deduplicate
func ScanRows(node nodes.Node, ctx scan.ScanContext) (shared.Rows, error) {
	scanner, err := node.Scanner(ctx)
	if err != nil {
		return shared.Rows{}, err
	}

	rows, err := shared.NewRows(node.Fields())
	if err != nil {
		return shared.Rows{}, err
	}

	return scan.ScanIntoRows(scanner, rows)
}
