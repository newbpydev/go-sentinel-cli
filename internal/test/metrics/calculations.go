package metrics

import (
	"math"
)

// calculateFileMetrics calculates aggregate metrics for a file
func (a *DefaultComplexityAnalyzer) calculateFileMetrics(file *FileComplexity) {
	if len(file.Functions) == 0 {
		return
	}

	// Calculate average cyclomatic complexity
	totalComplexity := 0
	for _, fn := range file.Functions {
		totalComplexity += fn.CyclomaticComplexity
	}
	file.AverageCyclomaticComplexity = float64(totalComplexity) / float64(len(file.Functions))

	// Calculate maintainability index
	file.MaintainabilityIndex = a.calculateMaintainabilityIndex(file)

	// Calculate technical debt in minutes
	file.TechnicalDebtMinutes = a.calculateTechnicalDebt(file)
}

// calculatePackageMetrics calculates aggregate metrics for a package
func (a *DefaultComplexityAnalyzer) calculatePackageMetrics(pkg *PackageComplexity) {
	if len(pkg.Files) == 0 {
		return
	}

	totalComplexity := 0.0
	totalMaintainability := 0.0
	totalTechnicalDebtMinutes := 0
	totalFunctions := 0

	for _, file := range pkg.Files {
		pkg.TotalLinesOfCode += file.LinesOfCode
		totalFunctions += len(file.Functions)
		totalComplexity += file.AverageCyclomaticComplexity * float64(len(file.Functions))
		totalMaintainability += file.MaintainabilityIndex
		totalTechnicalDebtMinutes += file.TechnicalDebtMinutes

		// Aggregate violations
		pkg.Violations = append(pkg.Violations, file.Violations...)
	}

	pkg.TotalFunctions = totalFunctions
	if totalFunctions > 0 {
		pkg.AverageCyclomaticComplexity = totalComplexity / float64(totalFunctions)
	}
	if len(pkg.Files) > 0 {
		pkg.MaintainabilityIndex = totalMaintainability / float64(len(pkg.Files))
	}
	pkg.TechnicalDebtHours = float64(totalTechnicalDebtMinutes) / 60.0
}

// calculateProjectSummary calculates the overall project complexity summary
func (a *DefaultComplexityAnalyzer) calculateProjectSummary(project *ProjectComplexity) {
	summary := &project.Summary

	totalComplexity := 0.0
	totalMaintainability := 0.0
	totalTechnicalDebtMinutes := 0
	totalFiles := 0
	totalFunctions := 0
	totalViolations := 0

	for _, pkg := range project.Packages {
		totalFiles += len(pkg.Files)
		summary.TotalLinesOfCode += pkg.TotalLinesOfCode
		totalFunctions += pkg.TotalFunctions
		totalComplexity += pkg.AverageCyclomaticComplexity * float64(pkg.TotalFunctions)
		totalMaintainability += pkg.MaintainabilityIndex * float64(len(pkg.Files))
		totalTechnicalDebtMinutes += int(pkg.TechnicalDebtHours * 60)
		totalViolations += len(pkg.Violations)
	}

	summary.TotalFiles = totalFiles
	summary.TotalFunctions = totalFunctions
	summary.ViolationCount = totalViolations

	if totalFunctions > 0 {
		summary.AverageCyclomaticComplexity = totalComplexity / float64(totalFunctions)
	}
	if totalFiles > 0 {
		summary.MaintainabilityIndex = totalMaintainability / float64(totalFiles)
	}

	summary.TechnicalDebtDays = float64(totalTechnicalDebtMinutes) / (60.0 * 8.0) // 8-hour work days
	summary.QualityGrade = a.calculateQualityGrade(summary)
}

// calculateMaintainabilityIndex calculates the maintainability index for a file
// Based on the standard Maintainability Index formula:
// MI = 171 - 5.2 * ln(V) - 0.23 * G - 16.2 * ln(LOC)
// Where: V = Halstead Volume, G = Cyclomatic Complexity, LOC = Lines of Code
func (a *DefaultComplexityAnalyzer) calculateMaintainabilityIndex(file *FileComplexity) float64 {
	if len(file.Functions) == 0 || file.LinesOfCode == 0 {
		return 100.0 // Perfect score for empty/trivial files
	}

	// Calculate average cyclomatic complexity
	avgComplexity := file.AverageCyclomaticComplexity

	// Simplified Halstead Volume estimation
	// In practice, this would require more detailed AST analysis
	estimatedVolume := float64(file.LinesOfCode) * 4.0 // Rough estimation

	// Apply the maintainability index formula
	mi := 171.0 -
		5.2*math.Log(estimatedVolume) -
		0.23*avgComplexity -
		16.2*math.Log(float64(file.LinesOfCode))

	// Normalize to 0-100 scale
	if mi < 0 {
		mi = 0
	} else if mi > 100 {
		mi = 100
	}

	return mi
}

