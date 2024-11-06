package keyfile

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Encoder struct {
	w                *bufio.Writer
	currentGroup     reflect.StructField
	currentGroupName string
	currentField     reflect.StructField
	writtenGroupName string
	groups           map[string]map[string]map[string]string
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:      bufio.NewWriter(w),
		groups: make(map[string]map[string]map[string]string),
	}
}

func (enc *Encoder) Encode(v any) error {
	rv := reflect.ValueOf(v)

	err := enc.validateParameter(rv)
	if err != nil {
		return err
	}

	err = enc.scan(rv)
	if err != nil {
		return err
	}

	err = enc.write()
	if err != nil {
		return err
	}

	return nil
}

func (enc *Encoder) validateParameter(rv reflect.Value) error {
	if rv.Kind() == reflect.Pointer {
		return enc.validateParameter(rv.Elem())
	}
	if rv.Kind() != reflect.Struct {
		return ErrInvalidParameterType
	}
	return nil
}

func (enc *Encoder) scan(rv reflect.Value) error {
	if rv.Kind() == reflect.Pointer {
		return enc.scan(rv.Elem())
	}

	for i := range rv.NumField() {
		field := rv.Field(i)
		enc.currentGroup = rv.Type().Field(i)

		// Skip unexported or ignored groups
		if (!enc.currentGroup.IsExported() || isIgnored(enc.currentGroup.Tag)) ||
			(isOmitempty(enc.currentGroup.Tag) && field.IsZero()) {
			continue
		}

		enc.currentGroupName = cmp.Or(getKeyName(enc.currentGroup.Tag), enc.currentGroup.Name)

		enc.groups[enc.currentGroupName] = make(map[string]map[string]string)

		err := enc.scanGroup(field)
		if err != nil {
			return err
		}
	}

	return nil
}

func (enc *Encoder) scanGroup(rv reflect.Value) error {
	if rv.Kind() == reflect.Pointer {
		return enc.scanGroup(rv.Elem())
	}

	if !rv.IsValid() {
		return nil
	}

	if rv.Kind() != reflect.Struct {
		return ErrInvalidGroupType{GroupName: enc.currentGroup.Name, GroupType: enc.currentGroup.Type.String()}
	}

	enc.currentGroupName = cmp.Or(getKeyName(enc.currentGroup.Tag), enc.currentGroup.Name)

	for i := range rv.NumField() {
		field := rv.Field(i)
		enc.currentField = rv.Type().Field(i)

		// Skip unexported or ignored groups
		if !enc.currentField.IsExported() || isIgnored(enc.currentField.Tag) ||
			isOmitempty(enc.currentField.Tag) && field.IsZero() {
			continue
		}

		v, err := enc.scanField(field)
		if err != nil {
			return err
		}

		key := cmp.Or(getKeyName(enc.currentField.Tag), enc.currentField.Name)
		enc.groups[enc.currentGroupName][key] = v
	}

	return nil
}

func (enc *Encoder) scanField(rv reflect.Value) (map[string]string, error) {
	if rv.Kind() == reflect.Pointer {
		return enc.scanField(rv.Elem())
	}

	if rv.Kind() == reflect.Map {
		value, err := enc.encodeMapValue(rv)
		if err != nil {
			if errors.Is(err, ErrUnsupportedValueType{}) {
				return nil, ErrUnsupportedValueType{
					FieldName: enc.currentField.Name,
					FieldType: enc.currentField.Type.String(),
				}
			}
			return nil, err
		}

		return value, nil
	}

	value, err := enc.encodeValue(rv)
	if err != nil {
		if errors.Is(err, ErrUnsupportedValueType{}) {
			return nil, ErrUnsupportedValueType{
				FieldName: enc.currentField.Name,
				FieldType: enc.currentField.Type.String(),
			}
		}
		return nil, err
	}

	return map[string]string{
		"": value,
	}, nil
}

func (enc *Encoder) encodeValue(rv reflect.Value) (string, error) {
	if !rv.IsValid() {
		return "", nil
	}

	switch rv.Kind() {
	case reflect.Pointer:
		return enc.encodeValue(rv.Elem())

	case reflect.String:
		return escape(rv.String()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10), nil

	case reflect.Bool:
		return strconv.FormatBool(rv.Bool()), nil

	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 64), nil

	case reflect.Complex64, reflect.Complex128:
		return strconv.FormatComplex(rv.Complex(), 'f', -1, 64), nil

	case reflect.Interface:
		return enc.encodeValue(rv.Elem())

	case reflect.Slice:
		result := make([]string, 0)
		sep := getSeperator(enc.currentField.Tag)
		for i := range rv.Len() {
			v, err := enc.encodeValue(rv.Index(i))
			if err != nil {
				return "", err
			}
			result = append(result, v)
		}
		return strings.Join(result, sep), nil

	default:
		return "", ErrUnsupportedValueType{}
	}
}

func (enc *Encoder) encodeMapValue(rv reflect.Value) (map[string]string, error) {
	result := make(map[string]string)
	iter := rv.MapRange()

	for iter.Next() {
		subkey := iter.Key().String()
		v, err := enc.encodeValue(iter.Value())
		if err != nil {
			return nil, err
		}
		result[subkey] = v
	}

	return result, nil
}

func (enc *Encoder) write() error {
	groupIndexes := make([]string, 0)
	for group := range enc.groups {
		groupIndexes = append(groupIndexes, group)
	}
	sort.Strings(groupIndexes)

	for i := range groupIndexes {
		if i > 0 {
			_, err := fmt.Fprintln(enc.w)
			if err != nil {
				return err
			}
		}
		_, err := fmt.Fprintln(enc.w, fmt.Sprintf("[%s]", groupIndexes[i]))
		if err != nil {
			return err
		}

		keyIndexes := make([]string, 0)
		for key := range enc.groups[groupIndexes[i]] {
			keyIndexes = append(keyIndexes, key)
		}
		sort.Strings(keyIndexes)

		for j := range keyIndexes {
			subkeyIndexes := make([]string, 0)
			for subkey := range enc.groups[groupIndexes[i]][keyIndexes[j]] {
				subkeyIndexes = append(subkeyIndexes, subkey)
			}
			sort.Strings(subkeyIndexes)

			for k := range subkeyIndexes {
				line := keyIndexes[j]

				if subkeyIndexes[k] != "" {
					line += fmt.Sprintf("[%s]", subkeyIndexes[k])
				}

				line += fmt.Sprintf("=%s", enc.groups[groupIndexes[i]][keyIndexes[j]][subkeyIndexes[k]])
				_, err := fmt.Fprintln(enc.w, line)
				if err != nil {
					return err
				}
			}
		}
	}

	return enc.w.Flush()
}
