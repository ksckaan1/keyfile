package keyfile

import (
	"errors"
	"reflect"
	"testing"
)

func TestDecoder(t *testing.T) {
	tests := []struct {
		name string
		src  string
		dst  any
		want any
		err  error
	}{
		// invalid
		{
			name: "parameter must be a pointer",
			src:  ``,
			dst:  struct{}{},
			want: struct{}{},
			err:  ErrParameterMustBePointer,
		},
		{
			name: "parameter must be a struct",
			src:  ``,
			dst:  new(string),
			want: new(string),
			err:  ErrInvalidParameterType,
		},
		{
			name: "parameter must not be nil",
			src:  ``,
			dst:  (*string)(nil),
			want: (*string)(nil),
			err:  ErrParameterMustNotBeNil,
		},
		// valid
		{
			name: "empty",
			src:  ``,
			dst:  new(struct{}),
			want: new(struct{}),
			err:  nil,
		},
		{
			name: "string field",
			src: `[example]
						key1 = value
						key2 =
						key3 = \s\n\t\r`,
			dst: &struct {
				Example struct {
					Key1 string `keyfile:"key1"`
					Key2 string `keyfile:"key2"`
					Key3 string `keyfile:"key3"`
				} `keyfile:"example"`
			}{},
			want: &struct {
				Example struct {
					Key1 string `keyfile:"key1"`
					Key2 string `keyfile:"key2"`
					Key3 string `keyfile:"key3"`
				} `keyfile:"example"`
			}{
				Example: struct {
					Key1 string `keyfile:"key1"`
					Key2 string `keyfile:"key2"`
					Key3 string `keyfile:"key3"`
				}{
					Key1: "value",
					Key2: "",
					Key3: " \n\t\r",
				},
			},
			err: nil,
		},
		{
			name: "bool field",
			src: `[example]
						key1 = true
						key2 = false
						key3 = TRUE
						key4 = FALSE`,
			dst: &struct {
				Example struct {
					Key1 bool `keyfile:"key1"`
					Key2 bool `keyfile:"key2"`
					Key3 bool `keyfile:"key3"`
					Key4 bool `keyfile:"key4"`
				} `keyfile:"example"`
			}{},
			want: &struct {
				Example struct {
					Key1 bool `keyfile:"key1"`
					Key2 bool `keyfile:"key2"`
					Key3 bool `keyfile:"key3"`
					Key4 bool `keyfile:"key4"`
				} `keyfile:"example"`
			}{
				Example: struct {
					Key1 bool `keyfile:"key1"`
					Key2 bool `keyfile:"key2"`
					Key3 bool `keyfile:"key3"`
					Key4 bool `keyfile:"key4"`
				}{
					Key1: true,
					Key2: false,
					Key3: true,
					Key4: false,
				},
			},
			err: nil,
		},
		{
			name: "int field",
			src: `[example]
						key1 = 42
						key2 = 42
						key3 = 42
						key4 = 42
						key5 = 42`,
			dst: &struct {
				Example struct {
					Key1 int   `keyfile:"key1"`
					Key2 int8  `keyfile:"key2"`
					Key3 int16 `keyfile:"key3"`
					Key4 int32 `keyfile:"key4"`
					Key5 int64 `keyfile:"key5"`
				} `keyfile:"example"`
			}{},
			want: &struct {
				Example struct {
					Key1 int   `keyfile:"key1"`
					Key2 int8  `keyfile:"key2"`
					Key3 int16 `keyfile:"key3"`
					Key4 int32 `keyfile:"key4"`
					Key5 int64 `keyfile:"key5"`
				} `keyfile:"example"`
			}{
				Example: struct {
					Key1 int   `keyfile:"key1"`
					Key2 int8  `keyfile:"key2"`
					Key3 int16 `keyfile:"key3"`
					Key4 int32 `keyfile:"key4"`
					Key5 int64 `keyfile:"key5"`
				}{
					Key1: 42,
					Key2: 42,
					Key3: 42,
					Key4: 42,
					Key5: 42,
				},
			},
			err: nil,
		},
		{
			name: "uint field",
			src: `[example]
						key1 = 42
						key2 = 42
						key3 = 42
						key4 = 42
						key5 = 42`,
			dst: &struct {
				Example struct {
					Key1 uint   `keyfile:"key1"`
					Key2 uint8  `keyfile:"key2"`
					Key3 uint16 `keyfile:"key3"`
					Key4 uint32 `keyfile:"key4"`
					Key5 uint64 `keyfile:"key5"`
				} `keyfile:"example"`
			}{},
			want: &struct {
				Example struct {
					Key1 uint   `keyfile:"key1"`
					Key2 uint8  `keyfile:"key2"`
					Key3 uint16 `keyfile:"key3"`
					Key4 uint32 `keyfile:"key4"`
					Key5 uint64 `keyfile:"key5"`
				} `keyfile:"example"`
			}{
				Example: struct {
					Key1 uint   `keyfile:"key1"`
					Key2 uint8  `keyfile:"key2"`
					Key3 uint16 `keyfile:"key3"`
					Key4 uint32 `keyfile:"key4"`
					Key5 uint64 `keyfile:"key5"`
				}{
					Key1: 42,
					Key2: 42,
					Key3: 42,
					Key4: 42,
					Key5: 42,
				},
			},
			err: nil,
		},
		{
			name: "float field",
			src: `[example]
						key1 = 42.5
						key2 = 42.5`,
			dst: &struct {
				Example struct {
					Key1 float32 `keyfile:"key1"`
					Key2 float64 `keyfile:"key2"`
				} `keyfile:"example"`
			}{},
			want: &struct {
				Example struct {
					Key1 float32 `keyfile:"key1"`
					Key2 float64 `keyfile:"key2"`
				} `keyfile:"example"`
			}{
				Example: struct {
					Key1 float32 `keyfile:"key1"`
					Key2 float64 `keyfile:"key2"`
				}{
					Key1: 42.5,
					Key2: 42.5,
				},
			},
			err: nil,
		},
		{
			name: "complex field",
			src: `[example]
						key1 = 10+11i 
						key2 = 10+11i
						key3 = (10+11i) 
						key4 = (10+11i)`,
			dst: &struct {
				Example struct {
					Key1 complex64  `keyfile:"key1"`
					Key2 complex128 `keyfile:"key2"`
					Key3 complex64  `keyfile:"key3"`
					Key4 complex128 `keyfile:"key4"`
				} `keyfile:"example"`
			}{},
			want: &struct {
				Example struct {
					Key1 complex64  `keyfile:"key1"`
					Key2 complex128 `keyfile:"key2"`
					Key3 complex64  `keyfile:"key3"`
					Key4 complex128 `keyfile:"key4"`
				} `keyfile:"example"`
			}{
				Example: struct {
					Key1 complex64  `keyfile:"key1"`
					Key2 complex128 `keyfile:"key2"`
					Key3 complex64  `keyfile:"key3"`
					Key4 complex128 `keyfile:"key4"`
				}{
					Key1: 10 + 11i,
					Key2: 10 + 11i,
					Key3: 10 + 11i,
					Key4: 10 + 11i,
				},
			},
			err: nil,
		},
		{
			name: "any field",
			src: `[example]
						key1 = 42 
						key2 = 42.5
						key3 = true
						key4 = 10+11i 
						key5 = value`,
			dst: &struct {
				Example struct {
					Key1 any `keyfile:"key1"`
					Key2 any `keyfile:"key2"`
					Key3 any `keyfile:"key3"`
					Key4 any `keyfile:"key4"`
					Key5 any `keyfile:"key5"`
				} `keyfile:"example"`
			}{},
			want: &struct {
				Example struct {
					Key1 any `keyfile:"key1"`
					Key2 any `keyfile:"key2"`
					Key3 any `keyfile:"key3"`
					Key4 any `keyfile:"key4"`
					Key5 any `keyfile:"key5"`
				} `keyfile:"example"`
			}{
				Example: struct {
					Key1 any `keyfile:"key1"`
					Key2 any `keyfile:"key2"`
					Key3 any `keyfile:"key3"`
					Key4 any `keyfile:"key4"`
					Key5 any `keyfile:"key5"`
				}{
					Key1: int64(42),
					Key2: float64(42.5),
					Key3: true,
					Key4: 10 + 11i,
					Key5: "value",
				},
			},
			err: nil,
		},
		{
			name: "slice field",
			src: `[example]
						key1 = a;b;c 
						key2 = 1,2,3
						key3 = 1;2;3
						key4 = 1,2,3
						key5 = 1;2;3
						key6 = 1,2,3
						key7 = 1;2;3
						key8 = 1,2,3
						key9 = 1;2;3
						key10 = 1,2,3
						key11 = 1;2;3
						key12 = 1,2,3
						key13 = 1;2;3
						key14 = 10+11i,10+11i,10+11i
						key15 = 10+11i;10+11i;10+11i
						key16 = true,false,true
						key17 = 42;42.5;true;10+11i;value`,
			dst: &struct {
				Example struct {
					Key1  []string     `keyfile:"key1"`
					Key2  []int        `keyfile:"key2;sep:,"`
					Key3  []int8       `keyfile:"key3"`
					Key4  []int16      `keyfile:"key4;sep:,"`
					Key5  []int32      `keyfile:"key5"`
					Key6  []int64      `keyfile:"key6;sep:,"`
					Key7  []uint       `keyfile:"key7"`
					Key8  []uint8      `keyfile:"key8;sep:,"`
					Key9  []uint16     `keyfile:"key9"`
					Key10 []uint32     `keyfile:"key10;sep:,"`
					Key11 []uint64     `keyfile:"key11"`
					Key12 []float32    `keyfile:"key12;sep:,"`
					Key13 []float64    `keyfile:"key13"`
					Key14 []complex64  `keyfile:"key14;sep:,"`
					Key15 []complex128 `keyfile:"key15"`
					Key16 []bool       `keyfile:"key16;sep:,"`
					Key17 []any        `keyfile:"key17"`
				} `keyfile:"example"`
			}{},
			want: &struct {
				Example struct {
					Key1  []string     `keyfile:"key1"`
					Key2  []int        `keyfile:"key2;sep:,"`
					Key3  []int8       `keyfile:"key3"`
					Key4  []int16      `keyfile:"key4;sep:,"`
					Key5  []int32      `keyfile:"key5"`
					Key6  []int64      `keyfile:"key6;sep:,"`
					Key7  []uint       `keyfile:"key7"`
					Key8  []uint8      `keyfile:"key8;sep:,"`
					Key9  []uint16     `keyfile:"key9"`
					Key10 []uint32     `keyfile:"key10;sep:,"`
					Key11 []uint64     `keyfile:"key11"`
					Key12 []float32    `keyfile:"key12;sep:,"`
					Key13 []float64    `keyfile:"key13"`
					Key14 []complex64  `keyfile:"key14;sep:,"`
					Key15 []complex128 `keyfile:"key15"`
					Key16 []bool       `keyfile:"key16;sep:,"`
					Key17 []any        `keyfile:"key17"`
				} `keyfile:"example"`
			}{
				Example: struct {
					Key1  []string     `keyfile:"key1"`
					Key2  []int        `keyfile:"key2;sep:,"`
					Key3  []int8       `keyfile:"key3"`
					Key4  []int16      `keyfile:"key4;sep:,"`
					Key5  []int32      `keyfile:"key5"`
					Key6  []int64      `keyfile:"key6;sep:,"`
					Key7  []uint       `keyfile:"key7"`
					Key8  []uint8      `keyfile:"key8;sep:,"`
					Key9  []uint16     `keyfile:"key9"`
					Key10 []uint32     `keyfile:"key10;sep:,"`
					Key11 []uint64     `keyfile:"key11"`
					Key12 []float32    `keyfile:"key12;sep:,"`
					Key13 []float64    `keyfile:"key13"`
					Key14 []complex64  `keyfile:"key14;sep:,"`
					Key15 []complex128 `keyfile:"key15"`
					Key16 []bool       `keyfile:"key16;sep:,"`
					Key17 []any        `keyfile:"key17"`
				}{
					Key1:  []string{"a", "b", "c"},
					Key2:  []int{1, 2, 3},
					Key3:  []int8{1, 2, 3},
					Key4:  []int16{1, 2, 3},
					Key5:  []int32{1, 2, 3},
					Key6:  []int64{1, 2, 3},
					Key7:  []uint{1, 2, 3},
					Key8:  []uint8{1, 2, 3},
					Key9:  []uint16{1, 2, 3},
					Key10: []uint32{1, 2, 3},
					Key11: []uint64{1, 2, 3},
					Key12: []float32{1, 2, 3},
					Key13: []float64{1, 2, 3},
					Key14: []complex64{10 + 11i, 10 + 11i, 10 + 11i},
					Key15: []complex128{10 + 11i, 10 + 11i, 10 + 11i},
					Key16: []bool{true, false, true},
					Key17: []any{int64(42), float64(42.5), true, 10 + 11i, "value"},
				},
			},
			err: nil,
		},
		{
			name: "map field",
			src: `[example]
						greet = hello
						greet[tr] = merhaba
						greet[de] = hallo
						greet[it] = ciao
						greet[fr] = bonjour`,
			dst: &struct {
				Example struct {
					Greet map[string]string `keyfile:"greet"`
				} `keyfile:"example"`
			}{},
			want: &struct {
				Example struct {
					Greet map[string]string `keyfile:"greet"`
				} `keyfile:"example"`
			}{
				Example: struct {
					Greet map[string]string `keyfile:"greet"`
				}{
					Greet: map[string]string{
						"":   "hello",
						"tr": "merhaba",
						"de": "hallo",
						"it": "ciao",
						"fr": "bonjour",
					},
				},
			},
			err: nil,
		},
		{
			name: "unexported group",
			src: `[example]
						key1 = 42`,
			dst: &struct {
				example struct {
					Key1 int `keyfile:"key1"`
				} `keyfile:"example"`
			}{},
			want: &struct {
				example struct {
					Key1 int `keyfile:"key1"`
				} `keyfile:"example"`
			}{
				example: struct {
					Key1 int `keyfile:"key1"`
				}{
					Key1: 0,
				},
			},
			err: nil,
		},
		{
			name: "ignored group",
			src: `[example]
						key1 = 42`,
			dst: &struct {
				Example struct {
					Key1 int `keyfile:"key1"`
				} `keyfile:"-"`
			}{},
			want: &struct {
				Example struct {
					Key1 int `keyfile:"key1"`
				} `keyfile:"-"`
			}{
				Example: struct {
					Key1 int `keyfile:"key1"`
				}{
					Key1: 0,
				},
			},
			err: nil,
		},
		{
			name: "unexported field",
			src: `[example]
						key1 = 42`,
			dst: &struct {
				Example struct {
					key1 int `keyfile:"key1"`
				} `keyfile:"example"`
			}{},
			want: &struct {
				Example struct {
					key1 int `keyfile:"key1"`
				} `keyfile:"example"`
			}{
				Example: struct {
					key1 int `keyfile:"key1"`
				}{
					key1: 0,
				},
			},
			err: nil,
		},
		{
			name: "ignored field",
			src: `[example]
						key1 = 42`,
			dst: &struct {
				Example struct {
					Key1 int `keyfile:"-"`
				} `keyfile:"example"`
			}{},
			want: &struct {
				Example struct {
					Key1 int `keyfile:"-"`
				} `keyfile:"example"`
			}{
				Example: struct {
					Key1 int `keyfile:"-"`
				}{
					Key1: 0,
				},
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.src), tt.dst)
			if !errors.Is(err, tt.err) {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tt.dst, tt.want) {
				t.Fatal(tt.dst, tt.want)
			}
		})
	}
}
