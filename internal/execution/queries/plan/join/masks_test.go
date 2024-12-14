package join

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSubsetMasks(t *testing.T) {
	assert.Equal(
		t,
		[]uint{
			0b0001, // popcount = 1
			0b0010,
			0b0100,
			0b1000,
			0b0011, // popcount = 2
			0b0101,
			0b1001,
			0b0110,
			0b1010,
			0b1100,
			0b0111, // popcount = 3
			0b1011,
			0b1101,
			0b1110,
			0b1111, // popcount = 4
		},
		generateSubsetMasks(4),
	)
}

func TestGenerateSubsetMasksMatchingPattern(t *testing.T) {
	assert.Equal(
		t,
		[]uint{
			0b000001, // popcount = 1
			0b000100,
			0b010000,
			0b100000,
			0b000101, // popcount = 2
			0b010001,
			0b100001,
			0b010100,
			0b100100,
			0b110000,
			0b010101, // popcount = 3
			0b100101,
			0b110001,
			0b110100,
			0b110101, // popcount = 4
		},
		generateSubsetMasksMatchingPattern(0b110101),
	)
}
