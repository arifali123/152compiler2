package compiler

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/arifali123/152compiler2/packages/analyzer"
)

var (
	validIdentifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

// CompileParser generates C code for the given struct and writes it to a file.
func CompileParser(cStruct analyzer.CStruct, outputDir string) error {
	// Validate struct
	if err := validateStruct(cStruct); err != nil {
		return fmt.Errorf("invalid struct: %v", err)
	}

	// Generate C code
	cCode, err := GenerateCCode(cStruct)
	if err != nil {
		return fmt.Errorf("failed to generate C code: %v", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Write the C code to files
	cFile := filepath.Join(outputDir, fmt.Sprintf("%s.c", cStruct.Name))
	headerFile := filepath.Join(outputDir, fmt.Sprintf("%s.h", cStruct.Name))

	// Split the code into header and implementation
	headerEnd := fmt.Sprintf("#endif // %s_H\n", cStruct.Name)
	headerEndIndex := len(headerEnd)
	for i := range cCode {
		if i+headerEndIndex <= len(cCode) && cCode[i:i+headerEndIndex] == headerEnd {
			// Write header file
			if err := os.WriteFile(headerFile, []byte(cCode[:i+headerEndIndex]), 0644); err != nil {
				return fmt.Errorf("failed to write header file: %v", err)
			}
			// Write implementation file
			if err := os.WriteFile(cFile, []byte(cCode[i+headerEndIndex:]), 0644); err != nil {
				return fmt.Errorf("failed to write C file: %v", err)
			}
			break
		}
	}

	return nil
}

// validateStruct checks if the CStruct is valid for code generation
func validateStruct(cStruct analyzer.CStruct) error {
	// Check struct name
	if cStruct.Name == "" {
		return fmt.Errorf("empty struct name")
	}
	if !validIdentifierRegex.MatchString(cStruct.Name) {
		return fmt.Errorf("invalid struct name: must be a valid C identifier")
	}

	// Check fields
	if len(cStruct.Fields) == 0 {
		return fmt.Errorf("struct has no fields")
	}

	// Check for duplicate field names and validate each field
	fieldNames := make(map[string]bool)
	for _, field := range cStruct.Fields {
		// Check field name
		if field.Name == "" {
			return fmt.Errorf("empty field name")
		}
		if !validIdentifierRegex.MatchString(field.Name) {
			return fmt.Errorf("invalid field name: %s (must be a valid C identifier)", field.Name)
		}
		if fieldNames[field.Name] {
			return fmt.Errorf("duplicate field name: %s", field.Name)
		}
		fieldNames[field.Name] = true

		// Validate CType
		switch field.CType {
		case "char*", "int", "bool":
			// These are the supported types
		case "":
			return fmt.Errorf("empty C type for field: %s", field.Name)
		default:
			return fmt.Errorf("unsupported C type: %s", field.CType)
		}
	}

	return nil
}

// CompiledParser represents a compiled JSON parser
type CompiledParser struct {
	execPath string
	cleanup  func()
	fields   []analyzer.FieldInfo // Store field information for parsing
}

// CompileAndBuild generates C code, compiles it into an executable, and returns a parser instance
func CompileAndBuild(cStruct analyzer.CStruct) (*CompiledParser, error) {
	// Create c_output directory in root if it doesn't exist
	outputDir := "c_output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate and write C code
	if err := CompileParser(cStruct, outputDir); err != nil {
		return nil, err
	}

	// Create main file using struct name
	mainFile := filepath.Join(outputDir, fmt.Sprintf("main_%s.c", cStruct.Name))
	mainCode := fmt.Sprintf(`
#include <stdio.h>
#include <stdlib.h>
#include "%s.h"

extern char* parse_and_serialize_json(const char* input);
extern void free_serialized(char* str);

int main(int argc, char *argv[]) {
    if (argc != 2) {
        fprintf(stderr, "Usage: %%s <json_string>\\n", argv[0]);
        return 1;
    }

    char* result = parse_and_serialize_json(argv[1]);
    if (result == NULL) {
        printf("ERROR|Failed to parse JSON\\n");
        return 1;
    }

    printf("%%s", result);  // No newline in output
    free_serialized(result);
    return 0;
}`, cStruct.Name)

	if err := os.WriteFile(mainFile, []byte(mainCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write main_%s.c: %v", cStruct.Name, err)
	}

	// Compile the program with struct-specific names
	outPath := filepath.Join(outputDir, fmt.Sprintf("parser_%s", cStruct.Name))
	cmd := exec.Command("gcc", "-o", outPath, mainFile, filepath.Join(outputDir, fmt.Sprintf("%s.c", cStruct.Name)))
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("compilation failed: %v\nOutput: %s", err, out)
	}

	return &CompiledParser{
		execPath: outPath,
		cleanup: func() {
			// Only remove the files specific to this struct
			slog.Info("Removing File", slog.String("file", mainFile))
			os.Remove(mainFile)
			slog.Info("Removing File", slog.String("file", filepath.Join(outputDir, fmt.Sprintf("%s.c", cStruct.Name))))
			os.Remove(filepath.Join(outputDir, fmt.Sprintf("%s.c", cStruct.Name)))
			slog.Info("Removing File", slog.String("file", filepath.Join(outputDir, fmt.Sprintf("%s.h", cStruct.Name))))
			os.Remove(filepath.Join(outputDir, fmt.Sprintf("%s.h", cStruct.Name)))
			slog.Info("Removing File", slog.String("file", outPath))
			os.Remove(outPath)
		},
		fields: cStruct.Fields,
	}, nil
}

// Parse parses a JSON string and returns the values as a map
func (p *CompiledParser) Parse(jsonStr string) (map[string]interface{}, error) {
	cmd := exec.Command(p.execPath, jsonStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("parser execution failed: %v\nOutput: %s", err, out)
	}
	slog.Info("C Parser Output", slog.String("output", string(out)))
	// Parse the pipe-delimited output and trim any whitespace/newlines
	parts := strings.Split(strings.TrimSpace(string(out)), "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	if parts[0] != "SUCCESS" {
		return nil, fmt.Errorf("parsing failed: %s", string(out))
	}

	// Convert to map using field information
	result := make(map[string]interface{})
	for i, field := range p.fields {
		if i+1 >= len(parts) {
			break
		}
		value := parts[i+1]

		// Convert value based on field type
		switch field.CType {
		case "int":
			// Keep as string for now, as the caller can parse it if needed
			result[field.Name] = value
		case "bool":
			result[field.Name] = value == "true"
		case "char*":
			result[field.Name] = value
		}
	}

	return result, nil
}

// Close releases resources associated with the parser
func (p *CompiledParser) Close() {
	if p.cleanup != nil {
		p.cleanup()
	}
}
