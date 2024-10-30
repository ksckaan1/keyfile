package keyfile

import "errors"

var (
	// Parameter
	ErrParameterMustBePointer = errors.New("keyfile: parameter must be a pointer")
	ErrParameterMustNotBeNil  = errors.New("keyfile: parameter must not be nil")
	ErrInvalidParameterType   = errors.New("keyfile: invalid parameter type")

	// Section
	ErrSectionKeyMustBeString = errors.New("keyfile: section key must be a string")

	// Key-Value Pair
	ErrKeyValuePairMustBeInSection   = errors.New("keyfile: key-value pair must be in a section")
	ErrKeyValuePairMustBeStructOrMap = errors.New("keyfile: key-value pair must be a struct or map")
	ErrKeyTypeMustBeString           = errors.New("keyfile: key type must be a string")
	ErrUnsupportedValueType          = errors.New("keyfile: unsupported value type")
)
