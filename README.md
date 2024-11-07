# Keyfile parser for Go

Keyfile is simular to INI file, but it is much flexible. It supports maps and lists as built-in types.

You can find more information about Keyfile format in [https://docs.gtk.org/glib/struct.KeyFile.html](https://docs.gtk.org/glib/struct.KeyFile.html)

## Installation

```sh
go get -u github.com/ksckaan1/keyfile@latest
```

## Usage

### Example keyfile data:

```keyfile
[profile]
name=John Doe
age=42
is_working=true
hobbies=swimming;reading
job=Software Developer
job[de]=Softwareentwickler
job[sp]=Desarrollador de software
job[tr]=Yazılım Geliştirici
```

### Example Model

```go
type Config struct{
  Profile struct{
    Name      string            `keyfile:"name"`
    Age       int               `keyfile:"age"`
    IsWorking bool              `keyfile:"is_working"`
    Hobbies   []string          `keyfile:"hobbies"`
    Job       map[string]string `keyfile:"job"
  } `keyfile:"example"`
}
```

### Unmarshal
```go
err := keyfile.Unmarshal([]byte(data), &config)
if err != nil {
 // handle error
}
```

### Marshal
```go
data, err := keyfile.Marshal(config)
if err != nil {
 // handle error
}
```

## Supported Types

- string
- int, int8, int16, int32, int64
- uint, uint8, uint16, uint32, uint64
- float32, float64
- complex64, complex128
- bool
- slice of supported types
- map of supported types

## Working with Custom Types

Custom types can be used with keyfile parser. If custom type is underlying supported type, no need to do anything. If it is not, must be implement `Unmarshaler` or `Marshaler` interface like in std `json` package.

**Example:**

```go
type CustomType struct {
  value string
}

func (c *CustomType) UnmarshalJSON(b []byte) error {
  c.value = string(b)
  return nil
}

func (c *CustomType) MarshalJSON() ([]byte, error) {
  return []byte(c.value), nil
}
```

## Struct Tags

The `keyfile` tag is used for the struct tag. It is similar to the JSON struct tag.

```go
type Config struct{
  Example struct{
    // Key1 is ignored
    Key1       string   `keyfile:"-"`
    // ignore Key2 if it's zero value
    Key2       string   `keyfile:"key2,omitempty"`
    // split Key3 with comma (default is semicolon)
    Key3       []string `keyfile:"key3;sep:,"`
    // its unexported, so it will not be included in the keyfile
    unexported string
  } `keyfile:"example"`
}
```

## Licence

MIT