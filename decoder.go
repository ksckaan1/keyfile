package keyfile

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Decoder struct {
	r                 *bufio.Reader
	currentLineNumber int
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
	}
}

func (dec *Decoder) Decode(v any) error {
	err := dec.validateParameter(v)
	if err != nil {
		return err
	}

	switch reflect.ValueOf(v).Elem().Kind() {
	case reflect.Map:
		return dec.decodeToMap(reflect.ValueOf(v).Elem())
	case reflect.Struct:
		return dec.decodeToStruct(reflect.ValueOf(v).Elem())
	}

	return nil
}

func (dec *Decoder) validateParameter(v any) error {
	wrapperVal := reflect.ValueOf(v)
	if wrapperVal.Kind() != reflect.Ptr {
		return ErrParameterMustBePointer
	}
	if wrapperVal.IsNil() {
		return ErrParameterMustNotBeNil
	}

	val := wrapperVal.Elem()

	switch val.Kind() {
	case reflect.Struct:
		err := dec.validateStruct(val)
		if err != nil {
			return err
		}
	case reflect.Map:
		err := dec.validateMap(val)
		if err != nil {
			return err
		}
	default:
		return ErrInvalidParameterType
	}
	return nil
}

func (dec *Decoder) validateMap(val reflect.Value) error {
	// section name must be a string
	// map[string]map[string]any
	//       ^
	if val.Type().Key().Kind() != reflect.String {
		return ErrSectionKeyMustBeString
	}

	// key value pair must be a map or a struct
	// map[string]map[string]any or map[string]struct{}
	//             ^                             ^
	switch val.Type().Elem().Kind() {
	case reflect.Map:
		err := dec.validateInnerMap(val.Type().Elem())
		if err != nil {
			return err
		}

	case reflect.Struct:
		err := dec.validateInnerStruct(val.Type().Elem())
		if err != nil {
			return err
		}
	default:
		return ErrKeyValuePairMustBeStructOrMap
	}

	return nil
}

func (dec *Decoder) validateInnerMap(val reflect.Type) error {
	// if section value is a map, key name must be a string
	// map[string]any
	//       ^
	if val.Key().Kind() != reflect.String {
		return ErrKeyTypeMustBeString
	}

	supportedValueTypes := []reflect.Kind{
		reflect.String, reflect.Bool, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Interface, reflect.Slice, reflect.Map, reflect.Complex64, reflect.Complex128,
	}

	// map[string]any
	//             ^
	if !dec.isKindMatches(val.Elem().Kind(), supportedValueTypes...) {
		return ErrUnsupportedValueType
	}

	// map[string][]any
	//               ^
	if val.Elem().Kind() == reflect.Slice &&
		!dec.isKindMatches(val.Elem().Elem().Kind(),
			reflect.String, reflect.Bool, reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Interface, reflect.Complex64, reflect.Complex128) {
		return ErrUnsupportedValueType
	}

	// map[string]map[string]any
	//             ^
	if val.Elem().Kind() == reflect.Map {
		err := dec.validateMapValue(val.Elem())
		if err != nil {
			return err
		}
	}

	return nil
}

func (dec *Decoder) validateStruct(val reflect.Value) error {
	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		// skip unexported or ignored fields
		if !fieldType.IsExported() ||
			fieldType.Tag.Get("keyfile") == "-" {
			continue
		}

		switch field.Kind() {
		case reflect.Struct:
			err := dec.validateInnerStruct(field.Type())
			if err != nil {
				return err
			}
		case reflect.Map:
			err := dec.validateInnerMap(field.Type())
			if err != nil {
				return err
			}
		default:
			return ErrKeyValuePairMustBeStructOrMap
		}

	}
	return nil
}

