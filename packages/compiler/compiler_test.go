package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arifali123/152compiler2/packages/analyzer"
)

func TestCompileParser(t *testing.T) {
	// Create a test struct
	testStruct := analyzer.CStruct{
		Name: "Person",
		Fields: []analyzer.FieldInfo{
			{Name: "name", CType: "char*"},
			{Name: "age", CType: "int"},
			{Name: "is_student", CType: "bool"},
		},
	}

	// Create temporary directory for test output
	tempDir := t.TempDir()

	// Generate the C code
	err := CompileParser(testStruct, tempDir)
	if err != nil {
		t.Fatalf("CompileParser failed: %v", err)
	}

	// Check if files were created
	headerPath := filepath.Join(tempDir, "Person.h")
	sourcePath := filepath.Join(tempDir, "Person.c")

	// Verify header file exists and contains expected content
	headerContent, err := os.ReadFile(headerPath)
	if err != nil {
		t.Fatalf("Failed to read header file: %v", err)
	}

	// Check header content
	expectedHeaderContent := []string{
		"#ifndef Person_H",
		"#define Person_H",
		"typedef struct {",
		"char* name;",
		"int age;",
		"bool is_student;",
		"} Person;",
	}

	for _, expected := range expectedHeaderContent {
		if !strings.Contains(string(headerContent), expected) {
			t.Errorf("Header file missing expected content: %s", expected)
		}
	}

	// Verify source file exists and contains expected content
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}

	// Check source content
	expectedSourceContent := []string{
		"int parse_json(const char* input, Person* out)",
		"strcmp(field, \"name\")",
		"strcmp(field, \"age\")",
		"strcmp(field, \"is_student\")",
		"parse_and_serialize_json",
	}

	for _, expected := range expectedSourceContent {
		if !strings.Contains(string(sourceContent), expected) {
			t.Errorf("Source file missing expected content: %s", expected)
		}
	}
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantValues  map[string]interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:  "valid json with all fields",
			input: `{"name": "John Doe", "age": 25, "is_student": true}`,
			wantValues: map[string]interface{}{
				"name":       "John Doe",
				"age":        "25",
				"is_student": true,
			},
		},
		{
			name:  "valid json with missing optional fields",
			input: `{"name": "John Doe"}`,
			wantValues: map[string]interface{}{
				"name":       "John Doe",
				"age":        "0",
				"is_student": false,
			},
		},
		{
			name:        "invalid json - missing opening brace",
			input:       `"name": "John Doe", "age": 25}`,
			wantErr:     true,
			errContains: "Failed to parse JSON",
		},
		{
			name:        "invalid json - missing closing brace",
			input:       `{"name": "John Doe", "age": 25`,
			wantErr:     true,
			errContains: "Failed to parse JSON",
		},
		{
			name:        "invalid json - missing quotes around field name",
			input:       `{name: "John Doe"}`,
			wantErr:     true,
			errContains: "Failed to parse JSON",
		},
		{
			name:        "invalid json - missing quotes around string value",
			input:       `{"name": John Doe}`,
			wantErr:     true,
			errContains: "Failed to parse JSON",
		},
		{
			name:        "invalid json - invalid boolean value",
			input:       `{"is_student": maybe}`,
			wantErr:     true,
			errContains: "Failed to parse JSON",
		},
		{
			name:        "invalid json - invalid integer value",
			input:       `{"age": twenty}`,
			wantErr:     true,
			errContains: "Failed to parse JSON",
		},
		{
			name:  "valid json with extra fields",
			input: `{"name": "John Doe", "age": 25, "is_student": true, "extra": "field"}`,
			wantValues: map[string]interface{}{
				"name":       "John Doe",
				"age":        "25",
				"is_student": true,
			},
		},
		{
			name:  "valid json with escaped quotes in string",
			input: `{"name": "John \"Johnny\" Doe", "age": 25, "is_student": true}`,
			wantValues: map[string]interface{}{
				"name":       "John \"Johnny\" Doe",
				"age":        "25",
				"is_student": true,
			},
		},
		{
			name: "valid json with whitespace",
			input: `{
				"name": "John Doe",
				"age": 25,
				"is_student": true
			}`,
			wantValues: map[string]interface{}{
				"name":       "John Doe",
				"age":        "25",
				"is_student": true,
			},
		},
	}

	// Create test struct
	testStruct := analyzer.CStruct{
		Name: "Person",
		Fields: []analyzer.FieldInfo{
			{Name: "name", CType: "char*"},
			{Name: "age", CType: "int"},
			{Name: "is_student", CType: "bool"},
		},
	}

	// Compile the parser once for all tests
	parser, err := CompileAndBuild(testStruct)
	if err != nil {
		t.Fatalf("Failed to compile parser: %v", err)
	}
	defer parser.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)

			// Check error cases
			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Parse() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			// Check success cases
			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			// Verify all expected values are present and correct
			for field, expected := range tt.wantValues {
				got, ok := result[field]
				if !ok {
					t.Errorf("Field %s missing from result", field)
					continue
				}
				if got != expected {
					t.Errorf("Field %s = %v, want %v", field, got, expected)
				}
			}

			// Verify no extra fields are present
			for field := range result {
				if _, ok := tt.wantValues[field]; !ok {
					t.Errorf("Unexpected field in result: %s", field)
				}
			}
		})
	}
}

