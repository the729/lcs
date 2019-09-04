# lcs

[![Build Status](https://travis-ci.org/the729/lcs.svg?branch=master)](https://travis-ci.org/the729/lcs)
[![codecov](https://codecov.io/gh/the729/lcs/branch/master/graph/badge.svg)](https://codecov.io/gh/the729/lcs)
[![Go Report Card](https://goreportcard.com/badge/github.com/the729/lcs)](https://goreportcard.com/report/github.com/the729/lcs)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/a70c457b8b7d44c0b69460b2a8704365)](https://www.codacy.com/app/the729/lcs?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=the729/lcs&amp;utm_campaign=Badge_Grade)

Go library for Libra canonical serialization (and deserialization). See [LCS Spec](https://github.com/libra/libra/tree/master/common/canonical_serialization).

## Installation

```bash
$ go get -u github.com/the729/lcs
```

## Usage

```golang
import "github.com/the729/lcs"
```

See [`example_test.go`](example_test.go) for complete examples.

### Basic types

You can serialize and deserialize the following basic types:
- bool
- int8, int16, int32, int64, uint8, uint16, uint32, uint64
- string
- slice, array, map

```golang
bytes, _ := lcs.Marshal("hello")

fmt.Printf("%x\n", bytes)
// Output: 050000068656c6c6f
```

```golang
myInt := int16(0)
lcs.Unmarshal([]byte{0x05, 0x00}, &myInt) // <- be careful to pass a pointer

fmt.Printf("%d\n", myInt)
// Output: 5
```

### Struct types

Simple struct can be serialized or deserialized directly. You can use struct field tags to change lcs behaviors.

```golang
type MyStruct struct {
    Boolean    bool
    Bytes      []byte
    Label      string `lcs:"-"` // "-" tagged field is ignored
    unexported uint32           // unexported field is ignored
}

// Serialize:
bytes, err := lcs.Marshal(&MyStruct{})

// Deserialize:
out := &MyStruct{}
err = lcs.Unmarshal(bytes, out)
```

### Struct with optional fields

Optional fields should be pointers with "optional" tag.

Slices, maps or interfaces can not be optional, unless they are wrapped in a wrapper struct, a pointer to which can be optional.

```golang
type MyStruct struct {
    Label  *string   `lcs:"optional"`
    Nested *MyStruct `lcs:"optional"`
}
```

### Enum types

Enum types are defined using interfaces. 

Enum types can only be struct fields, with "enum" tags. Stand-alone enum types are not supported.

```golang
// isOption is an enum type. It has a dummy function to distinguish its variants.
type isOption interface {
	isOption()
}

// *Option0, *Option1 and Option2 are variants of isOption
// Use pointers for struct types.
type Option0 struct {
	Data uint32
}
type Option1 struct {
	Data uint64
}
type Option2 bool
func (*Option0) isOption() {}
func (*Option1) isOption() {}
func (Option2) isOption()  {}

// MyStruct contains the enum type Option
type MyStruct struct {
    Name   string
    Option isOption     `lcs:"enum:option"` // tag in "enum:name" format
    List2D [][]isOption `lcs:"enum:option"` // support multi-dim slices
}

// EnumTypes implement lcs.EnumTypeUser. It returns the ingredients used for
// all enum types.
func (*Option) EnumTypes() []EnumVariant {
	return []EnumVariant{
		{
			Name:     "option",        // name should match the tags
			Value:    0,               // numeric value of this variant
			Template: (*Option0)(nil), // zero value of this variant type
		},
		{
			Name:     "option",
			Value:    1,
			Template: (*Option1)(nil),
		},
		{
			Name:     "option",
			Value:    2,
			Template: Option2(false),
		},
	}
}
```
