package serde

import (
	"errors"
	"fmt"
	"iter"
	"strconv"
)

// StringValue adapts a `string` to the [SourceValue] interface, providing a straightforward
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
// as strings but needs to be decoded into native Go types. Developers can use `StringValue`
// directly or embed it within custom [SourceValue] implementations to inherit its parsing
// functionality.
//
// Example:
//
//	sv := serde.StringValue("42")
//	i, err := sv.Int() // Parses "42" into an integer value
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(i) // Outputs: 42
//
//	bsv := serde.StringValue("true")
//	b, err := bsv.Bool() // Parses "true" into a boolean value
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(b) // Outputs: true
//
// This lightweight implementation makes it easy to handle scenarios where serialized data
// is stored in string form, such as text-based configurations or simple protocols.
type StringValue string

var _ BinarySourceValue = StringValue("")

func (s StringValue) Int8() (int8, error) {
	intValue, err := strconv.ParseInt(string(s), 10, 8)
	return handleSyntaxErr(string(s), int8(intValue), err)
}

func (s StringValue) Int16() (int16, error) {
	intValue, err := strconv.ParseInt(string(s), 10, 16)
	return handleSyntaxErr(string(s), int16(intValue), err)
}

func (s StringValue) Int32() (int32, error) {
	intValue, err := strconv.ParseInt(string(s), 10, 32)
	return handleSyntaxErr(string(s), int32(intValue), err)
}

func (s StringValue) Int64() (int64, error) {
	intValue, err := strconv.ParseInt(string(s), 10, 64)
	return handleSyntaxErr(string(s), intValue, err)
}

func (s StringValue) Uint8() (uint8, error) {
	intValue, err := strconv.ParseUint(string(s), 10, 8)
	return handleSyntaxErr(string(s), uint8(intValue), err)
}

func (s StringValue) Uint16() (uint16, error) {
	intValue, err := strconv.ParseUint(string(s), 10, 16)
	return handleSyntaxErr(string(s), uint16(intValue), err)
}

func (s StringValue) Uint32() (uint32, error) {
	intValue, err := strconv.ParseUint(string(s), 10, 32)
	return handleSyntaxErr(string(s), uint32(intValue), err)
}

func (s StringValue) Uint64() (uint64, error) {
	intValue, err := strconv.ParseUint(string(s), 10, 64)
	return handleSyntaxErr(string(s), intValue, err)
}

func (s StringValue) Bool() (bool, error) {
	parsedValue, err := strconv.ParseBool(string(s))
	return handleSyntaxErr(string(s), parsedValue, err)
}

func (s StringValue) Int() (int64, error) {
	return s.Int64()
}

func (s StringValue) Uint() (uint64, error) {
	return s.Uint64()
}

func (s StringValue) Float() (float64, error) {
	return s.Float64()
}

func (s StringValue) Float32() (float32, error) {
	parsedValue, err := strconv.ParseFloat(string(s), 32)
	return handleSyntaxErr(string(s), float32(parsedValue), err)
}

func (s StringValue) Float64() (float64, error) {
	parsedValue, err := strconv.ParseFloat(string(s), 64)
	return handleSyntaxErr(string(s), parsedValue, err)
}

func (s StringValue) String() (string, error) {
	return string(s), nil
}

func (s StringValue) Get(key string) (SourceValue, error) {
	return nil, ErrNotSupported
}

func (s StringValue) KeyValues() (iter.Seq2[SourceValue, SourceValue], error) {
	return nil, ErrNotSupported
}

func (s StringValue) Iter() (iter.Seq[SourceValue], error) {
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
