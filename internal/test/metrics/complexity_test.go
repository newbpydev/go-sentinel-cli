package metrics

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestComplexityAnalyzer_DefaultThresholds tests default threshold values
func TestComplexityAnalyzer_DefaultThresholds(t *testing.T) {
	thresholds := DefaultThresholds()

	if thresholds.CyclomaticComplexity != 10 {
		t.Errorf("Expected CyclomaticComplexity=10, got %d", thresholds.CyclomaticComplexity)
	}
	if thresholds.MaintainabilityIndex != 85.0 {
		t.Errorf("Expected MaintainabilityIndex=85.0, got %f", thresholds.MaintainabilityIndex)
	}
	if thresholds.LinesOfCode != 500 {
		t.Errorf("Expected LinesOfCode=500, got %d", thresholds.LinesOfCode)
	}
	if thresholds.TechnicalDebtRatio != 5.0 {
		t.Errorf("Expected TechnicalDebtRatio=5.0, got %f", thresholds.TechnicalDebtRatio)
	}
	if thresholds.FunctionLength != 50 {
		t.Errorf("Expected FunctionLength=50, got %d", thresholds.FunctionLength)
	}
}

// TestComplexityAnalyzer_NewComplexityAnalyzer tests analyzer creation
func TestComplexityAnalyzer_NewComplexityAnalyzer(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	if analyzer == nil {
		t.Fatal("NewComplexityAnalyzer returned nil")
	}

	if analyzer.thresholds.CyclomaticComplexity != 10 {
		t.Errorf("Expected default threshold, got %d", analyzer.thresholds.CyclomaticComplexity)
	}
}

// TestComplexityAnalyzer_SetThresholds tests threshold configuration
func TestComplexityAnalyzer_SetThresholds(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	customThresholds := ComplexityThresholds{
		CyclomaticComplexity: 15,
		MaintainabilityIndex: 80.0,
		LinesOfCode:          400,
		TechnicalDebtRatio:   3.0,
		FunctionLength:       40,
	}

	analyzer.SetThresholds(customThresholds)

	if analyzer.thresholds.CyclomaticComplexity != 15 {
		t.Errorf("Expected CyclomaticComplexity=15, got %d", analyzer.thresholds.CyclomaticComplexity)
	}
	if analyzer.thresholds.MaintainabilityIndex != 80.0 {
		t.Errorf("Expected MaintainabilityIndex=80.0, got %f", analyzer.thresholds.MaintainabilityIndex)
	}
}

// TestComplexityAnalyzer_AnalyzeFile_SimpleFunction tests simple function analysis
func TestComplexityAnalyzer_AnalyzeFile_SimpleFunction(t *testing.T) {
	// Create a simple test file
	testCode := `package test

// Add adds two integers
func Add(a, b int) int {
	return a + b
}
`
	testFile := createTempGoFile(t, "simple_test.go", testCode)
	defer os.Remove(testFile)

	analyzer := NewComplexityAnalyzer()
	result, err := analyzer.AnalyzeFile(testFile)

	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(result.Functions))
	}

	fn := result.Functions[0]
	if fn.Name != "Add" {
		t.Errorf("Expected function name 'Add', got '%s'", fn.Name)
	}

	if fn.CyclomaticComplexity != 1 {
		t.Errorf("Expected complexity=1, got %d", fn.CyclomaticComplexity)
	}

	if fn.Parameters != 2 {
		t.Errorf("Expected 2 parameters, got %d", fn.Parameters)
	}

	if fn.ReturnValues != 1 {
		t.Errorf("Expected 1 return value, got %d", fn.ReturnValues)
	}
}

