package lcs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
)

const (
	MaxByteSliceSize    = 100 * 1024 * 1024
	SliceAndMapInitSize = 100
)

type Decoder struct {
	//d []byte
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: r,
	}
}

func (d *Decoder) Decode(v interface{}) error {
	err := d.decode(reflect.Indirect(reflect.ValueOf(v)))
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

func (d *Decoder) decode(rv reflect.Value) (err error) {
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
		err = d.decodeSlice(rv)
	case reflect.Array:
		err = d.decodeArray(rv)
	case reflect.String:
		err = d.decodeString(rv)
	case reflect.Struct:
		err = d.decodeStruct(rv)
	case reflect.Map:
		err = d.decodeMap(rv)
	case reflect.Ptr:
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		err = d.decode(rv.Elem())
	default:
		err = errors.New("not supported")
	}
	return
}

func (d *Decoder) decodeByteSlice() (b []byte, err error) {
	l := uint32(0)
	if err = binary.Read(d.r, binary.LittleEndian, &l); err != nil {
		return
	}
	if l > MaxByteSliceSize {
		return nil, errors.New("byte slice longer than 100MB not supported")
	}
	b = make([]byte, l)
	if _, err = io.ReadFull(d.r, b); err != nil {
		return
	}
	return
}

func (d *Decoder) decodeSlice(rv reflect.Value) (err error) {
	if !rv.CanSet() {
		return errors.New("slice cannot set")
	}
	if rv.Type() == reflect.TypeOf([]byte{}) {
		var b []byte
		if b, err = d.decodeByteSlice(); err != nil {
			return
		}
		rv.SetBytes(b)
		return
	}

	l := uint32(0)
	if err = binary.Read(d.r, binary.LittleEndian, &l); err != nil {
		return
	}
	cap := int(l)
	if cap > SliceAndMapInitSize {
		cap = SliceAndMapInitSize
	}
	s := reflect.MakeSlice(rv.Type(), 0, cap)
	for i := 0; i < int(l); i++ {
		v := reflect.New(rv.Type().Elem())
		if err = d.decode(v); err != nil {
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
	if cap > SliceAndMapInitSize {
		cap = SliceAndMapInitSize
	}
	m := reflect.MakeMapWithSize(rv.Type(), cap)
	for i := 0; i < int(l); i++ {
		k := reflect.New(rv.Type().Key())
		v := reflect.New(rv.Type().Elem())
		if err = d.decode(k); err != nil {
			return
		}
		if err = d.decode(v); err != nil {
			return
		}
		m.SetMapIndex(k.Elem(), v.Elem())
	}
	rv.Set(m)
	return
}

func (d *Decoder) decodeArray(rv reflect.Value) (err error) {
	if !rv.CanSet() {
		return errors.New("array cannot set")
	}
	if rv.Type().Elem() == reflect.TypeOf(byte(0)) {
		var b []byte
		if b, err = d.decodeByteSlice(); err != nil {
			return
		}
		if len(b) != rv.Len() {
			return errors.New("length mismatch")
		}
		reflect.Copy(rv, reflect.ValueOf(b))
		return
	}

	l := uint32(0)
	if err = binary.Read(d.r, binary.LittleEndian, &l); err != nil {
		return
	}
	if int(l) != rv.Len() {
		return errors.New("length mismatch")
	}
	for i := 0; i < int(l); i++ {
		if err = d.decode(rv.Index(i)); err != nil {
			return
		}
	}
	return
}

func (d *Decoder) decodeString(rv reflect.Value) (err error) {
	if !rv.CanSet() {
		return errors.New("string cannot set")
	}
	var b []byte
	if b, err = d.decodeByteSlice(); err != nil {
		return
	}
	rv.SetString(string(b))
	return
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
		if fv.Kind() == reflect.Interface || fv.Kind() == reflect.Ptr {
			if rt.Field(i).Tag.Get(lcsTagName) == "optional" {
				rb := reflect.New(reflect.TypeOf(false))
				if err = d.decode(rb); err != nil {
					return
				}
				if !rb.Elem().Bool() {
					fv.Set(reflect.Zero(fv.Type()))
					continue
				}
			}
		}
		if err = d.decode(fv); err != nil {
			return
		}
	}
	return
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
