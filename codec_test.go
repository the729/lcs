package lcs

import (
	"encoding/hex"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	v             interface{}
	b             []byte
	skipMarshal   bool
	skipUnmarshal bool
	errMarshal    error
	errUnmarshal  error
	name          string
}

func runTest(t *testing.T, cases []*testCase) {
	var b []byte
	var err error

	for idx, c := range cases {
		var name string
		if c.name == "" {
			name = strconv.Itoa(idx)
		} else {
			name = c.name
		}
		if !c.skipMarshal {
			t.Run(name+"_marshal", func(t *testing.T) {
				b, err = Marshal(c.v)
				if c.errMarshal != nil {
					assert.EqualError(t, err, c.errMarshal.Error())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, c.b, b)
				}
				// t.Logf("Case #%d(%s) marshal: Done", idx, c.name)
			})
		}
		if !c.skipUnmarshal {
			t.Run(name+"_unmarshal", func(t *testing.T) {
				v := reflect.New(reflect.TypeOf(c.v))
				err = Unmarshal(c.b, v.Interface())
				if c.errUnmarshal != nil {
					assert.EqualError(t, err, c.errUnmarshal.Error())
				} else {
					assert.NoError(t, err)
					assert.Equal(t, c.v, v.Elem().Interface())
				}
				// t.Logf("Case #%d(%s) unmarshal: Done", idx, c.name)
			})
		}
	}
}

func hexMustDecode(s string) []byte {
	b, err := hex.DecodeString(strings.ReplaceAll(s, " ", ""))
	if err != nil {
		panic(err)
	}
	return b
}

func TestBool(t *testing.T) {
	vTrue := true
	type Bool bool
	vTrue2 := Bool(true)

	runTest(t, []*testCase{
		{
			v:    bool(true),
			b:    []byte{1},
			name: "bool true",
		},
		{
			v:    bool(false),
			b:    []byte{0},
			name: "bool false",
		},
		{
			v:    &vTrue,
			b:    []byte{1},
			name: "ptr to bool",
		},
		{
			v:    &vTrue2,
			b:    []byte{1},
			name: "ptr to alias of bool",
		},
	})
}

func TestInts(t *testing.T) {
	runTest(t, []*testCase{
		{
			v:    int8(-1),
			b:    []byte{0xFF},
			name: "int8 neg",
		},
		{
			v:    uint8(1),
			b:    []byte{1},
			name: "uint8 pos",
		},
		{
			v:    int16(-4660),
			b:    hexMustDecode("CCED"),
			name: "int16 neg",
		},
		{
			v:    uint16(4660),
			b:    hexMustDecode("3412"),
			name: "uint16 pos",
		},
		{
			v:    int32(-305419896),
			b:    hexMustDecode("88A9CBED"),
			name: "int32 neg",
		},
		{
			v:    uint32(305419896),
			b:    hexMustDecode("78563412"),
			name: "uint32 pos",
		},
		{
			v:    int64(-1311768467750121216),
			b:    hexMustDecode("0011325487A9CBED"),
			name: "int64 neg",
		},
		{
			v:    uint64(1311768467750121216),
			b:    hexMustDecode("00EFCDAB78563412"),
			name: "uint64 pos",
		},
	})
}

func TestBasicSlice(t *testing.T) {
	runTest(t, []*testCase{
		{
			v:    []byte{0x11, 0x22, 0x33, 0x44, 0x55},
			b:    hexMustDecode("05000000 11 22 33 44 55"),
			name: "byte slice",
		},
		{
			v:    []uint16{0x11, 0x22},
			b:    hexMustDecode("02000000 1100 2200"),
			name: "uint16 slice",
		},
		{
			v:    "ሰማይ አይታረስ ንጉሥ አይከሰስ።",
			b:    hexMustDecode("36000000E188B0E1889BE18BAD20E18AA0E18BADE189B3E188A8E188B520E18A95E18C89E188A520E18AA0E18BADE18AA8E188B0E188B5E18DA2"),
			name: "utf8 string",
		},
	})
}

func Test2DSlice(t *testing.T) {
	runTest(t, []*testCase{
		{
			v:    [][]byte{{0x01, 0x02}, {0x11, 0x12, 0x13}, {0x21}},
			b:    hexMustDecode("03000000 02000000 0102 03000000 111213 01000000 21"),
			name: "2d byte slice",
		},
		{
			v:    []string{"hello", "world"},
			b:    hexMustDecode("02000000 05000000 68656c6c6f 05000000 776f726c64"),
			name: "string slice",
		},
	})
}

