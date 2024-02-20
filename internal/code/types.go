package code

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
)

func retrieveTypes(inputsString, functionName string) ([]types.Type, error) {

	fileSet := token.NewFileSet() // positions are relative to fileSet

	// Parse the file.
	file, err := parser.ParseFile(fileSet, "", inputsString, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var funcDecl *ast.FuncDecl

	// Search for the function declaration.
	ast.Inspect(file, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok && fd.Name.Name == functionName {
			funcDecl = fd
			return false // Stop the search
		}
		return true // Continue the search
	})

	if funcDecl == nil {
		return nil, fmt.Errorf("Function %s not found\n", functionName)
	}

	// Print the input types of the function.
	fmt.Printf("Input types for function %s:\n", functionName)
	for _, field := range funcDecl.Type.Params.List {
		fmt.Printf("%s\n", field.Type)
	}

	fmt.Printf("\nInput types for function %s:\n", functionName)
	for _, field := range funcDecl.Type.Results.List {
		fmt.Printf("%s\n", field.Type)
	}

	return nil, nil

}
