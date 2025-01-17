# unravel

[![Build Status](https://github.com/go-gum/unravel/actions/workflows/go.yml/badge.svg)](https://github.com/go-gum/unravel/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-gum/unravel#section-documentation.svg)](https://pkg.go.dev/github.com/go-gum/unravel#section-documentation)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-gum/unravel)](https://goreportcard.com/report/github.com/go-gum/unravel)

Unravel is a Go package that provides an abstraction layer for working with serialized or structured data. At its core,
it introduces the `Source` interface, which defines an abstract data model for source data. The package also
includes an `Unmarshal` function, enabling you to decode data from a `Source` into a target Go type.

## Features

- **Flexible Data Access**: The `Source` interface allows for diverse implementations, making it easy to adapt to
  various data formats and structures.
- **Unmarshal Functionality**: Decode data from a `Source` into Go types like structs, slices, strings, and more.
- **Extensibility**: Build custom `Source` implementations to handle unique data sources or formats.

## Example

```go
// Example: Unmarshaling data from query values to a struct
var qv url.Values = ... // parsed from /user?name=Alice&age=30

source := urlSource(qv)
var person Person

err := unravel.Unmarshal(source, &person)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Name: %s, Age: %d\n", person.Name, person.Age)
```

## Flexible Implementations

Unravel keeps the `Source` interface deliberately flexible, allowing developers to implement it for various data
sources. Here are some potential implementations:

- **`PathParamSource`**: Provides access to path parameters in an `http.Request`.
- **`UrlValuesSource`**: Adapts `url.Values` for query parameter decoding.
- **`BinarySource`**: Reads binary data using `binary.Encoding`, ideal for decoding binary protocols.
- **`FakerSource`**: Supplies fake values for every data access, useful for testing and simulations.
- **`CopySource`**: Copies data from a Go struct or map into another struct, similar to the functionality provided
  by the `mapstructure` package.

## Further Reading

For more examples on how to use unravel, you can read this [article](https://stuff.narf.zone/posts/unravel/).

To dive deeper into the groundwork behind unravel, check out this [post](https://stuff.narf.zone/posts/unmarshal/).

