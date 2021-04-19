package nodes

import (
	"fmt"
	"io"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type combinationNode struct {
	left             Node
	right            Node
	fields           []shared.Field
	groupedRowFilter groupedRowFilterFunc
	distinct         bool
}

var _ Node = &combinationNode{}

type sourcedRow struct {
	index int
	row   shared.Row
}

type groupedRowFilterFunc func(rows []sourcedRow) bool

func NewIntersect(left Node, right Node, distinct bool) (Node, error) {
	return newCombination(left, right, intersectFilter, distinct)
}
func NewExcept(left Node, right Node, distinct bool) (Node, error) {
	return newCombination(left, right, exceptFilter, distinct)
}

func newCombination(left Node, right Node, groupedRowFilter groupedRowFilterFunc, distinct bool) (Node, error) {
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

func (n *combinationNode) PushDownFilter(filter expressions.Expression) bool {
	// TODO - only push down if they're valid (see join)
	// pushedLeft := n.left.PushDownFilter(filter)
	// pushedRight := n.right.PushDownFilter(filter)
	// return pushedLeft && pushedRight
	return false
}

func (n *combinationNode) Scan(visitor VisitorFunc) error {
	hash := map[string][]sourcedRow{}
	if err := n.left.Scan(hashVisitor(hash, 0)); err != nil {
		return err
	}
	if err := n.right.Scan(hashVisitor(hash, 1)); err != nil {
		return err
	}

outer:
	for _, rows := range hash {
		if !n.groupedRowFilter(rows) {
			continue outer
		}

		for _, row := range rows {
			row, err := shared.NewRow(n.Fields(), row.row.Values)
			if err != nil {
				return err
			}

			ok, err := visitor(row)
			if err != nil {
				return err
			}
			if !ok {
				break outer
			}

			if n.distinct {
				break
			}
		}
	}

	return nil
}

func hashVisitor(hash map[string][]sourcedRow, index int) VisitorFunc {
	return func(row shared.Row) (bool, error) {
		key := hashValues(row.Values)

		hash[key] = append(hash[key], sourcedRow{
			index: index,
			row:   row,
		})

		return true, nil
	}
}

func hashValues(values []interface{}) string {
	strValues := make([]string, 0, len(values))
	for _, value := range values {
		strValues = append(strValues, fmt.Sprintf("%v", value))
	}

	return strings.Join(strValues, ":")
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
