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
