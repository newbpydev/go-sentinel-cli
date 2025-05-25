// Package metrics provides code quality and complexity analysis tools for the Go Sentinel CLI.
// This package implements cyclomatic complexity measurement, maintainability index calculation,
// and technical debt tracking as part of Phase 4 - Task 8.
package metrics

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ComplexityAnalyzer provides code complexity analysis capabilities
type ComplexityAnalyzer interface {
	AnalyzeFile(filePath string) (*FileComplexity, error)
	AnalyzePackage(packagePath string) (*PackageComplexity, error)
	AnalyzeProject(projectRoot string) (*ProjectComplexity, error)
	SetThresholds(thresholds ComplexityThresholds)
	GenerateReport(complexity *ProjectComplexity, output io.Writer) error
}

// ComplexityThresholds defines acceptable complexity limits
type ComplexityThresholds struct {
	CyclomaticComplexity int     `json:"cyclomatic_complexity"`
	MaintainabilityIndex float64 `json:"maintainability_index"`
	LinesOfCode          int     `json:"lines_of_code"`
	TechnicalDebtRatio   float64 `json:"technical_debt_ratio"`
	FunctionLength       int     `json:"function_length"`
}

// DefaultThresholds returns industry-standard complexity thresholds
func DefaultThresholds() ComplexityThresholds {
	return ComplexityThresholds{
		CyclomaticComplexity: 10,
		MaintainabilityIndex: 85.0,
		LinesOfCode:          500,
		TechnicalDebtRatio:   5.0,
		FunctionLength:       50,
	}
}

// FileComplexity represents complexity metrics for a single file
type FileComplexity struct {
	FilePath                    string                `json:"file_path"`
	LinesOfCode                 int                   `json:"lines_of_code"`
	Functions                   []FunctionMetrics     `json:"functions"`
	AverageCyclomaticComplexity float64               `json:"average_cyclomatic_complexity"`
	MaintainabilityIndex        float64               `json:"maintainability_index"`
	TechnicalDebtMinutes        int                   `json:"technical_debt_minutes"`
	Violations                  []ComplexityViolation `json:"violations"`
}

// FunctionMetrics represents complexity metrics for a single function
type FunctionMetrics struct {
	Name                 string `json:"name"`
	LinesOfCode          int    `json:"lines_of_code"`
	CyclomaticComplexity int    `json:"cyclomatic_complexity"`
	Parameters           int    `json:"parameters"`
	ReturnValues         int    `json:"return_values"`
	Nesting              int    `json:"nesting"`
	StartLine            int    `json:"start_line"`
	EndLine              int    `json:"end_line"`
}

// PackageComplexity represents complexity metrics for a package
type PackageComplexity struct {
	PackagePath                 string                `json:"package_path"`
	Files                       []FileComplexity      `json:"files"`
	TotalLinesOfCode            int                   `json:"total_lines_of_code"`
	TotalFunctions              int                   `json:"total_functions"`
	AverageCyclomaticComplexity float64               `json:"average_cyclomatic_complexity"`
	MaintainabilityIndex        float64               `json:"maintainability_index"`
	TechnicalDebtHours          float64               `json:"technical_debt_hours"`
	Violations                  []ComplexityViolation `json:"violations"`
}

// ProjectComplexity represents complexity metrics for the entire project
type ProjectComplexity struct {
	ProjectRoot string              `json:"project_root"`
	Packages    []PackageComplexity `json:"packages"`
	Summary     ComplexitySummary   `json:"summary"`
	GeneratedAt time.Time           `json:"generated_at"`
}

// ComplexitySummary provides high-level project complexity metrics
type ComplexitySummary struct {
	TotalFiles                  int     `json:"total_files"`
	TotalLinesOfCode            int     `json:"total_lines_of_code"`
	TotalFunctions              int     `json:"total_functions"`
	AverageCyclomaticComplexity float64 `json:"average_cyclomatic_complexity"`
	MaintainabilityIndex        float64 `json:"maintainability_index"`
	TechnicalDebtDays           float64 `json:"technical_debt_days"`
	ViolationCount              int     `json:"violation_count"`
	QualityGrade                string  `json:"quality_grade"`
}

// ComplexityViolation represents a code quality violation
type ComplexityViolation struct {
	Type           string      `json:"type"`
	Severity       string      `json:"severity"`
	Message        string      `json:"message"`
	FilePath       string      `json:"file_path"`
	FunctionName   string      `json:"function_name,omitempty"`
	LineNumber     int         `json:"line_number"`
	ActualValue    interface{} `json:"actual_value"`
	ThresholdValue interface{} `json:"threshold_value"`
}

// DefaultComplexityAnalyzer implements ComplexityAnalyzer
type DefaultComplexityAnalyzer struct {
	thresholds ComplexityThresholds
	fileSet    *token.FileSet
}

