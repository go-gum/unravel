package examples

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"github.com/go-gum/unravel"
	"github.com/stretchr/testify/require"
	"io"
	"iter"
	"testing"
)

type binarySource struct {
	unravel.EmptySource
	r io.Reader
}

func (b binarySource) Iter() (iter.Seq[unravel.Source], error) {
	it := func(yield func(unravel.Source) bool) {
		for {
			if !yield(b) {
				break
			}
		}
	}

	return it, nil
}

func (b binarySource) Get(key string) (unravel.Source, error) {
	return b, nil
}

func (b binarySource) KeyValues() (iter.Seq2[unravel.Source, unravel.Source], error) {
	return nil, unravel.ErrNotSupported
}

func (b binarySource) Int8() (int8, error) {
	var buf [1]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return int8(buf[0]), nil
}

func (b binarySource) Int16() (int16, error) {
	var buf [2]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return int16(binary.LittleEndian.Uint16(buf[:])), nil
}

func (b binarySource) Int32() (int32, error) {
	var buf [4]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return int32(binary.LittleEndian.Uint32(buf[:])), nil
}

func (b binarySource) Int64() (int64, error) {
	var buf [8]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return int64(binary.LittleEndian.Uint64(buf[:])), nil
}

func (b binarySource) Uint8() (uint8, error) {
	var buf [1]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return buf[0], nil
}

func (b binarySource) Uint16() (uint16, error) {
	var buf [2]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(buf[:]), nil
}

func (b binarySource) Uint32() (uint32, error) {
	var buf [4]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(buf[:]), nil
}

func (b binarySource) Uint64() (uint64, error) {
	var buf [8]byte
	if _, err := b.r.Read(buf[:]); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(buf[:]), nil
}

func (b binarySource) Float32() (float32, error) {
	return 0, unravel.ErrNotSupported
}

func (b binarySource) Float64() (float64, error) {
	return 0, unravel.ErrNotSupported
}

func TestBinarySource(t *testing.T) {
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

	source := binarySource{r: bytes.NewReader(values)}
	parsed, err := unravel.UnmarshalNew[Struct](source)
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

	// a 3x5 px bitmap header
	buf, _ := base64.StdEncoding.DecodeString(`Qk3GAAAAAAAAAIoAAAB8AAAAAwAAAAUAAAABABgAAAAAADwAAAAAAAAAAAAAAAAAAAAAAAAAAAD/AAD/AAD/AAAAAAAA/0JHUnOPwvUoUbgeFR6F6wEzMzMTZmZmJmZmZgaZmZkJPQrXAyhcjzIAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA`)
	source := binarySource{r: bytes.NewReader(buf)}

	parsed, err := unravel.UnmarshalNew[Header](source)
	require.Equal(t, err, nil)

	expected := Header{
		File: BitmapFileHeader{Signature: [2]byte{66, 77}, FileSize: 198, Reserved1: 0, Reserved2: 0, PixelArrayOffset: 138},
		Info: BitmapInfoHeader{HeaderSize: 124, Width: 3, Height: 5, Planes: 1, BitsPerPixel: 24, Compression: 0, ImageSize: 60, XPixelsPerMeter: 0, YPixelsPerMeter: 0, ColorsUsed: 0, ImportantColors: 0},
	}

	require.Equal(t, parsed, expected)
}
