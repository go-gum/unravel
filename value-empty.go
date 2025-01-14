package serde

import "iter"

// EmptyValue is a minimal implementation of the [SourceValue] interface that returns
// [ErrNotSupported] for all conversion methods. It serves as a no-op base for scenarios
// where no meaningful data extraction is possible or required.
//
// This implementation is particularly useful as an embedded component in custom [SourceValue]
// implementations, providing default behavior for unsupported operations. By embedding
// `EmptyValue`, you can focus on implementing only the methods relevant to your specific use case.
//
// Example:
//
//	type CustomSourceValue struct {
//	    serde.EmptyValue // Embed EmptyValue for unsupported operations
//	}
//
//	func (c *CustomSourceValue) Get(key string) (serde.SourceValue, error) {
//	    // Implement only the methods you need
//	    return nil, fmt.Errorf("custom logic for Get is not implemented")
//	}
//
// Usage:
//
//	ev := serde.EmptyValue{}
//	_, err := ev.Int() // Returns ErrNotSupported
//	if errors.Is(err, serde.ErrNotSupported) {
//	    fmt.Println("Operation not supported")
//	}
//
// By using `EmptyValue`, you can ensure your custom implementation gracefully handles
// unsupported conversions while maintaining compatibility with the [SourceValue] interface.
type EmptyValue struct{}

var _ SourceValue = EmptyValue{}

func (i EmptyValue) Bool() (bool, error) {
	return false, ErrNotSupported
}

func (i EmptyValue) Int() (int64, error) {
	return 0, ErrNotSupported
}

func (i EmptyValue) Uint() (uint64, error) {
	return 0, ErrNotSupported
}

func (i EmptyValue) Float() (float64, error) {
	return 0, ErrNotSupported
}

func (i EmptyValue) String() (string, error) {
	return "", ErrNotSupported
}

func (i EmptyValue) Get(key string) (SourceValue, error) {
	return nil, ErrNotSupported
}

func (i EmptyValue) KeyValues() (iter.Seq2[SourceValue, SourceValue], error) {
	return nil, ErrNotSupported
}

func (i EmptyValue) Iter() (iter.Seq[SourceValue], error) {
	return nil, ErrNotSupported
}