func (dec *Decoder) validateInnerStruct(val reflect.Type) error {
	supportedValueTypes := []reflect.Kind{
		reflect.String, reflect.Bool, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Interface, reflect.Slice, reflect.Map, reflect.Complex64, reflect.Complex128,
	}

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := field.Type

		// if field type is a pointer, get the underlying type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if reflect.PointerTo(fieldType).Implements(reflect.TypeOf((*Unmarshaler)(nil)).Elem()) {
			continue
		}

		// skip unexported or ignored fields
		if !field.IsExported() ||
			field.Tag.Get("keyfile") == "-" {
			continue
		}

		// check field type is supported
		if !dec.isKindMatches(
			fieldType.Kind(),
			supportedValueTypes...,
		) {
			return ErrUnsupportedValueType
		}

		// check slice type is supported, if it's slice
		if fieldType.Kind() == reflect.Slice &&
			!dec.isKindMatches(fieldType.Elem().Kind(),
				reflect.String, reflect.Bool, reflect.Float32, reflect.Float64,
				reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Interface, reflect.Complex64, reflect.Complex128, reflect.Pointer) {
			return ErrUnsupportedValueType
		}

		if fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.Pointer &&
			!dec.isKindMatches(fieldType.Elem().Elem().Kind(),
				reflect.String, reflect.Bool, reflect.Float32, reflect.Float64,
				reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Interface, reflect.Complex64, reflect.Complex128) {
			return ErrUnsupportedValueType
		}

		// check map type is supported, if it's map
		if fieldType.Kind() == reflect.Map {
			err := dec.validateMapValue(fieldType)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (dec *Decoder) validateMapValue(val reflect.Type) error {
	// map[string]any
	//       ^
	if val.Key().Kind() != reflect.String {
		return ErrUnsupportedValueType
	}

	// map[string]any
	//             ^
	if !dec.isKindMatches(val.Elem().Kind(),
		reflect.String, reflect.Bool, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Interface, reflect.Slice, reflect.Complex64, reflect.Complex128) {
		return ErrUnsupportedValueType
	}

	// map[string][]any
	//               ^
	if val.Elem().Kind() == reflect.Slice &&
		!dec.isKindMatches(val.Elem().Elem().Kind(),
			reflect.String, reflect.Bool, reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Interface, reflect.Complex64, reflect.Complex128) {
		return ErrUnsupportedValueType
	}
	return nil
}

func (dec *Decoder) isKindMatches(k reflect.Kind, types ...reflect.Kind) bool {
	for i := range types {
		if k == types[i] {
			return true
		}
	}
	return false
}

func (dec *Decoder) decodeToMap(val reflect.Value) error {
	var currentSection string
	var currentSectionMap reflect.Value
	sections := make(map[string]map[string]string)

	for {
		lineRaw, err := dec.r.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("read line: %w", err)
		}

		dec.currentLineNumber++

		if err == io.EOF && lineRaw == "" {
			break
		}

		line := strings.TrimRight(lineRaw, "\r\n")
		line = strings.TrimSpace(line)

		// ignore empty lines
		if line == "" {
			continue
		}

		// ignore comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// [section]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			switch val.Type().Elem().Kind() {
			case reflect.Map:
				currentSectionMap = reflect.MakeMap(val.Type().Elem())
				val.SetMapIndex(reflect.ValueOf(currentSection), currentSectionMap)
			case reflect.Struct:
				sections[currentSection] = make(map[string]string)
			}
			continue
		}

		// key=value
		split := strings.SplitN(line, "=", 2)
		if len(split) != 2 {
			return fmt.Errorf("line: %d: invalid key-value pair", dec.currentLineNumber)
		}

		key := strings.TrimSpace(split[0])
		value := strings.TrimSpace(split[1])

		if val.Type().Elem().Kind() == reflect.Struct {
			sections[currentSection][key] = value
			continue
		}

		// map[string]map[string]any
		//                        ^
		v, err := dec.decodeValue(val.Type().Elem().Elem(), key, value)
		if err != nil {
			return err
		}

		if val.Type().Kind() == reflect.Map {
			subkey := ""
			match := subkeyRgx.FindStringSubmatch(key)
			if len(match) == 3 {
				key = match[1]
				subkey = match[2]
			}

			dstKey := currentSectionMap.MapIndex(reflect.ValueOf(key))
			if dstKey.IsValid() {
				dstKey.SetMapIndex(reflect.ValueOf(subkey), v.MapIndex(reflect.ValueOf(subkey)))
				v = dstKey
			}
		}

		currentSectionMap.SetMapIndex(reflect.ValueOf(key), v)
	}

	if val.Type().Elem().Kind() == reflect.Struct {
		return dec.fillStructsFromMap(val, sections)
	}

	return nil
}

