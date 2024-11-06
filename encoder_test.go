package keyfile

import (
	"errors"
	"testing"
	"time"
)

func TestEncoder(t *testing.T) {
	tests := []struct {
		name  string
		model any
		want  string
		err   error
	}{
		// invalid
		{
			name:  "invalid parameter",
			model: nil,
			want:  "",
			err:   ErrInvalidParameterType,
		},
		{
			name:  "nil parameter",
			model: (*struct{})(nil),
			want:  "",
			err:   ErrInvalidParameterType,
		},
		{
			name: "invalid group type",
			model: struct {
				ExampleField string
			}{},
			want: "",
			err: ErrInvalidGroupType{
				GroupName: "ExampleField",
				GroupType: "string",
			},
		},
		{
			name: "invalid field type",
			model: struct {
				ExampleGroup struct {
					ExampleField time.Time `keyfile:"example_field"`
				} `keyfile:"example_group"`
			}{},
			want: "",
			err: ErrUnsupportedValueType{
				FieldName: "ExampleField",
				FieldType: "time.Time",
			},
		},
		// valid
		{
			name: "string field",
			model: struct {
				Group struct {
					Key1 string `keyfile:"key1"`
				} `keyfile:"group"`
			}{
				Group: struct {
					Key1 string "keyfile:\"key1\""
				}{
					Key1: "value1",
				},
			},
			want: "[group]\nkey1=value1\n",
			err:  nil,
		},
		{
			name: "int field",
			model: struct {
				Group struct {
					Key1 int   `keyfile:"key1"`
					Key2 int8  `keyfile:"key2"`
					Key3 int16 `keyfile:"key3"`
					Key4 int32 `keyfile:"key4"`
					Key5 int64 `keyfile:"key5"`
				} `keyfile:"group"`
			}{
				Group: struct {
					Key1 int   "keyfile:\"key1\""
					Key2 int8  "keyfile:\"key2\""
					Key3 int16 "keyfile:\"key3\""
					Key4 int32 "keyfile:\"key4\""
					Key5 int64 "keyfile:\"key5\""
				}{
					Key1: 42,
					Key2: 42,
					Key3: 42,
					Key4: 42,
					Key5: 42,
				},
			},
			want: "[group]\nkey1=42\nkey2=42\nkey3=42\nkey4=42\nkey5=42\n",
			err:  nil,
		},
		{
			name: "uint field",
			model: struct {
				Group struct {
					Key1 uint   `keyfile:"key1"`
					Key2 uint8  `keyfile:"key2"`
					Key3 uint16 `keyfile:"key3"`
					Key4 uint32 `keyfile:"key4"`
					Key5 uint64 `keyfile:"key5"`
				} `keyfile:"group"`
			}{
				Group: struct {
					Key1 uint   "keyfile:\"key1\""
					Key2 uint8  "keyfile:\"key2\""
					Key3 uint16 "keyfile:\"key3\""
					Key4 uint32 "keyfile:\"key4\""
					Key5 uint64 "keyfile:\"key5\""
				}{
					Key1: 42,
					Key2: 42,
					Key3: 42,
					Key4: 42,
					Key5: 42,
				},
			},
			want: "[group]\nkey1=42\nkey2=42\nkey3=42\nkey4=42\nkey5=42\n",
			err:  nil,
		},
		{
			name: "float field",
			model: struct {
				Group struct {
					Key1 float32 `keyfile:"key1"`
					Key2 float32 `keyfile:"key2"`
				} `keyfile:"group"`
			}{
				Group: struct {
					Key1 float32 "keyfile:\"key1\""
					Key2 float32 "keyfile:\"key2\""
				}{
					Key1: 42.5,
					Key2: 42.5,
				},
			},
			want: "[group]\nkey1=42.5\nkey2=42.5\n",
			err:  nil,
		},
		{
			name: "complex field",
			model: struct {
				Group struct {
					Key1 complex64  `keyfile:"key1"`
					Key2 complex128 `keyfile:"key2"`
				} `keyfile:"group"`
			}{
				Group: struct {
					Key1 complex64  "keyfile:\"key1\""
					Key2 complex128 "keyfile:\"key2\""
				}{
					Key1: 10 + 11i,
					Key2: 10 + 11i,
				},
			},
			want: "[group]\nkey1=(10+11i)\nkey2=(10+11i)\n",
			err:  nil,
		},
		{
			name: "any field",
			model: struct {
				Group struct {
					Key1 any `keyfile:"key1"`
					Key2 any `keyfile:"key2"`
					Key3 any `keyfile:"key3"`
					Key4 any `keyfile:"key4"`
					Key5 any `keyfile:"key5"`
					Key6 any `keyfile:"key6"`
				} `keyfile:"group"`
			}{
				Group: struct {
					Key1 any "keyfile:\"key1\""
					Key2 any "keyfile:\"key2\""
					Key3 any "keyfile:\"key3\""
					Key4 any "keyfile:\"key4\""
					Key5 any "keyfile:\"key5\""
					Key6 any "keyfile:\"key6\""
				}{
					Key1: int8(42),
					Key2: float64(42.5),
					Key3: "value",
					Key4: true,
					Key5: []int{1, 2, 3},
					Key6: uint(42),
				},
			},
			want: "[group]\nkey1=42\nkey2=42.5\nkey3=value\nkey4=true\nkey5=1;2;3\nkey6=42\n",
			err:  nil,
		},
		{
			name: "nil field",
			model: struct {
				Group struct {
					Key1 *string `keyfile:"key1"`
				} `keyfile:"group"`
			}{
				Group: struct {
					Key1 *string "keyfile:\"key1\""
				}{
					Key1: nil,
				},
			},
			want: "[group]\nkey1=\n",
			err:  nil,
		},
		{
			name: "omitempty field",
			model: struct {
				Group struct {
					Key1 *string `keyfile:"key1,omitempty"`
				} `keyfile:"group"`
			}{
				Group: struct {
					Key1 *string "keyfile:\"key1,omitempty\""
				}{
					Key1: nil,
				},
			},
			want: "[group]\n",
			err:  nil,
		},
		{
			name: "escaped string field",
			model: struct {
				Group struct {
					Key1 string `keyfile:"key1,omitempty"`
				} `keyfile:"group"`
			}{
				Group: struct {
					Key1 string "keyfile:\"key1,omitempty\""
				}{
					Key1: "   example value\n\thello",
				},
			},
			want: "[group]\nkey1=\\s\\s\\sexample value\\n\\thello\n",
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.model)
			if !errors.Is(err, tt.err) {
				t.Fatal(err)
			}

			if string(got) != tt.want {
				t.Fatalf("got %s, want %s", string(got), tt.want)
			}
		})
	}
}
