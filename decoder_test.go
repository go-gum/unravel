package unravel

import (
	"encoding"
	"github.com/stretchr/testify/require"
	"iter"
	"net"
	"reflect"
	"strings"
	"testing"
)

func TestUnmarshalStruct(t *testing.T) {
	type Address struct {
		City    string
		ZipCode int32 `json:"zip,omitempty"`
	}

	//goland:noinspection ALL
	type Student struct {
		Name       string
		AgeInYears int64  `json:"age"`
		SkipThis   string `json:"-"`
		Tags       Tags
		Address    *Address
		Height     float32
		Accepted   bool

		// not exported, must not be set
		note string
	}

	source := dummySource{
		Path: "$",

		Values: map[string]any{
			"$.Name": "Albert",
			"$.age":  int64(21),

			"$.Height": 1.76,

			"$.Tags":         "foo,bar",
			"$.Address.City": "Zürich",
			"$.Address.zip":  int64(8015),
			"$.Accepted":     true,

			// should not be used
			"$.SkipThis": "FOOBAR",
			"$.-":        "FOOBAR",
		},
	}

	stud, err := UnmarshalNew[Student](source)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Student{
		Name:       "Albert",
		AgeInYears: 21,
		Tags:       Tags{"foo", "bar"},
		Height:     1.76,
		Accepted:   true,
		Address: &Address{
			City:    "Zürich",
			ZipCode: 8015,
		},
	})
}

func TestUnmarshalStructWithMap(t *testing.T) {
	type Struct struct {
		Type   string
		Values map[string]string
	}

	source := dummySource{
		Path: "$",

		Values: map[string]any{
			"$.Type":       "Foo",
			"$.Values.One": "Eins",
			"$.Values.Two": "Zwei",
		},
	}

	stud, err := UnmarshalNew[Struct](source)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{
		Type: "Foo",
		Values: map[string]string{
			"One": "Eins",
			"Two": "Zwei",
		},
	})
}

func TestNaming_JsonTagExplicit(t *testing.T) {
	type Struct struct {
		A string
		B string `json:"A"`
	}

	source := dummySource{
		Values: map[string]any{
			".A": "A",
			".B": "B",
		},
	}

	stud, err := UnmarshalNew[Struct](source)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{B: "A"})
}

func TestNaming_JsonTagSkip(t *testing.T) {
	type Struct struct {
		A string
		B string `json:"-"`
	}

	source := dummySource{
		Values: map[string]any{
			".A": "A",
			".B": "B",
		},
	}

	stud, err := UnmarshalNew[Struct](source)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{A: "A"})
}

func TestNaming_JsonTagNoName(t *testing.T) {
	type Struct struct {
		A string
		B string `json:",omitempty"` // same as no json tag
	}

	source := dummySource{
		Values: map[string]any{
			".A": "A",
			".B": "B",
		},
	}

	stud, err := UnmarshalNew[Struct](source)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{A: "A", B: "B"})
}

func TestNaming_EmbeddedNamingConflict(t *testing.T) {
	type First struct{ A string }
	type Second struct{ A string }

	type Struct struct {
		First
		Second
	}

	source := dummySource{
		Values: map[string]any{
			".A": "A",
		},
	}

	stud, err := UnmarshalNew[Struct](source)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{
		// naming conflict, nothing deserializes
	})
}

func TestNaming_EmbeddedNamingExplicitWinsOnSameNesting(t *testing.T) {
	type First struct {
		A string
	}
	type Second struct {
		A string `json:"A"` // this one wins
	}

	type Struct struct {
		First
		Second
	}

	source := dummySource{
		Values: map[string]any{
			".A": "A",
		},
	}

	stud, err := UnmarshalNew[Struct](source)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{Second: Second{A: "A"}})
}

func TestNaming_EmbeddedLowerNestingWins(t *testing.T) {
	type First struct{ A string }

	type Struct struct {
		First
		A string // this one wins
	}

	source := dummySource{
		Values: map[string]any{
			".A": "A",
		},
	}

	stud, err := UnmarshalNew[Struct](source)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{A: "A"})
}