// NewComplexityAnalyzer creates a new complexity analyzer with default thresholds
func NewComplexityAnalyzer() *DefaultComplexityAnalyzer {
	return &DefaultComplexityAnalyzer{
		thresholds: DefaultThresholds(),
		fileSet:    token.NewFileSet(),
	}
}

// SetThresholds updates the complexity thresholds
func (a *DefaultComplexityAnalyzer) SetThresholds(thresholds ComplexityThresholds) {
	a.thresholds = thresholds
}

// AnalyzeFile analyzes a single Go file for complexity metrics
func (a *DefaultComplexityAnalyzer) AnalyzeFile(filePath string) (*FileComplexity, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	file, err := parser.ParseFile(a.fileSet, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	complexity := &FileComplexity{
		FilePath:   filePath,
		Functions:  make([]FunctionMetrics, 0),
		Violations: make([]ComplexityViolation, 0),
	}

	// Count lines of code (excluding comments and empty lines)
	complexity.LinesOfCode = a.countLinesOfCode(string(content))

	// Analyze functions
	visitor := &complexityVisitor{
		fileSet:    a.fileSet,
		file:       file,
		thresholds: a.thresholds,
		filePath:   filePath,
	}
	ast.Walk(visitor, file)

	complexity.Functions = visitor.functions
	complexity.Violations = append(complexity.Violations, visitor.violations...)

	// Calculate aggregate metrics
	a.calculateFileMetrics(complexity)

	// Check file-level violations
	a.checkFileViolations(complexity)

	return complexity, nil
}

// AnalyzePackage analyzes all Go files in a package
func (a *DefaultComplexityAnalyzer) AnalyzePackage(packagePath string) (*PackageComplexity, error) {
	files, err := filepath.Glob(filepath.Join(packagePath, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("failed to list files in package %s: %w", packagePath, err)
	}

	// Filter out test files for main complexity analysis
	var mainFiles []string
	for _, file := range files {
		if !strings.HasSuffix(file, "_test.go") {
			mainFiles = append(mainFiles, file)
		}
	}

	packageComplexity := &PackageComplexity{
		PackagePath: packagePath,
		Files:       make([]FileComplexity, 0, len(mainFiles)),
		Violations:  make([]ComplexityViolation, 0),
	}

	for _, filePath := range mainFiles {
		fileComplexity, err := a.AnalyzeFile(filePath)
		if err != nil {
			// Log error but continue with other files
			packageComplexity.Violations = append(packageComplexity.Violations, ComplexityViolation{
				Type:       "AnalysisError",
				Severity:   "Warning",
				Message:    fmt.Sprintf("Failed to analyze file: %s", err.Error()),
				FilePath:   filePath,
				LineNumber: 0,
			})
			continue
		}
		packageComplexity.Files = append(packageComplexity.Files, *fileComplexity)
	}

	// Calculate package-level metrics
	a.calculatePackageMetrics(packageComplexity)

	return packageComplexity, nil
}

// AnalyzeProject analyzes the entire project for complexity metrics
func (a *DefaultComplexityAnalyzer) AnalyzeProject(projectRoot string) (*ProjectComplexity, error) {
	project := &ProjectComplexity{
		ProjectRoot: projectRoot,
		Packages:    make([]PackageComplexity, 0),
		GeneratedAt: time.Now(),
	}

	// Find all Go packages in the project
	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && a.containsGoFiles(path) {
			// Skip vendor, .git, and other excluded directories
			if a.shouldSkipDirectory(path) {
				return filepath.SkipDir
			}

			packageComplexity, err := a.AnalyzePackage(path)
			if err != nil {
				// Log error but continue with other packages
				return nil
			}

			if len(packageComplexity.Files) > 0 {
				project.Packages = append(project.Packages, *packageComplexity)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk project directory: %w", err)
	}

	// Calculate project-level summary
	a.calculateProjectSummary(project)

	return project, nil
}

// countLinesOfCode counts non-empty, non-comment lines
func (a *DefaultComplexityAnalyzer) countLinesOfCode(content string) int {
	lines := strings.Split(content, "\n")
	count := 0
	inBlockComment := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle block comments
		if strings.Contains(line, "/*") {
			inBlockComment = true
		}
		if strings.Contains(line, "*/") {
			inBlockComment = false
			continue
		}
		if inBlockComment {
			continue
		}

		// Skip single-line comments
		if strings.HasPrefix(line, "//") {
			continue
		}

		count++
	}

	return count
}

// containsGoFiles checks if a directory contains Go files
func (a *DefaultComplexityAnalyzer) containsGoFiles(dir string) bool {
	files, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") {
			return true
		}
	}
	return false
}

// shouldSkipDirectory determines if a directory should be skipped during analysis
func (a *DefaultComplexityAnalyzer) shouldSkipDirectory(path string) bool {
	skipDirs := []string{
		"vendor", ".git", ".trunk", "node_modules", ".windsurf",
		"coverage", ".cache", "tmp", "temp",
	}

	for _, skipDir := range skipDirs {
		if strings.Contains(path, skipDir) {
			return true
		}
	}
	return false
}
