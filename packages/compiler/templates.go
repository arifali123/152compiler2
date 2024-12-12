package compiler

const ParserTemplate = `
#include <string.h>
#include <stdbool.h>
#include <stdlib.h>
#include <stdio.h>
#include "{{.Header}}"

// Function declarations
int parse_json(const char* input, {{.StructName}}* out);
char* parse_and_serialize_json(const char* input);
void free_serialized(char* str);

// Parse JSON and return values in a pipe-delimited format that Go can read
char* parse_and_serialize_json(const char* input) {
    {{.StructName}} out;
    memset(&out, 0, sizeof({{.StructName}}));  // Initialize struct to zero

    int result = parse_json(input, &out);
    if (result != 0) {
        return NULL;
    }

    // Allocate buffer for serialized output (adjust size as needed)
    char* serialized = (char*)malloc(1024);
    if (serialized == NULL) return NULL;

    // Format: SUCCESS|field1|field2|...
    snprintf(serialized, 1024, "SUCCESS");
    {{range .Fields}}
    {{if eq .CType "char*"}}
    if (out.{{.Name}} != NULL) {
        strcat(serialized, "|");
        strcat(serialized, out.{{.Name}});
        free(out.{{.Name}});  // Free the strdup'd string
    } else {
        strcat(serialized, "|");
    }
    {{else if eq .CType "int"}}
    char numStr[32];
    snprintf(numStr, sizeof(numStr), "|%d", out.{{.Name}});
    strcat(serialized, numStr);
    {{else if eq .CType "bool"}}
    strcat(serialized, "|");
    strcat(serialized, out.{{.Name}} ? "true" : "false");
    {{end}}
    {{end}}

    return serialized;
}

// Free the serialized string after use
void free_serialized(char* str) {
    if (str != NULL) {
        free(str);
    }
}

// Helper function to check if a character is escaped
static bool is_escaped(const char* str, const char* pos) {
    int backslashes = 0;
    while (pos > str && *(pos - 1) == '\\') {
        backslashes++;
        pos--;
    }
    return (backslashes % 2) == 1;  // Odd number of backslashes means the character is escaped
}

// Parse JSON into the C struct
int parse_json(const char* input, {{.StructName}}* out) {
    // Simple and naive JSON parser implementation
    const char* ptr = input;

    // Parse opening brace
    if (*ptr != '{') return -1;
    ptr++;

    while (*ptr != '}' && *ptr != '\0') {
        // Skip whitespace
        while (*ptr && (*ptr == ' ' || *ptr == '\n' || *ptr == '\t' || *ptr == ',')) ptr++;

        // Check for end of object
        if (*ptr == '}') break;

        // Expecting a string (field name)
        if (*ptr != '"') return -1;
        ptr++;
        // Read field name
        char field[256];
        int i = 0;
        while (*ptr != '\0' && (*ptr != '"' || is_escaped(input, ptr))) {
            if (*ptr == '\\' && *(ptr + 1) == '"') {
                field[i++] = '"';
                ptr += 2;
            } else {
                field[i++] = *ptr++;
            }
        }
        field[i] = '\0';
        if (*ptr != '"') return -1;
        ptr++; // Skip closing quote

        // Expecting colon
        while (*ptr && (*ptr == ' ' || *ptr == '\n' || *ptr == '\t')) ptr++;
        if (*ptr != ':') return -1;
        ptr++;

        // Skip whitespace
        while (*ptr && (*ptr == ' ' || *ptr == '\n' || *ptr == '\t')) ptr++;

        // Handle different types
        {{range .Fields}}
        if (strcmp(field, "{{.Name}}") == 0) {
            {{if eq .CType "char*"}}
                if (*ptr != '"') return -1;
                ptr++;
                char value[256];
                int j = 0;
                while (*ptr != '\0' && (*ptr != '"' || is_escaped(input, ptr))) {
                    if (*ptr == '\\' && *(ptr + 1) == '"') {
                        value[j++] = '"';
                        ptr += 2;
                    } else {
                        value[j++] = *ptr++;
                    }
                }
                value[j] = '\0';
                if (*ptr != '"') return -1;
                ptr++;
                out->{{.Name}} = strdup(value);
            {{else if eq .CType "int"}}
                // Simple integer parsing
                char number[20];
                int j = 0;
                if (*ptr == '-') {
                    number[j++] = *ptr++;
                }
                while (*ptr >= '0' && *ptr <= '9') {
                    number[j++] = *ptr++;
                }
                number[j] = '\0';
                if (j == 0 || (j == 1 && number[0] == '-')) return -1;
                out->{{.Name}} = atoi(number);
            {{else if eq .CType "bool"}}
                if (strncmp(ptr, "true", 4) == 0) {
                    out->{{.Name}} = true;
                    ptr += 4;
                } else if (strncmp(ptr, "false", 5) == 0) {
                    out->{{.Name}} = false;
                    ptr += 5;
                } else {
                    return -1;
                }
            {{end}}
            continue;
        }
        {{end}}

        // Skip unknown field value
        int depth = 0;
        bool in_string = false;
        while (*ptr != '\0') {
            if (!in_string) {
                if (*ptr == '{' || *ptr == '[') depth++;
                if (*ptr == '}' || *ptr == ']') {
                    if (depth == 0) break;
                    depth--;
                }
                if (*ptr == ',' && depth == 0) break;
            }
            if (*ptr == '"' && !is_escaped(input, ptr)) {
                in_string = !in_string;
            }
            ptr++;
        }
    }

    // Parse closing brace
    while (*ptr && (*ptr == ' ' || *ptr == '\n' || *ptr == '\t' || *ptr == ',')) ptr++;
    if (*ptr != '}') return -1;

    return 0; // success
}
`
