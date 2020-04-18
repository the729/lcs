package lcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Enum1 interface {
	isEnum1()
}

type Enum1Opt0 struct {
	Data uint32
}
type Enum1Opt1 bool
type Enum1Opt2 []byte
type Enum1Opt3 []Enum1

func (*Enum1Opt0) isEnum1() {}
func (Enum1Opt1) isEnum1()  {}
func (Enum1Opt2) isEnum1()  {}
func (Enum1Opt3) isEnum1()  {}

func TestInterfaceAsEnum(t *testing.T) {
	RegisterEnum((*Enum1)(nil),
		(*Enum1Opt0)(nil),
		Enum1Opt1(false),
		Enum1Opt2(nil),
		Enum1Opt3(nil),
	)
	e0 := Enum1(&Enum1Opt0{3})
	e1 := Enum1(Enum1Opt1(true))
	e2 := Enum1(Enum1Opt2([]byte{0x11, 0x22}))
	e3 := Enum1(Enum1Opt3([]Enum1{e0, e1}))
	runTest(t, []*testCase{
		{
			v:    &e0,
			b:    hexMustDecode("00 03000000"),
			name: "struct pointer",
		},
		{
			v:    &e1,
			b:    hexMustDecode("01 01"),
			name: "bool",
		},
		{
			v:    &e2,
			b:    hexMustDecode("02 02 1122"),
			name: "[]byte",
		},
		{
			v:    &e3,
			b:    hexMustDecode("03 02 00 03000000 01 01"),
			name: "enum slice of self",
		},
	})
}

func TestRegisterNonInterfacePtrShouldPanic(t *testing.T) {
	assert.Panics(t, func() {
		RegisterEnum(uint32(0))
	})
	assert.Panics(t, func() {
		RegisterEnum(Enum1(nil))
	})
	assert.Panics(t, func() {
		RegisterEnum((*Enum1Opt0)(nil))
	})
}

func TestRegisterWrongVariantShouldPanic(t *testing.T) {
	assert.Panics(t, func() {
		RegisterEnum((*Enum1)(nil), uint32(0))
	})
}