// calculateTechnicalDebt estimates technical debt in minutes based on violations
func (a *DefaultComplexityAnalyzer) calculateTechnicalDebt(file *FileComplexity) int {
	totalDebt := 0

	for _, violation := range file.Violations {
		switch violation.Type {
		case "CyclomaticComplexity":
			// High complexity functions take longer to understand and modify
			if complexity, ok := violation.ActualValue.(int); ok {
				excess := complexity - a.thresholds.CyclomaticComplexity
				totalDebt += excess * 10 // 10 minutes per excess complexity point
			}
		case "FunctionLength":
			// Long functions are harder to maintain
			if length, ok := violation.ActualValue.(int); ok {
				excess := length - a.thresholds.FunctionLength
				totalDebt += excess / 2 // 0.5 minutes per excess line
			}
		case "ParameterCount":
			// Functions with many parameters are harder to use
			if params, ok := violation.ActualValue.(int); ok {
				excess := params - 5    // Standard threshold
				totalDebt += excess * 5 // 5 minutes per excess parameter
			}
		case "NestingDepth":
			// Deep nesting makes code harder to follow
			if nesting, ok := violation.ActualValue.(int); ok {
				excess := nesting - 4    // Standard threshold
				totalDebt += excess * 15 // 15 minutes per excess nesting level
			}
		}
	}

	return totalDebt
}

// calculateQualityGrade assigns a letter grade based on overall project quality
func (a *DefaultComplexityAnalyzer) calculateQualityGrade(summary *ComplexitySummary) string {
	// Calculate complexity score with more reasonable scaling
	// Excellent (1-3): 100-90, Good (4-7): 89-70, Fair (8-10): 69-50, Poor (>10): <50
	var complexityScore float64
	if summary.AverageCyclomaticComplexity <= 3.0 {
		complexityScore = 100.0 - (summary.AverageCyclomaticComplexity-1.0)*5.0
	} else if summary.AverageCyclomaticComplexity <= 7.0 {
		complexityScore = 90.0 - (summary.AverageCyclomaticComplexity-3.0)*5.0
	} else if summary.AverageCyclomaticComplexity <= 10.0 {
		complexityScore = 70.0 - (summary.AverageCyclomaticComplexity-7.0)*6.67
	} else {
		complexityScore = 50.0 - (summary.AverageCyclomaticComplexity-10.0)*5.0
	}

	maintainabilityScore := summary.MaintainabilityIndex

	// Technical debt score - more gradual penalty
	debtScore := 100.0 - (summary.TechnicalDebtDays * 8.0) // 8 points per day instead of 10

	// Violation score - more reasonable penalty for small violation counts
	violationScore := 100.0 - math.Min(float64(summary.ViolationCount)*1.5, 50.0) // Cap at 50 point penalty

	// Ensure scores don't go below 0
	complexityScore = math.Max(0, complexityScore)
	debtScore = math.Max(0, debtScore)
	violationScore = math.Max(0, violationScore)

	// Calculate weighted average (maintainability has highest weight)
	overallScore := (maintainabilityScore*0.4 + complexityScore*0.3 + debtScore*0.2 + violationScore*0.1)

	// Assign letter grades
	if overallScore >= 90 {
		return "A"
	} else if overallScore >= 80 {
		return "B"
	} else if overallScore >= 70 {
		return "C"
	} else if overallScore >= 60 {
		return "D"
	} else {
		return "F"
	}
}

// checkFileViolations checks for file-level violations
func (a *DefaultComplexityAnalyzer) checkFileViolations(file *FileComplexity) {
	// Check file length
	if file.LinesOfCode > a.thresholds.LinesOfCode {
		file.Violations = append(file.Violations, ComplexityViolation{
			Type:           "FileLength",
			Severity:       a.getFileSeverity("FileLength", file.LinesOfCode),
			Message:        "File is too long and should be split into smaller files",
			FilePath:       file.FilePath,
			LineNumber:     1,
			ActualValue:    file.LinesOfCode,
			ThresholdValue: a.thresholds.LinesOfCode,
		})
	}

	// Check maintainability index
	if file.MaintainabilityIndex < a.thresholds.MaintainabilityIndex {
		file.Violations = append(file.Violations, ComplexityViolation{
			Type:           "MaintainabilityIndex",
			Severity:       "Warning",
			Message:        "File has low maintainability index",
			FilePath:       file.FilePath,
			LineNumber:     1,
			ActualValue:    file.MaintainabilityIndex,
			ThresholdValue: a.thresholds.MaintainabilityIndex,
		})
	}

	// Check technical debt ratio
	debtRatio := float64(file.TechnicalDebtMinutes) / float64(file.LinesOfCode) * 100
	if debtRatio > a.thresholds.TechnicalDebtRatio {
		file.Violations = append(file.Violations, ComplexityViolation{
			Type:           "TechnicalDebtRatio",
			Severity:       "Major",
			Message:        "File has high technical debt ratio",
			FilePath:       file.FilePath,
			LineNumber:     1,
			ActualValue:    debtRatio,
			ThresholdValue: a.thresholds.TechnicalDebtRatio,
		})
	}
}

// getFileSeverity determines severity for file-level violations
func (a *DefaultComplexityAnalyzer) getFileSeverity(metricType string, value int) string {
	var threshold int
	switch metricType {
	case "FileLength":
		threshold = a.thresholds.LinesOfCode
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
