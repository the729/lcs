package lcs

const (
	lcsTagName = "lcs"

	// maxByteSliceSize is the maximum size of byte slice that can be decoded.
	//
	// When decoding a byte slice, we will get the length first, and then we will allocate
	// enough space according to the length. We don't want a wrong length leads to out-of-memory
	// error. So maxByteSliceSize is a hard limit.
	//
	// It is set to 100MB by default.
	maxByteSliceSize = 100 * 1024 * 1024

	// sliceAndMapInitSize is the initial allocation size for non-byte slices.
	//
	// When decoding a non-byte slice, we will allocate an initial space, and then append to it.
	sliceAndMapInitSize = 100
)

// EnumVariant is a definition of a variant of enum type.
type EnumVariant struct {
	// Name of the enum type. Different variants of a same enum type should have same name.
	// This name should match the name defined in the struct field tag.
	Name string

	// Value is the numeric value of the enum variant. Should be unique within the same enum type.
	Value int32

	// Template object for the enum variant. Should be the zero value of the variant type.
	//
	// Example values: (*SomeStruct)(nil), MyUint32(0).
	Template interface{}
}

// EnumTypeUser is an interface of struct with enum type definition. Struct with enum type should
// implement this interface.
type EnumTypeUser interface {
	// EnumTypes return the ingredients used for all enum types in the struct.
	EnumTypes() []EnumVariant
}
