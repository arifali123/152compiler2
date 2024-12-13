# JSON Parser Generator

A Go-based JSON parser generator that creates specialized C parsers for specific struct types.

## Project Structure

```
152compiler2/
├── packages/
│ ├── analyzer/ # Struct analysis
│ │ ├── analyzer_test.go
│ │ ├── reflect.go # Struct reflection
│ │ └── types.go # Type mappings
│ ├── compiler/ # Parser generation
│ │ ├── compiler.go # Compilation logic
│ │ ├── generator.go # C code generation
│ │ └── templates.go # C code templates
│ └── lexer/ # JSON lexing
│ │ ├── lexer.go # Tokenization
│ │ └── lexer_test.go # Lexer test
│ ├── main.go # Main entry point
```

## Implementation Overview

The project consists of three main components:

1. **Struct Analysis** - Analyzes Go structs using reflection
2. **Code Generation** - Generates C code for parsing
3. **Compilation** - Compiles and loads the generated parser

### 1. Struct Analysis

The analyzer examines Go structs and maps their fields to C types:

```go
type FieldInfo struct {
    Name   string       // JSON tag or field name
    GoName string       // Original Go field name
    Type   reflect.Type
    Offset uintptr
    CType  string      // Mapped C type
    Kind   string      // Kind as string
}
```

Reference to type mappings:

```go
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
```

### 2. Code Generation

Generates C code for parsing JSON into the analyzed struct. The generator creates:

- Header file (.h) with struct definition
- Implementation file (.c) with parsing logic

Example usage:

```go
type Student struct {
    FirstName         string `json:"first_name"`
    LastName          string `json:"last_name"`
    StudentID         int    `json:"student_id"`
    CurrentlyEnrolled bool   `json:"currently_enrolled"`
}
```

### 3. Parser Compilation

The compiler:

1. Validates the struct definition
2. Generates C code
3. Compiles to executable
4. Provides interface for parsing

Reference to compilation process:

```go
func CompileAndBuild(cStruct analyzer.CStruct) (*CompiledParser, error) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "jsonparser_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}

	// Generate and write C code
	if err := CompileParser(cStruct, tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		return nil, err
	}

	// Create main.c that uses our parser
	mainFile := filepath.Join(tmpDir, "main.c")
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
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to write main.c: %v", err)
	}

	// Compile the program
	outPath := filepath.Join(tmpDir, "parser")
	cmd := exec.Command("gcc", "-o", outPath, mainFile, filepath.Join(tmpDir, fmt.Sprintf("%s.c", cStruct.Name)))
	if out, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("compilation failed: %v\nOutput: %s", err, out)
	}

	return &CompiledParser{
		execPath: outPath,
		cleanup: func() {
			os.RemoveAll(tmpDir)
		},
		fields: cStruct.Fields,
	}, nil
}
```

## Usage Example

```go
// 1. Define your struct
type Student struct {
    FirstName         string `json:"first_name"`
    LastName          string `json:"last_name"`
    StudentID         int    `json:"student_id"`
    CurrentlyEnrolled bool   `json:"currently_enrolled"`
}

// 2. Create parser
structType := reflect.TypeOf(Student{})
fieldInfos, err := analyzer.AnalyzeStruct(structType)
if err != nil {
    log.Fatal(err)
}

// 3. Generate C struct
cStruct := analyzer.CStruct{
    Name:   "Student",
    Fields: fieldInfos,
}

// 4. Compile parser
parser, err := compiler.CompileAndBuild(cStruct)
if err != nil {
    log.Fatal(err)
}
defer parser.Close()

// 5. Parse JSON
jsonData := `{
    "first_name": "John",
    "last_name": "Doe",
    "student_id": 12345,
    "currently_enrolled": true
}`

result, err := parser.Parse(jsonData)
if err != nil {
    log.Fatal(err)
}
```

## Features

- **Type Support**:

  - Strings (`char*`)
  - Integers (`int`)
  - Booleans (`bool`)

- **Validation**:

  - JSON syntax checking
  - Type validation
  - Struct field validation

- **Error Handling**:
  - Detailed error messages
  - Invalid JSON detection
  - Type mismatch reporting

## Testing

The project includes comprehensive tests for:

- Lexer functionality
- Struct analysis
- Parser generation
- JSON parsing

## Limitations

1. Currently supports only:

   - Basic scalar types (string, int, bool)
   - Single-level JSON objects
   - Non-nested structures

2. No support for:
   - Arrays/slices
   - Maps
   - Nested objects
   - Custom types

## Future Improvements

1. Add support for:

   - Nested structures
   - Arrays and slices
   - More numeric types
   - Custom type mappings

2. Performance optimizations:

   - SIMD instructions
   - Parser caching
   - Memory pooling

3. Additional features:
   - Custom unmarshaling
   - Streaming support
   - Pretty printing

## Inspiration & Learning Goals

This project is inspired by ByteDance's [sonic](https://github.com/bytedance/sonic) JSON parser, serving as a learning implementation to understand JIT (Just-In-Time) compiler concepts. While sonic is a production-grade parser with advanced features, this project focuses on:

### Key Learning Areas

1. **Basic JIT Compilation**

   - Understanding how to generate specialized C code
   - Runtime compilation using gcc
   - Dynamic loading of compiled parsers

2. **Performance Optimization Concepts**

   - Direct memory access through C structs
   - Avoiding reflection during parsing
   - Reduced allocations through pre-compiled parsers

3. **Differences from sonic**
   - sonic uses SIMD instructions for parsing
   - sonic implements a full JIT compiler in Go
   - This project uses C as an intermediate step for learning

### Educational Focus

This implementation deliberately simplifies many concepts to focus on learning:

- Uses C compilation instead of direct machine code generation
- Focuses on basic JSON parsing without optimizations
- Provides a clear path to understanding JIT compilation basics

## Compiler Theory Components

### 1. Regular Language and Grammar

The JSON parser's lexical structure can be described using regular expressions:

```
STRING  -> "[^"]*"
NUMBER  -> -?[0-9]+(\.[0-9]+)?
BOOLEAN -> true|false
NULL    -> null
WS      -> [ \t\n\r]+
```

### 2. Context-Free Grammar

```
json     → object
object   → '{' members '}'
members  → pair (',' pair)* | ε
pair     → STRING ':' value
value    → STRING | NUMBER | BOOLEAN | NULL
```

### 3. Chomsky Normal Form

Transforming our grammar to CNF:

```
S → AB          /* json -> { members } */
A → OB          /* object transformation */
B → CD          /* pair production */
C → STRING      /* terminal production */
D → EF          /* value split */
E → COLON       /* terminal */
F → VALUE       /* terminal */
```

### 4. Automata Analysis

#### DFA for String Tokenization

```
States: Q = {START, IN_STRING, ESCAPED, DONE, ERROR}
Start state: START
Accept state: DONE
Transitions:
  START     + " → IN_STRING
  IN_STRING + \ → ESCAPED
  IN_STRING + " → DONE
  ESCAPED   + any → IN_STRING
```

#### State Transition Table

```
State      | "  | \  | other
-----------|----|----|-------
START      | 1  | E  | E
IN_STRING  | 3  | 2  | 1
ESCAPED    | 1  | 1  | 1
DONE       | E  | E  | E

Where: 1=IN_STRING, 2=ESCAPED, 3=DONE, E=ERROR
```

### 5. Production Rules and Parsing

Left-most derivation example for `{"name":"John"}`:

```
json ⇒ object
    ⇒ '{' members '}'
    ⇒ '{' pair '}'
    ⇒ '{' STRING ':' value '}'
    ⇒ '{"name":"John"}'
```
