package filter

import (
	"fmt"
	"io"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

type filterNode struct {
	queries.Node
	filter expressions.Expression
}

var _ queries.Node = &filterNode{}

func NewFilter(node queries.Node, filter expressions.Expression) queries.Node {
	return &filterNode{
		Node:   node,
		filter: filter,
	}
}

func (n *filterNode) Serialize(w io.Writer, indentationLevel int) {
	if n.filter == nil {
		n.Node.Serialize(w, indentationLevel)
		return
	}

	io.WriteString(w, fmt.Sprintf("%sfilter by %s\n", indent(indentationLevel), n.filter))
	n.Node.Serialize(w, indentationLevel+1)
}

func (n *filterNode) Optimize() {
	if n.filter != nil {
		n.filter = n.filter.Fold()
		n.Node.AddFilter(n.filter)
	}

	n.Node.Optimize()
	n.filter = filterDifference(n.filter, n.Node.Filter())
}

func (n *filterNode) AddFilter(filter expressions.Expression) {
	n.filter = unionFilters(n.filter, filter)
}

func (n *filterNode) AddOrder(order expressions.OrderExpression) {
	n.Node.AddOrder(order)
}

func (n *filterNode) Filter() expressions.Expression {
	return unionFilters(n.filter, n.Node.Filter())
}

func (n *filterNode) Ordering() expressions.OrderExpression {
	return n.Node.Ordering()
}

func (n *filterNode) SupportsMarkRestore() bool {
	return false
}

func (n *filterNode) Scanner(ctx scan.ScanContext) (scan.Scanner, error) {
	scanner, err := n.Node.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return NewFilterScanner(ctx, scanner, n.filter)
}

func NewFilterScanner(ctx scan.ScanContext, scanner scan.Scanner, filter expressions.Expression) (scan.Scanner, error) {
	if filter == nil {
		return scanner, nil
	}

	return scan.ScannerFunc(func() (shared.Row, error) {
		for {
			row, err := scanner.Scan()
			if err != nil {
				return shared.Row{}, err
			}

			if ok, err := shared.EnsureBool(ctx.Evaluate(filter, row)); err != nil {
				return shared.Row{}, err
			} else if !ok {
				continue
			}

			return row, nil
		}
	}), nil
}

// TODO - deduplicate

func filterDifference(filter, childFilter expressions.Expression) expressions.Expression {
	return combineFilters(filter, childFilter, func(conjunctions, childConjunctions []expressions.Expression) {
		for i, f1 := range conjunctions {
			for _, f2 := range childConjunctions {
				if f1.Equal(f2) {
					conjunctions[i] = nil
					break
				}
			}
		}
	})
}

func combineFilters(filter, childFilter expressions.Expression, filterConjunctions func(conjunctions, childConjunctions []expressions.Expression)) expressions.Expression {
	if filter == nil {
		return nil
	}
	if childFilter == nil {
		return filter
	}

	conjunctions := filter.Conjunctions()
	filterConjunctions(conjunctions, childFilter.Conjunctions())
	return unionFilters(conjunctions...)
}

func unionFilters(filters ...expressions.Expression) expressions.Expression {
	var conjunctions []expressions.Expression
	for _, expression := range filters {
		if expression == nil {
			continue
		}

		conjunctions = append(conjunctions, expression.Conjunctions()...)
	}
	if len(conjunctions) == 0 {
		return nil
	}

	for i, c1 := range conjunctions {
		for j, c2 := range conjunctions {
			if c1 == nil || c2 == nil || j <= i {
				continue
			}

			if c1.Equal(c2) {
				conjunctions[j] = nil
			}
		}
	}

	filter := conjunctions[0]
	for _, conjunction := range conjunctions[1:] {
		if conjunction == nil {
			continue
		}

		filter = expressions.NewAnd(filter, conjunction)
	}

	return filter
}

func indent(level int) string {
	return strings.Repeat(" ", level*4)
}
