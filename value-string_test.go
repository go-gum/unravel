package serde

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math"
	"strconv"
	"testing"
	"unsafe"
)

func TestStringValue(t *testing.T) {
	runSourceValueTests(t, toStringSource, math.MaxUint64)
}

func TestSimpleStringValue(t *testing.T) {
	runSourceValueTests(t, toSimpleStringValue, math.MaxUint64)
}

func runSourceValueTests(t *testing.T, toStringSource func(value string) SourceValue, maxUint64 uint64) {
	if unsafe.Sizeof(int(int8(0))) == 8 {
		parseTest(t, toStringSource, stringValueTestValues[int]{
			MinIn:        "-9223372036854775808",
			MinOut:       -9223372036854775808,
			MaxIn:        "9223372036854775807",
			MaxOut:       9223372036854775807,
			OutOfRange:   []string{"-9223372036854775809", "9223372036854775808"},
			NotSupported: []string{"foobar", "", "1e4"},
		})

		parseTest(t, toStringSource, stringValueTestValues[uint]{
			MinIn:        "0",
			MinOut:       0,
			MaxIn:        strconv.FormatUint(maxUint64, 10),
			MaxOut:       uint(maxUint64),
			OutOfRange:   []string{"18446744073709551616"},
			NotSupported: []string{"foobar", "", "1e4", "-1"},
		})
	}

	parseTest(t, toStringSource, stringValueTestValues[int8]{
		MinIn:        "-128",
		MinOut:       -128,
		MaxIn:        "127",
		MaxOut:       127,
		OutOfRange:   []string{"-129", "128"},
		NotSupported: []string{"foobar", "", "1e4"},
	})

	parseTest(t, toStringSource, stringValueTestValues[int16]{
		MinIn:        "-32768",
		MinOut:       -32768,
		MaxIn:        "32767",
		MaxOut:       32767,
		OutOfRange:   []string{"-32769", "32768"},
		NotSupported: []string{"foobar", "", "1e4"},
	})

	parseTest(t, toStringSource, stringValueTestValues[int32]{
		MinIn:        "-2147483648",
		MinOut:       -2147483648,
		MaxIn:        "2147483647",
		MaxOut:       2147483647,
		OutOfRange:   []string{"-2147483649", "2147483648"},
		NotSupported: []string{"foobar", "", "1e4"},
	})

	parseTest(t, toStringSource, stringValueTestValues[int64]{
		MinIn:        "-9223372036854775808",
		MinOut:       -9223372036854775808,
		MaxIn:        "9223372036854775807",
		MaxOut:       9223372036854775807,
		OutOfRange:   []string{"-9223372036854775809", "9223372036854775808"},
		NotSupported: []string{"foobar", "", "1e4"},
	})

	parseTest(t, toStringSource, stringValueTestValues[uint8]{
		MinIn:        "0",
		MinOut:       0,
		MaxIn:        "255",
		MaxOut:       255,
		OutOfRange:   []string{"256"},
		NotSupported: []string{"foobar", "", "1e4", "-1"},
	})

	parseTest(t, toStringSource, stringValueTestValues[uint16]{
		MinIn:        "0",
		MinOut:       0,
		MaxIn:        "65535",
		MaxOut:       65535,
		OutOfRange:   []string{"65536"},
		NotSupported: []string{"foobar", "", "1e4", "-1"},
	})

	parseTest(t, toStringSource, stringValueTestValues[uint32]{
		MinIn:        "0",
		MinOut:       0,
		MaxIn:        "4294967295",
		MaxOut:       4294967295,
		OutOfRange:   []string{"4294967296"},
		NotSupported: []string{"foobar", "", "1e4", "-1"},
	})

	parseTest(t, toStringSource, stringValueTestValues[uint64]{
		MinIn:        "0",
		MinOut:       0,
		MaxIn:        strconv.FormatUint(maxUint64, 10),
		MaxOut:       maxUint64,
		OutOfRange:   []string{"18446744073709551616"},
		NotSupported: []string{"foobar", "", "1e4", "-1"},
	})

	parseTest(t, toStringSource, stringValueTestValues[bool]{
		MinIn:        "true",
		MinOut:       true,
		MaxIn:        "false",
		MaxOut:       false,
		NotSupported: []string{"foobar", "", "1e4", "-1"},
	})

	parseTest(t, toStringSource, stringValueTestValues[float64]{
		MinIn:        "-1234.5",
		MinOut:       -1234.5,
		MaxIn:        "1235.5",
		MaxOut:       1235.5,
		Valid:        []string{"1e4", "-1", "0.0024"},
		NotSupported: []string{"foobar", ""},
	})
}

type stringValueTestValues[T any] struct {
	MinIn  string
	MinOut T

	MaxIn  string
	MaxOut T

	OutOfRange   []string
	NotSupported []string
	Valid        []string
}

func parseTest[T any](t *testing.T, toSource func(string) SourceValue, v stringValueTestValues[T]) {
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

func toStringSource(value string) SourceValue {
	return StringValue(value)
}

func toSimpleStringValue(value string) SourceValue {
	return simpleStringValue{Value: value}
}

type simpleStringValue struct {
	Value string
}

func (s simpleStringValue) Bool() (bool, error) {
	return StringValue(s.Value).Bool()
}

func (s simpleStringValue) Float() (float64, error) {
	return StringValue(s.Value).Float()
}

func (s simpleStringValue) Int() (int64, error) {
	return StringValue(s.Value).Int()
}
func (s simpleStringValue) Uint() (uint64, error) {
	return StringValue(s.Value).Uint()
}

func (s simpleStringValue) String() (string, error) {
	return "", ErrNotSupported
}
