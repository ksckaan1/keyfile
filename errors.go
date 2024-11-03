package keyfile

import (
	"errors"
	"fmt"
)

var (
	// Parameter
	ErrParameterMustBePointer = errors.New("keyfile: parameter must be a pointer")
	ErrParameterMustNotBeNil  = errors.New("keyfile: parameter must not be nil")
	ErrInvalidParameterType   = errors.New("keyfile: invalid parameter type")

	// Key-Value Pair
	ErrUnsupportedValueType = errors.New("keyfile: unsupported value type")
	ErrInvalidMapKeyType    = errors.New("invalid map key type, must be a string")
)

type ErrInvalidGroupName struct {
	Line       string
	LineNumber int
}

func (e ErrInvalidGroupName) Error() string {
	return fmt.Sprintf("keyfile: line[%d] -> invalid group name: %q", e.LineNumber, e.Line)
}

type ErrKeyValuePairMustBeContainedInAGroup struct {
	Line       string
	LineNumber int
}

func (e ErrKeyValuePairMustBeContainedInAGroup) Error() string {
	return fmt.Sprintf("keyfile: line[%d] -> key-value pair must be contained in a group: %q", e.LineNumber, e.Line)
}

type ErrInvalidEntry struct {
	Line       string
	LineNumber int
}

func (e ErrInvalidEntry) Error() string {
	return fmt.Sprintf("keyfile: line[%d] -> invalid entry: %q", e.LineNumber, e.Line)
}

type ErrInvalidKey struct {
	Line       string
	LineNumber int
}

func (e ErrInvalidKey) Error() string {
	return fmt.Sprintf("keyfile: line[%d] -> invalid key: %q", e.LineNumber, e.Line)
}

type ErrInvalidGroupType struct {
	GroupName string
	GroupType string
}

func (e ErrInvalidGroupType) Error() string {
	return fmt.Sprintf("keyfile: invalid group type: %s %s", e.GroupName, e.GroupType)
}

type ErrCanNotParsed struct {
	Err        error
	SourceKey  string
	TargetName string
	TargetType string
}

func (e ErrCanNotParsed) Error() string {
	return fmt.Sprintf("keyfile: can not parsed: from %q to \"%s %s\": %s", e.SourceKey, e.TargetName, e.TargetType, e.Err)
}
