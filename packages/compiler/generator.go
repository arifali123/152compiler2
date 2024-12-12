package compiler

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/arifali123/152compiler2/packages/analyzer"
)

// GenerateCCode generates C code for the given struct.
func GenerateCCode(cStruct analyzer.CStruct) (string, error) {
	// Create the header file content
	header := generateCHeader(cStruct)

	// Prepare the data for the parser template
	data := struct {
		Header     string
		StructName string
		Fields     []analyzer.FieldInfo
	}{
		Header:     fmt.Sprintf("%s.h", cStruct.Name),
		StructName: cStruct.Name,
		Fields:     cStruct.Fields,
	}

	// Parse the parser template
	tmpl, err := template.New("parser").Parse(ParserTemplate)
	if err != nil {
		return "", err
	}

	// Execute the template
	var parserBuffer bytes.Buffer
	err = tmpl.Execute(&parserBuffer, data)
	if err != nil {
		return "", err
	}

	// Combine header and parser code
	fullCode := fmt.Sprintf("%s\n%s", header, parserBuffer.String())
	return fullCode, nil
}

// generateCHeader creates a C header file for the struct.
func generateCHeader(cStruct analyzer.CStruct) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("#ifndef %s_H\n", cStruct.Name))
	buffer.WriteString(fmt.Sprintf("#define %s_H\n\n", cStruct.Name))
	buffer.WriteString("#include <stdint.h>\n#include <stdbool.h>\n\n")
	buffer.WriteString(fmt.Sprintf("typedef struct {\n"))
	for _, field := range cStruct.Fields {
		buffer.WriteString(fmt.Sprintf("    %s %s;\n", field.CType, field.Name))
	}
	buffer.WriteString(fmt.Sprintf("} %s;\n\n", cStruct.Name))
	buffer.WriteString(fmt.Sprintf("#endif // %s_H\n", cStruct.Name))
	return buffer.String()
}