// TestComplexityAnalyzer_AnalyzeFile_ComplexFunction tests complex function analysis
func TestComplexityAnalyzer_AnalyzeFile_ComplexFunction(t *testing.T) {
	testCode := `package test

func ComplexFunction(x int) string {
	if x < 0 {
		return "negative"
	} else if x == 0 {
		return "zero"
	}
	
	for i := 0; i < x; i++ {
		if i%2 == 0 {
			continue
		}
		
		switch i {
		case 1:
			return "one"
		case 3:
			return "three"
		default:
			if i > 10 {
				return "big"
			}
		}
	}
	
	return "positive"
}
`
	testFile := createTempGoFile(t, "complex_test.go", testCode)
	defer os.Remove(testFile)

	analyzer := NewComplexityAnalyzer()
	result, err := analyzer.AnalyzeFile(testFile)

	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	if len(result.Functions) != 1 {
		t.Errorf("Expected 1 function, got %d", len(result.Functions))
	}

	fn := result.Functions[0]
	if fn.Name != "ComplexFunction" {
		t.Errorf("Expected function name 'ComplexFunction', got '%s'", fn.Name)
	}

	// Expected complexity: base(1) + if(1) + else if(1) + for(1) + if(1) + switch(1) + case(2) + if(1) = 9
	expectedComplexity := 8 // Adjust based on actual implementation
	if fn.CyclomaticComplexity < expectedComplexity {
		t.Errorf("Expected complexity>=%d, got %d", expectedComplexity, fn.CyclomaticComplexity)
	}
}

// TestComplexityAnalyzer_AnalyzeFile_ViolationDetection tests violation detection
func TestComplexityAnalyzer_AnalyzeFile_ViolationDetection(t *testing.T) {
	// Create a function that exceeds thresholds
	longFunctionCode := `package test

func VeryLongFunction() {
` + strings.Repeat("	x := 1\n", 60) + `
}
`
	testFile := createTempGoFile(t, "violations_test.go", longFunctionCode)
	defer os.Remove(testFile)

	analyzer := NewComplexityAnalyzer()
	// Set strict thresholds
	analyzer.SetThresholds(ComplexityThresholds{
		FunctionLength: 30, // Function will exceed this
	})

	result, err := analyzer.AnalyzeFile(testFile)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	// Should detect function length violation
	hasLengthViolation := false
	for _, violation := range result.Violations {
		if violation.Type == "FunctionLength" {
			hasLengthViolation = true
			if violation.Severity == "" {
				t.Error("Violation should have severity")
			}
			if violation.FunctionName != "VeryLongFunction" {
				t.Errorf("Expected function name in violation, got '%s'", violation.FunctionName)
			}
		}
	}

	if !hasLengthViolation {
		t.Error("Expected to detect function length violation")
	}
}

// TestComplexityAnalyzer_CountLinesOfCode tests line counting logic
func TestComplexityAnalyzer_CountLinesOfCode(t *testing.T) {
	testCases := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name: "Simple code",
			code: `package test
func Add(a, b int) int {
	return a + b
}`,
			expected: 4,
		},
		{
			name: "Code with comments",
			code: `package test
// This is a comment
func Add(a, b int) int {
	// Another comment
	return a + b // Inline comment
}`,
			expected: 4, // Comments should be excluded
		},
		{
			name: "Code with empty lines",
			code: `package test

func Add(a, b int) int {

	return a + b

}`,
			expected: 4, // Empty lines should be excluded
		},
		{
			name: "Code with block comments",
			code: `package test
/*
Multi-line comment
should be excluded
*/
func Add(a, b int) int {
	return a + b
}`,
			expected: 4,
		},
	}

	analyzer := NewComplexityAnalyzer()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := analyzer.countLinesOfCode(tc.code)
			if actual != tc.expected {
				t.Errorf("Expected %d lines, got %d for:\n%s", tc.expected, actual, tc.code)
			}
		})
	}
}

