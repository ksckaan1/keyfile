package keyfile

import "bytes"

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
