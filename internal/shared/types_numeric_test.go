package shared

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefineNumericType(t *testing.T) {
	type testCase struct {
		name      string
		typ       Type
		input     any
		ok        bool
		converted any
	}
	passing := func(name string, typ Type, input, converted any) testCase {
		return testCase{name: name, typ: typ, input: input, ok: true, converted: converted}
	}
	failing := func(name string, typ Type, input any) testCase {
		return testCase{name: name, typ: typ, input: input, ok: false}
	}

	var (
		one             = new(big.Float).SetUint64(1)
		max16PlusOne    = new(big.Float).Add(new(big.Float).SetUint64(math.MaxInt16), one)
		max32PlusOne    = new(big.Float).Add(new(big.Float).SetUint64(math.MaxInt32), one)
		huge, _         = new(big.Float).SetString("1e25")
		floatAfterMax32 = math.Nextafter(math.MaxFloat32, math.MaxFloat64)
	)

	for _, testCase := range []testCase{
		passing("smallint -> smallint", TypeSmallInteger, int16(200), int16(200)),
		passing("smallint -> integer", TypeInteger, int16(200), int32(200)),
		passing("smallint -> bigint", TypeBigInteger, int16(200), int64(200)),
		passing("smallint -> real", TypeReal, int16(200), float32(200)),
		passing("smallint -> double precision", TypeDoublePrecision, int16(200), float64(200)),
		passing("smallint -> numeric", TypeNumeric, int16(200), big.NewFloat(200)),
		passing("integer -> smallint", TypeSmallInteger, int32(200), int16(200)),
		passing("integer -> integer", TypeInteger, int32(200), int32(200)),
		passing("integer -> bigint", TypeBigInteger, int32(200), int64(200)),
		passing("integer -> real", TypeReal, int32(200), float32(200)),
		passing("integer -> double precision", TypeDoublePrecision, int32(200), float64(200)),
		passing("integer -> numeric", TypeNumeric, int32(200), big.NewFloat(200)),
		passing("bigint -> smallint", TypeSmallInteger, int64(200), int16(200)),
		passing("bigint -> integer", TypeInteger, int64(200), int32(200)),
		passing("bigint -> bigint", TypeBigInteger, int64(200), int64(200)),
		passing("bigint -> real", TypeReal, int64(200), float32(200)),
		passing("bigint -> double precision", TypeDoublePrecision, int64(200), float64(200)),
		passing("bigint -> numeric", TypeNumeric, int64(200), big.NewFloat(200)),
		passing("real -> smallint", TypeSmallInteger, float32(200), int16(200)),
		passing("real -> integer", TypeInteger, float32(200), int32(200)),
		passing("real -> bigint", TypeBigInteger, float32(200), int64(200)),
		passing("real -> real", TypeReal, float32(200), float32(200)),
		passing("real -> double precision", TypeDoublePrecision, float32(200), float64(200)),
		passing("real -> numeric", TypeNumeric, float32(200), big.NewFloat(200)),
		passing("double precision -> smallint", TypeSmallInteger, float64(200), int16(200)),
		passing("double precision -> integer", TypeInteger, float64(200), int32(200)),
		passing("double precision -> bigint", TypeBigInteger, float64(200), int64(200)),
		passing("double precision -> real", TypeReal, float64(200), float32(200)),
		passing("double precision -> double precision", TypeDoublePrecision, float64(200), float64(200)),
		passing("double precision -> numeric", TypeNumeric, float64(200), big.NewFloat(200)),
		passing("numeric -> smallint", TypeSmallInteger, big.NewFloat(200), int16(200)),
		passing("numeric -> integer", TypeInteger, big.NewFloat(200), int32(200)),
		passing("numeric -> bigint", TypeBigInteger, big.NewFloat(200), int64(200)),
		passing("numeric -> real", TypeReal, big.NewFloat(200), float32(200)),
		passing("numeric -> double precision", TypeDoublePrecision, big.NewFloat(200), float64(200)),
		passing("numeric -> numeric", TypeNumeric, big.NewFloat(200), big.NewFloat(200)),
		failing("(large) integer -> smallint", TypeSmallInteger, int32(math.MaxInt16+1)),
		failing("(large) bigint -> smallint", TypeSmallInteger, int64(math.MaxInt16+1)),
		failing("(large) bigint -> integer", TypeInteger, int64(math.MaxInt32+1)),
		failing("(large) real -> smallint", TypeSmallInteger, float32(math.MaxInt16)+1),
		failing("(large) real -> integer", TypeInteger, float32(math.MaxInt32)+1),
		failing("(large) double precision -> smallint", TypeSmallInteger, float64(math.MaxInt16)+1),
		failing("(large) double precision -> integer", TypeInteger, float64(math.MaxInt32)+1),
		failing("(large) double precision -> real", TypeReal, floatAfterMax32),
		failing("(large) numeric -> smallint", TypeSmallInteger, max16PlusOne),
		failing("(large) numeric -> integer", TypeInteger, max32PlusOne),
		failing("(large) numeric -> bigint", TypeBigInteger, huge),
		failing("(large) numeric -> real", TypeReal, big.NewFloat(floatAfterMax32)),
		failing("(large) numeric -> double precision", TypeDoublePrecision, huge),
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if _, value, ok := testCase.typ.Refine(testCase.input); testCase.ok {
				require.True(t, ok)
				assert.Equal(t, testCase.converted, value)
			} else {
				assert.False(t, ok)
			}
		})
	}
}

func TestNumericPromotion(t *testing.T) {
	t.Skip()
	type pair struct {
		left  any
		right any
	}
	type testCase struct {
		name string
		from pair
		to   pair
	}
	test := func(name string, left, right, leftPromoted any) testCase {
		return testCase{
			name: name,
			from: pair{left, right},
			to:   pair{leftPromoted, right},
		}
	}

	testCases := []testCase{
		test("smallint/integer", int16(1), int32(2), int32(1)),
		test("smallint/bigint", int16(1), int64(2), int64(1)),
		test("smallint/real", int16(1), float32(2), float32(1)),
		test("smallint/double precision", int16(1), float64(2), float64(1)),
		test("smallint/numeric", int16(1), big.NewFloat(2), big.NewFloat(1)),
		test("integer/bigint", int32(1), int64(2), int64(1)),
		test("integer/real", int32(1), float32(2), float32(1)),
		test("integer/double precision", int32(1), float64(2), float64(1)),
		test("integer/numeric", int32(1), big.NewFloat(2), big.NewFloat(1)),
		test("bigint/real", int64(1), float32(2), float32(1)),
		test("bigint/double precision", int64(1), float64(2), float64(1)),
		test("bigint/numeric", int64(1), big.NewFloat(2), big.NewFloat(1)),
		test("real/numeric", float32(1), big.NewFloat(2), big.NewFloat(1)),
		test("real/double precision", float32(1), float64(2), float64(1)),
		test("double precision/numeric", float64(1), big.NewFloat(2), big.NewFloat(1)),
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			left, right, err := PromoteToCommonNumericValues(testCase.from.left, testCase.from.right)
			require.NoError(t, err)
			assert.Equal(t, testCase.to.left, left)
			assert.Equal(t, testCase.to.right, right)
		})
	}

	for _, testCase := range testCases {
		t.Run(testCase.name+"-symmetric", func(t *testing.T) {
			right, left, err := PromoteToCommonNumericValues(testCase.from.right, testCase.from.left)
			require.NoError(t, err)
			assert.Equal(t, testCase.to.right, right)
			assert.Equal(t, testCase.to.left, left)
		})
	}
}