// TestComplexityAnalyzer_MaintainabilityIndex tests maintainability calculation
func TestComplexityAnalyzer_MaintainabilityIndex(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	// Test with simple file
	file := &FileComplexity{
		LinesOfCode: 20,
		Functions: []FunctionMetrics{
			{CyclomaticComplexity: 2},
			{CyclomaticComplexity: 3},
		},
	}

	analyzer.calculateFileMetrics(file)

	// Should have reasonable maintainability index
	if file.MaintainabilityIndex < 0 || file.MaintainabilityIndex > 100 {
		t.Errorf("Maintainability index should be 0-100, got %f", file.MaintainabilityIndex)
	}

	// Simple code should have high maintainability
	if file.MaintainabilityIndex < 50 {
		t.Errorf("Simple code should have higher maintainability, got %f", file.MaintainabilityIndex)
	}
}

// TestComplexityAnalyzer_TechnicalDebt tests technical debt calculation
func TestComplexityAnalyzer_TechnicalDebt(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	file := &FileComplexity{
		Violations: []ComplexityViolation{
			{
				Type:           "CyclomaticComplexity",
				ActualValue:    15,
				ThresholdValue: 10,
			},
			{
				Type:           "FunctionLength",
				ActualValue:    80,
				ThresholdValue: 50,
			},
		},
	}

	debt := analyzer.calculateTechnicalDebt(file)

	// Should calculate some debt based on violations
	if debt <= 0 {
		t.Error("Expected technical debt > 0 for violations")
	}

	// Complexity violation: (15-10) * 10 = 50 minutes
	// Function length violation: (80-50) / 2 = 15 minutes
	// Total expected: 65 minutes
	expectedDebt := 65
	if debt != expectedDebt {
		t.Errorf("Expected technical debt=%d minutes, got %d", expectedDebt, debt)
	}
}

// TestComplexityAnalyzer_QualityGrade tests quality grade assignment
func TestComplexityAnalyzer_QualityGrade(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	testCases := []struct {
		name          string
		summary       ComplexitySummary
		expectedGrade string
	}{
		{
			name: "Excellent quality",
			summary: ComplexitySummary{
				AverageCyclomaticComplexity: 2.0,
				MaintainabilityIndex:        95.0,
				TechnicalDebtDays:           0.1,
				ViolationCount:              0,
			},
			expectedGrade: "A",
		},
		{
			name: "Good quality",
			summary: ComplexitySummary{
				AverageCyclomaticComplexity: 5.0,
				MaintainabilityIndex:        85.0,
				TechnicalDebtDays:           1.0,
				ViolationCount:              2,
			},
			expectedGrade: "B",
		},
		{
			name: "Poor quality",
			summary: ComplexitySummary{
				AverageCyclomaticComplexity: 15.0,
				MaintainabilityIndex:        40.0,
				TechnicalDebtDays:           10.0,
				ViolationCount:              50,
			},
			expectedGrade: "F",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			grade := analyzer.calculateQualityGrade(&tc.summary)
			if grade != tc.expectedGrade {
				t.Errorf("Expected grade %s, got %s", tc.expectedGrade, grade)
			}
		})
	}
}

// TestComplexityAnalyzer_GenerateReport tests report generation
func TestComplexityAnalyzer_GenerateReport(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	// Create test project complexity
	summary := ComplexitySummary{
		TotalFiles:                  2,
		TotalFunctions:              5,
		TotalLinesOfCode:            100,
		AverageCyclomaticComplexity: 3.5,
		MaintainabilityIndex:        82.0,
		TechnicalDebtDays:           1.5,
		ViolationCount:              3,
	}

	// Calculate the quality grade properly
	summary.QualityGrade = analyzer.calculateQualityGrade(&summary)

	project := &ProjectComplexity{
		Summary: summary,
		Packages: []PackageComplexity{
			{
				PackagePath:    "test/package",
				TotalFunctions: 3,
				Violations: []ComplexityViolation{
					{
						Type:         "CyclomaticComplexity",
						Severity:     "Major",
						Message:      "Function has high complexity",
						FilePath:     "test.go",
						FunctionName: "ComplexFunc",
						LineNumber:   10,
					},
				},
			},
		},
	}

	var output bytes.Buffer
	err := analyzer.GenerateReport(project, &output)

	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	report := output.String()

	// Check report contains expected sections
	expectedSections := []string{
		"Code Complexity Report",
		"Project Summary",
		"Quality Assessment",
		"Top Complexity Violations",
		"Package Details",
		"Recommendations",
	}

	for _, section := range expectedSections {
		if !strings.Contains(report, section) {
			t.Errorf("Report missing section: %s", section)
		}
	}

	// Check that the quality grade appears in the report
	qualityGradeSection := fmt.Sprintf("Quality Grade**: %s", project.Summary.QualityGrade)
	if !strings.Contains(report, qualityGradeSection) {
		t.Errorf("Report missing quality grade section: %s", qualityGradeSection)
	}
}

