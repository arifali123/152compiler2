package main

import (
	"fmt"
	"reflect"

	"github.com/arifali123/152compiler2/packages/analyzer"
	"github.com/arifali123/152compiler2/packages/compiler"
)

// Student represents a student record
type Student struct {
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	StudentID         int    `json:"student_id"`
	CurrentlyEnrolled bool   `json:"currently_enrolled"`
}

func main() {
	// Example JSON data for testing
	jsonData := `{
		"first_name": "John",
		"last_name": "Doe",
		"student_id": 12345,
		"currently_enrolled": true
	}`

	// 1. Analyze the struct using reflection
	structType := reflect.TypeOf(Student{})
	fieldInfos, err := analyzer.AnalyzeStruct(structType)
	if err != nil {
		fmt.Printf("Error analyzing struct: %v\n", err)
		return
	}

	// 2. Create CStruct representation
	cStruct := analyzer.CStruct{
		Name:   "Student",
		Fields: fieldInfos,
	}

	// 3. Compile the parser
	parser, err := compiler.CompileAndBuild(cStruct)
	if err != nil {
		fmt.Printf("Error compiling parser: %v\n", err)
		return
	}
	defer parser.Close()

	// 4. Parse the JSON data
	result, err := parser.Parse(jsonData)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	// 5. Print the results
	fmt.Printf("Parsed Student Information:\n")
	fmt.Printf("First Name: %v\n", result["first_name"])
	fmt.Printf("Last Name: %v\n", result["last_name"])
	fmt.Printf("Student ID: %v\n", result["student_id"])
	fmt.Printf("Currently Enrolled: %v\n", result["currently_enrolled"])
}
