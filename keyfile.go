package keyfile

import "bytes"

const structTag = "keyfile"

type Unmarshaler interface {
	UnmarshalKeyFile(data []byte) error
}

func Unmarshal(data []byte, v any) error {
	err := NewDecoder(bytes.NewReader(data)).Decode(v)
	if err != nil {
		return err
	}
	return nil
}

type Marshaler interface {
	MarshalKeyFile() ([]byte, error)
}

func Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	err := NewEncoder(&buf).Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