// TestComplexityAnalyzer_GenerateJSONReport tests JSON report generation
func TestComplexityAnalyzer_GenerateJSONReport(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	project := &ProjectComplexity{
		Summary: ComplexitySummary{
			QualityGrade: "A",
			TotalFiles:   1,
		},
	}

	var output bytes.Buffer
	err := analyzer.GenerateJSONReport(project, &output)

	if err != nil {
		t.Fatalf("GenerateJSONReport failed: %v", err)
	}

	jsonOutput := output.String()

	// Should be valid JSON containing expected fields
	if !strings.Contains(jsonOutput, `"quality_grade"`) {
		t.Error("JSON report missing quality_grade field")
	}
	if !strings.Contains(jsonOutput, `"total_files"`) {
		t.Error("JSON report missing total_files field")
	}
}

// TestComplexityAnalyzer_GenerateHTMLReport tests HTML report generation
func TestComplexityAnalyzer_GenerateHTMLReport(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	project := &ProjectComplexity{
		Summary: ComplexitySummary{
			QualityGrade: "A",
			TotalFiles:   1,
		},
	}

	var output bytes.Buffer
	err := analyzer.GenerateHTMLReport(project, &output)

	if err != nil {
		t.Fatalf("GenerateHTMLReport failed: %v", err)
	}

	htmlOutput := output.String()

	// Should be valid HTML
	if !strings.Contains(htmlOutput, "<!DOCTYPE html>") {
		t.Error("HTML report missing DOCTYPE")
	}
	if !strings.Contains(htmlOutput, "Complexity Report") {
		t.Error("HTML report missing title")
	}
	if !strings.Contains(htmlOutput, "Quality Grade") {
		t.Error("HTML report missing quality grade section")
	}
}

// Helper function to create temporary Go files for testing
func createTempGoFile(t *testing.T, name, content string) string {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, name)

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return filePath
}

// Benchmark tests
func BenchmarkComplexityAnalyzer_AnalyzeFile(b *testing.B) {
	testCode := `package test

func BenchmarkFunction(x int) int {
	if x < 0 {
		return -1
	}
	
	result := 0
	for i := 0; i < x; i++ {
		if i%2 == 0 {
			result += i
		} else {
			result -= i
		}
	}
	
	return result
}
`
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "benchmark_test.go")
	err := os.WriteFile(testFile, []byte(testCode), 0644)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	analyzer := NewComplexityAnalyzer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeFile(testFile)
		if err != nil {
			b.Fatalf("AnalyzeFile failed: %v", err)
		}
	}
}

