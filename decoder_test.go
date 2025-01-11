package serde

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"iter"
	"net"
	"reflect"
	"strings"
	"testing"
	"unsafe"
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

	sourceValue := dummySourceValue{
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

	stud, err := UnmarshalNew[Student](sourceValue)
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

	sourceValue := dummySourceValue{
		Path: "$",

		Values: map[string]any{
			"$.Type":       "Foo",
			"$.Values.One": "Eins",
			"$.Values.Two": "Zwei",
		},
	}

	stud, err := UnmarshalNew[Struct](sourceValue)
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

	sourceValue := dummySourceValue{
		Values: map[string]any{
			".A": "A",
			".B": "B",
		},
	}

	stud, err := UnmarshalNew[Struct](sourceValue)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{B: "A"})
}

func TestNaming_JsonTagSkip(t *testing.T) {
	type Struct struct {
		A string
		B string `json:"-"`
	}

	sourceValue := dummySourceValue{
		Values: map[string]any{
			".A": "A",
			".B": "B",
		},
	}

	stud, err := UnmarshalNew[Struct](sourceValue)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{A: "A"})
}

func TestNaming_JsonTagNoName(t *testing.T) {
	type Struct struct {
		A string
		B string `json:",omitempty"` // same as no json tag
	}

	sourceValue := dummySourceValue{
		Values: map[string]any{
			".A": "A",
			".B": "B",
		},
	}

	stud, err := UnmarshalNew[Struct](sourceValue)
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

	sourceValue := dummySourceValue{
		Values: map[string]any{
			".A": "A",
		},
	}

	stud, err := UnmarshalNew[Struct](sourceValue)
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

	sourceValue := dummySourceValue{
		Values: map[string]any{
			".A": "A",
		},
	}

	stud, err := UnmarshalNew[Struct](sourceValue)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{Second: Second{A: "A"}})
}

func TestNaming_EmbeddedLowerNestingWins(t *testing.T) {
	type First struct{ A string }

	type Struct struct {
		First
		A string // this one wins
	}

	sourceValue := dummySourceValue{
		Values: map[string]any{
			".A": "A",
		},
	}

	stud, err := UnmarshalNew[Struct](sourceValue)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{A: "A"})
}

func TestNaming_NoEmbeddingWithExplicitTag(t *testing.T) {
	type First struct{ A string }

	type Struct struct {
		First `json:"First"`
		A     string
	}

	sourceValue := dummySourceValue{
		Values: map[string]any{
			".A":       "A",
			".First.A": "FirstA",
		},
	}

	stud, err := UnmarshalNew[Struct](sourceValue)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{A: "A", First: First{A: "FirstA"}})
}

func TestNaming_EmbeddingWithExplicitNameWins(t *testing.T) {
	type First struct{ A string }

	type Struct struct {
		First `json:"A"` // wins over A string
		A     string
	}

	sourceValue := dummySourceValue{
		Values: map[string]any{
			".A.A": "FirstA",
		},
	}

	stud, err := UnmarshalNew[Struct](sourceValue)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{First: First{A: "FirstA"}})
}

func TestNaming_NoEmbeddingWithPointer(t *testing.T) {
	type First struct{ A string }

	type Struct struct {
		*First
	}

	sourceValue := dummySourceValue{}

	stud, err := UnmarshalNew[Struct](sourceValue)
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

	sourceValue := dummySourceValue{
		Values: map[string]any{
			".B": "FirstB",
			".C": "SecondB",
			".D": "FirstD",
		},
	}

	stud, err := UnmarshalNew[Struct](sourceValue)
	require.Equal(t, err, nil)
	require.Equal(t, stud, Struct{
		First:  First{B: "FirstB", D: "FirstD"},
		Second: Second{B: "SecondB"},
	})
}

func TestUnsupportedType(t *testing.T) {
	type Struct struct{ A any }

	_, err := UnmarshalNew[Struct](dummySourceValue{})

	var notSupportedError NotSupportedError
	require.ErrorAs(t, err, &notSupportedError)
	require.Equal(t, notSupportedError.Type, reflect.TypeFor[any]())
}

