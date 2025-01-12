package serde

import (
	"errors"
	"fmt"
	"strconv"
)

// StringValue adapts a `string` to a SourceValue.
// It parses primitive values using strconv.ParseInt, strconv.ParseFloat
// and strconv.ParseBool. string values are returned as is.
type StringValue string

var _ IntSourceValue = StringValue("")

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
	parsedValue, err := strconv.ParseInt(string(s), 10, 64)
	return handleSyntaxErr(string(s), parsedValue, err)
}

func (s StringValue) Uint() (uint64, error) {
	parsedValue, err := strconv.ParseUint(string(s), 10, 64)
	return handleSyntaxErr(string(s), parsedValue, err)
}

func (s StringValue) Float() (float64, error) {
	parsedValue, err := strconv.ParseFloat(string(s), 64)
	return handleSyntaxErr(string(s), parsedValue, err)
}

func (s StringValue) String() (string, error) {
	return string(s), nil
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