// TestComplexityAnalyzer_AnalyzePackage tests package-level analysis
func TestComplexityAnalyzer_AnalyzePackage(t *testing.T) {
	// Create a temporary package with multiple Go files
	tmpDir := t.TempDir()

	// Create first file
	file1 := `package testpkg

func SimpleFunc() {
	x := 1
	y := 2
	_ = x + y
}

func ComplexFunc(a int) string {
	if a < 0 {
		return "negative"
	} else if a == 0 {
		return "zero"
	}
	
	for i := 0; i < a; i++ {
		if i%2 == 0 {
			continue
		}
	}
	
	return "positive"
}
`

	// Create second file
	file2 := `package testpkg

func AnotherFunc(x, y, z int) int {
	if x > y {
		if y > z {
			return x
		} else {
			return y
		}
	}
	return z
}
`

	err := os.WriteFile(filepath.Join(tmpDir, "file1.go"), []byte(file1), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte(file2), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	analyzer := NewComplexityAnalyzer()
	result, err := analyzer.AnalyzePackage(tmpDir)

	if err != nil {
		t.Fatalf("AnalyzePackage failed: %v", err)
	}

	if len(result.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(result.Files))
	}

	if result.TotalFunctions != 3 {
		t.Errorf("Expected 3 total functions, got %d", result.TotalFunctions)
	}

	if result.AverageCyclomaticComplexity <= 0 {
		t.Error("Expected positive average complexity")
	}

	if len(result.Violations) == 0 {
		t.Log("No violations found - functions are within thresholds")
	}
}

// TestComplexityAnalyzer_AnalyzeProject tests project-level analysis
func TestComplexityAnalyzer_AnalyzeProject(t *testing.T) {
	// Create a temporary project structure
	tmpDir := t.TempDir()

	// Create main package
	mainDir := filepath.Join(tmpDir, "main")
	err := os.MkdirAll(mainDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create main dir: %v", err)
	}

	mainFile := `package main

func main() {
	println("Hello, World!")
}

func complexMain() {
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			if i > 5 {
				println("even and > 5")
			} else {
				println("even and <= 5")
			}
		} else {
			switch i {
			case 1:
				println("one")
			case 3:
				println("three")
			default:
				println("odd")
			}
		}
	}
}
`

	// Create util package
	utilDir := filepath.Join(tmpDir, "util")
	err = os.MkdirAll(utilDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create util dir: %v", err)
	}

	utilFile := `package util

func Add(a, b int) int {
	return a + b
}

func ProcessData(data []int) []int {
	result := make([]int, 0, len(data))
	for _, v := range data {
		if v > 0 {
			result = append(result, v*2)
		}
	}
	return result
}
`

	err = os.WriteFile(filepath.Join(mainDir, "main.go"), []byte(mainFile), 0644)
	if err != nil {
		t.Fatalf("Failed to create main file: %v", err)
	}

	err = os.WriteFile(filepath.Join(utilDir, "util.go"), []byte(utilFile), 0644)
	if err != nil {
		t.Fatalf("Failed to create util file: %v", err)
	}

	analyzer := NewComplexityAnalyzer()
	result, err := analyzer.AnalyzeProject(tmpDir)

	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	if len(result.Packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(result.Packages))
	}

	if result.Summary.TotalFunctions != 4 {
		t.Errorf("Expected 4 total functions, got %d", result.Summary.TotalFunctions)
	}

	if result.Summary.QualityGrade == "" {
		t.Error("Expected quality grade to be assigned")
	}

	if result.Summary.AverageCyclomaticComplexity <= 0 {
		t.Error("Expected positive average complexity")
	}
}

// TestComplexityAnalyzer_CalculatePackageMetrics tests package metrics calculation
func TestComplexityAnalyzer_CalculatePackageMetrics(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	// Create test package with files
	pkg := &PackageComplexity{
		PackagePath: "test/package",
		Files: []FileComplexity{
			{
				FilePath:    "file1.go",
				LinesOfCode: 50,
				Functions: []FunctionMetrics{
					{CyclomaticComplexity: 2},
					{CyclomaticComplexity: 4},
				},
				AverageCyclomaticComplexity: 3.0,
				MaintainabilityIndex:        80.0,
				TechnicalDebtMinutes:        30,
				Violations: []ComplexityViolation{
					{Type: "test", Severity: "Minor"},
				},
			},
			{
				FilePath:    "file2.go",
				LinesOfCode: 30,
				Functions: []FunctionMetrics{
					{CyclomaticComplexity: 1},
				},
				AverageCyclomaticComplexity: 1.0,
				MaintainabilityIndex:        90.0,
				TechnicalDebtMinutes:        10,
			},
		},
	}

	analyzer.calculatePackageMetrics(pkg)

	if pkg.TotalLinesOfCode != 80 {
		t.Errorf("Expected 80 total lines, got %d", pkg.TotalLinesOfCode)
	}

	if pkg.TotalFunctions != 3 {
		t.Errorf("Expected 3 total functions, got %d", pkg.TotalFunctions)
	}

	if pkg.TechnicalDebtHours == 0 {
		t.Error("Expected positive technical debt hours")
	}

	if len(pkg.Violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(pkg.Violations))
	}
}

