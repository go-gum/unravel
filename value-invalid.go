package serde

// InvalidTypeValue is a SourceValue that returns ErrInvalidType for all conversion functions.
// It is useful as an embedded base for your own custom SourceValue implementation.
type InvalidTypeValue struct{}

var _ SourceValue = InvalidTypeValue{}

func (i InvalidTypeValue) Bool() (bool, error) {
	return false, ErrInvalidType
}

func (i InvalidTypeValue) Int() (int64, error) {
	return 0, ErrInvalidType
}

func (i InvalidTypeValue) Float() (float64, error) {
	return 0, ErrInvalidType
}

func (i InvalidTypeValue) String() (string, error) {
	return "", ErrInvalidType
}