func (dec *Decoder) decodeToStruct(val reflect.Value) error {
	sections := make(map[string]map[string]string)
	var currentSection string

	for {
		lineRaw, err := dec.r.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("read line: %w", err)
		}

		dec.currentLineNumber++

		if err == io.EOF && lineRaw == "" {
			break
		}

		line := strings.TrimRight(lineRaw, "\r\n")
		line = strings.TrimSpace(line)

		// ignore empty lines
		if line == "" {
			continue
		}

		// ignore comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// [section]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			sections[currentSection] = make(map[string]string)
			continue
		}

		if currentSection == "" {
			return ErrKeyValuePairMustBeInSection
		}

		// key=value
		split := strings.SplitN(line, "=", 2)
		if len(split) != 2 {
			return fmt.Errorf("line: %d: invalid key-value pair", dec.currentLineNumber)
		}

		key := strings.TrimSpace(split[0])
		value := strings.TrimSpace(split[1])

		sections[currentSection][key] = value
	}

	return dec.fillStruct(val, sections)
}

func (dec *Decoder) fillStructsFromMap(val reflect.Value, sections map[string]map[string]string) error {
	for sectionKey, sectionData := range sections {
		if !val.MapIndex(reflect.ValueOf(sectionKey)).IsValid() {
			// init struct

			v := reflect.New(val.Type().Elem()).Elem()
			err := dec.fillStructField(v, sectionData)
			if err != nil {
				return err
			}
			val.SetMapIndex(reflect.ValueOf(sectionKey), v)
		}
	}

	return nil
}

func (dec *Decoder) fillStruct(val reflect.Value, sections map[string]map[string]string) error {
	for i := range val.NumField() {
		// skip unexported and ignored fields
		if !val.Type().Field(i).IsExported() ||
			val.Type().Field(i).Tag.Get("keyfile") == "-" {
			continue
		}

		field := val.Field(i)
		sectionName := val.Type().Field(i).Name
		if tag := val.Type().Field(i).Tag.Get("keyfile"); tag != "" {
			sectionName = tag
		}

		sectionData, ok := sections[sectionName]
		if !ok || sectionData == nil {
			continue
		}

		switch field.Kind() {
		case reflect.Struct:
			err := dec.fillStructField(field, sectionData)
			if err != nil {
				return err
			}
		case reflect.Map:
			err := dec.fillMapField(field, sectionData)
			if err != nil {
				return err
			}
		default:
			return ErrKeyValuePairMustBeStructOrMap
		}
	}
	return nil
}

func (dec *Decoder) fillStructField(val reflect.Value, data map[string]string) error {
	typ := val.Type()

	for i := range val.NumField() {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		key := fieldType.Name
		if tag := fieldType.Tag.Get("keyfile"); tag != "" {
			key = tag
		}

		if field.Type().Kind() == reflect.Map ||
			(field.Type().Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Map) {
			for k, strV := range data {
				if k != key && !strings.HasPrefix(k, key+"[") && !strings.HasSuffix(k, "]") {
					continue
				}

				subkey := ""
				match := subkeyRgx.FindStringSubmatch(k)
				if len(match) == 3 {
					subkey = match[2]
				}

				if field.Type().Kind() == reflect.Map { // if it's a regular map
					if field.IsNil() {
						field.Set((reflect.MakeMap(field.Type())))
					}
					v, err := dec.decodeValue(field.Type().Elem(), key, strV)
					if err != nil {
						return err
					}
					field.SetMapIndex(reflect.ValueOf(subkey), v)
				} else { // if it's a pointer to a map
					if field.IsNil() {
						field.Set(reflect.New(field.Type().Elem()))
					}
					if field.Elem().IsNil() {
						field.Elem().Set(reflect.MakeMap(field.Elem().Type()))
					}
					v, err := dec.decodeValue(field.Elem().Type().Elem(), key, strV)
					if err != nil {
						return err
					}
					field.Elem().SetMapIndex(reflect.ValueOf(subkey), v)
				}
			}

		} else {
			strValue, ok := data[key]
			if !ok {
				continue
			}

			v, err := dec.decodeValue(field.Type(), key, strValue)
			if err != nil {
				return err
			}
			field.Set(v)
		}
	}

	return nil
}