func TestTypeUint(t *testing.T) {
	type Struct struct{ A uint }

	parsed, err := UnmarshalNew[Struct](dummySourceValue{
		Values: map[string]any{".A": int64(1234)},
	})

	require.NoError(t, err)
	require.Equal(t, parsed, Struct{A: 1234})
}

func TestNegativeUIntValues(t *testing.T) {
	type Struct struct{ A uint64 }

	_, err := UnmarshalNew[Struct](dummySourceValue{
		Values: map[string]any{".A": int64(-1)},
	})

	require.Error(t, err)
}

func TestDecoderWithStructTag(t *testing.T) {
	type Struct struct {
		Foo string `url:"foo"`
	}

	sourceValue := dummySourceValue{
		Values: map[string]any{".foo": "Works"},
	}

	dec := NewDecoder().WithTag("url")

	parsed, err := UnmarshalNewWith[Struct](dec, sourceValue)
	require.NoError(t, err)
	require.Equal(t, parsed, Struct{Foo: "Works"})
}

func TestDecoderRequireValues(t *testing.T) {
	type Struct struct {
		Foo string
	}

	sourceValue := emptySourceValue{}

	dec := NewDecoder().RequireValues()

	_, err := UnmarshalNewWith[Struct](dec, sourceValue)
	require.ErrorIs(t, err, ErrNoValue)
}

func TestDecoderTextUnmarshalerInterface(t *testing.T) {
	type Struct struct {
		Foo encoding.TextUnmarshaler
	}

	_, err := UnmarshalNew[Struct](dummySourceValue{})
	require.ErrorIs(t, err, NotSupportedError{Type: reflect.TypeFor[encoding.TextUnmarshaler]()})
}

type emptySourceValue struct{ EmptyValue }

func (e emptySourceValue) Get(key string) (SourceValue, error) {
	return nil, ErrNoValue
}

func (e emptySourceValue) KeyValues() (iter.Seq2[SourceValue, SourceValue], error) {
	return nil, ErrNotSupported
}

type Tags []string

func (t *Tags) UnmarshalText(text []byte) error {
	*t = strings.Split(string(text), ",")
	return nil
}

