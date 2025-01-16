package unravel

import (
	"errors"
	"fmt"
	"iter"
	"strconv"
)

// StringSource adapts a `string` to the [Source] interface, providing a straightforward
// implementation for working with textual data. It enables parsing of primitive types from
// the underlying string representation using the following functions from the `strconv` package:
//
//   - `strconv.ParseInt` for parsing integers
//   - `strconv.ParseFloat` for parsing floating-point numbers
//   - `strconv.ParseBool` for parsing boolean values
//
// When requested as a string, the original value is returned as-is without modification.
//
// This implementation is particularly useful for source values where the data is represented
// as strings but needs to be decoded into native Go types. Developers can use `StringSource`
// directly or embed it within custom [Source] implementations to inherit its parsing
// functionality.
//
// Example:
//
//	sv := unravel.StringSource("42")
//	i, err := sv.Int() // Parses "42" into an integer value
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(i) // Outputs: 42
//
//	bsv := unravel.StringSource("true")
//	b, err := bsv.Bool() // Parses "true" into a boolean value
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(b) // Outputs: true
//
// This lightweight implementation makes it easy to handle scenarios where serialized data
// is stored in string form, such as text-based configurations or simple protocols.
type StringSource string

var _ Source = StringSource("")
var _ BinarySource = StringSource("")

func (s StringSource) Int8() (int8, error) {
	intValue, err := strconv.ParseInt(string(s), 10, 8)
	return handleSyntaxErr(string(s), int8(intValue), err)
}

func (s StringSource) Int16() (int16, error) {
	intValue, err := strconv.ParseInt(string(s), 10, 16)
	return handleSyntaxErr(string(s), int16(intValue), err)
}

func (s StringSource) Int32() (int32, error) {
	intValue, err := strconv.ParseInt(string(s), 10, 32)
	return handleSyntaxErr(string(s), int32(intValue), err)
}

func (s StringSource) Int64() (int64, error) {
	intValue, err := strconv.ParseInt(string(s), 10, 64)
	return handleSyntaxErr(string(s), intValue, err)
}

func (s StringSource) Uint8() (uint8, error) {
	intValue, err := strconv.ParseUint(string(s), 10, 8)
	return handleSyntaxErr(string(s), uint8(intValue), err)
}

func (s StringSource) Uint16() (uint16, error) {
	intValue, err := strconv.ParseUint(string(s), 10, 16)
	return handleSyntaxErr(string(s), uint16(intValue), err)
}

func (s StringSource) Uint32() (uint32, error) {
	intValue, err := strconv.ParseUint(string(s), 10, 32)
	return handleSyntaxErr(string(s), uint32(intValue), err)
}

func (s StringSource) Uint64() (uint64, error) {
	intValue, err := strconv.ParseUint(string(s), 10, 64)
	return handleSyntaxErr(string(s), intValue, err)
}

func (s StringSource) Bool() (bool, error) {
	parsedValue, err := strconv.ParseBool(string(s))
	return handleSyntaxErr(string(s), parsedValue, err)
}

func (s StringSource) Int() (int64, error) {
	return s.Int64()
}

func (s StringSource) Uint() (uint64, error) {
	return s.Uint64()
}

func (s StringSource) Float() (float64, error) {
	return s.Float64()
}

func (s StringSource) Float32() (float32, error) {
	parsedValue, err := strconv.ParseFloat(string(s), 32)
	return handleSyntaxErr(string(s), float32(parsedValue), err)
}

func (s StringSource) Float64() (float64, error) {
	parsedValue, err := strconv.ParseFloat(string(s), 64)
	return handleSyntaxErr(string(s), parsedValue, err)
}

func (s StringSource) String() (string, error) {
	return string(s), nil
}

func (s StringSource) Get(key string) (Source, error) {
	return nil, ErrNotSupported
}

func (s StringSource) KeyValues() (iter.Seq2[Source, Source], error) {
	return nil, ErrNotSupported
}

func (s StringSource) Iter() (iter.Seq[Source], error) {
	return nil, ErrNotSupported
}
func handleSyntaxErr[T any](inputValue string, value T, err error) (T, error) {
	var zeroValue T
	if errors.Is(err, strconv.ErrSyntax) {
		err := fmt.Errorf("parse number %q: %w", inputValue, err)
		return zeroValue, errors.Join(err, ErrNotSupported)
	}

	if err != nil {
		return zeroValue, err
	}

	return value, nil
}