func TestCompileParserErrors(t *testing.T) {
	tests := []struct {
		name        string
		cStruct     analyzer.CStruct
		outputDir   string
		wantErr     bool
		errContains string
	}{
		{
			name: "empty struct name",
			cStruct: analyzer.CStruct{
				Name: "",
				Fields: []analyzer.FieldInfo{
					{Name: "name", CType: "char*"},
				},
			},
			outputDir:   t.TempDir(),
			wantErr:     true,
			errContains: "empty struct name",
		},
		{
			name: "no fields",
			cStruct: analyzer.CStruct{
				Name:   "Person",
				Fields: []analyzer.FieldInfo{},
			},
			outputDir:   t.TempDir(),
			wantErr:     true,
			errContains: "struct has no fields",
		},
		{
			name: "empty field name",
			cStruct: analyzer.CStruct{
				Name: "Person",
				Fields: []analyzer.FieldInfo{
					{Name: "", CType: "char*"},
				},
			},
			outputDir:   t.TempDir(),
			wantErr:     true,
			errContains: "empty field name",
		},
		{
			name: "duplicate field names",
			cStruct: analyzer.CStruct{
				Name: "Person",
				Fields: []analyzer.FieldInfo{
					{Name: "name", CType: "char*"},
					{Name: "name", CType: "int"},
				},
			},
			outputDir:   t.TempDir(),
			wantErr:     true,
			errContains: "duplicate field name",
		},
		{
			name: "unsupported C type",
			cStruct: analyzer.CStruct{
				Name: "Person",
				Fields: []analyzer.FieldInfo{
					{Name: "name", CType: "float"},
				},
			},
			outputDir:   t.TempDir(),
			wantErr:     true,
			errContains: "unsupported C type",
		},
		{
			name: "empty C type",
			cStruct: analyzer.CStruct{
				Name: "Person",
				Fields: []analyzer.FieldInfo{
					{Name: "name", CType: ""},
				},
			},
			outputDir:   t.TempDir(),
			wantErr:     true,
			errContains: "empty C type",
		},
		{
			name: "invalid output directory",
			cStruct: analyzer.CStruct{
				Name: "Person",
				Fields: []analyzer.FieldInfo{
					{Name: "name", CType: "char*"},
				},
			},
			outputDir:   "/nonexistent/directory/that/should/not/exist",
			wantErr:     true,
			errContains: "failed to create output directory",
		},
		{
			name: "nil fields",
			cStruct: analyzer.CStruct{
				Name:   "Person",
				Fields: nil,
			},
			outputDir:   t.TempDir(),
			wantErr:     true,
			errContains: "struct has no fields",
		},
		{
			name: "special characters in struct name",
			cStruct: analyzer.CStruct{
				Name: "Person!@#",
				Fields: []analyzer.FieldInfo{
					{Name: "name", CType: "char*"},
				},
			},
			outputDir:   t.TempDir(),
			wantErr:     true,
			errContains: "invalid struct name",
		},
		{
			name: "special characters in field name",
			cStruct: analyzer.CStruct{
				Name: "Person",
				Fields: []analyzer.FieldInfo{
					{Name: "name!@#", CType: "char*"},
				},
			},
			outputDir:   t.TempDir(),
			wantErr:     true,
			errContains: "invalid field name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CompileParser(tt.cStruct, tt.outputDir)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CompileParser() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("CompileParser() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("CompileParser() unexpected error: %v", err)
			}
		})
	}
}
