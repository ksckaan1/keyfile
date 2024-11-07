package keyfile

import (
	"bufio"
	"cmp"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Decoder struct {
	r                *bufio.Reader
	lineNumber       int
	groups           map[string]map[string]map[string]string // map[groupName]map[key]map[subkey]value
	currentGroupName string
	currentKeyName   string
	currentField     reflect.StructField
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:      bufio.NewReader(r),
		groups: make(map[string]map[string]map[string]string),
	}
}

func (dec *Decoder) Decode(v any) error {
	rv := reflect.ValueOf(v)

	err := dec.validateParameter(rv)
	if err != nil {
		return err
	}

	err = dec.decode(rv.Elem())
	if err != nil {
		return err
	}

	return nil
}

func (dec *Decoder) validateParameter(rv reflect.Value) error {
	if rv.Kind() != reflect.Ptr {
		return ErrParameterMustBePointer
	}

	if rv.IsNil() {
		return ErrParameterMustNotBeNil
	}

	if rv.Elem().Kind() != reflect.Struct {
		return ErrInvalidParameterType
	}

	return nil
}

func (dec *Decoder) decode(rv reflect.Value) error {
	err := dec.scanDocument()
	if err != nil {
		return err
	}

	err = dec.fillModel(rv)
	if err != nil {
		return err
	}

	return nil
}

func (dec *Decoder) scanDocument() error {
	for {
		// Read line
		lineRaw, err := dec.r.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("read line: %w", err)
		}
		dec.lineNumber++

		// The end of the file
		if err == io.EOF && lineRaw == "" {
			break
		}

		// Trim line
		line := strings.TrimRight(lineRaw, "\r\n")
		line = strings.TrimSpace(line)

		// Ignore empty line
		if line == "" {
			continue
		}

		// Ignore comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Get group name
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			dec.currentGroupName = strings.Trim(line, "[]")
			dec.currentGroupName = strings.TrimSpace(dec.currentGroupName)
			if dec.currentGroupName == "" {
				return ErrInvalidGroupName{Line: line, LineNumber: dec.lineNumber}
			}

			dec.groups[dec.currentGroupName] = make(map[string]map[string]string)
			continue
		}

		// Get key-value pair
		parts := strings.SplitN(line, "=", 2)

		if len(parts) != 2 {
			return ErrInvalidEntry{Line: line, LineNumber: dec.lineNumber}
		}

		key := strings.TrimSpace(parts[0])
		subkey := ""

		if sm := mapValueRgx.FindStringSubmatch(key); len(sm) == 3 {
			key = sm[1]
			subkey = sm[2]
		}

		if key == "" {
			return ErrInvalidKey{Line: line, LineNumber: dec.lineNumber}
		}

		if dec.currentGroupName == "" {
			return ErrKeyValuePairMustBeContainedInAGroup{Line: line, LineNumber: dec.lineNumber}
		}

		if _, ok := dec.groups[dec.currentGroupName][key]; !ok {
			dec.groups[dec.currentGroupName][key] = make(map[string]string)
		}

		dec.groups[dec.currentGroupName][key][subkey] = unescape(strings.TrimSpace(parts[1]))
	}

	return nil
}

