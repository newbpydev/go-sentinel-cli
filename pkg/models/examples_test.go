package models

import (
	"testing"
)

// TestExample_errorHandling tests the error handling example
func TestExample_errorHandling(t *testing.T) {
	t.Parallel()

	// This test ensures the example function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Example_errorHandling panicked: %v", r)
		}
	}()

	Example_errorHandling()
}

// TestExample_testResults tests the test results example
func TestExample_testResults(t *testing.T) {
	t.Parallel()

	// This test ensures the example function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Example_testResults panicked: %v", r)
		}
	}()

	Example_testResults()
}

// TestExample_fileChanges tests the file changes example
func TestExample_fileChanges(t *testing.T) {
	t.Parallel()

	// This test ensures the example function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Example_fileChanges panicked: %v", r)
		}
	}()

	Example_fileChanges()
}

// TestExample_configuration tests the configuration example
func TestExample_configuration(t *testing.T) {
	t.Parallel()

	// This test ensures the example function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Example_configuration panicked: %v", r)
		}
	}()

	Example_configuration()
}

// TestExample_testStatus tests the test status example
func TestExample_testStatus(t *testing.T) {
	t.Parallel()

	// This test ensures the example function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Example_testStatus panicked: %v", r)
		}
	}()

	Example_testStatus()
}

// TestExample_coverage tests the coverage example
func TestExample_coverage(t *testing.T) {
	t.Parallel()

	// This test ensures the example function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Example_coverage panicked: %v", r)
		}
	}()

	Example_coverage()
}

// TestCreateExamplePassingTest tests the createExamplePassingTest helper
func TestCreateExamplePassingTest(t *testing.T) {
	t.Parallel()

	test := createExamplePassingTest()

	if test == nil {
		t.Fatal("createExamplePassingTest should not return nil")
	}

	if test.Status != TestStatusPassed {
		t.Errorf("Expected status %v, got %v", TestStatusPassed, test.Status)
	}

	if test.Name == "" {
		t.Error("Test name should not be empty")
	}
}

// TestCreateExampleFailingTest tests the createExampleFailingTest helper
func TestCreateExampleFailingTest(t *testing.T) {
	t.Parallel()

	test := createExampleFailingTest()

	if test == nil {
		t.Fatal("createExampleFailingTest should not return nil")
	}

	if test.Status != TestStatusFailed {
		t.Errorf("Expected status %v, got %v", TestStatusFailed, test.Status)
	}

	if test.Name == "" {
		t.Error("Test name should not be empty")
	}
}

// TestCreateExamplePackageResult tests the createExamplePackageResult helper
func TestCreateExamplePackageResult(t *testing.T) {
	t.Parallel()

	test := createExamplePassingTest()
	failingTest := createExampleFailingTest()
	pkg := createExamplePackageResult(test, failingTest)

	if pkg == nil {
		t.Fatal("createExamplePackageResult should not return nil")
	}

	if pkg.Package == "" {
		t.Error("Package name should not be empty")
	}

	if len(pkg.Tests) == 0 {
		t.Error("Package should have tests")
	}
}

// TestCreateExampleTestSummary tests the createExampleTestSummary helper
func TestCreateExampleTestSummary(t *testing.T) {
	t.Parallel()

	test := createExamplePassingTest()
	failingTest := createExampleFailingTest()
	pkg := createExamplePackageResult(test, failingTest)
	summary := createExampleTestSummary(pkg)

	if summary == nil {
		t.Fatal("createExampleTestSummary should not return nil")
	}

	if summary.TotalTests == 0 {
		t.Error("Summary should have total tests")
	}
}

// TestDisplayTestResults tests the displayTestResults helper
func TestDisplayTestResults(t *testing.T) {
	t.Parallel()

	// This test ensures the function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayTestResults panicked: %v", r)
		}
	}()

	test := createExamplePassingTest()
	failingTest := createExampleFailingTest()
	pkg := createExamplePackageResult(test, failingTest)
	summary := createExampleTestSummary(pkg)
	displayTestResults(pkg, summary, test, failingTest)
}

// TestCreateExampleTestConfiguration tests the createExampleTestConfiguration helper
func TestCreateExampleTestConfiguration(t *testing.T) {
	t.Parallel()

	config := createExampleTestConfiguration()

	if config == nil {
		t.Fatal("createExampleTestConfiguration should not return nil")
	}

	if len(config.Packages) == 0 {
		t.Error("Configuration should have packages")
	}
}

// TestCreateExampleWatchConfiguration tests the createExampleWatchConfiguration helper
func TestCreateExampleWatchConfiguration(t *testing.T) {
	t.Parallel()

	config := createExampleWatchConfiguration()

	if config == nil {
		t.Fatal("createExampleWatchConfiguration should not return nil")
	}

	if len(config.Paths) == 0 {
		t.Error("Watch configuration should have paths")
	}
}

