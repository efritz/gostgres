package indexes

import (
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

func (i *btreeIndex) Scanner(ctx types.Context, opts BtreeIndexScanOptions) (types.TIDScanner, error) {
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
		for current != nil || len(stack) > 0 {
			for current != nil {
				stack = append(stack, current)

				if opts.scanDirection == ScanDirectionForward && checkBounds(current.values, lowerBounds, shared.OrderTypeAfter) {
					current = current.left
				} else if opts.scanDirection == ScanDirectionBackward && checkBounds(current.values, upperBounds, shared.OrderTypeBefore) {
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

			lowerOk := checkBounds(node.values, lowerBounds, shared.OrderTypeAfter)
			upperOk := checkBounds(node.values, upperBounds, shared.OrderTypeBefore)

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

func resolveScanBounds(ctx types.Context, scanBounds [][]scanBound) ([][]resolvedScanBound, error) {
	var resolvedScanBounds [][]resolvedScanBound
	for _, bounds := range scanBounds {
		resolvedBounds := []resolvedScanBound{}
		for _, bound := range bounds {
			value, err := queries.Evaluate(ctx, bound.expression, shared.Row{})
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

func checkBounds(nodeValues []any, bounds [][]resolvedScanBound, expected shared.OrderType) bool {
	for j, bounds := range bounds[:min(len(bounds), len(nodeValues))] {
		for _, bound := range bounds {
			orderType := shared.CompareValues(nodeValues[j], bound.value)
			if !(orderType == expected || (orderType == shared.OrderTypeEqual && bound.inclusive)) {
				return false
			}
		}
	}

	return true
}
