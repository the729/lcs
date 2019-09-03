package lcs

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"strings"
)

const (
	lcsTagName = "lcs"
)

type Encoder struct {
	w     *bufio.Writer
	enums map[reflect.Type]map[string]map[reflect.Type]int32
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:     bufio.NewWriter(w),
		enums: make(map[reflect.Type]map[string]map[reflect.Type]int32),
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
		tag := rt.Field(i).Tag.Get(lcsTagName)
		if fv.Kind() == reflect.Interface && strings.HasPrefix(tag, "enum:") {
			enumName := tag[5:]
			evsAll, ok := e.enums[rv.Type()]
			if !ok {
				if evsAll = e.getEnumVariants(rv); evsAll != nil {
					e.enums[rv.Type()] = evsAll
				}
			}
			if evsAll == nil {
				return errors.New("enum variants not defined")
			}
			evs, ok := evsAll[enumName]
			if !ok {
				return errors.New("enum variants not defined for enum name: " + enumName)
			}
			fv = fv.Elem()
			if fv.Kind() == reflect.Ptr {
				fv = fv.Elem()
			}
			ev, ok := evs[fv.Type()]
			if !ok {
				return errors.New("enum variants not defined for type: " + fv.Elem().Type().String())
			}
			if err = binary.Write(e.w, binary.LittleEndian, ev); err != nil {
				return err
			}
		} else if fv.Kind() == reflect.Interface || fv.Kind() == reflect.Ptr {
			if tag == "optional" {
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

func (e *Encoder) getEnumVariants(rv reflect.Value) map[string]map[reflect.Type]int32 {
	vv, ok := rv.Interface().(EnumTypeUser)
	if !ok {
		vv, ok = reflect.New(reflect.PtrTo(rv.Type())).Elem().Interface().(EnumTypeUser)
		if !ok {
			return nil
		}
	}
	r := make(map[string]map[reflect.Type]int32)
	evs := vv.EnumTypes()
	for _, ev := range evs {
		evt := reflect.TypeOf(ev.Template)
		if evt.Kind() == reflect.Ptr {
			evt = evt.Elem()
		}
		if r[ev.Name] == nil {
			r[ev.Name] = make(map[reflect.Type]int32)
		}
		r[ev.Name][evt] = ev.Value
	}
	return r
}

func Marshal(v interface{}) ([]byte, error) {
	var b bytes.Buffer
	e := NewEncoder(&b)
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
