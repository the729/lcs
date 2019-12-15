package lcs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type Decoder struct {
	r     io.Reader
	enums map[reflect.Type]map[string]map[int32]reflect.Type
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:     r,
		enums: make(map[reflect.Type]map[string]map[int32]reflect.Type),
	}
}

func (d *Decoder) Decode(v interface{}) error {
	err := d.decode(reflect.Indirect(reflect.ValueOf(v)), nil, 0)
	if err != nil {
		return err
	}
	return nil
}

func (d *Decoder) EOF() bool {
	_, err := d.r.Read(make([]byte, 1))
	if err == io.EOF {
		return true
	}
	return false
}

func (d *Decoder) decode(rv reflect.Value, enumVariants map[int32]reflect.Type, fixedLen int) (err error) {
	switch rv.Kind() {
	case reflect.Bool:
		if !rv.CanSet() {
			return errors.New("bool value cannot set")
		}
		v8 := uint8(0)
		if err = binary.Read(d.r, binary.LittleEndian, &v8); err != nil {
			return
		}
		if v8 == 1 {
			rv.SetBool(true)
		} else if v8 == 0 {
			rv.SetBool(false)
		} else {
			return errors.New("unexpected value for bool")
		}
	case /*reflect.Int,*/ reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		/*reflect.Uint,*/ reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if !rv.CanSet() {
			return errors.New("integer value cannot set")
		}
		err = binary.Read(d.r, binary.LittleEndian, rv.Addr().Interface())
	case reflect.Slice:
		err = d.decodeSlice(rv, enumVariants, fixedLen)
	case reflect.Array:
		err = d.decodeArray(rv, enumVariants, fixedLen)
	case reflect.String:
		err = d.decodeString(rv, fixedLen)
	case reflect.Struct:
		err = d.decodeStruct(rv)
	case reflect.Map:
		err = d.decodeMap(rv)
	case reflect.Ptr:
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		err = d.decode(rv.Elem(), enumVariants, fixedLen)
	case reflect.Interface:
		err = d.decodeInterface(rv, enumVariants)
	default:
		err = errors.New("not supported kind: " + rv.Kind().String())
	}
	return
}

func (d *Decoder) decodeByteSlice(fixedLen int) (b []byte, err error) {
	l := uint32(fixedLen)
	if l == 0 {
		if err = binary.Read(d.r, binary.LittleEndian, &l); err != nil {
			return
		}
		if l > maxByteSliceSize {
			return nil, errors.New("byte slice longer than 100MB not supported")
		}
	}
	b = make([]byte, l)
	if _, err = io.ReadFull(d.r, b); err != nil {
		return
	}
	return
}

func (d *Decoder) decodeSlice(rv reflect.Value, enumVariants map[int32]reflect.Type, fixedLen int) (err error) {
	if !rv.CanSet() {
		return errors.New("slice cannot set")
	}
	if rv.Type() == reflect.TypeOf([]byte{}) {
		var b []byte
		if b, err = d.decodeByteSlice(fixedLen); err != nil {
			return
		}
		rv.SetBytes(b)
		return
	}

	l := uint32(fixedLen)
	var cap int
	if l == 0 {
		if err = binary.Read(d.r, binary.LittleEndian, &l); err != nil {
			return
		}
		cap = int(l)
		if cap > sliceAndMapInitSize {
			cap = sliceAndMapInitSize
		}
	}
	s := reflect.MakeSlice(rv.Type(), 0, cap)
	for i := 0; i < int(l); i++ {
		v := reflect.New(rv.Type().Elem())
		if err = d.decode(v.Elem(), enumVariants, 0); err != nil {
			return
		}
		s = reflect.Append(s, v.Elem())
	}
	rv.Set(s)
	return
}

func (d *Decoder) decodeMap(rv reflect.Value) (err error) {
	if !rv.CanSet() {
		return errors.New("map cannot set")
	}

	l := uint32(0)
	if err = binary.Read(d.r, binary.LittleEndian, &l); err != nil {
		return
	}
	cap := int(l)
	if cap > sliceAndMapInitSize {
		cap = sliceAndMapInitSize
	}
	m := reflect.MakeMapWithSize(rv.Type(), cap)
	for i := 0; i < int(l); i++ {
		k := reflect.New(rv.Type().Key())
		v := reflect.New(rv.Type().Elem())
		if err = d.decode(k, nil, 0); err != nil {
			return
		}
		if err = d.decode(v, nil, 0); err != nil {
			return
		}
		m.SetMapIndex(k.Elem(), v.Elem())
	}
	rv.Set(m)
	return
}

