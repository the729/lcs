package lcs

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"sort"
	"strconv"
)

type Encoder struct {
	w     *bufio.Writer
	enums map[reflect.Type]map[string]map[reflect.Type]EnumKeyType
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:     bufio.NewWriter(w),
		enums: make(map[reflect.Type]map[string]map[reflect.Type]EnumKeyType),
	}
}

func (e *Encoder) Encode(v interface{}) error {
	if err := e.encode(reflect.Indirect(reflect.ValueOf(v)), nil, 0); err != nil {
		return err
	}
	e.w.Flush()
	return nil
}

func (e *Encoder) encode(rv reflect.Value, enumVariants map[reflect.Type]EnumKeyType, fixedLen int) (err error) {
	// rv = indirect(rv)
	switch rv.Kind() {
	case reflect.Bool,
		/*reflect.Int,*/ reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		/*reflect.Uint,*/ reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = binary.Write(e.w, binary.LittleEndian, rv.Interface())
	case reflect.Slice, reflect.Array, reflect.String:
		err = e.encodeSlice(rv, enumVariants, fixedLen)
	case reflect.Struct:
		err = e.encodeStruct(rv)
	case reflect.Map:
		err = e.encodeMap(rv)
	case reflect.Ptr:
		err = e.encode(rv.Elem(), enumVariants, 0)
	case reflect.Interface:
		err = e.encodeInterface(rv, enumVariants)
	default:
		err = errors.New("not supported kind: " + rv.Kind().String())
	}
	if err != nil {
		return err
	}
	return nil
}

func (e *Encoder) encodeSlice(rv reflect.Value, enumVariants map[reflect.Type]EnumKeyType, fixedLen int) (err error) {
	if rv.Kind() == reflect.Array {
		// ignore fixedLen
	} else if fixedLen == 0 {
		if _, err := writeVarUint(e.w, uint64(rv.Len())); err != nil {
			return err
		}
	} else if fixedLen != rv.Len() {
		return errors.New("actual len not equal to fixed len")
	}
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i)
		if err = e.encode(item, enumVariants, 0); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeInterface(rv reflect.Value, enumVariants map[reflect.Type]EnumKeyType) (err error) {
	if rv.IsNil() {
		return errors.New("non-optional enum value is nil")
	}

	ev, ok := enumGetIdxByType(rv.Type(), rv.Elem().Type())
	rvReal := rv.Elem()
	if !ok {
		ev, ok = enumVariants[rvReal.Type()]
		if !ok {
			return errors.New("enum " + rv.Type().String() + " does not have variant of type " + rvReal.Type().String())
		}
	}
	if _, err = writeVarUint(e.w, ev); err != nil {
		return
	}
	if err = e.encode(rvReal, nil, 0); err != nil {
		return err
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
		tag := parseTag(rt.Field(i).Tag.Get(lcsTagName))

		var evs map[reflect.Type]EnumKeyType
		if enumName, ok := tag["enum"]; ok {
			evsAll, ok := e.enums[rv.Type()]
			if !ok {
				if evsAll = e.getEnumVariants(rv); evsAll != nil {
					e.enums[rv.Type()] = evsAll
				}
			}
			if evsAll == nil {
				return errors.New("enum variants not defined")
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
			if err = e.encode(reflect.ValueOf(!fv.IsNil()), nil, 0); err != nil {
				return err
			}
			if fv.IsNil() {
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
		if err = e.encode(fv, evs, fixedLen); err != nil {
			return
		}
	}
	return nil
}

func (e *Encoder) encodeMap(rv reflect.Value) (err error) {
	_, err = writeVarUint(e.w, uint64(rv.Len()))
	if err != nil {
		return err
	}

	keys := make([]string, 0, rv.Len())
	marshaledMap := make(map[string][]byte)
	for iter := rv.MapRange(); iter.Next(); {
		k := iter.Key()
		v := iter.Value()
		kb, err := Marshal(k.Interface())
		if err != nil {
			return err
		}
		vb, err := Marshal(v.Interface())
		if err != nil {
			return err
		}
		keys = append(keys, string(kb))
		marshaledMap[string(kb)] = vb
	}

	sort.Strings(keys)
	for _, k := range keys {
		e.w.Write([]byte(k))
		e.w.Write(marshaledMap[k])
	}

	return nil
}

func (e *Encoder) getEnumVariants(rv reflect.Value) map[string]map[reflect.Type]EnumKeyType {
	vv, ok := rv.Interface().(EnumTypeUser)
	if !ok {
		vv, ok = reflect.New(reflect.PtrTo(rv.Type())).Elem().Interface().(EnumTypeUser)
		if !ok {
			return nil
		}
	}
	r := make(map[string]map[reflect.Type]EnumKeyType)
	evs := vv.EnumTypes()
	for _, ev := range evs {
		evt := reflect.TypeOf(ev.Template)
		if r[ev.Name] == nil {
			r[ev.Name] = make(map[reflect.Type]EnumKeyType)
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
