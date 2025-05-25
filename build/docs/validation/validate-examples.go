package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	examplesDir := "pkg/models"
	if len(os.Args) > 1 {
		examplesDir = os.Args[1]
	}

	validationResults := make(map[string][]string)

	err := filepath.WalkDir(examplesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, "_test.go") && strings.HasSuffix(path, ".go") {
			results := validateGoFile(path)
			if len(results) > 0 {
				validationResults[path] = results
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	if len(validationResults) == 0 {
		fmt.Println("✅ All example code validated successfully")
		return
	}

	fmt.Printf("⚠️  Found %d files with validation issues:\n", len(validationResults))
	for file, issues := range validationResults {
		fmt.Printf("\n📁 %s:\n", file)
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
	}

	os.Exit(1)
}

func validateGoFile(filepath string) []string {
	var issues []string

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		issues = append(issues, fmt.Sprintf("Parse error: %v", err))
		return issues
	}

	// Check for exported functions without documentation
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Name.IsExported() && fn.Doc == nil {
				issues = append(issues, fmt.Sprintf("Exported function %s lacks documentation", fn.Name.Name))
			}
		}

		if gen, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range gen.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					if ts.Name.IsExported() && gen.Doc == nil {
						issues = append(issues, fmt.Sprintf("Exported type %s lacks documentation", ts.Name.Name))
					}
				}
			}
		}
	}

	return issues
}