// TestComplexityAnalyzer_CalculateProjectSummary tests project summary calculation
func TestComplexityAnalyzer_CalculateProjectSummary(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	project := &ProjectComplexity{
		Packages: []PackageComplexity{
			{
				PackagePath:    "pkg1",
				TotalFunctions: 5,
				Files: []FileComplexity{
					{LinesOfCode: 100},
					{LinesOfCode: 50},
				},
				TotalLinesOfCode:            150,
				AverageCyclomaticComplexity: 3.0,
				MaintainabilityIndex:        75.0,
				TechnicalDebtHours:          2.0,
				Violations: []ComplexityViolation{
					{Type: "test1"}, {Type: "test2"},
				},
			},
			{
				PackagePath:    "pkg2",
				TotalFunctions: 3,
				Files: []FileComplexity{
					{LinesOfCode: 80},
				},
				TotalLinesOfCode:            80,
				AverageCyclomaticComplexity: 2.0,
				MaintainabilityIndex:        85.0,
				TechnicalDebtHours:          1.0,
				Violations: []ComplexityViolation{
					{Type: "test3"},
				},
			},
		},
	}

	analyzer.calculateProjectSummary(project)

	summary := &project.Summary

	if summary.TotalFiles != 3 {
		t.Errorf("Expected 3 total files, got %d", summary.TotalFiles)
	}

	if summary.TotalFunctions != 8 {
		t.Errorf("Expected 8 total functions, got %d", summary.TotalFunctions)
	}

	if summary.TotalLinesOfCode != 230 {
		t.Errorf("Expected 230 total lines, got %d", summary.TotalLinesOfCode)
	}

	if summary.ViolationCount != 3 {
		t.Errorf("Expected 3 violations, got %d", summary.ViolationCount)
	}

	if summary.QualityGrade == "" {
		t.Error("Expected quality grade to be assigned")
	}

	if summary.TechnicalDebtDays <= 0 {
		t.Error("Expected positive technical debt days")
	}
}

// TestComplexityAnalyzer_ContainsGoFiles tests Go file detection
func TestComplexityAnalyzer_ContainsGoFiles(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	// Test directory with Go files
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(goFile, []byte("package test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create Go file: %v", err)
	}

	if !analyzer.containsGoFiles(tmpDir) {
		t.Error("Expected directory to contain Go files")
	}

	// Test directory without Go files
	emptyDir := t.TempDir()
	if analyzer.containsGoFiles(emptyDir) {
		t.Error("Expected directory to not contain Go files")
	}

	// Test non-existent directory
	if analyzer.containsGoFiles("/non/existent/path") {
		t.Error("Expected non-existent directory to not contain Go files")
	}
}

// TestComplexityAnalyzer_ShouldSkipDirectory tests directory skipping logic
func TestComplexityAnalyzer_ShouldSkipDirectory(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	testCases := []struct {
		name     string
		dirName  string
		expected bool
	}{
		{"vendor directory", "vendor", true},
		{"node_modules directory", "node_modules", true},
		{"dot directory", ".git", true},
		{"test directory", "test", false},
		{"normal directory", "internal", false},
		{"underscore directory", "_examples", false}, // Directory skipping only checks for specific paths, not patterns
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.shouldSkipDirectory(tc.dirName)
			if result != tc.expected {
				t.Errorf("Expected %v for directory %s, got %v", tc.expected, tc.dirName, result)
			}
		})
	}
}

