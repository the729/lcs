package lcs

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
)

const (
	lcsTagName = "lcs"
)

type Encoder struct {
	w *bufio.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: bufio.NewWriter(w),
	}
}

func (e *Encoder) Encode(v interface{}) error {
	if err := e.encode(reflect.Indirect(reflect.ValueOf(v))); err != nil {
		return err
	}
	e.w.Flush()
	return nil
}

func (e *Encoder) encode(rv reflect.Value) (err error) {
	// rv = indirect(rv)
	switch rv.Kind() {
	case reflect.Bool,
		/*reflect.Int,*/ reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		/*reflect.Uint,*/ reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = binary.Write(e.w, binary.LittleEndian, rv.Interface())
	case reflect.Slice, reflect.Array, reflect.String:
		err = e.encodeSlice(rv)
	case reflect.Struct:
		err = e.encodeStruct(rv)
	case reflect.Map:
		err = e.encodeMap(rv)
	default:
		err = errors.New("not supported")
	}
	if err != nil {
		return err
	}
	return nil
}

func (e *Encoder) encodeSlice(rv reflect.Value) (err error) {
	if err = binary.Write(e.w, binary.LittleEndian, uint32(rv.Len())); err != nil {
		return err
	}
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		if err = e.encode(item); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeStruct(rv reflect.Value) (err error) {
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		fv := rv.Field(i)
		if !fv.CanInterface() {
			continue
		}
		if rt.Field(i).Tag.Get(lcsTagName) == "-" {
			continue
		}
		if fv.Kind() == reflect.Interface || fv.Kind() == reflect.Ptr {
			if rt.Field(i).Tag.Get(lcsTagName) == "optional" {
				if err = e.encode(reflect.ValueOf(!fv.IsNil())); err != nil {
					return err
				}
				if fv.IsNil() {
					continue
				}
			}
			fv = fv.Elem()
		}
		if err = e.encode(fv); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeMap(rv reflect.Value) (err error) {
	err = binary.Write(e.w, binary.LittleEndian, uint32(rv.Len()))
	if err != nil {
		return err
	}
	for iter := rv.MapRange(); iter.Next(); {
		k := iter.Key()
		v := iter.Value()
		if err = e.encode(k); err != nil {
			return err
		}
		if err = e.encode(v); err != nil {
			return err
		}
	}
	return nil
}

// func indirect(rv reflect.Value) reflect.Value {
// 	switch rv.Kind() {
// 	case reflect.Ptr, reflect.Interface:
// 		return indirect(rv.Elem())
// 	default:
// 		return rv
// 	}
// }

func Marshal(v interface{}) ([]byte, error) {
	var b bytes.Buffer
	e := NewEncoder(&b)
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
