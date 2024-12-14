package join

import "math/bits"

func bit(i uint) uint           { return 1 << i }                    // {i}
func all(n uint) uint           { return (1 << n) - 1 }              // {0, 1, ..., n-1}
func complement(n, m uint) uint { return difference(all(n), m) }     // ¬m
func set(m, pos uint) uint      { return union(m, bit(pos)) }        // m ∪ {i}
func has(m, i uint) bool        { return intersect(m, bit(i)) != 0 } // {i} ∈ mask
func union(a, b uint) uint      { return a | b }                     // a ∪ b
func intersect(a, b uint) uint  { return a & b }                     // a ∩ b
func difference(a, b uint) uint { return a & ^b }                    // a \ b
func isSubset(a, b uint) bool   { return intersect(a, b) == a }      // a ⊆ b
func overlaps(a, b uint) bool   { return intersect(a, b) != 0 }      // a ∩ b ≠ ∅

func coalesce(a, b uint) uint {
	if a != 0 {
		return a
	}

	return b
}

// generateSubsetMasks generates all possible non-empty masks of n bits.
//
// The output is ordered by the number of bits set in the mask. This can be
// useful for dyanmic programming algorithms, where larger subsets can make
// use of the cached results of already-processed smaller subsets.
func generateSubsetMasks(n uint) []uint {
	var results []uint
	for popcount := uint(1); popcount <= n; popcount++ {
		results = generateSubsetMasksWithKBits(results, n, popcount, 0, 0)
	}

	return results
}

func generateSubsetMasksWithKBits(results []uint, n, k, pos, num uint) []uint {
	popcount := uint(bits.OnesCount(num))

	if popcount == k {
		// Found a mask with k bits set; add it to the list.
		// Stop searching here as we can't set any additional bits
		// without overshooting our target popcount.
		return append(results, num)
	}

	if pos >= n {
		// We ran off the edge of the mask; no more bits to set.
		return results
	}

	if k-popcount > n-pos {
		// We don't have enough bits left to reach our target popcount.
		return results
	}

	// Two recursive calls:
	// (1) Set the current bit and generate combinations of upper bits.
	// (2) Keep the current bit unset and generate combinations of upper bits.
	results = generateSubsetMasksWithKBits(results, n, k, pos+1, set(num, pos))
	results = generateSubsetMasksWithKBits(results, n, k, pos+1, num)
	return results
}

// generateSubsetMasksMatchingPattern generates all possible non-empty subsets
// of set bits in the given mask.
//
// This function is equivalent to (but not as inefficient as) generating all
// possible subset masks, applying the given mask to each result, and removing
// duplicates.

func generateSubsetMasksMatchingPattern(m uint) []uint {
	n := uint(bits.OnesCount(m))
	l := uint(bits.Len(m))

	var bits []uint
	for i := uint(0); i < l; i++ {
		if has(m, i) {
			bits = append(bits, i)
		}
	}

	var results []uint
	for _, mask := range generateSubsetMasks(n) {
		result := uint(0)

		// Each bit in the subset mask will correspond to one of the bits set
		// in the input mask. If the i-th bit is set in the subset mask, then
		// bits[i]-th bit will be set in the result.
		for i, j := range bits {
			if has(mask, uint(i)) {
				result = set(result, j)
			}
		}

		results = append(results, result)
	}

	return results
}
