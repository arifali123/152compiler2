package analyzer

import (
	"errors"
	"reflect"
)

type FieldInfo struct {
	Name   string // JSON tag or field name
	GoName string // Original Go field name
	Type   reflect.Type
	Offset uintptr
	CType  string // Mapped C type
	Kind   string // Kind as string, e.g., "String", "Int", "Bool"
}

// AnalyzeStruct analyzes a Go struct type and returns information about its fields.
func AnalyzeStruct(t reflect.Type) ([]FieldInfo, error) {
	if t.Kind() != reflect.Struct {
		return nil, errors.New("AnalyzeStruct: provided type is not a struct")
	}

	var fields []FieldInfo
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		}

		cType, ok := TypeMapping[field.Type.Kind()]
		if !ok {
			return nil, errors.New("unsupported field type: " + field.Type.Kind().String())
		}

		fields = append(fields, FieldInfo{
			Name:   jsonTag,
			GoName: field.Name,
			Type:   field.Type,
			Offset: field.Offset,
			CType:  cType,
			Kind:   field.Type.Kind().String(),
		})
	}

	return fields, nil
}
