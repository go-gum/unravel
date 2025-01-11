package serde

import (
	"encoding"
	"errors"
	"fmt"
	"golang.org/x/exp/constraints"
	"iter"
	"math"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

var ErrNoValue = errors.New("no value")
var ErrNotSupported = errors.New("not supported")

type NotSupportedError struct {
	Type reflect.Type
}

func (n NotSupportedError) Error() string {
	return fmt.Sprintf("type %q is not supported", n.Type)
}

func Unmarshal(source SourceValue, target any) error {
	return dec.Unmarshal(source, target)
}

func UnmarshalNew[T any](source SourceValue) (T, error) {
	return UnmarshalNewWith[T](&dec, source)
}

func UnmarshalNewWith[T any](dec *Decoder, source SourceValue) (T, error) {
	var target T
	err := dec.Unmarshal(source, &target)
	return target, err
}

// A setter sets the reflect.Value to a value extracted from the given SourceValue
type setter func(SourceValue, reflect.Value) error

// A set of types that are currently in construction
type typeSet map[reflect.Type]struct{}

var tyTextUnmarshaler = reflect.TypeFor[encoding.TextUnmarshaler]()

// The default Decoder instance.
var dec Decoder

// Decoder can be used to customize unmarshalling. This type is typesafe.
type Decoder struct {
	// the struct tag that is used
	structTag string

	// Cache for setters, indexed by reflect.Type
	setterCache sync.Map

	// Require values for fields. Set to true to fail with ErrNoValue
	// if a value is missing in a MapSourceValue
	requireValues bool
}

func NewDecoder() *Decoder {
	return &Decoder{
		structTag: "json",
	}
}

func (d *Decoder) WithTag(structTag string) *Decoder {
	if d.structTag == structTag {
		return d
	}

	return &Decoder{
		structTag:     structTag,
		requireValues: d.requireValues,
	}
}

func (d *Decoder) RequireValues() *Decoder {
	if d.requireValues {
		return d
	}

	return &Decoder{
		structTag:     d.structTag,
		requireValues: true,
	}
}

func (d *Decoder) Unmarshal(source SourceValue, target any) error {
	targetValue := reflect.ValueOf(target).Elem()

	// build the setter for the targets type
	setter, err := d.setterOf(typeSet{}, targetValue.Type())
	if err != nil {
		return err
	}

	return setter(source, targetValue)
}

func (d *Decoder) setterOf(inConstruction typeSet, ty reflect.Type) (setter, error) {
	if cached, ok := d.setterCache.Load(ty); ok {
		return cached.(setter), nil
	}

	if _, ok := inConstruction[ty]; ok {
		// detected a cycle. return a setter that does a cache lookup when executed.
		// we assume that the actual setter will be in the cache once this setter is executed.
		lazySetter := func(source SourceValue, target reflect.Value) error {
			cached, _ := d.setterCache.Load(ty)
			return cached.(setter)(source, target)
		}

		return lazySetter, nil
	}

	inConstruction[ty] = struct{}{}

	setter, err := d.makeSetterOf(inConstruction, ty)
	if err != nil {
		return nil, err
	}

	d.setterCache.Store(ty, setter)

	return setter, nil
}

func (d *Decoder) makeSetterOf(inConstruction typeSet, ty reflect.Type) (setter, error) {
	if reflect.PointerTo(ty).Implements(tyTextUnmarshaler) {
		return setTextUnmarshaler, nil
	}

	switch ty.Kind() {
	case reflect.Bool:
		return setBool, nil

	case reflect.Int:
		switch unsafe.Sizeof(int(int8(0))) {
		case 4:
			return makeSetInt(IntSourceValue.Int32, reflect.Value.SetInt, math.MinInt, math.MaxInt, false), nil
		case 8:
			return makeSetInt(IntSourceValue.Int64, reflect.Value.SetInt, math.MinInt, math.MaxInt, false), nil
		default:
			panic("int must be 4 or 8 byte")
		}

	case reflect.Int8:
		return makeSetInt(IntSourceValue.Int8, reflect.Value.SetInt, math.MinInt8, math.MaxInt8, false), nil

	case reflect.Int16:
		return makeSetInt(IntSourceValue.Int16, reflect.Value.SetInt, math.MinInt16, math.MaxInt16, false), nil

	case reflect.Int32:
		return makeSetInt(IntSourceValue.Int32, reflect.Value.SetInt, math.MinInt32, math.MaxInt32, false), nil

	case reflect.Int64:
		return makeSetInt(IntSourceValue.Int64, reflect.Value.SetInt, math.MinInt64, math.MaxInt64, false), nil

	case reflect.Uint:
		switch unsafe.Sizeof(uint(0)) {
		case 4:
			return makeSetInt(IntSourceValue.Uint32, reflect.Value.SetUint, 0, math.MaxUint, true), nil
		case 8:
			return makeSetInt(IntSourceValue.Uint64, reflect.Value.SetUint, 0, math.MaxUint, true), nil
		default:
			panic("uint must be 4 or 8 byte")
		}

	case reflect.Uint8:
		return makeSetInt(IntSourceValue.Uint8, reflect.Value.SetUint, 0, math.MaxUint8, true), nil

	case reflect.Uint16:
		return makeSetInt(IntSourceValue.Uint16, reflect.Value.SetUint, 0, math.MaxUint16, true), nil

	case reflect.Uint32:
		return makeSetInt(IntSourceValue.Uint32, reflect.Value.SetUint, 0, math.MaxUint32, true), nil

	case reflect.Uint64:
		return makeSetInt(IntSourceValue.Uint64, reflect.Value.SetUint, 0, math.MaxUint64, true), nil

	case reflect.Float32, reflect.Float64:
		return setFloat, nil

	case reflect.String:
		return setString, nil

	case reflect.Pointer:
		return d.makeSetPointer(inConstruction, ty)

	case reflect.Struct:
		return d.makeSetStruct(inConstruction, ty)

	case reflect.Slice:
		return d.makeSetSlice(inConstruction, ty)

	case reflect.Array:
		return d.makeSetArray(inConstruction, ty)

	case reflect.Map:
		return d.makeSetMap(inConstruction, ty)

	default:
		return nil, NotSupportedError{Type: ty}
	}
}

func (d *Decoder) makeSetStruct(inConstruction typeSet, ty reflect.Type) (setter, error) {
	var setters []setter

	structTag := d.structTag
	if structTag == "" {
		structTag = "json"
	}

	fields := fieldsToSerialize(ty, structTag)

	for _, field := range fields {
		de, err := d.setterOf(inConstruction, field.Type)
		if err != nil {
			return nil, fmt.Errorf("setter for field %q: %w", field.Name, err)
		}

		setters = append(setters, de)
	}

	setter := func(source SourceValue, target reflect.Value) error {
		mapSource, ok := source.(MapSourceValue)
		if !ok {
			return ErrNotSupported
		}

		for idx, field := range fields {
			fieldSource, err := mapSource.Get(field.Name)
			switch {
			case errors.Is(err, ErrNoValue):
				if d.requireValues {
					return fmt.Errorf("field %q: %w", field.Name, err)
				}
				// It is okay to not get a value at all,
				// in that case we just skip the field
				continue
			case err != nil:
				return fmt.Errorf("lookup child %q: %w", field.Name, err)
			}

			fieldValue := target.FieldByIndex(field.Index)
			if err := setters[idx](fieldSource, fieldValue); err != nil {
				return fmt.Errorf("set field %q on %q: %w", field.Name, target.Type(), err)
			}
		}

		return nil
	}

	return setter, nil
}

func (d *Decoder) makeSetMap(inConstruction typeSet, ty reflect.Type) (setter, error) {
	keySetter, err := d.setterOf(inConstruction, ty.Key())
	if err != nil {
		return nil, fmt.Errorf("setter for key type %q: %w", ty, err)
	}

	valueSetter, err := d.setterOf(inConstruction, ty.Elem())
	if err != nil {
		return nil, fmt.Errorf("setter for value type %q: %w", ty, err)
	}

	keyType := ty.Key()
	valueType := ty.Elem()

	setter := func(source SourceValue, target reflect.Value) error {
		mapSource, ok := source.(MapSourceValue)
		if !ok {
			return ErrNotSupported
		}

		keyValues, err := mapSource.KeyValues()
		if err != nil {
			return fmt.Errorf("iterate key/value pairs: %w", err)
		}

		mapTarget := reflect.MakeMap(ty)

		for keySource, valueSource := range keyValues {
			keyTarget := reflect.New(keyType).Elem()
			if err := keySetter(keySource, keyTarget); err != nil {
				return fmt.Errorf("set key: %w", err)
			}

			valueTarget := reflect.New(valueType).Elem()
			if err := valueSetter(valueSource, valueTarget); err != nil {
				return fmt.Errorf("set key: %w", err)
			}

			mapTarget.SetMapIndex(keyTarget, valueTarget)
		}

		target.Set(mapTarget)

		return nil
	}

	return setter, nil
}

func (d *Decoder) makeSetSlice(inConstruction typeSet, ty reflect.Type) (setter, error) {
	elementSetter, err := d.setterOf(inConstruction, ty.Elem())
	if err != nil {
		return nil, fmt.Errorf("setter for element type %q: %w", ty, err)
	}

	// a empty element
	placeholderValue := reflect.New(ty.Elem()).Elem()

	setter := func(source SourceValue, target reflect.Value) error {
		sliceSource, ok := source.(SliceSourceValue)
		if !ok {
			return ErrNotSupported
		}

		sourceIter, err := sliceSource.Iter()
		if err != nil {
			return fmt.Errorf("as iter: %w", err)
		}

		for elementSource := range sourceIter {
			// add an empty element to grow the list
			target.Set(reflect.Append(target, placeholderValue))

			idx := target.Len() - 1
			elementValue := target.Index(idx)
			if err := elementSetter(elementSource, elementValue); err != nil {
				return fmt.Errorf("set element idx=%d: %w", idx, err)
			}
		}

		return nil
	}

	return setter, nil
}

func (d *Decoder) makeSetArray(inConstruction typeSet, ty reflect.Type) (setter, error) {
	elementSetter, err := d.setterOf(inConstruction, ty.Elem())
	if err != nil {
		return nil, fmt.Errorf("setter for element type %q: %w", ty, err)
	}

	// number of elements in the array
	elementCount := ty.Len()

	setter := func(source SourceValue, target reflect.Value) error {
		sliceSource, ok := source.(SliceSourceValue)
		if !ok {
			return ErrNotSupported
		}

		sourceIter, err := sliceSource.Iter()
		if err != nil {
			return fmt.Errorf("as iter: %w", err)
		}

		next, stop := iter.Pull(sourceIter)
		defer stop()

		for idx := 0; idx < elementCount; idx++ {
			elementSource, ok := next()
			if !ok {
				break
			}

			elementValue := target.Index(idx)
			if err := elementSetter(elementSource, elementValue); err != nil {
				return fmt.Errorf("set element idx=%d: %w", idx, err)
			}
		}

		return nil
	}

	return setter, nil
}

func (d *Decoder) makeSetPointer(inConstruction typeSet, ty reflect.Type) (setter, error) {
	pointeeType := ty.Elem()

	pointeeSetter, err := d.setterOf(inConstruction, pointeeType)
	if err != nil {
		return nil, err
	}

	setter := func(source SourceValue, target reflect.Value) error {
		// newValue is now a pointer to an instance of the pointeeType
		newValue := reflect.New(pointeeType)
		if err := pointeeSetter(source, newValue.Elem()); err != nil {
			return err
		}

		// set pointer to the new value
		target.Set(newValue)

		return nil
	}

	return setter, err
}

func setBool(source SourceValue, target reflect.Value) error {
	boolValue, err := source.Bool()
	if err != nil {
		return fmt.Errorf("get bool value: %w", err)
	}

	target.SetBool(boolValue)
	return nil
}

func makeSetInt[T constraints.Integer | constraints.Unsigned, V uint64 | int64](
	parse func(IntSourceValue) (T, error),
	setValue func(reflect.Value, V),
	minValue, maxValue V,
	isUnsigned bool,
) setter {
	return func(source SourceValue, target reflect.Value) error {
		if intSource, ok := source.(IntSourceValue); ok {
			parsedValue, err := parse(intSource)
			if err != nil {
				return fmt.Errorf("get %T value: %w", parsedValue, err)
			}

			setValue(target, V(parsedValue))
			return nil
		}

		// no int source, need to fallback to SourceValue.Int
		intValue, err := source.Int()
		if err != nil {
			return fmt.Errorf("get int value: %w", err)
		}

		var vZero V

		if isUnsigned && intValue < 0 {
			return ErrNotSupported
		}

		if V(intValue) < minValue {
			return fmt.Errorf("invalid %T value %d: %w", vZero, intValue, strconv.ErrRange)
		}

		if V(intValue) > maxValue {
			return fmt.Errorf("invalid %T value: %d: %w", vZero, intValue, strconv.ErrRange)
		}

		setValue(target, V(intValue))
		return nil
	}
}

func setFloat(source SourceValue, target reflect.Value) error {
	floatValue, err := source.Float()
	if err != nil {
		return fmt.Errorf("get float value: %w", err)
	}

	target.SetFloat(floatValue)
	return nil
}

func setString(source SourceValue, target reflect.Value) error {
	stringValue, err := source.String()
	if err != nil {
		return fmt.Errorf("get string value: %w", err)
	}

	target.SetString(stringValue)

	return nil
}

func setTextUnmarshaler(source SourceValue, target reflect.Value) error {
	text, err := source.String()
	if err != nil {
		return fmt.Errorf("get string value: %w", err)
	}

	m := target.Addr().Interface().(encoding.TextUnmarshaler)
	return m.UnmarshalText([]byte(text))
}
