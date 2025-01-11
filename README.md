# unravel

Unravel provides a set of interfaces named `SourceValue` to define an abstract data model for source data.
It then offers a `Unmarshal` function very similar to `json.Unmarshal` to unmarshal
data from a `SourceValue` into a target type.

It tries hard to keep the source value flexible to allow all kind of implementations.
Some possible implementations are provided:

* `StringValue` parses `string` values into the supported primitive types.
* `PathParamSourceValue` gives access to path parameters of a `http.Request`.
* `UrlValuesSourceValue` is an adapter for `url.Values`
* `BinarySourceValue` can read binary data using `binary.Encoding`
* `FakeSourceValue` Provides fake values for every data access, useful for testing
* `GoSourceValue` provides access to all kinds of go values

## Further reading

* Initial groundwork for unravel: https://stuff.narf.zone/posts/unmarshal
