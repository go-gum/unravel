package serde

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

func (i EmptyValue) Float() (float64, error) {
	return 0, ErrNotSupported
}

func (i EmptyValue) String() (string, error) {
	return "", ErrNotSupported
}