func (dec *Decoder) fillModel(model reflect.Value) error {
	for i := range model.NumField() {
		group := model.Field(i)
		groupType := model.Type().Field(i)

		// Skip unexported or ignored groups
		if !groupType.IsExported() || isIgnored(groupType.Tag) {
			continue
		}

		// check group type
		if !(group.Kind() == reflect.Struct ||
			(group.Kind() == reflect.Pointer && group.Type().Elem().Kind() == reflect.Struct)) {
			return ErrInvalidGroupType{GroupName: groupType.Name, GroupType: group.Kind().String()}
		}

		// get group name
		dec.currentGroupName = cmp.Or(getKeyName(groupType.Tag), groupType.Name)

		// check group exists
		if _, ok := dec.groups[dec.currentGroupName]; !ok {
			continue
		}

		// if group is a pointer to struct
		if group.Kind() == reflect.Ptr {
			// Initialize group
			group.Set(reflect.New(group.Type().Elem()))

			err := dec.fillGroup(group.Elem())
			if err != nil {
				return err
			}

			continue
		}

		// if group is a struct's itself
		err := dec.fillGroup(group)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dec *Decoder) fillGroup(group reflect.Value) error {
	for i := range group.NumField() {
		field := group.Field(i)
		fieldType := group.Type().Field(i)
		// Skip unexported or ignored groups
		if !fieldType.IsExported() || isIgnored(fieldType.Tag) {
			continue
		}

		dec.currentField = fieldType

		err := dec.fillField(field)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dec *Decoder) fillField(field reflect.Value) error {
	// get key
	dec.currentKeyName = cmp.Or(getKeyName(dec.currentField.Tag), dec.currentField.Name)

	// check key exists
	if !dec.isKeyExists(dec.currentGroupName, dec.currentKeyName, field.Kind() == reflect.Map) {
		return nil
	}

	val, err := dec.decodeValue(field.Type(), dec.getValue(dec.currentGroupName, dec.currentKeyName))
	if err != nil {
		return ErrCanNotParsed{
			Err:        err,
			SourceKey:  dec.currentKeyName,
			TargetName: dec.currentField.Name,
			TargetType: dec.currentField.Type.String(),
		}
	}

	field.Set(val)

	return nil
}

func (dec *Decoder) decodeValue(rt reflect.Type, value string) (reflect.Value, error) {
	if reflect.PointerTo(rt).Implements(reflect.TypeFor[Unmarshaler]()) {
		v := reflect.New(rt)
		result := v.MethodByName("UnmarshalKeyFile").Call([]reflect.Value{
			reflect.ValueOf([]byte(value)),
		})
		if !result[0].IsNil() {
			return reflect.Value{}, result[0].Interface().(error)
		}
		return v.Elem(), nil
	}

	switch rt.Kind() {
	case reflect.Interface:
		return dec.decodeAnyValue(value), nil

	case reflect.String:
		return reflect.ValueOf(value).Convert(rt), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(v).Convert(rt), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(v).Convert(rt), nil

	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(v).Convert(rt), nil

	case reflect.Complex64, reflect.Complex128:
		v, err := strconv.ParseComplex(value, 128)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(v).Convert(rt), nil

	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(v).Convert(rt), nil

	case reflect.Slice:
		sep := cmp.Or(getSeperator(dec.currentField.Tag), ";")
		elems := split(value, sep)
		slice := reflect.MakeSlice(rt, 0, len(elems))

		for i := range elems {
			elem := strings.TrimSpace(elems[i])
			v, err := dec.decodeValue(rt.Elem(), elem)
			if err != nil {
				return reflect.Value{}, err
			}
			slice = reflect.Append(slice, v)
		}
		return slice, nil

	case reflect.Map:
		if rt.Key().Kind() != reflect.String {
			return reflect.Value{}, ErrInvalidMapKeyType
		}
		m := reflect.MakeMap(rt)
		for subkey, mapValue := range dec.getMapValue(dec.currentGroupName, dec.currentKeyName) {
			v, err := dec.decodeValue(rt.Elem(), mapValue)
			if err != nil {
				return reflect.Value{}, err
			}
			m.SetMapIndex(reflect.ValueOf(subkey), v)
		}
		return m, nil

	case reflect.Pointer:
		ptr := reflect.New(rt.Elem())
		v, err := dec.decodeValue(ptr.Type().Elem(), value)
		if err != nil {
			return reflect.Value{}, err
		}
		ptr.Elem().Set(v)
		return ptr, nil
	}

	return reflect.Value{}, ErrUnsupportedValueType{}
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

var mapValueRgx = regexp.MustCompile(`(.*)\[(.*)\]`)

func (dec *Decoder) isKeyExists(groupName, key string, isMap bool) bool {
	_, ok := dec.groups[groupName]
	if !ok {
		return false
	}

	if !isMap {
		_, ok = dec.groups[groupName][key]
		return ok
	}

	for k := range dec.groups[groupName] {
		if k == key {
			return true
		}

		sm := mapValueRgx.FindStringSubmatch(k)
		if len(sm) == 3 && sm[1] == key {
			return true
		}
	}

	return false
}

func (dec *Decoder) getValue(groupName, key string) string {
	return dec.groups[groupName][key][""]
}

func (dec *Decoder) getMapValue(groupName, key string) map[string]string {
	return dec.groups[groupName][key]
}
