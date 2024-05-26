package indexes

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/ordering"
	"github.com/efritz/gostgres/internal/shared/rows"
)

func (i *btreeIndex) Scanner(ctx impls.Context, opts BtreeIndexScanOptions) (impls.TIDScanner, error) {
	ctx.Log("Building BTree Index scanner")

	stack := []*btreeNode{}
	current := i.root

	lowerBounds, err := resolveScanBounds(ctx, opts.lowerBounds)
	if err != nil {
		return nil, err
	}

	upperBounds, err := resolveScanBounds(ctx, opts.upperBounds)
	if err != nil {
		return nil, err
	}

	return tidScannerFunc(func() (int64, error) {
		ctx.Log("Scanning BTree Index")

		for current != nil || len(stack) > 0 {
			for current != nil {
				stack = append(stack, current)

				if opts.scanDirection == ScanDirectionForward && checkBounds(current.values, lowerBounds, ordering.OrderTypeAfter) {
					current = current.left
				} else if opts.scanDirection == ScanDirectionBackward && checkBounds(current.values, upperBounds, ordering.OrderTypeBefore) {
					current = current.right
				} else {
					current = nil
				}
			}

			if len(stack) == 0 {
				break
			}

			idx := len(stack) - 1
			node := stack[idx]
			stack = stack[:idx]

			lowerOk := checkBounds(node.values, lowerBounds, ordering.OrderTypeAfter)
			upperOk := checkBounds(node.values, upperBounds, ordering.OrderTypeBefore)

			if opts.scanDirection == ScanDirectionForward && upperOk {
				current = node.right
			} else if opts.scanDirection == ScanDirectionBackward && lowerOk {
				current = node.left
			} else {
				current = nil
			}

			if upperOk && lowerOk {
				return node.tid, nil
			}
		}

		return 0, scan.ErrNoRows
	}), nil
}

type resolvedScanBound struct {
	value     any
	inclusive bool
}

func resolveScanBounds(ctx impls.Context, scanBounds [][]scanBound) ([][]resolvedScanBound, error) {
	var resolvedScanBounds [][]resolvedScanBound
	for _, bounds := range scanBounds {
		resolvedBounds := []resolvedScanBound{}
		for _, bound := range bounds {
			value, err := queries.Evaluate(ctx, bound.expression, rows.Row{})
			if err != nil {
				return nil, err
			}

			resolvedBounds = append(resolvedBounds, resolvedScanBound{
				value:     value,
				inclusive: bound.inclusive,
			})
		}

		resolvedScanBounds = append(resolvedScanBounds, resolvedBounds)
	}

	return resolvedScanBounds, nil
}

func checkBounds(nodeValues []any, bounds [][]resolvedScanBound, expected ordering.OrderType) bool {
	for j, bounds := range bounds[:min(len(bounds), len(nodeValues))] {
		for _, bound := range bounds {
			orderType := ordering.CompareValues(nodeValues[j], bound.value)
			if !(orderType == expected || (orderType == ordering.OrderTypeEqual && bound.inclusive)) {
				return false
			}
		}
	}

	return true
}