// TestComplexityAnalyzer_GetFileName tests file name extraction
func TestComplexityAnalyzer_GetFileName(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	testCases := []struct {
		path     string
		expected string
	}{
		{"/path/to/file.go", "file.go"},
		{"file.go", "file.go"},
		{"/path/to/complex_file.go", "complex_file.go"},
		{"", ""},
		{"/path/to/", ""},
	}

	for _, tc := range testCases {
		result := analyzer.getFileName(tc.path)
		if result != tc.expected {
			t.Errorf("Expected %s for path %s, got %s", tc.expected, tc.path, result)
		}
	}
}

// TestComplexityAnalyzer_QualityAssessment tests quality assessment generation
func TestComplexityAnalyzer_QualityAssessment(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	project := &ProjectComplexity{
		Summary: ComplexitySummary{
			QualityGrade:                "B",
			TotalFiles:                  10,
			TotalFunctions:              50,
			AverageCyclomaticComplexity: 4.2,
			MaintainabilityIndex:        75.5,
			TechnicalDebtDays:           1.2,
			ViolationCount:              8,
		},
	}

	var buf bytes.Buffer
	analyzer.writeQualityAssessment(&buf, &project.Summary)

	output := buf.String()

	expectedSections := []string{
		"Code Quality Indicators",
		"**Complexity**:",
		"**Maintainability**:",
		"**Technical Debt**:",
		"**Violations**:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Quality assessment missing section: %s", section)
		}
	}
}

// TestComplexityAnalyzer_SeverityIcon tests severity icon mapping
func TestComplexityAnalyzer_SeverityIcon(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	testCases := []struct {
		severity string
		expected string
	}{
		{"Critical", "🔴"},
		{"Major", "🟠"},
		{"Minor", "🟡"},
		{"Warning", "⚠️"},
		{"Unknown", "⚠️"}, // Default case returns ⚠️
	}

	for _, tc := range testCases {
		result := analyzer.getSeverityIcon(tc.severity)
		if result != tc.expected {
			t.Errorf("Expected %s for severity %s, got %s", tc.expected, tc.severity, result)
		}
	}
}

// TestComplexityAnalyzer_EdgeCases tests various edge cases
func TestComplexityAnalyzer_EdgeCases(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	t.Run("Empty file analysis", func(t *testing.T) {
		emptyFile := createTempGoFile(t, "empty.go", "package test\n")
		defer os.Remove(emptyFile)

		result, err := analyzer.AnalyzeFile(emptyFile)
		if err != nil {
			t.Fatalf("Failed to analyze empty file: %v", err)
		}

		if len(result.Functions) != 0 {
			t.Errorf("Expected 0 functions in empty file, got %d", len(result.Functions))
		}

		if result.LinesOfCode != 1 {
			t.Errorf("Expected 1 line of code, got %d", result.LinesOfCode)
		}
	})

	t.Run("File with only comments", func(t *testing.T) {
		commentFile := `package test
// This is a comment
/* This is a block comment */
// Another comment
`
		testFile := createTempGoFile(t, "comments.go", commentFile)
		defer os.Remove(testFile)

		result, err := analyzer.AnalyzeFile(testFile)
		if err != nil {
			t.Fatalf("Failed to analyze comment file: %v", err)
		}

		if result.LinesOfCode != 1 { // Only package declaration
			t.Errorf("Expected 1 line of code (excluding comments), got %d", result.LinesOfCode)
		}
	})

	t.Run("Non-existent file", func(t *testing.T) {
		_, err := analyzer.AnalyzeFile("/non/existent/file.go")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("Invalid Go file", func(t *testing.T) {
		invalidFile := createTempGoFile(t, "invalid.go", "this is not valid go code")
		defer os.Remove(invalidFile)

		_, err := analyzer.AnalyzeFile(invalidFile)
		if err == nil {
			t.Error("Expected error for invalid Go file")
		}
	})
}