// TestDisplayConfigurations tests the displayConfigurations helper
func TestDisplayConfigurations(t *testing.T) {
	t.Parallel()

	// This test ensures the function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayConfigurations panicked: %v", r)
		}
	}()

	testConfig := createExampleTestConfiguration()
	watchConfig := createExampleWatchConfiguration()
	displayConfigurations(testConfig, watchConfig)
}

// TestCreateExampleFunctionCoverage tests the createExampleFunctionCoverage helper
func TestCreateExampleFunctionCoverage(t *testing.T) {
	t.Parallel()

	coverage := createExampleFunctionCoverage()

	if coverage == nil {
		t.Fatal("createExampleFunctionCoverage should not return nil")
	}

	if coverage.Name == "" {
		t.Error("Function coverage should have a name")
	}
}

// TestCreateExampleFileCoverage tests the createExampleFileCoverage helper
func TestCreateExampleFileCoverage(t *testing.T) {
	t.Parallel()

	coverage := createExampleFileCoverage()

	if coverage == nil {
		t.Fatal("createExampleFileCoverage should not return nil")
	}

	if coverage.FilePath == "" {
		t.Error("File coverage should have a file path")
	}
}

// TestCreateExamplePackageCoverage tests the createExamplePackageCoverage helper
func TestCreateExamplePackageCoverage(t *testing.T) {
	t.Parallel()

	fileCoverage := createExampleFileCoverage()
	functionCoverage := createExampleFunctionCoverage()
	coverage := createExamplePackageCoverage(fileCoverage, functionCoverage)

	if coverage == nil {
		t.Fatal("createExamplePackageCoverage should not return nil")
	}

	if coverage.Package == "" {
		t.Error("Package coverage should have a package name")
	}
}

// TestCreateExampleTestCoverage tests the createExampleTestCoverage helper
func TestCreateExampleTestCoverage(t *testing.T) {
	t.Parallel()

	fileCoverage := createExampleFileCoverage()
	coverage := createExampleTestCoverage(fileCoverage)

	if coverage == nil {
		t.Fatal("createExampleTestCoverage should not return nil")
	}

	if coverage.Percentage < 0 || coverage.Percentage > 100 {
		t.Errorf("Coverage percentage should be between 0 and 100, got %f", coverage.Percentage)
	}
}

// TestDisplayCoverageInformation tests the displayCoverageInformation helper
func TestDisplayCoverageInformation(t *testing.T) {
	t.Parallel()

	// This test ensures the function runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayCoverageInformation panicked: %v", r)
		}
	}()

	fileCoverage := createExampleFileCoverage()
	functionCoverage := createExampleFunctionCoverage()
	packageCoverage := createExamplePackageCoverage(fileCoverage, functionCoverage)
	testCoverage := createExampleTestCoverage(fileCoverage)

	displayCoverageInformation(packageCoverage, fileCoverage, functionCoverage, testCoverage)
}

// TestDisplayCoverageInformation_BelowThreshold tests the coverage threshold failure branch
func TestDisplayCoverageInformation_BelowThreshold(t *testing.T) {
	t.Parallel()

	// This test ensures the function runs without panicking and covers the threshold failure branch
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayCoverageInformation panicked: %v", r)
		}
	}()

	// Create coverage data that's below the 80% threshold
	funcCoverage := &FunctionCoverage{
		Name:       "LowCoverageFunction",
		FilePath:   "low_coverage.go",
		StartLine:  1,
		EndLine:    10,
		Percentage: 50.0,
		IsCovered:  true,
		CallCount:  5,
	}

	fileCoverage := &FileCoverage{
		FilePath:          "low_coverage.go",
		Percentage:        60.0,
		CoveredLines:      60,
		TotalLines:        100,
		CoveredStatements: 30,
		TotalStatements:   50,
		LinesCovered:      []int{1, 2, 3},
		LinesUncovered:    []int{4, 5, 6},
	}

	pkgCoverage := &PackageCoverage{
		Package:           "github.com/example/low",
		Percentage:        65.0,
		CoveredLines:      65,
		TotalLines:        100,
		CoveredStatements: 32,
		TotalStatements:   50,
		Files: map[string]*FileCoverage{
			"low_coverage.go": fileCoverage,
		},
		Functions: map[string]*FunctionCoverage{
			"LowCoverageFunction": funcCoverage,
		},
	}

	// Create test coverage below threshold (80%)
	testCoverage := &TestCoverage{
		Percentage:        70.0, // Below 80% threshold
		CoveredLines:      70,
		TotalLines:        100,
		CoveredStatements: 35,
		TotalStatements:   50,
		Files: map[string]*FileCoverage{
			"low_coverage.go": fileCoverage,
		},
	}

	// This should trigger the "below threshold" branch
	displayCoverageInformation(pkgCoverage, fileCoverage, funcCoverage, testCoverage)
}
