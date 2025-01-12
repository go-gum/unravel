package serde

import "iter"

// SourceValue describes a source value that can be feed into the Unmarshal function.
// It defines the abstract data model used for serialization. To ease implementation
// the SourceValue interface is defined for primitive types only. It can be extended
// due to the magic of interface stuffing, by also implementing MapSourceValue and
// SliceSourceValue if needed.
type SourceValue interface {
	// Bool returns the current value as a bool.
	// Returns error ErrNotSupported if the value can not be represented as such.
	Bool() (bool, error)

	// Int returns the current value as an int64.
	// Returns error ErrNotSupported if the value can not be represented as such.
	Int() (int64, error)

	// Uint returns the current value as an uint64.
	// Returns error ErrNotSupported if the value can not be represented as such.
	Uint() (uint64, error)

	// Float returns the current value as a float64.
	// Returns error ErrNotSupported if the value can not be represented as such.
	Float() (float64, error)

	// String returns the current value as a string.
	// Returns error ErrNotSupported if the value can not be represented as such.
	String() (string, error)
}

type MapSourceValue interface {
	// Get returns a child value of this SourceValue if it exists.
	// Returns error ErrNotSupported if the current SourceValue does not have any
	// child values. If the SourceValue does have children, but just not the
	// requested child, ErrNoValue must be returned.
	Get(key string) (SourceValue, error)

	// KeyValues interprets the SourceValue as a map and iterates over the
	// elements within.
	// Returns ErrNotSupported if the SourceValue is not iterable.
	KeyValues() (iter.Seq2[SourceValue, SourceValue], error)
}

type SliceSourceValue interface {
	// Iter interprets the SourceValue as a slice and iterates over the
	// elements within.
	// Returns ErrNotSupported if the SourceValue is not iterable.
	Iter() (iter.Seq[SourceValue], error)
}

// IntSourceValue extends a SourceValue by functions to query for specific
// IntXX and UintXX values. This can be useful when decoding different sized Int
// values is relevant, e.g. for binary formats. The Decoder will prefer the
// specific Int8, Int16, ... methods to the generic SourceValue.Int method.
type IntSourceValue interface {
	Int8() (int8, error)
	Int16() (int16, error)
	Int32() (int32, error)
	Int64() (int64, error)

	Uint8() (uint8, error)
	Uint16() (uint16, error)
	Uint32() (uint32, error)
	Uint64() (uint64, error)
}
