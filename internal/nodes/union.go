package nodes

import (
	"fmt"
	"io"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type unionNode struct {
	left   Node
	right  Node
	fields []shared.Field
}

var _ Node = &unionNode{}

func NewUnion(left Node, right Node) (Node, error) {
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
		left:   left,
		right:  right,
		fields: leftFields,
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

func (n *unionNode) PushDownFilter(filter expressions.Expression) bool {
	// TODO - only push down if they're valid (see join)

	// pushedLeft := n.left.PushDownFilter(filter)
	// pushedRight := n.right.PushDownFilter(filter)
	// return pushedLeft && pushedRight
	return false
}

func (n *unionNode) Scan(visitor VisitorFunc) error {
	if err := n.left.Scan(visitor); err != nil {
		return err
	}

	return n.right.Scan(func(row shared.Row) (bool, error) {
		row, err := shared.NewRow(n.Fields(), row.Values)
		if err != nil {
			return false, err
		}

		return visitor(row)
	})
}
