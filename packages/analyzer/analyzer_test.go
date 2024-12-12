package analyzer

import (
	"reflect"
	"testing"
)

type Sample struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Active  bool    `json:"active"`
	Balance float64 `json:"balance"`
}

// SampleWithUnexported contains unexported and unsupported fields
type SampleWithUnexported struct {
	ID      int64      `json:"id"`
	private string     // unexported field
	NoTag   bool       // field without json tag
	Complex complex128 // unsupported type
}

func TestAnalyzeStruct_SpecialCases(t *testing.T) {
	// Test unexported and unsupported fields
	t.Run("unexported and unsupported", func(t *testing.T) {
		typ := reflect.TypeOf(SampleWithUnexported{})
		fieldInfos, err := AnalyzeStruct(typ)
		if err == nil {
			t.Error("expected error for unsupported type, got nil")
		}
		if err.Error() != "unsupported field type: complex128" {
			t.Errorf("unexpected error message: %v", err)
		}
		if fieldInfos != nil {
			t.Errorf("expected nil fieldInfos, got %v", fieldInfos)
		}
	})

	// Test struct with only unexported field
	t.Run("only unexported", func(t *testing.T) {
		type OnlyUnexported struct {
			private string
		}
		typ := reflect.TypeOf(OnlyUnexported{})
		fieldInfos, err := AnalyzeStruct(typ)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(fieldInfos) != 0 {
			t.Errorf("expected 0 fields, got %d", len(fieldInfos))
		}
	})

	// Test struct with field missing json tag
	t.Run("missing json tag", func(t *testing.T) {
		type NoJsonTag struct {
			Name string // no json tag
		}
		typ := reflect.TypeOf(NoJsonTag{})
		fieldInfos, err := AnalyzeStruct(typ)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(fieldInfos) != 1 {
			t.Fatalf("expected 1 field, got %d", len(fieldInfos))
		}
		if fieldInfos[0].Name != "Name" {
			t.Errorf("expected field name 'Name', got %q", fieldInfos[0].Name)
		}
	})
}

func TestAnalyzeStruct_NonStruct(t *testing.T) {
	// Test with non-struct types
	testCases := []struct {
		name        string
		value       interface{}
		expectedErr string
	}{
		{"int", 42, "AnalyzeStruct: provided type is not a struct"},
		{"string", "test", "AnalyzeStruct: provided type is not a struct"},
		{"slice", []int{1, 2, 3}, "AnalyzeStruct: provided type is not a struct"},
		{"map", map[string]int{}, "AnalyzeStruct: provided type is not a struct"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			typ := reflect.TypeOf(tc.value)
			_, err := AnalyzeStruct(typ)
			if err == nil {
				t.Errorf("expected error for type %v", tc.name)
			}
			if err.Error() != tc.expectedErr {
				t.Errorf("expected error %q, got %q", tc.expectedErr, err.Error())
			}
		})
	}
}

func TestAnalyzeStruct(t *testing.T) {
	typ := reflect.TypeOf(Sample{})
	fieldInfos, err := AnalyzeStruct(typ)
	if err != nil {
		t.Fatalf("AnalyzeStruct failed: %v", err)
	}

	expected := []FieldInfo{
		{Name: "id", Type: reflect.TypeOf(int64(0)), CType: "int64_t"},
		{Name: "name", Type: reflect.TypeOf(""), CType: "char*"},
		{Name: "active", Type: reflect.TypeOf(true), CType: "bool"},
		{Name: "balance", Type: reflect.TypeOf(0.0), CType: "double"},
	}

	if len(fieldInfos) != len(expected) {
		t.Fatalf("expected %d fields, got %d", len(expected), len(fieldInfos))
	}

	for i, fi := range fieldInfos {
		exp := expected[i]
		if fi.Name != exp.Name || fi.Type != exp.Type || fi.CType != exp.CType {
			t.Errorf("field %d - expected (%s, %v, %s), got (%s, %v, %s)",
				i, exp.Name, exp.Type, exp.CType, fi.Name, fi.Type, fi.CType)
		}
	}
}
