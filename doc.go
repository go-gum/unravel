// Package serde offers a flexible and generic framework for working with serialized data.
// At its core, it provides tools to map serialized data to native Go types, such as structs,
// slices, strings, and more. This is achieved using the [Decoder] type, which facilitates
// the [Unmarshal] process, much like the familiar [encoding/json.Unmarshal].
//
// A key component of the package is the [SourceValue], which serves as an abstraction over
// the underlying serialized data. It provides a suite of methods to extract values in their
// intended forms, such as [serde.SourceValue.Int] for integers and [serde.SourceValue.String]
// for strings. During the unmarshaling process, the [Decoder] navigates the structure of the
// target Go type and retrieves the corresponding data from the [SourceValue]. This design
// makes it easy to work with a variety of serialized formats while maintaining a consistent
// API for decoding.
package serde
