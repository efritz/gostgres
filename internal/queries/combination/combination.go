package combination

import (
	"fmt"
	"io"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/nodes"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

type combinationNode struct {
	left             nodes.Node
	right            nodes.Node
	fields           []shared.Field
	groupedRowFilter groupedRowFilterFunc
	distinct         bool
}

var _ nodes.Node = &combinationNode{}

type sourcedRow struct {
	index int
	row   shared.Row
}

type groupedRowFilterFunc func(rows []sourcedRow) bool

func NewIntersect(left nodes.Node, right nodes.Node, distinct bool) (nodes.Node, error) {
	return newCombination(left, right, intersectFilter, distinct)
}

func NewExcept(left nodes.Node, right nodes.Node, distinct bool) (nodes.Node, error) {
	return newCombination(left, right, exceptFilter, distinct)
}

func newCombination(left nodes.Node, right nodes.Node, groupedRowFilter groupedRowFilterFunc, distinct bool) (nodes.Node, error) {
	leftFields := left.Fields()
	rightFields := right.Fields()

	if len(leftFields) != len(rightFields) {
		return nil, fmt.Errorf("unexpected combination columns")
	}
	for i, leftField := range leftFields {
		if leftField.TypeKind != rightFields[i].TypeKind {
			// TODO - refine type if possible
			return nil, fmt.Errorf("unexpected combination types")
		}
	}

	return &combinationNode{
		left:             left,
		right:            right,
		fields:           leftFields,
		groupedRowFilter: groupedRowFilter,
		distinct:         distinct,
	}, nil
}

func (n *combinationNode) Name() string {
	return ""
}

func (n *combinationNode) Fields() []shared.Field {
	return copyFields(n.fields)
}

func (n *combinationNode) Serialize(w io.Writer, indentationLevel int) {
	indentation := indent(indentationLevel)
	io.WriteString(w, fmt.Sprintf("%scombination\n", indentation))
	n.left.Serialize(w, indentationLevel+1)
	io.WriteString(w, fmt.Sprintf("%swith\n", indentation))
	n.right.Serialize(w, indentationLevel+1)
}

func (n *combinationNode) Optimize() {
	n.left.Optimize()
	n.right.Optimize()
}

func (n *combinationNode) AddFilter(filter expressions.Expression) {
	lowerFilter(filter, n.left, n.right)
}

func (n *combinationNode) AddOrder(order expressions.OrderExpression) {
	lowerOrder(order, n.left, n.right)
}

func (n *combinationNode) Filter() expressions.Expression {
	return n.left.Filter()
}

func (n *combinationNode) Ordering() expressions.OrderExpression {
	return nil
}

func (n *combinationNode) SupportsMarkRestore() bool {
	return false
}

func (n *combinationNode) Scanner(ctx scan.ScanContext) (scan.Scanner, error) {
	leftScanner, err := n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}
	rightScanner, err := n.right.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	hash := map[string][]sourcedRow{}
	if err := scan.VisitRows(leftScanner, hashVisitor(hash, 0)); err != nil {
		return nil, err
	}
	if err := scan.VisitRows(rightScanner, hashVisitor(hash, 1)); err != nil {
		return nil, err
	}

	var selection []sourcedRow

	return scan.ScannerFunc(func() (shared.Row, error) {
	outer:
		for {
			if len(selection) > 0 {
				row := selection[0]
				selection = selection[1:]
				return shared.NewRow(n.Fields(), row.row.Values)
			}

			for key, rows := range hash {
				if n.groupedRowFilter(rows) {
					if n.distinct {
						selection = rows[:1]
					} else {
						selection = rows
					}
				}

				delete(hash, key)
				continue outer
			}

			break
		}

		return shared.Row{}, scan.ErrNoRows
	}), nil
}

func hashVisitor(hash map[string][]sourcedRow, index int) scan.VisitorFunc {
	return func(row shared.Row) (bool, error) {
		key := hashValues(row.Values)

		hash[key] = append(hash[key], sourcedRow{
			index: index,
			row:   row,
		})

		return true, nil
	}
}

func intersectFilter(rows []sourcedRow) bool {
	indexes := map[int]bool{}
	for _, r := range rows {
		indexes[r.index] = true
	}

	if !indexes[0] || !indexes[1] {
		return false
	}

	return true
}

func exceptFilter(rows []sourcedRow) bool {
	for _, r := range rows {
		if r.index == 1 {
			return false
		}
	}

	return true
}

// TODO - deduplicate

func copyFields(fields []shared.Field) []shared.Field {
	c := make([]shared.Field, len(fields))
	copy(c, fields)
	return c
}

func lowerFilter(filter expressions.Expression, nodes ...nodes.Node) {
	for _, expression := range filter.Conjunctions() {
		missing := make([]bool, len(nodes))
		for _, field := range expression.Fields() {
			for i, node := range nodes {
				if _, err := shared.FindMatchingFieldIndex(field, node.Fields()); err != nil {
					missing[i] = true
				}
			}
		}

		for i, missing := range missing {
			if !missing {
				nodes[i].AddFilter(expression)
			}
		}
	}
}

func lowerOrder(order expressions.OrderExpression, nodes ...nodes.Node) {
	orderExpressions := order.Expressions()

	for _, node := range nodes {
		filteredExpressions := make([]expressions.ExpressionWithDirection, 0, len(orderExpressions))
	exprLoop:
		for _, expression := range orderExpressions {
			for _, field := range expression.Expression.Fields() {
				if _, err := shared.FindMatchingFieldIndex(field, node.Fields()); err != nil {
					continue exprLoop
				}
			}

			filteredExpressions = append(filteredExpressions, expression)
		}

		if len(filteredExpressions) != 0 {
			node.AddOrder(expressions.NewOrderExpression(filteredExpressions))
		}
	}
}

func indent(level int) string {
	return strings.Repeat(" ", level*4)
}

func hashValues(values []any) string {
	strValues := make([]string, 0, len(values))
	for _, value := range values {
		strValues = append(strValues, fmt.Sprintf("%v", value))
	}

	return strings.Join(strValues, ":")
}
