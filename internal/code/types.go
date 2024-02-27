package code

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func (c *Code) retrieveTypes() error {

	fileSet := token.NewFileSet() // positions are relative to fileSet

	// Parse the file.
	file, err := parser.ParseFile(fileSet, "", c.CodeContent, parser.ParseComments)
	if err != nil {
		return err
	}

	var funcDecl *ast.FuncDecl

	// Search for the function declaration.
	ast.Inspect(file, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok && fd.Name.Name == c.FuncName {
			funcDecl = fd
			return false // Stop the search
		}
		return true // Continue the search
	})

	if funcDecl == nil {
		return fmt.Errorf("function %s not found", c.FuncName)
	}

	for _, field := range funcDecl.Type.Params.List {
		c.InputTypes[fmt.Sprintf("%s", field.Type)] = len(field.Names)
	}

	for _, field := range funcDecl.Type.Results.List {
		c.OutputTypes[fmt.Sprintf("%s", field.Type)] = len(field.Names)
	}

	return nil

}
