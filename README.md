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
- slice, map

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

Arrays are treated as fixed length lists.

You can also specify fixed length for struct members with `len` tag. Slices and strings are supported.


```golang
type MyStruct struct {
	Str           string `lcs:"len=2"`
	Bytes         []byte `lcs:"len=4"`
	OptionalBytes []byte `lcs:"len=4,optional"`
}
```

### Enum types

Enum types are golang interfaces.

(The [old enum API](https://github.com/the729/lcs/blob/v0.1.4/README.md#enum-types) is deprecated.)

```golang
// Enum1 is an enum type.
type Enum1 interface {
	isEnum1()	// optional: a dummy function to identify its variants.
}

// *Enum1Opt0, Enum1Opt1, Enum1Opt2, Enum1Opt3 are variants of Enum1
// Use pointer for non-empty struct.
// Use empty struct for a variant without contents.
type Enum1Opt0 struct {
	Data uint32
}
type Enum1Opt1 bool
type Enum1Opt2 []byte
type Enum1Opt3 []Enum1	// self reference is OK

// Variants should implement Enum1
func (*Enum1Opt0) isEnum1() {}
func (Enum1Opt1) isEnum1()  {}
func (Enum1Opt2) isEnum1()  {}
func (Enum1Opt3) isEnum1()  {}

// Register Enum1 with LCS. Will be available globaly.
var _ = lcs.RegisterEnum(
	// nil pointer to the enum interface type:
	(*Enum1)(nil),
	// zero-values of each variants
	(*Enum1Opt0)(nil),
	Enum1Opt1(false),
	Enum1Opt2(nil),
	Enum1Opt3(nil),
)

// Usage: Marshal the enum alone, must use pointer
e1 := Enum1(Enum1Opt1(true))
bytes, err := lcs.Marshal(&e1)

// Use Enum1 within other structs
type Wrapper struct {
	Enum Enum1
}
bytes, err := lcs.Marshal(&Wrapper{
	Enum: Enum1Opt1(true),
})

```
