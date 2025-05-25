package metrics

import (
	"go/ast"
	"go/token"
)

// complexityVisitor walks the AST and calculates complexity metrics
type complexityVisitor struct {
	fileSet    *token.FileSet
	file       *ast.File
	thresholds ComplexityThresholds
	filePath   string
	functions  []FunctionMetrics
	violations []ComplexityViolation
}

// Visit implements ast.Visitor interface
func (v *complexityVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Name != nil && n.Name.IsExported() || n.Name != nil {
			v.analyzeFunction(n)
		}
	}
	return v
}

// analyzeFunction calculates complexity metrics for a function
func (v *complexityVisitor) analyzeFunction(fn *ast.FuncDecl) {
	metrics := FunctionMetrics{
		Name:      fn.Name.Name,
		StartLine: v.fileSet.Position(fn.Pos()).Line,
		EndLine:   v.fileSet.Position(fn.End()).Line,
	}

	// Calculate lines of code for the function
	metrics.LinesOfCode = metrics.EndLine - metrics.StartLine + 1

	// Count parameters
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			metrics.Parameters += len(field.Names)
		}
	}

	// Count return values
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			if len(field.Names) == 0 {
				metrics.ReturnValues++
			} else {
				metrics.ReturnValues += len(field.Names)
			}
		}
	}

	// Calculate cyclomatic complexity
	if fn.Body != nil {
		complexity := v.calculateCyclomaticComplexity(fn.Body)
		metrics.CyclomaticComplexity = complexity

		// Calculate maximum nesting depth
		metrics.Nesting = v.calculateMaxNesting(fn.Body, 0)
	}

	v.functions = append(v.functions, metrics)

	// Check for violations
	v.checkFunctionViolations(metrics)
}

// calculateCyclomaticComplexity calculates the cyclomatic complexity of a function body
func (v *complexityVisitor) calculateCyclomaticComplexity(body *ast.BlockStmt) int {
	complexity := 1 // Base complexity

	ast.Inspect(body, func(node ast.Node) bool {
		switch node.(type) {
		case *ast.IfStmt:
			complexity++
		case *ast.ForStmt:
			complexity++
		case *ast.RangeStmt:
			complexity++
		case *ast.SwitchStmt:
			complexity++
		case *ast.TypeSwitchStmt:
			complexity++
		case *ast.SelectStmt:
			complexity++
		case *ast.CaseClause:
			// Don't count default case
			if n := node.(*ast.CaseClause); len(n.List) > 0 {
				complexity++
			}
		case *ast.CommClause:
			// Don't count default case
			if n := node.(*ast.CommClause); n.Comm != nil {
				complexity++
			}
		}
		return true
	})

	return complexity
}