func (d *Decoder) decodeArray(rv reflect.Value, enumVariants map[int32]reflect.Type, fixedLen int) (err error) {
	if !rv.CanSet() {
		return errors.New("array cannot set")
	}
	if rv.Kind() == reflect.Array {
		fixedLen = rv.Len()
	}
	if rv.Type().Elem() == reflect.TypeOf(byte(0)) {
		var b []byte
		if b, err = d.decodeByteSlice(fixedLen); err != nil {
			return
		}
		if len(b) != rv.Len() {
			return errors.New("length mismatch")
		}
		reflect.Copy(rv, reflect.ValueOf(b))
		return
	}

	l := uint32(fixedLen)
	if l == 0 {
		if err = binary.Read(d.r, binary.LittleEndian, &l); err != nil {
			return
		}
	}
	if int(l) != rv.Len() {
		return errors.New("length mismatch")
	}
	for i := 0; i < int(l); i++ {
		if err = d.decode(rv.Index(i), enumVariants, 0); err != nil {
			return
		}
	}
	return
}

func (d *Decoder) decodeString(rv reflect.Value, fixedLen int) (err error) {
	if !rv.CanSet() {
		return errors.New("string cannot set")
	}
	var b []byte
	if b, err = d.decodeByteSlice(fixedLen); err != nil {
		return
	}
	rv.SetString(string(b))
	return
}

func (d *Decoder) decodeInterface(rv reflect.Value, enumVariants map[int32]reflect.Type) (err error) {
	if enumVariants == nil {
		return errors.New("enum variants not defined for interface: " + rv.Type().String())
	}
	typeVal := int32(0)
	if err = binary.Read(d.r, binary.LittleEndian, &typeVal); err != nil {
		return
	}
	tpl, ok := enumVariants[typeVal]
	if !ok {
		return fmt.Errorf("enum variant value %d unknown for interface: %s", typeVal, rv.Type())
	}
	if tpl.Kind() == reflect.Ptr {
		rv1 := reflect.New(tpl.Elem())
		if err = d.decode(rv1, nil, 0); err != nil {
			return
		}
		rv.Set(rv1)
	} else {
		rv1 := reflect.New(tpl)
		if err = d.decode(rv1, nil, 0); err != nil {
			return
		}
		rv.Set(rv1.Elem())
	}
	return nil
}

func (d *Decoder) decodeStruct(rv reflect.Value) (err error) {
	if !rv.CanSet() {
		return errors.New("struct cannot set")
	}
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		fv := rv.Field(i)
		if !fv.CanSet() {
			continue
		}
		if rt.Field(i).Tag.Get(lcsTagName) == "-" {
			continue
		}
		tag := parseTag(rt.Field(i).Tag.Get(lcsTagName))
		var evs map[int32]reflect.Type
		if enumName, ok := tag["enum"]; ok {
			evsAll, ok := d.enums[rv.Type()]
			if !ok {
				if evsAll = d.getEnumVariants(rv); evsAll != nil {
					d.enums[rv.Type()] = evsAll
				}
			}
			if evsAll == nil {
				return fmt.Errorf("struct (%s) does not implement EnumTypeUser", rv.Type())
			}
			evs, ok = evsAll[enumName]
			if !ok {
				return errors.New("enum variants not defined for enum name: " + enumName)
			}
		}

		if _, ok := tag["optional"]; ok &&
			(fv.Kind() == reflect.Ptr ||
				fv.Kind() == reflect.Slice ||
				fv.Kind() == reflect.Map ||
				fv.Kind() == reflect.Interface) {
			rb := reflect.New(reflect.TypeOf(false))
			if err = d.decode(rb, nil, 0); err != nil {
				return
			}
			if !rb.Elem().Bool() {
				fv.Set(reflect.Zero(fv.Type()))
				continue
			}
		}
		fixedLen := 0
		if fixedLenStr, ok := tag["len"]; ok && (fv.Kind() == reflect.Slice || fv.Kind() == reflect.String) {
			fixedLen, err = strconv.Atoi(fixedLenStr)
			if err != nil {
				return errors.New("tag len parse error: " + err.Error())
			}
		}
		if err = d.decode(fv, evs, fixedLen); err != nil {
			return
		}
	}
	return
}

func (d *Decoder) getEnumVariants(rv reflect.Value) map[string]map[int32]reflect.Type {
	vv, ok := rv.Interface().(EnumTypeUser)
	if !ok {
		vv, ok = reflect.New(reflect.PtrTo(rv.Type())).Elem().Interface().(EnumTypeUser)
		if !ok {
			return nil
		}
	}
	r := make(map[string]map[int32]reflect.Type)
	evs := vv.EnumTypes()
	for _, ev := range evs {
		evt := reflect.TypeOf(ev.Template)
		if r[ev.Name] == nil {
			r[ev.Name] = make(map[int32]reflect.Type)
		}
		r[ev.Name][ev.Value] = evt
	}
	return r
}

func Unmarshal(data []byte, v interface{}) error {
	d := NewDecoder(bytes.NewReader(data))
	if err := d.Decode(v); err != nil {
		return err
	}
	if !d.EOF() {
		return errors.New("unexpected data")
	}
	return nil
}
