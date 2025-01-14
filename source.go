package serde

import "iter"

// SourceValue represents the abstract interface to a serialized data source, designed to work
// seamlessly with the [Unmarshal] function. It defines a flexible data model for interpreting
// and accessing various types of serialized data.
//
// A [SourceValue] provides methods to interpret the source data in different forms:
//   - **Primitive types**: Supports conversion to basic Go types such as `bool`, `int`, `uint`,
//     `float`, and `string`.
//   - **Objects**: Accesses nested data structures using [SourceValue.Get], which retrieves
//     a value corresponding to a specified key.
//   - **Slices**: Iterates over list-like structures using [SourceValue.Iter], enabling sequential
//     processing of elements.
//   - **Maps**: Handles key-value pairs via [SourceValue.KeyValue], facilitating traversal of
//     dictionary-like data.
//
// If converting the [SourceValue] into a particular type isn't possible, the method must return
// [ErrNotSupported] as the error. This signals that the requested operation is not valid for
// the underlying data representation.
//
// Notably, there is no requirement for [SourceValue] methods to be idempotent. This design choice
// enables creative implementations, such as a `FakerSourceValue` that generates dynamic values
// on demand or a binary `SourceValue` that streams data from an [io.Reader]. Such flexibility
// allows for diverse and innovative uses of the [SourceValue] abstraction.
//
// BinarySourceValue is an extension of the [SourceValue] interface, designed to provide
// support for accessing sized integer and float types. This feature is particularly useful
// for implementing binary protocols, where precise control over data representation is required.
//
// To facilitate the creation of custom [SourceValue] implementations, the package includes
// two ready-to-use implementations:
//
//  1. **[StringValue]**: This implementation leverages the `strconv` package to parse strings
//     into various target types, such as integers, floats, and booleans. It serves as a practical
//     foundation for source values that operate on textual data.
//
//  2. **[EmptyValue]**: A minimalist implementation that returns [ErrNotSupported] for all methods.
//     This is ideal as a fallback or placeholder for unsupported operations or as a starting
//     point for developing new [SourceValue] implementations.
//
// These implementations can be embedded or delegated to within your custom [SourceValue]
// implementation, allowing you to focus on extending functionality rather than re-implementing
// common behaviors.
//
// Example:
//
//	type MySourceValue struct {
//	    serde.StringValue // Embed StringValue for string parsing support
//	}
//
//	func (m *MySourceValue) Get(key string) (serde.SourceValue, error) {
//	    // Custom logic for handling object fields
//	}
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

	// Get returns a child value of this [SourceValue] if it exists.
	// Returns error [ErrNotSupported] if the current [SourceValue] does not have any
	// child values. If the [SourceValue] does have children, but just not the
	// requested child, [ErrNoValue] must be returned.
	Get(key string) (SourceValue, error)

	// KeyValues interprets the [SourceValue] as a map and iterates over the
	// elements within. It yields a pair of key and value [SourceValue] instances.
	// Returns [ErrNotSupported] if the [SourceValue] is not iterable.
	KeyValues() (iter.Seq2[SourceValue, SourceValue], error)

	// Iter interprets the [SourceValue] as a slice and iterates over the
	// elements within.
	// Returns [ErrNotSupported] if the [SourceValue] is not iterable.
	Iter() (iter.Seq[SourceValue], error)
}

// BinarySourceValue extends the [SourceValue] interface by adding methods for extracting
// integer, unsigned integer, and floating-point values of specific bit sizes. This extension
// is particularly valuable for decoding binary formats where precise control over data size
// is essential, such as parsing protocol buffers, binary file formats, or custom serialization schemes.
//
// The additional methods include support for bit-specific types such as Int8, Int16, Int32,
// Uint8, Uint16, Uint32, and so on, as well as floating-point types like Float32 and Float64.
// These methods enable decoding of data with the exact size constraints dictated by the binary format.
//
// When using [Unmarshal], it will prioritize these specific methods (e.g., `Int8`, `Uint16`, etc.)
// over the more generic [SourceValue.Int] or [SourceValue.Float] methods. This behavior ensures
// that the decoded values adhere to the intended size and precision.
type BinarySourceValue interface {
	Int8() (int8, error)
	Int16() (int16, error)
	Int32() (int32, error)
	Int64() (int64, error)

	Uint8() (uint8, error)
	Uint16() (uint16, error)
	Uint32() (uint32, error)
	Uint64() (uint64, error)

	Float32() (float32, error)
	Float64() (float64, error)
}