func TestTextUnmarshaler(t *testing.T) {
	studentSource := dummySourceValue{
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

	sourceValue := dummySourceValue{
		Values: map[string]any{
			".Sha1":                 "aaaa",
			".Parent.Sha1":          "bbbb",
			".Parent.Parent.Sha1":   "cccc",
			".Parent.Parent.Parent": nil,
		},
	}

	value, err := UnmarshalNew[GitCommit](sourceValue)
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

	sourceValue := dummySourceValue{
		Values: map[string]any{
			".Text": "some long text",
			".Tags": []string{
				"first",
				"second",
				"third",
			},
		},
	}

	value, err := UnmarshalNew[Article](sourceValue)
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
	sourceValue := dummySourceValue{
		Values: map[string]any{
			"": []string{
				"first",
				"second",
				"third",
			},
		},
	}

	tags4, err := UnmarshalNew[[4]string](sourceValue)
	require.Equal(t, err, nil)
	require.Equal(t, tags4, [4]string{"first", "second", "third", ""})

	tags2, err := UnmarshalNew[[2]string](sourceValue)
	require.Equal(t, err, nil)
	require.Equal(t, tags2, [2]string{"first", "second"})
}

type dummySourceValue struct {
	Values map[string]any
	Path   string
}

func (d dummySourceValue) KeyValues() (iter.Seq2[SourceValue, SourceValue], error) {
	return func(yield func(SourceValue, SourceValue) bool) {
		for key, value := range d.Values {
			if !strings.HasPrefix(key, d.Path+".") {
				continue
			}

			key = strings.TrimPrefix(key, d.Path+".")

			if !yield(StringValue(key), StringValue(value.(string))) {
				return
			}
		}
	}, nil
}

func (d dummySourceValue) Float() (float64, error) {
	if value, ok := d.Values[d.Path]; ok {
		if floatValue, ok := value.(float64); ok {
			return floatValue, nil
		}

		return 0, ErrNotSupported
	}

	return 3.14159, nil
}

func (d dummySourceValue) Bool() (bool, error) {
	return true, nil
}

func (d dummySourceValue) Iter() (iter.Seq[SourceValue], error) {
	if value, ok := d.Values[d.Path]; ok {
		if sliceValue, ok := value.([]string); ok {
			valuesIter := func(yield func(SourceValue) bool) {
				for _, value := range sliceValue {
					elementSource := dummySourceValue{Values: map[string]any{"": value}}
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

func (d dummySourceValue) Int() (int64, error) {
	if value, ok := d.Values[d.Path]; ok {
		if intValue, ok := value.(int64); ok {
			return intValue, nil
		}

		return 0, ErrNotSupported
	}

	return 1234, nil
}

func (d dummySourceValue) String() (string, error) {
	if value, ok := d.Values[d.Path]; ok {
		if strValue, ok := value.(string); ok {
			return strValue, nil
		}

		return "", ErrNotSupported
	}

	return "foobar", nil
}

func (d dummySourceValue) Get(key string) (SourceValue, error) {
	path := d.Path + "." + key
	if value, ok := d.Values[path]; ok && value == nil {
		return nil, ErrNoValue
	}

	return dummySourceValue{Values: d.Values, Path: path}, nil
}

type binarySourceValue struct {
	r io.Reader
}

func (b binarySourceValue) Iter() (iter.Seq[SourceValue], error) {
	it := func(yield func(SourceValue) bool) {
		for {
			if !yield(b) {
				break
			}
		}
	}

	return it, nil
}

func (b binarySourceValue) Get(key string) (SourceValue, error) {
	return b, nil
}

func (b binarySourceValue) KeyValues() (iter.Seq2[SourceValue, SourceValue], error) {
	return nil, ErrNotSupported
}

func (b binarySourceValue) Bool() (bool, error) {
	var buf [1]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return false, err
	}

	return buf[0] != 0, nil
}

func (b binarySourceValue) Int() (int64, error) {
	// no support for unsized int types
	return 0, ErrNotSupported
}

func (b binarySourceValue) Float() (float64, error) {
	// no support for unsized int types
	return 0, ErrNotSupported
}

func (b binarySourceValue) String() (string, error) {
	bc, err := b.Int32()
	if err != nil {
		return "", err
	}

	buf := make([]byte, bc)
	if _, err := b.r.Read(buf); err != nil {
		return "", err
	}

	return unsafe.String(&buf[0], bc), nil
}

func (b binarySourceValue) Int8() (int8, error) {
	var buf [1]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return int8(buf[0]), nil
}

func (b binarySourceValue) Int16() (int16, error) {
	var buf [2]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return int16(binary.LittleEndian.Uint16(buf[:])), nil
}

func (b binarySourceValue) Int32() (int32, error) {
	var buf [4]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return int32(binary.LittleEndian.Uint32(buf[:])), nil
}

func (b binarySourceValue) Int64() (int64, error) {
	var buf [8]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return int64(binary.LittleEndian.Uint64(buf[:])), nil
}

func (b binarySourceValue) Uint8() (uint8, error) {
	var buf [1]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return buf[0], nil
}

func (b binarySourceValue) Uint16() (uint16, error) {
	var buf [2]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(buf[:]), nil
}

func (b binarySourceValue) Uint32() (uint32, error) {
	var buf [4]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(buf[:]), nil
}

func (b binarySourceValue) Uint64() (uint64, error) {
	var buf [8]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(buf[:]), nil
}

func TestBinarySourceValue(t *testing.T) {
	var values []byte
	for idx := range 256 {
		values = append(values, byte(idx))
	}

	type Struct struct {
		Int8  int8
		Int16 int16
		Int32 int32
		Int64 int64

		Uint8  uint8
		Uint16 uint16
		Uint32 uint32
		Uint64 uint64
	}

	expected := Struct{
		Int8:  0,
		Int16: 0x0201,
		Int32: 0x06050403,
		Int64: 0x0e0d0c0b0a090807,

		Uint8:  0x0f,
		Uint16: 0x1110,
		Uint32: 0x15141312,
		Uint64: 0x1d1c1b1a19181716,
	}

	sourceValue := binarySourceValue{r: bytes.NewReader(values)}
	parsed, err := UnmarshalNew[Struct](sourceValue)
	require.Equal(t, err, nil)
	require.Equal(t, parsed, expected)
}

func TestDecodeBitmapHeader(t *testing.T) {
	type BitmapFileHeader struct {
		Signature        [2]byte // Signature ("BM" for Bitmap files)
		FileSize         uint32  // File size in bytes
		Reserved1        uint16  // Reserved, must be zero
		Reserved2        uint16  // Reserved, must be zero
		PixelArrayOffset uint32  // Offset to the start of the pixel array
	}

	// BitmapInfoHeader represents the DIB Header (40 bytes for BITMAPINFOHEADER)
	type BitmapInfoHeader struct {
		HeaderSize      uint32 // Size of this header (40 bytes)
		Width           int32  // Bitmap width in pixels
		Height          int32  // Bitmap height in pixels
		Planes          uint16 // Number of color planes (must be 1)
		BitsPerPixel    uint16 // Bits per pixel (1, 4, 8, 16, 24, or 32)
		Compression     uint32 // Compression method (0 = BI_RGB, no compression)
		ImageSize       uint32 // Image size (may be 0 for BI_RGB)
		XPixelsPerMeter int32  // Horizontal resolution (pixels per meter)
		YPixelsPerMeter int32  // Vertical resolution (pixels per meter)
		ColorsUsed      uint32 // Number of colors in the color palette
		ImportantColors uint32 // Number of important colors used (0 = all)
	}

	type Header struct {
		File BitmapFileHeader
		Info BitmapInfoHeader
	}

	buf, _ := base64.StdEncoding.DecodeString(`Qk3GAAAAAAAAAIoAAAB8AAAAAwAAAAUAAAABABgAAAAAADwAAAAAAAAAAAAAAAAAAAAAAAAAAAD/AAD/AAD/AAAAAAAA/0JHUnOPwvUoUbgeFR6F6wEzMzMTZmZmJmZmZgaZmZkJPQrXAyhcjzIAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA`)
	sourceValue := binarySourceValue{bytes.NewReader(buf)}

	parsed, err := UnmarshalNew[Header](sourceValue)
	require.Equal(t, err, nil)

	expected := Header{
		File: BitmapFileHeader{Signature: [2]byte{66, 77}, FileSize: 198, Reserved1: 0, Reserved2: 0, PixelArrayOffset: 138},
		Info: BitmapInfoHeader{HeaderSize: 124, Width: 3, Height: 5, Planes: 1, BitsPerPixel: 24, Compression: 0, ImageSize: 60, XPixelsPerMeter: 0, YPixelsPerMeter: 0, ColorsUsed: 0, ImportantColors: 0},
	}

	require.Equal(t, parsed, expected)
}

type rawJsonSource struct {
	value any
}

func (r rawJsonSource) Bool() (bool, error) {
	switch value := r.value.(type) {
	case bool:
		return value, nil

	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return value != 0, nil

	case float32, float64:
		return value != 0, nil

	case string:
		return value != "false" && value != "" && value != "0", nil

	default:
		return r.value != nil, nil
	}
}

func (r rawJsonSource) Int() (int64, error) {
	switch value := r.value.(type) {
	case int:
		return int64(value), nil
	case int8:
		return int64(value), nil
	case int16:
		return int64(value), nil
	case int32:
		return int64(value), nil
	case int64:
		return value, nil
	case uint:
		return int64(value), nil
	case uint8:
		return int64(value), nil
	case uint16:
		return int64(value), nil
	case uint32:
		return int64(value), nil
	case uint64:
		return int64(value), nil

	case float32:
		return int64(value), nil

	case float64:
		return int64(value), nil

	case string:
		return StringValue(value).Int()

	case json.Number:
		return value.Int64()

	default:
		return 0, ErrNotSupported
	}
}

func (r rawJsonSource) Float() (float64, error) {
	switch value := r.value.(type) {
	case float32:
		return float64(value), nil

	case float64:
		return value, nil

	case string:
		return StringValue(value).Float()

	case json.Number:
		return value.Float64()

	default:
		intValue, err := r.Int()
		return float64(intValue), err
	}
}

func (r rawJsonSource) String() (string, error) {
	switch value := r.value.(type) {
	case json.Number:
		return value.String(), nil

	case map[string]any, []any:
		return "", ErrNotSupported

	default:
		return fmt.Sprintf("%v", r.value), nil
	}
}
