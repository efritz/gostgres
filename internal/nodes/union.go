package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

type unionNode struct {
	left     Node
	right    Node
	fields   []shared.Field
	distinct bool
}

var _ Node = &unionNode{}

func NewUnion(left Node, right Node, distinct bool) (Node, error) {
	leftFields := left.Fields()
	rightFields := right.Fields()

	if len(leftFields) != len(rightFields) {
		return nil, fmt.Errorf("unexpected union columns")
	}
	for i, leftField := range leftFields {
		if leftField.TypeKind != rightFields[i].TypeKind {
			// TODO - refine type if possible
			return nil, fmt.Errorf("unexpected union types")
		}
	}

	return &unionNode{
		left:     left,
		right:    right,
		fields:   leftFields,
		distinct: distinct,
	}, nil
}

func (n *unionNode) Name() string {
	return ""
}

func (n *unionNode) Fields() []shared.Field {
	return copyFields(n.fields)
}

func (n *unionNode) Serialize(w io.Writer, indentationLevel int) {
	indentation := indent(indentationLevel)
	io.WriteString(w, fmt.Sprintf("%sunion\n", indentation))
	n.left.Serialize(w, indentationLevel+1)
	io.WriteString(w, fmt.Sprintf("%swith\n", indentation))
	n.right.Serialize(w, indentationLevel+1)
}

func (n *unionNode) Optimize() {
	n.left.Optimize()
	n.right.Optimize()
}

func (n *unionNode) AddFilter(filter expressions.Expression) {
	lowerFilter(filter, n.left, n.right)
}

func (n *unionNode) AddOrder(order expressions.OrderExpression) {
	lowerOrder(order, n.left, n.right)
}

func (n *unionNode) Filter() expressions.Expression {
	return filterIntersection(n.left.Filter(), n.right.Filter())
}

func (n *unionNode) Ordering() expressions.OrderExpression {
	return nil
}

func (n *unionNode) SupportsMarkRestore() bool {
	return false
}

func (n *unionNode) Scanner(ctx scan.ScanContext) (scan.Scanner, error) {
	hash := map[string]struct{}{}
	mark := func(row shared.Row) bool {
		key := hashValues(row.Values)
		if _, ok := hash[key]; ok {
			return !n.distinct
		}

		hash[key] = struct{}{}
		return true
	}

	leftScanner, err := n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var rightScanner scan.Scanner

	return scan.ScannerFunc(func() (shared.Row, error) {
		for leftScanner != nil {
			row, err := leftScanner.Scan()
			if err != nil {
				if err == scan.ErrNoRows {
					leftScanner = nil
					continue
				}

				return shared.Row{}, err
			}

			if mark(row) {
				return row, nil
			}
		}

		if rightScanner == nil {
			rightScanner, err = n.right.Scanner(ctx)
			if err != nil {
				return shared.Row{}, err
			}
		}

		for {
			row, err := rightScanner.Scan()
			if err != nil {
				return shared.Row{}, err
			}

			if mark(row) {
				return shared.NewRow(n.Fields(), row.Values)
			}
		}
	}), nil
}