func TestNaming_NoEmbeddingWithExplicitTag(t *testing.T) {
	type First struct{ A string }

	type Struct struct {
		First `json:"First"`
		A     string
	}

	source := dummySource{
		Values: map[string]any{
			".A":       "A",
			".First.A": "FirstA",
		},
	}

	stud, err := UnmarshalNew[Struct](source)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{A: "A", First: First{A: "FirstA"}})
}

func TestNaming_EmbeddingWithExplicitNameWins(t *testing.T) {
	type First struct{ A string }

	type Struct struct {
		First `json:"A"` // wins over A string
		A     string
	}

	source := dummySource{
		Values: map[string]any{
			".A.A": "FirstA",
		},
	}

	stud, err := UnmarshalNew[Struct](source)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{First: First{A: "FirstA"}})
}

func TestNaming_NoEmbeddingWithPointer(t *testing.T) {
	type First struct{ A string }

	type Struct struct {
		*First
	}

	source := dummySource{}

	stud, err := UnmarshalNew[Struct](source)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{})
}

func TestNaming_MultipleEmbeddedTypes(t *testing.T) {
	type First struct {
		A string
		B string
		D string `json:"D"`
	}

	type Second struct {
		A string // neither First.A, nor Second.A are filled
		B string `json:"C"` // First.B and Second.B are both filled
		D string // Only first.D is filled
	}

	type Struct struct {
		First
		Second
	}

	source := dummySource{
		Values: map[string]any{
			".B": "FirstB",
			".C": "SecondB",
			".D": "FirstD",
		},
	}

	stud, err := UnmarshalNew[Struct](source)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{
		First:  First{B: "FirstB", D: "FirstD"},
		Second: Second{B: "SecondB"},
	})
}

func TestUnsupportedType(t *testing.T) {
	type Struct struct{ A any }

	_, err := UnmarshalNew[Struct](dummySource{})

	var notSupportedError NotSupportedError
	require.ErrorAs(t, err, &notSupportedError)
	require.Equal(t, notSupportedError.Type, reflect.TypeFor[any]())
}

func TestTypeUint(t *testing.T) {
	type Struct struct{ A uint }

	parsed, err := UnmarshalNew[Struct](dummySource{
		Values: map[string]any{".A": int64(1234)},
	})

	require.NoError(t, err)
	require.Equal(t, parsed, Struct{A: 1234})
}

func TestDecoderWithStructTag(t *testing.T) {
	type Struct struct {
		Foo string `url:"foo" json:"bar"`
	}

	source := dummySource{
		Values: map[string]any{".foo": "Url", ".bar": "Json"},
	}

	dec := NewDecoder().WithTag("json")
	parsed, err := UnmarshalNewWith[Struct](dec, source)
	require.NoError(t, err)
	require.Equal(t, parsed, Struct{Foo: "Json"})

	dec = dec.WithTag("url")

	parsed, err = UnmarshalNewWith[Struct](dec, source)
	require.NoError(t, err)
	require.Equal(t, parsed, Struct{Foo: "Url"})
}

func TestDecoderRequireValues(t *testing.T) {
	type Struct struct {
		Foo string
	}

	source := emptySource{}

	dec := NewDecoder().RequireValues()

	_, err := UnmarshalNewWith[Struct](dec, source)
	require.ErrorIs(t, err, ErrNoValue)
}

func TestDecoderTextUnmarshalerInterface(t *testing.T) {
	type Struct struct {
		Foo encoding.TextUnmarshaler
	}

	_, err := UnmarshalNew[Struct](dummySource{})
	require.ErrorIs(t, err, NotSupportedError{Type: reflect.TypeFor[encoding.TextUnmarshaler]()})
}

type emptySource struct{ EmptySource }

func (e emptySource) Get(key string) (Source, error) {
	return nil, ErrNoValue
}

func (e emptySource) KeyValues() (iter.Seq2[Source, Source], error) {
	return nil, ErrNotSupported
}

type Tags []string

func (t *Tags) UnmarshalText(text []byte) error {
	*t = strings.Split(string(text), ",")
	return nil
}

