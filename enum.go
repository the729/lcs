package lcs

import (
	"reflect"
)

var regEnumTypeToIdx map[reflect.Type]map[reflect.Type]EnumKeyType
var regEnumIdxToType map[reflect.Type][]reflect.Type

// RegisterEnum register an enum type with its available variants. If the enum type
// was registered, it will be overwriten.
//
// This function panics on errors. The returned error is always nil.
func RegisterEnum(enumTypePtr interface{}, types ...interface{}) (err error) {
	rEnumType := reflect.TypeOf(enumTypePtr)
	if rEnumType.Kind() != reflect.Ptr {
		panic("enumType should be a pointer to a nil interface")
	}
	rEnumType = rEnumType.Elem()
	if rEnumType.Kind() != reflect.Interface {
		panic("enumType should be a pointer to a nil interface")
	}
	if regEnumTypeToIdx == nil {
		regEnumTypeToIdx = make(map[reflect.Type]map[reflect.Type]uint64)
	}
	if regEnumIdxToType == nil {
		regEnumIdxToType = make(map[reflect.Type][]reflect.Type)
	}
	regEnumIdxToType[rEnumType] = make([]reflect.Type, 0, len(types))
	regEnumTypeToIdx[rEnumType] = make(map[reflect.Type]uint64)
	for i, t := range types {
		rType := reflect.TypeOf(t)
		if !rType.Implements(rEnumType) {
			panic(rType.String() + " does not implement " + rEnumType.String())
		}
		regEnumIdxToType[rEnumType] = append(regEnumIdxToType[rEnumType], rType)
		regEnumTypeToIdx[rEnumType][rType] = EnumKeyType(i)
	}
	// log.Printf("registered: %v", rEnumType.String())
	return
}

func enumGetTypeByIdx(enumType reflect.Type, idx EnumKeyType) (reflect.Type, bool) {
	// log.Printf("enumGetTypeByIdx: %v", enumType.String())
	m, ok := regEnumIdxToType[enumType]
	if !ok {
		return nil, false
	}
	if len(m) <= int(idx) {
		return nil, false
	}
	return m[int(idx)], true
}

func enumGetIdxByType(enumType, vType reflect.Type) (EnumKeyType, bool) {
	// log.Printf("enumGetIdxByType: %v", enumType.String())
	m, ok := regEnumTypeToIdx[enumType]
	if !ok {
		return 0, false
	}
	idx, ok := m[vType]
	return idx, ok
}
