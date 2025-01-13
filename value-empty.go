package serde

import "iter"

// EmptyValue is a SourceValue that returns ErrNotSupported for all conversion functions.
// It is useful as an embedded base for your own custom SourceValue implementation.
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
