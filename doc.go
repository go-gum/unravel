// Package serde provides a generic data model to describe access to serialized data. It further
// provides a [Decoder] type to [Unmarshal] data onto go types (e.g. structs, slices, strings, etc)
// similar to [json.Unmarshal].
//
// The [SourceValue] defines access to a serialized value. It adapts the underlying representation
// using a set of functions. The [Decoder.Unmarshal] function walks the target type and pulls data out
// of the [SourceValue] using functions like [SourceValue.Int], [SourceValue.String], etc.
package serde
