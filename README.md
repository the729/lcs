# lcs

[![Build Status](https://travis-ci.org/the729/lcs.svg?branch=master)](https://travis-ci.org/the729/lcs)
[![codecov](https://codecov.io/gh/the729/lcs/branch/master/graph/badge.svg)](https://codecov.io/gh/the729/lcs)
[![Go Report Card](https://goreportcard.com/badge/github.com/the729/lcs)](https://goreportcard.com/report/github.com/the729/lcs)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/a70c457b8b7d44c0b69460b2a8704365)](https://www.codacy.com/app/the729/lcs?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=the729/lcs&amp;utm_campaign=Badge_Grade)

Go library for Libra canonical serialization (and deserialization). See [LCS Spec](https://github.com/libra/libra/tree/master/common/canonical_serialization).

For types defined and used in actual Libra blockchain, please visit [go-libra](https://github.com/the729/go-libra): Libra client library with crypto verifications.

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

Optional fields should be pointers, slices or maps with "optional" tag.

```golang
type MyStruct struct {
    Label  *string          `lcs:"optional"`
    Nested *MyStruct        `lcs:"optional"`
    Slice  []byte           `lcs:"optional"`
    Map    map[uint8]uint8  `lcs:"optional"`
}
```

### Fixed length lists

Fixed length lists are supported as member of structs with "len" tag.


```golang
type MyStruct struct {
	Bytes         []byte `lcs:"len=4"`
	OptionalBytes []byte `lcs:"len=4,optional"`
}
```

### Enum types

Enum types are defined using interfaces. 

Enum types can only be struct fields, with "enum" tags. Stand-alone enum types are not supported.

```golang
// isOption is an enum type. It has a dummy function to identify its variants.
type isOption interface {
	isOption()
}

// *Option0, Option1 and Option2 are variants of isOption
// Use pointer for (non-empty) struct.
type Option0 struct {
	Data uint32
}
type Option1 struct {} // Empty enum variant
type Option2 bool
// Variants should implement isOption
func (*Option0) isOption() {}
func (Option1) isOption() {}
func (Option2) isOption()  {}

// MyStruct contains the enum type Option
type MyStruct struct {
    Name   string
    Option isOption     `lcs:"enum=dummy"` // tag in "enum=name" format
    List2D [][]isOption `lcs:"enum=dummy"` // support multi-dim slices
}

// EnumTypes implement lcs.EnumTypeUser. It returns all the ingredients that can be 
// used for all enum fields in the receiver struct type.
func (*Option) EnumTypes() []EnumVariant {
	return []EnumVariant{
		{
			Name:     "dummy",         // name should match the tags
			Value:    0,               // numeric value of this variant
			Template: (*Option0)(nil), // zero value of this variant type
		},
		{
			Name:     "dummy",
			Value:    1,
			Template: Option1{},
		},
		{
			Name:     "dummy",
			Value:    2,
			Template: Option2(false),
		},
	}
}
```