func TestTextUnmarshaler(t *testing.T) {
	studentSource := dummySource{
		Values: map[string]any{
			".Host": "127.0.0.1",
			".Port": int64(80),
		},
	}

	type Host struct {
		Host net.IP
		Port *int
	}

	http := 80

	value, err := UnmarshalNew[Host](studentSource)
	require.Equal(t, err, nil)
	require.Equal(t, value, Host{
		Host: net.IPv4(127, 0, 0, 1),
		Port: &http,
	})
}

func TestUnmarshalGitCommit(t *testing.T) {
	type GitCommit struct {
		Sha1   string
		Parent *GitCommit
	}

	source := dummySource{
		Values: map[string]any{
			".Sha1":                 "aaaa",
			".Parent.Sha1":          "bbbb",
			".Parent.Parent.Sha1":   "cccc",
			".Parent.Parent.Parent": nil,
		},
	}

	value, err := UnmarshalNew[GitCommit](source)
	require.Equal(t, err, nil)
	require.Equal(t, value, GitCommit{
		Sha1: "aaaa",
		Parent: &GitCommit{
			Sha1: "bbbb",
			Parent: &GitCommit{
				Sha1:   "cccc",
				Parent: nil,
			},
		},
	})
}

func TestUnmarshalSliceValue(t *testing.T) {
	type Article struct {
		Text string
		Tags []string
	}

	source := dummySource{
		Values: map[string]any{
			".Text": "some long text",
			".Tags": []string{
				"first",
				"second",
				"third",
			},
		},
	}

	value, err := UnmarshalNew[Article](source)
	require.Equal(t, err, nil)
	require.Equal(t, value, Article{
		Text: "some long text",
		Tags: []string{
			"first",
			"second",
			"third",
		},
	})
}

func TestUnmarshalArrayValue(t *testing.T) {
	source := dummySource{
		Values: map[string]any{
			"": []string{
				"first",
				"second",
				"third",
			},
		},
	}

	tags4, err := UnmarshalNew[[4]string](source)
	require.Equal(t, err, nil)
	require.Equal(t, tags4, [4]string{"first", "second", "third", ""})

	tags2, err := UnmarshalNew[[2]string](source)
	require.Equal(t, err, nil)
	require.Equal(t, tags2, [2]string{"first", "second"})
}

type dummySource struct {
	Values map[string]any
	Path   string
}

func (d dummySource) KeyValues() (iter.Seq2[Source, Source], error) {
	return func(yield func(Source, Source) bool) {
		for key, value := range d.Values {
			if !strings.HasPrefix(key, d.Path+".") {
				continue
			}

			key = strings.TrimPrefix(key, d.Path+".")

			if !yield(StringSource(key), StringSource(value.(string))) {
				return
			}
		}
	}, nil
}

func (d dummySource) Float() (float64, error) {
	if value, ok := d.Values[d.Path]; ok {
		if floatValue, ok := value.(float64); ok {
			return floatValue, nil
		}

		return 0, ErrNotSupported
	}

	return 3.14159, nil
}

func (d dummySource) Bool() (bool, error) {
	return true, nil
}

func (d dummySource) Iter() (iter.Seq[Source], error) {
	if value, ok := d.Values[d.Path]; ok {
		if sliceValue, ok := value.([]string); ok {
			valuesIter := func(yield func(Source) bool) {
				for _, value := range sliceValue {
					elementSource := dummySource{Values: map[string]any{"": value}}
					if !yield(elementSource) {
						break
					}
				}
			}
			return valuesIter, nil
		}
	}

	return nil, ErrNotSupported
}

func (d dummySource) Int() (int64, error) {
	if value, ok := d.Values[d.Path]; ok {
		if intValue, ok := value.(int64); ok {
			return intValue, nil
		}

		return 0, ErrNotSupported
	}

	return 1234, nil
}

func (d dummySource) Uint() (uint64, error) {
	value, err := d.Int()
	return uint64(value), err
}

func (d dummySource) String() (string, error) {
	if value, ok := d.Values[d.Path]; ok {
		if strValue, ok := value.(string); ok {
			return strValue, nil
		}

		return "", ErrNotSupported
	}

	return "foobar", nil
}

func (d dummySource) Get(key string) (Source, error) {
	path := d.Path + "." + key
	if value, ok := d.Values[path]; ok && value == nil {
		return nil, ErrNoValue
	}

	return dummySource{Values: d.Values, Path: path}, nil
}
