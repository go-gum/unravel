package unravel

import "iter"

// EmptyValue is a minimal implementation of the [Source] interface that returns
// [ErrNotSupported] for all conversion methods. It serves as a no-op base for scenarios
// where no meaningful data extraction is possible or required.
//
// This implementation is particularly useful as an embedded component in custom [Source]
// implementations, providing default behavior for unsupported operations. By embedding
// `EmptyValue`, you can focus on implementing only the methods relevant to your specific use case.
//
// Example:
//
//	type CustomSource struct {
//	    unravel.EmptyValue // Embed EmptyValue for unsupported operations
//	}
//
//	func (c *CustomSource) Get(key string) (unravel.Source, error) {
//	    // Implement only the methods you need
//	    return nil, fmt.Errorf("custom logic for Get is not implemented")
//	}
//
// Usage:
//
//	ev := unravel.EmptyValue{}
//	_, err := ev.Int() // Returns ErrNotSupported
//	if errors.Is(err, unravel.ErrNotSupported) {
//	    fmt.Println("Operation not supported")
//	}
//
// By using `EmptyValue`, you can ensure your custom implementation gracefully handles
// unsupported conversions while maintaining compatibility with the [Source] interface.
type EmptyValue struct{}

var _ Source = EmptyValue{}

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

func (i EmptyValue) Get(key string) (Source, error) {
	return nil, ErrNotSupported
}

func (i EmptyValue) KeyValues() (iter.Seq2[Source, Source], error) {
	return nil, ErrNotSupported
}

func (i EmptyValue) Iter() (iter.Seq[Source], error) {
	return nil, ErrNotSupported
}