func TestBasicStruct(t *testing.T) {
	type MyStruct struct {
		Boolean    bool
		Bytes      []byte
		Label      string
		unexported uint32
	}
	type Wrapper struct {
		Inner *MyStruct
		Name  string
	}

	runTest(t, []*testCase{
		{
			v:    struct{}{},
			b:    nil,
			name: "empty struct",
		},
		{
			v: MyStruct{
				Boolean: true,
				Bytes:   []byte{0x11, 0x22},
				Label:   "hello",
			},
			b:    hexMustDecode("01 02000000 11 22 05000000 68656c6c6f"),
			name: "struct with unexported fields",
		},
		{
			v: &MyStruct{
				Boolean: true,
				Bytes:   []byte{0x11, 0x22},
				Label:   "hello",
			},
			b:    hexMustDecode("01 02000000 11 22 05000000 68656c6c6f"),
			name: "pointer to struct",
		},
		{
			v: Wrapper{
				Inner: &MyStruct{
					Boolean: true,
					Bytes:   []byte{0x11, 0x22},
					Label:   "hello",
				},
				Name: "world",
			},
			b:    hexMustDecode("01 02000000 11 22 05000000 68656c6c6f 05000000 776f726c64"),
			name: "nested struct",
		},
		{
			v: &Wrapper{
				Inner: &MyStruct{
					Boolean: true,
					Bytes:   []byte{0x11, 0x22},
					Label:   "hello",
				},
				Name: "world",
			},
			b:    hexMustDecode("01 02000000 11 22 05000000 68656c6c6f 05000000 776f726c64"),
			name: "pointer to nested struct",
		},
	})
}

func TestStructWithFixedLenMember(t *testing.T) {
	type MyStruct struct {
		Str           string `lcs:"len=2"`
		Bytes         []byte `lcs:"len=4"`
		OptionalBytes []byte `lcs:"len=4,optional"`
	}

	runTest(t, []*testCase{
		{
			v: MyStruct{
				Str:   "12",
				Bytes: []byte{0x11, 0x22},
			},
			errMarshal:    errors.New("actual len not equal to fixed len"),
			name:          "struct with wrong fixed len (bytes)",
			skipUnmarshal: true,
		},
		{
			v: MyStruct{
				Str:   "",
				Bytes: []byte{0x11, 0x22, 0x33, 0x44},
			},
			errMarshal:    errors.New("actual len not equal to fixed len"),
			name:          "struct with wrong fixed len (string)",
			skipUnmarshal: true,
		},
		{
			v: MyStruct{
				Str:   "12",
				Bytes: []byte{0x11, 0x22, 0x33, 0x44},
			},
			b:    hexMustDecode("31 32 11223344 00"),
			name: "struct with fixed len",
		},
		{
			v: MyStruct{
				Str:           "12",
				Bytes:         []byte{0x11, 0x22, 0x33, 0x44},
				OptionalBytes: []byte{0x55, 0x66, 0x77, 0x88},
			},
			b:    hexMustDecode("3132 11223344 01 55667788"),
			name: "struct with optional fixed len",
		},
	})
}

func TestArray(t *testing.T) {
	runTest(t, []*testCase{
		{
			v:    [4]byte{0x11, 0x22, 0x33, 0x44},
			b:    hexMustDecode("11223344"),
			name: "byte array",
		},
		{
			v:    [2]uint32{0x11, 0x22},
			b:    hexMustDecode("11000000 22000000"),
			name: "uint32 array",
		},
	})
}

func TestRecursiveStruct(t *testing.T) {
	type StructTag struct {
		Address    []byte
		Module     string
		Name       string
		TypeParams []*StructTag
	}

	runTest(t, []*testCase{
		{
			v: &StructTag{
				Address:    []byte{0x11, 0x22},
				Module:     "hello",
				Name:       "world",
				TypeParams: []*StructTag{},
			},
			b:    hexMustDecode("02000000 11 22 05000000 68656c6c6f 05000000 776f726c64 00000000"),
			name: "recursive struct",
		},
	})
}

func TestOptional(t *testing.T) {
	type Wrapper struct {
		Ignored int     `lcs:"-"`
		Name    *string `lcs:"optional"`
	}
	hello := "hello"

	type Wrapper2 struct {
		Slice []byte `lcs:"optional"`
	}
	type Wrapper3 struct {
		Map map[uint8]uint8 `lcs:"optional"`
	}

	runTest(t, []*testCase{
		{
			v: Wrapper{
				Name: &hello,
			},
			b:    hexMustDecode("01 05000000 68656c6c6f"),
			name: "struct with set optional fields",
		},
		{
			v:    Wrapper{},
			b:    hexMustDecode("00"),
			name: "struct with unset optional fields",
		},
		{
			v: Wrapper2{
				Slice: []byte(hello),
			},
			b:    hexMustDecode("01 05000000 68656c6c6f"),
			name: "struct with set optional slice",
		},
		{
			v:    Wrapper2{},
			b:    hexMustDecode("00"),
			name: "struct with unset optional slice",
		},
		{
			v: Wrapper3{
				Map: map[uint8]uint8{1: 2},
			},
			b:    hexMustDecode("01 01000000 01 02"),
			name: "struct with set optional map",
		},
		{
			v:    Wrapper3{},
			b:    hexMustDecode("00"),
			name: "struct with unset optional map",
		},
	})
}