func (dec *Decoder) fillMapField(val reflect.Value, data map[string]string) error {
	if val.IsNil() {
		val.Set(reflect.MakeMap(val.Type()))
	}

	for k, strV := range data {
		if val.Type().Elem().Kind() == reflect.Map ||
			(val.Type().Elem().Kind() == reflect.Pointer && val.Type().Elem().Elem().Kind() == reflect.Map) {
			key := k
			subkey := ""
			sm := subkeyRgx.FindStringSubmatch(k)
			if len(sm) == 3 {
				key = sm[1]
				subkey = sm[2]
			}

			v, err := dec.decodeValue(val.Type().Elem().Elem(), subkey, strV)
			if err != nil {
				return err
			}

			if !val.MapIndex(reflect.ValueOf(key)).IsValid() {
				val.SetMapIndex(reflect.ValueOf(key), reflect.MakeMap(val.Type().Elem()))
			}

			val.MapIndex(reflect.ValueOf(key)).SetMapIndex(reflect.ValueOf(subkey), v)
		} else {
			v, err := dec.decodeValue(val.Type().Elem(), k, strV)
			if err != nil {
				return err
			}
			val.SetMapIndex(reflect.ValueOf(k), v)
		}
	}
	return nil
}

var subkeyRgx = regexp.MustCompile(`(.*)\[(.*)\]`)

func (dec *Decoder) decodeValue(t reflect.Type, key, value string) (reflect.Value, error) {
	if reflect.PointerTo(t).Implements(reflect.TypeOf((*Unmarshaler)(nil)).Elem()) {
		val := reflect.New(t)
		returnVal := val.MethodByName("UnmarshalKeyFile").Call([]reflect.Value{reflect.ValueOf([]byte(value))})

		if !returnVal[0].IsNil() {
			err := returnVal[0].Interface().(error)
			if err != nil {
				return reflect.Value{}, err
			}
		}
		return val.Elem(), nil
	}

	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(value).Convert(t), nil

	case reflect.Interface:
		return dec.decodeAnyValue(value), nil

	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("decode value: %w", err)
		}
		return reflect.ValueOf(v).Convert(t), nil

	case reflect.Float64, reflect.Float32:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("decode value: %w", err)
		}
		return reflect.ValueOf(v).Convert(t), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("decode value: %w", err)
		}
		return reflect.ValueOf(v).Convert(t), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("decode value: %w", err)
		}
		return reflect.ValueOf(v).Convert(t), nil

	case reflect.Complex64, reflect.Complex128:
		v, err := strconv.ParseComplex(value, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("decode value: %w", err)
		}
		return reflect.ValueOf(v).Convert(t), nil

	case reflect.Slice:
		slice := reflect.MakeSlice(t, 0, 0)
		for _, v := range strings.Split(value, ";") {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			sliceElem, err := dec.decodeValue(t.Elem(), "", v)
			if err != nil {
				return reflect.Value{}, err
			}
			slice = reflect.Append(slice, sliceElem)
		}
		return slice.Convert(t), nil

	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return reflect.Value{}, ErrKeyTypeMustBeString
		}

		m := reflect.MakeMap(t)
		subkey := ""
		match := subkeyRgx.FindStringSubmatch(key)
		if len(match) == 3 {
			subkey = match[2]
		}

		mapVal, err := dec.decodeValue(t.Elem(), "", value)
		if err != nil {
			return reflect.Value{}, ErrUnsupportedValueType
		}
		m.SetMapIndex(reflect.ValueOf(subkey), mapVal)
		return m.Convert(t), nil

	case reflect.Pointer:
		v, err := dec.decodeValue(t.Elem(), key, value)
		if err != nil {
			return reflect.Value{}, err
		}
		vPtr := reflect.New(t.Elem())
		vPtr.Elem().Set(v)
		return vPtr, nil
	}
	return reflect.Value{}, ErrUnsupportedValueType
}

func (dec *Decoder) decodeAnyValue(value string) reflect.Value {
	if val, err := strconv.ParseInt(value, 10, 64); err == nil {
		return reflect.ValueOf(val)
	}
	if val, err := strconv.ParseFloat(value, 64); err == nil {
		return reflect.ValueOf(val)
	}
	if val, err := strconv.ParseBool(value); err == nil {
		return reflect.ValueOf(val)
	}
	if val, err := strconv.ParseComplex(value, 64); err == nil {
		return reflect.ValueOf(val)
	}
	return reflect.ValueOf(value)
}
