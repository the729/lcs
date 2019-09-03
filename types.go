package lcs

type EnumVariant struct {
	// Name of the enum type. Different variants of a same enum type should have same name.
	Name string
	// Value of the enum variant. Should be unique within the same enum type.
	Value int32
	// Template object pointer for the enum variant.
	Template interface{}
}

type EnumTypeUser interface {
	EnumTypes() []EnumVariant
}