// calculateMaxNesting calculates the maximum nesting depth of control structures
func (v *complexityVisitor) calculateMaxNesting(node ast.Node, currentDepth int) int {
	if node == nil {
		return currentDepth
	}

	maxDepth := currentDepth

	// Handle block statements first
	if blockStmt, ok := node.(*ast.BlockStmt); ok {
		for _, stmt := range blockStmt.List {
			depth := v.calculateMaxNesting(stmt, currentDepth)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		return maxDepth
	}

	// Handle control structures that increase nesting
	controlStructures := v.getControlStructureBodies(node, currentDepth)
	for _, body := range controlStructures {
		depth := v.calculateMaxNesting(body.node, body.depth)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth
}

// nestingNode represents a node and its nesting depth
type nestingNode struct {
	node  ast.Node
	depth int
}

// getControlStructureBodies extracts bodies from control structures
func (v *complexityVisitor) getControlStructureBodies(node ast.Node, currentDepth int) []nestingNode {
	var bodies []nestingNode

	switch n := node.(type) {
	case *ast.IfStmt:
		if n.Body != nil {
			bodies = append(bodies, nestingNode{n.Body, currentDepth + 1})
		}
		if n.Else != nil {
			bodies = append(bodies, nestingNode{n.Else, currentDepth + 1})
		}
	case *ast.ForStmt:
		if n.Body != nil {
			bodies = append(bodies, nestingNode{n.Body, currentDepth + 1})
		}
	case *ast.RangeStmt:
		if n.Body != nil {
			bodies = append(bodies, nestingNode{n.Body, currentDepth + 1})
		}
	case *ast.SwitchStmt:
		if n.Body != nil {
			bodies = append(bodies, nestingNode{n.Body, currentDepth + 1})
		}
	case *ast.TypeSwitchStmt:
		if n.Body != nil {
			bodies = append(bodies, nestingNode{n.Body, currentDepth + 1})
		}
	case *ast.SelectStmt:
		if n.Body != nil {
			bodies = append(bodies, nestingNode{n.Body, currentDepth + 1})
		}
	case *ast.CaseClause:
		for _, stmt := range n.Body {
			bodies = append(bodies, nestingNode{stmt, currentDepth})
		}
	case *ast.CommClause:
		for _, stmt := range n.Body {
			bodies = append(bodies, nestingNode{stmt, currentDepth})
		}
	}

	return bodies
}

// checkFunctionViolations checks if a function violates complexity thresholds
func (v *complexityVisitor) checkFunctionViolations(metrics FunctionMetrics) {
	// Check cyclomatic complexity
	if metrics.CyclomaticComplexity > v.thresholds.CyclomaticComplexity {
		v.violations = append(v.violations, ComplexityViolation{
			Type:           "CyclomaticComplexity",
			Severity:       v.getSeverity("CyclomaticComplexity", metrics.CyclomaticComplexity),
			Message:        "Function has high cyclomatic complexity",
			FilePath:       v.filePath,
			FunctionName:   metrics.Name,
			LineNumber:     metrics.StartLine,
			ActualValue:    metrics.CyclomaticComplexity,
			ThresholdValue: v.thresholds.CyclomaticComplexity,
		})
	}

	// Check function length
	if metrics.LinesOfCode > v.thresholds.FunctionLength {
		v.violations = append(v.violations, ComplexityViolation{
			Type:           "FunctionLength",
			Severity:       v.getSeverity("FunctionLength", metrics.LinesOfCode),
			Message:        "Function is too long",
			FilePath:       v.filePath,
			FunctionName:   metrics.Name,
			LineNumber:     metrics.StartLine,
			ActualValue:    metrics.LinesOfCode,
			ThresholdValue: v.thresholds.FunctionLength,
		})
	}

	// Check parameter count (industry standard: <=5 parameters)
	maxParameters := 5
	if metrics.Parameters > maxParameters {
		v.violations = append(v.violations, ComplexityViolation{
			Type:           "ParameterCount",
			Severity:       "Warning",
			Message:        "Function has too many parameters",
			FilePath:       v.filePath,
			FunctionName:   metrics.Name,
			LineNumber:     metrics.StartLine,
			ActualValue:    metrics.Parameters,
			ThresholdValue: maxParameters,
		})
	}

	// Check nesting depth (industry standard: <=4 levels)
	maxNesting := 4
	if metrics.Nesting > maxNesting {
		v.violations = append(v.violations, ComplexityViolation{
			Type:           "NestingDepth",
			Severity:       "Warning",
			Message:        "Function has excessive nesting depth",
			FilePath:       v.filePath,
			FunctionName:   metrics.Name,
			LineNumber:     metrics.StartLine,
			ActualValue:    metrics.Nesting,
			ThresholdValue: maxNesting,
		})
	}
}

// getSeverity determines the severity level based on how much a metric exceeds the threshold
func (v *complexityVisitor) getSeverity(metricType string, value int) string {
	var threshold int
	switch metricType {
	case "CyclomaticComplexity":
		threshold = v.thresholds.CyclomaticComplexity
	case "FunctionLength":
		threshold = v.thresholds.FunctionLength
	default:
		return "Warning"
	}

	ratio := float64(value) / float64(threshold)
	if ratio >= 2.0 {
		return "Critical"
	} else if ratio >= 1.5 {
		return "Major"
	} else {
		return "Minor"
	}
}