func TestMap(t *testing.T) {
	runTest(t, []*testCase{
		{
			v:    map[uint8]string{1: "hello", 2: "world"},
			b:    hexMustDecode("02000000 01 05000000 68656c6c6f 02 05000000 776f726c64"),
			name: "map[uint8]string",
		},
		{
			v:    map[string]uint8{"hello": 1, "world": 2},
			b:    hexMustDecode("02000000 05000000 68656c6c6f 01 05000000 776f726c64 02"),
			name: "map[string]uint8",
		},
	})
}

type Option0 struct {
	Data uint32
}
type Option1 struct{}
type Option2 bool
type isOption interface {
	isOption()
}
type Option struct {
	Option isOption `lcs:"enum=option"`
}
type OptionalOption struct {
	Option isOption `lcs:"optional,enum=option"`
}

func (*Option0) isOption() {}
func (Option1) isOption()  {}
func (Option2) isOption()  {}

var optionEnumDef = []EnumVariant{
	{
		Name:     "option",
		Value:    0,
		Template: (*Option0)(nil),
	},
	{
		Name:     "option",
		Value:    1,
		Template: Option1{},
	},
	{
		Name:     "option",
		Value:    2,
		Template: Option2(false),
	},
}

func (*Option) EnumTypes() []EnumVariant         { return optionEnumDef }
func (*OptionalOption) EnumTypes() []EnumVariant { return optionEnumDef }

func TestEnum(t *testing.T) {
	runTest(t, []*testCase{
		{
			v: &Option{
				Option: &Option0{5},
			},
			b:    hexMustDecode("0000 0000 0500 0000"),
			name: "ptr to struct with ptr enum variant",
		},
		{
			v: &Option{
				Option: Option1{},
			},
			b:    hexMustDecode("0100 0000"),
			name: "ptr to struct with non-ptr empty variant",
		},
		{
			v: &Option{
				Option: Option2(true),
			},
			b:    hexMustDecode("0200 0000 01"),
			name: "ptr to struct with real value as enum variant",
		},
		{
			v: Option{
				Option: Option1{},
			},
			b:    hexMustDecode("0100 0000"),
			name: "non-ptr struct with variant",
		},
		{
			v:    OptionalOption{},
			name: "nil enum on optional field",
			b:    hexMustDecode("00"),
		},
		{
			v:             Option{},
			name:          "nil variant on non-optional field",
			skipUnmarshal: true,
			errMarshal:    errors.New("non-optional enum value is nil"),
		},
	})
}

type Wrapper struct {
	Option []isOption `lcs:"enum=option"`
}

func (*Wrapper) EnumTypes() []EnumVariant {
	return []EnumVariant{
		{
			Name:     "option",
			Value:    0,
			Template: (*Option0)(nil),
		},
		{
			Name:     "option",
			Value:    1,
			Template: Option1{},
		},
		{
			Name:     "option",
			Value:    2,
			Template: Option2(false),
		},
	}
}

func TestEnumSlice(t *testing.T) {
	runTest(t, []*testCase{
		{
			v: &Wrapper{
				Option: []isOption{
					&Option0{5},
					Option1{},
					Option2(true),
				},
			},
			b:    hexMustDecode("03000000 00000000 05000000 01000000 02000000 01"),
			name: "enum slice",
		},
	})
}

type Wrapper2 struct {
	Option [][]isOption `lcs:"enum=option"`
}

func (*Wrapper2) EnumTypes() []EnumVariant {
	return []EnumVariant{
		{
			Name:     "option",
			Value:    0,
			Template: (*Option0)(nil),
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

func TestEnum2DSlice(t *testing.T) {
	runTest(t, []*testCase{
		{
			v: &Wrapper2{
				Option: [][]isOption{
					{
						&Option0{5},
						Option2(true),
					},
					{
						Option2(false),
					},
				},
			},
			b:    hexMustDecode("02000000 02000000 00000000 05000000 02000000 01 01000000 02000000 00"),
			name: "2D enum slice",
		},
	})
}
