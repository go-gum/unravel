package unravel

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math"
	"strconv"
	"testing"
	"unsafe"
)

func TestStringSource(t *testing.T) {
	runSourceTests(t, toStringSource, math.MaxUint64)
}

func TestSimpleStringSource(t *testing.T) {
	runSourceTests(t, toSimpleStringSource, math.MaxUint64)
}

func runSourceTests(t *testing.T, toStringSource func(value string) Source, maxUint64 uint64) {
	if unsafe.Sizeof(int(int8(0))) == 8 {
		parseTest(t, toStringSource, stringSourceTestValues[int]{
			MinIn:        "-9223372036854775808",
			MinOut:       -9223372036854775808,
			MaxIn:        "9223372036854775807",
			MaxOut:       9223372036854775807,
			OutOfRange:   []string{"-9223372036854775809", "9223372036854775808"},
			NotSupported: []string{"foobar", "", "1e4"},
		})

		parseTest(t, toStringSource, stringSourceTestValues[uint]{
			MinIn:        "0",
			MinOut:       0,
			MaxIn:        strconv.FormatUint(maxUint64, 10),
			MaxOut:       uint(maxUint64),
			OutOfRange:   []string{"18446744073709551616"},
			NotSupported: []string{"foobar", "", "1e4", "-1"},
		})
	}

	parseTest(t, toStringSource, stringSourceTestValues[int8]{
		MinIn:        "-128",
		MinOut:       -128,
		MaxIn:        "127",
		MaxOut:       127,
		OutOfRange:   []string{"-129", "128"},
		NotSupported: []string{"foobar", "", "1e4"},
	})

	parseTest(t, toStringSource, stringSourceTestValues[int16]{
		MinIn:        "-32768",
		MinOut:       -32768,
		MaxIn:        "32767",
		MaxOut:       32767,
		OutOfRange:   []string{"-32769", "32768"},
		NotSupported: []string{"foobar", "", "1e4"},
	})

	parseTest(t, toStringSource, stringSourceTestValues[int32]{
		MinIn:        "-2147483648",
		MinOut:       -2147483648,
		MaxIn:        "2147483647",
		MaxOut:       2147483647,
		OutOfRange:   []string{"-2147483649", "2147483648"},
		NotSupported: []string{"foobar", "", "1e4"},
	})

	parseTest(t, toStringSource, stringSourceTestValues[int64]{
		MinIn:        "-9223372036854775808",
		MinOut:       -9223372036854775808,
		MaxIn:        "9223372036854775807",
		MaxOut:       9223372036854775807,
		OutOfRange:   []string{"-9223372036854775809", "9223372036854775808"},
		NotSupported: []string{"foobar", "", "1e4"},
	})

	parseTest(t, toStringSource, stringSourceTestValues[uint8]{
		MinIn:        "0",
		MinOut:       0,
		MaxIn:        "255",
		MaxOut:       255,
		OutOfRange:   []string{"256"},
		NotSupported: []string{"foobar", "", "1e4", "-1"},
	})

	parseTest(t, toStringSource, stringSourceTestValues[uint16]{
		MinIn:        "0",
		MinOut:       0,
		MaxIn:        "65535",
		MaxOut:       65535,
		OutOfRange:   []string{"65536"},
		NotSupported: []string{"foobar", "", "1e4", "-1"},
	})

	parseTest(t, toStringSource, stringSourceTestValues[uint32]{
		MinIn:        "0",
		MinOut:       0,
		MaxIn:        "4294967295",
		MaxOut:       4294967295,
		OutOfRange:   []string{"4294967296"},
		NotSupported: []string{"foobar", "", "1e4", "-1"},
	})

	parseTest(t, toStringSource, stringSourceTestValues[uint64]{
		MinIn:        "0",
		MinOut:       0,
		MaxIn:        strconv.FormatUint(maxUint64, 10),
		MaxOut:       maxUint64,
		OutOfRange:   []string{"18446744073709551616"},
		NotSupported: []string{"foobar", "", "1e4", "-1"},
	})

	parseTest(t, toStringSource, stringSourceTestValues[bool]{
		MinIn:        "true",
		MinOut:       true,
		MaxIn:        "false",
		MaxOut:       false,
		NotSupported: []string{"foobar", "", "1e4", "-1"},
	})

	parseTest(t, toStringSource, stringSourceTestValues[float64]{
		MinIn:        "-1234.5",
		MinOut:       -1234.5,
		MaxIn:        "1235.5",
		MaxOut:       1235.5,
		Valid:        []string{"1e4", "-1", "0.0024"},
		NotSupported: []string{"foobar", ""},
	})
}

type stringSourceTestValues[T any] struct {
	MinIn  string
	MinOut T

	MaxIn  string
	MaxOut T

	OutOfRange   []string
	NotSupported []string
	Valid        []string
}

func parseTest[T any](t *testing.T, toSource func(string) Source, v stringSourceTestValues[T]) {
	var tZero T

	t.Run(fmt.Sprintf("parse to %T", tZero), func(t *testing.T) {
		actual, err := UnmarshalNew[T](toSource(v.MinIn))
		require.NoError(t, err)
		require.Equal(t, actual, v.MinOut)

		actual, err = UnmarshalNew[T](toSource(v.MaxIn))
		require.NoError(t, err)
		require.Equal(t, actual, v.MaxOut)

		for _, value := range v.OutOfRange {
			actual, err = UnmarshalNew[T](toSource(value))
			require.ErrorIs(t, err, strconv.ErrRange)
			require.Equal(t, actual, tZero)
		}

		for _, value := range v.NotSupported {
			actual, err = UnmarshalNew[T](toSource(value))
			require.ErrorIs(t, err, ErrNotSupported)
			require.Equal(t, actual, tZero)
		}

		for _, value := range v.Valid {
			actual, err = UnmarshalNew[T](toSource(value))
			require.NoError(t, err)
		}
	})
}

func toStringSource(value string) Source {
	return StringSource(value)
}

func toSimpleStringSource(value string) Source {
	return simpleStringSource{Value: value}
}

type simpleStringSource struct {
	EmptySource
	Value string
}

func (s simpleStringSource) Bool() (bool, error) {
	return StringSource(s.Value).Bool()
}

func (s simpleStringSource) Float() (float64, error) {
	return StringSource(s.Value).Float()
}

func (s simpleStringSource) Int() (int64, error) {
	return StringSource(s.Value).Int()
}

func (s simpleStringSource) Uint() (uint64, error) {
	return StringSource(s.Value).Uint()
}
