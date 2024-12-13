package analyzer

import "reflect"

// TypeMapping maps Go types to C types.
var TypeMapping = map[reflect.Kind]string{
	reflect.Int:     "int",
	reflect.Int8:    "int8_t",
	reflect.Int16:   "int16_t",
	reflect.Int32:   "int32_t",
	reflect.Int64:   "int64_t",
	reflect.Uint:    "unsigned int",
	reflect.Uint8:   "uint8_t",
	reflect.Uint16:  "uint16_t",
	reflect.Uint32:  "uint32_t",
	reflect.Uint64:  "uint64_t",
	reflect.Float32: "float",
	reflect.Float64: "double",
	reflect.Bool:    "bool",
	reflect.String:  "char*",
}

type FieldInfo struct {
	Name   string // JSON tag or field name
	GoName string // Original Go field name
	Type   reflect.Type
	Offset uintptr
	CType  string // Mapped C type
	Kind   string // Kind as string, e.g., "String", "Int", "Bool"
}

// CStruct represents a C struct with its name and fields
type CStruct struct {
	Name   string
	Fields []FieldInfo
}
