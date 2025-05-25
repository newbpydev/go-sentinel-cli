package metrics

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// GenerateReport generates a detailed complexity report
func (a *DefaultComplexityAnalyzer) GenerateReport(complexity *ProjectComplexity, output io.Writer) error {
	return a.generateTextReport(complexity, output)
}

// GenerateJSONReport generates a JSON format report
func (a *DefaultComplexityAnalyzer) GenerateJSONReport(complexity *ProjectComplexity, output io.Writer) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(complexity)
}

// GenerateHTMLReport generates an HTML format report
func (a *DefaultComplexityAnalyzer) GenerateHTMLReport(complexity *ProjectComplexity, output io.Writer) error {
	return a.generateHTMLReport(complexity, output)
}

// generateTextReport generates a human-readable text report
func (a *DefaultComplexityAnalyzer) generateTextReport(complexity *ProjectComplexity, output io.Writer) error {
	fmt.Fprintf(output, "# Go Sentinel CLI - Code Complexity Report\n")
	fmt.Fprintf(output, "Generated: %s\n\n", complexity.GeneratedAt.Format(time.RFC3339))

	// Project Summary
	summary := complexity.Summary
	fmt.Fprintf(output, "## Project Summary\n")
	fmt.Fprintf(output, "**Quality Grade**: %s\n", summary.QualityGrade)
	fmt.Fprintf(output, "**Total Files**: %d\n", summary.TotalFiles)
	fmt.Fprintf(output, "**Total Functions**: %d\n", summary.TotalFunctions)
	fmt.Fprintf(output, "**Total Lines of Code**: %d\n", summary.TotalLinesOfCode)
	fmt.Fprintf(output, "**Average Cyclomatic Complexity**: %.2f\n", summary.AverageCyclomaticComplexity)
	fmt.Fprintf(output, "**Maintainability Index**: %.2f\n", summary.MaintainabilityIndex)
	fmt.Fprintf(output, "**Technical Debt**: %.2f days\n", summary.TechnicalDebtDays)
	fmt.Fprintf(output, "**Total Violations**: %d\n\n", summary.ViolationCount)

	// Quality Assessment
	fmt.Fprintf(output, "## Quality Assessment\n")
	a.writeQualityAssessment(output, &summary)
	fmt.Fprintf(output, "\n")

	// Top Violations
	fmt.Fprintf(output, "## Top Complexity Violations\n")
	a.writeTopViolations(output, complexity)
	fmt.Fprintf(output, "\n")

	// Package Details
	fmt.Fprintf(output, "## Package Details\n")
	for _, pkg := range complexity.Packages {
		a.writePackageDetails(output, &pkg)
	}

	// Recommendations
	fmt.Fprintf(output, "## Recommendations\n")
	a.writeRecommendations(output, complexity)

	return nil
}

// writeQualityAssessment writes quality assessment details
func (a *DefaultComplexityAnalyzer) writeQualityAssessment(output io.Writer, summary *ComplexitySummary) {
	fmt.Fprintf(output, "### Code Quality Indicators\n")

	// Complexity assessment
	if summary.AverageCyclomaticComplexity <= 5 {
		fmt.Fprintf(output, "✅ **Complexity**: Excellent (≤5.0)\n")
	} else if summary.AverageCyclomaticComplexity <= 10 {
		fmt.Fprintf(output, "⚠️  **Complexity**: Good (≤10.0)\n")
	} else {
		fmt.Fprintf(output, "❌ **Complexity**: Needs Improvement (>10.0)\n")
	}

	// Maintainability assessment
	if summary.MaintainabilityIndex >= 85 {
		fmt.Fprintf(output, "✅ **Maintainability**: Excellent (≥85.0)\n")
	} else if summary.MaintainabilityIndex >= 70 {
		fmt.Fprintf(output, "⚠️  **Maintainability**: Good (≥70.0)\n")
	} else {
		fmt.Fprintf(output, "❌ **Maintainability**: Needs Improvement (<70.0)\n")
	}

	// Technical debt assessment
	if summary.TechnicalDebtDays <= 1 {
		fmt.Fprintf(output, "✅ **Technical Debt**: Low (≤1 day)\n")
	} else if summary.TechnicalDebtDays <= 5 {
		fmt.Fprintf(output, "⚠️  **Technical Debt**: Moderate (≤5 days)\n")
	} else {
		fmt.Fprintf(output, "❌ **Technical Debt**: High (>5 days)\n")
	}

	// Violation assessment
	violationRatio := float64(summary.ViolationCount) / float64(summary.TotalFunctions) * 100
	if violationRatio <= 10 {
		fmt.Fprintf(output, "✅ **Violations**: Low (%.1f%% of functions)\n", violationRatio)
	} else if violationRatio <= 25 {
		fmt.Fprintf(output, "⚠️  **Violations**: Moderate (%.1f%% of functions)\n", violationRatio)
	} else {
		fmt.Fprintf(output, "❌ **Violations**: High (%.1f%% of functions)\n", violationRatio)
	}
}

// writeTopViolations writes the most critical violations
func (a *DefaultComplexityAnalyzer) writeTopViolations(output io.Writer, complexity *ProjectComplexity) {
	allViolations := make([]ComplexityViolation, 0)

	// Collect all violations
	for _, pkg := range complexity.Packages {
		allViolations = append(allViolations, pkg.Violations...)
	}

	// Sort by severity (Critical > Major > Minor > Warning)
	sort.Slice(allViolations, func(i, j int) bool {
		severityOrder := map[string]int{
			"Critical": 4,
			"Major":    3,
			"Minor":    2,
			"Warning":  1,
		}
		return severityOrder[allViolations[i].Severity] > severityOrder[allViolations[j].Severity]
	})

	// Show top 10 violations
	maxViolations := 10
	if len(allViolations) < maxViolations {
		maxViolations = len(allViolations)
	}

	for i := 0; i < maxViolations; i++ {
		v := allViolations[i]
		icon := a.getSeverityIcon(v.Severity)

		fmt.Fprintf(output, "%s **%s** - %s\n", icon, v.Severity, v.Message)
		fmt.Fprintf(output, "   📁 %s", a.getRelativePath(v.FilePath))
		if v.FunctionName != "" {
			fmt.Fprintf(output, " → %s()", v.FunctionName)
		}
		fmt.Fprintf(output, " (line %d)\n", v.LineNumber)
		fmt.Fprintf(output, "   📊 %v (threshold: %v)\n\n", v.ActualValue, v.ThresholdValue)
	}

	if len(allViolations) > maxViolations {
		fmt.Fprintf(output, "... and %d more violations\n", len(allViolations)-maxViolations)
	}
}

// writePackageDetails writes detailed package information
func (a *DefaultComplexityAnalyzer) writePackageDetails(output io.Writer, pkg *PackageComplexity) {
	fmt.Fprintf(output, "### Package: %s\n", a.getRelativePath(pkg.PackagePath))
	fmt.Fprintf(output, "- **Files**: %d\n", len(pkg.Files))
	fmt.Fprintf(output, "- **Functions**: %d\n", pkg.TotalFunctions)
	fmt.Fprintf(output, "- **Lines of Code**: %d\n", pkg.TotalLinesOfCode)
	fmt.Fprintf(output, "- **Avg Complexity**: %.2f\n", pkg.AverageCyclomaticComplexity)
	fmt.Fprintf(output, "- **Maintainability**: %.2f\n", pkg.MaintainabilityIndex)
	fmt.Fprintf(output, "- **Technical Debt**: %.2f hours\n", pkg.TechnicalDebtHours)
	fmt.Fprintf(output, "- **Violations**: %d\n\n", len(pkg.Violations))

	// Show files with issues
	for _, file := range pkg.Files {
		if len(file.Violations) > 0 {
			fmt.Fprintf(output, "  ⚠️  **%s** (%d violations)\n", a.getFileName(file.FilePath), len(file.Violations))
		}
	}
	fmt.Fprintf(output, "\n")
}

// writeRecommendations writes actionable recommendations
func (a *DefaultComplexityAnalyzer) writeRecommendations(output io.Writer, complexity *ProjectComplexity) {
	summary := complexity.Summary

	fmt.Fprintf(output, "### Priority Actions\n\n")

	// High complexity functions
	if summary.AverageCyclomaticComplexity > 10 {
		fmt.Fprintf(output, "1. **Reduce Cyclomatic Complexity**\n")
		fmt.Fprintf(output, "   - Break down complex functions into smaller, focused functions\n")
		fmt.Fprintf(output, "   - Use early returns to reduce nesting\n")
		fmt.Fprintf(output, "   - Consider using strategy pattern for complex conditional logic\n\n")
	}

	// Low maintainability
	if summary.MaintainabilityIndex < 70 {
		fmt.Fprintf(output, "2. **Improve Maintainability**\n")
		fmt.Fprintf(output, "   - Add comprehensive comments and documentation\n")
		fmt.Fprintf(output, "   - Refactor large functions and files\n")
		fmt.Fprintf(output, "   - Improve variable and function naming\n\n")
	}

	// High technical debt
	if summary.TechnicalDebtDays > 2 {
		fmt.Fprintf(output, "3. **Address Technical Debt**\n")
		fmt.Fprintf(output, "   - Prioritize refactoring based on violation severity\n")
		fmt.Fprintf(output, "   - Set up automated complexity monitoring\n")
		fmt.Fprintf(output, "   - Establish coding standards and enforce them\n\n")
	}

	// General recommendations
	fmt.Fprintf(output, "### Best Practices\n")
	fmt.Fprintf(output, "- Keep functions under 50 lines\n")
	fmt.Fprintf(output, "- Limit cyclomatic complexity to 10 or less\n")
	fmt.Fprintf(output, "- Maintain files under 500 lines\n")
	fmt.Fprintf(output, "- Use descriptive variable and function names\n")
	fmt.Fprintf(output, "- Write unit tests for complex functions\n")
	fmt.Fprintf(output, "- Regular code reviews focusing on complexity\n")
}

// Helper functions
func (a *DefaultComplexityAnalyzer) getSeverityIcon(severity string) string {
	switch severity {
	case "Critical":
		return "🔴"
	case "Major":
		return "🟠"
	case "Minor":
		return "🟡"
	default:
		return "⚠️"
	}
}

func (a *DefaultComplexityAnalyzer) getRelativePath(fullPath string) string {
	// Simple path shortening for readability
	parts := strings.Split(fullPath, "/")
	if len(parts) > 3 {
		return ".../" + strings.Join(parts[len(parts)-2:], "/")
	}
	return fullPath
}

func (a *DefaultComplexityAnalyzer) getFileName(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullPath
}

// generateHTMLReport generates an HTML complexity report
func (a *DefaultComplexityAnalyzer) generateHTMLReport(complexity *ProjectComplexity, output io.Writer) error {
	summary := complexity.Summary

	fmt.Fprintf(output, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go Sentinel CLI - Complexity Report</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1, h2, h3 { color: #333; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin: 20px 0; }
        .metric { background: #f8f9fa; padding: 20px; border-radius: 6px; border-left: 4px solid #007bff; }
        .metric h3 { margin-top: 0; font-size: 14px; text-transform: uppercase; color: #666; }
        .metric .value { font-size: 24px; font-weight: bold; color: #333; }
        .grade-A { border-left-color: #28a745; }
        .grade-B { border-left-color: #ffc107; }
        .grade-C { border-left-color: #fd7e14; }
        .grade-D, .grade-F { border-left-color: #dc3545; }
        .violations { margin: 20px 0; }
        .violation { background: #fff; border: 1px solid #dee2e6; border-radius: 4px; margin: 10px 0; padding: 15px; }
        .severity-Critical { border-left: 4px solid #dc3545; }
        .severity-Major { border-left: 4px solid #fd7e14; }
        .severity-Minor { border-left: 4px solid #ffc107; }
        .severity-Warning { border-left: 4px solid #6c757d; }
        .packages { margin: 20px 0; }
        .package { margin: 15px 0; padding: 15px; background: #f8f9fa; border-radius: 4px; }
        .recommendations { background: #e7f3ff; padding: 20px; border-radius: 6px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔍 Go Sentinel CLI - Complexity Report</h1>
        <p><strong>Generated:</strong> %s</p>
        
        <div class="summary">
            <div class="metric grade-%s">
                <h3>Quality Grade</h3>
                <div class="value">%s</div>
            </div>
            <div class="metric">
                <h3>Total Functions</h3>
                <div class="value">%d</div>
            </div>
            <div class="metric">
                <h3>Lines of Code</h3>
                <div class="value">%d</div>
            </div>
            <div class="metric">
                <h3>Avg Complexity</h3>
                <div class="value">%.2f</div>
            </div>
            <div class="metric">
                <h3>Maintainability</h3>
                <div class="value">%.2f</div>
            </div>
            <div class="metric">
                <h3>Technical Debt</h3>
                <div class="value">%.2f days</div>
            </div>
        </div>`,
		complexity.GeneratedAt.Format("2006-01-02 15:04:05"),
		summary.QualityGrade, summary.QualityGrade,
		summary.TotalFunctions,
		summary.TotalLinesOfCode,
		summary.AverageCyclomaticComplexity,
		summary.MaintainabilityIndex,
		summary.TechnicalDebtDays)

	// Add violations section
	fmt.Fprintf(output, `
        <h2>🚨 Top Violations</h2>
        <div class="violations">`)

	// Collect and sort violations
	allViolations := make([]ComplexityViolation, 0)
	for _, pkg := range complexity.Packages {
		allViolations = append(allViolations, pkg.Violations...)
	}

	sort.Slice(allViolations, func(i, j int) bool {
		severityOrder := map[string]int{"Critical": 4, "Major": 3, "Minor": 2, "Warning": 1}
		return severityOrder[allViolations[i].Severity] > severityOrder[allViolations[j].Severity]
	})

	maxViolations := 10
	if len(allViolations) < maxViolations {
		maxViolations = len(allViolations)
	}

	for i := 0; i < maxViolations; i++ {
		v := allViolations[i]
		fmt.Fprintf(output, `
            <div class="violation severity-%s">
                <strong>%s</strong> - %s<br>
                <small>📁 %s`, v.Severity, v.Severity, v.Message, a.getRelativePath(v.FilePath))

		if v.FunctionName != "" {
			fmt.Fprintf(output, " → %s()", v.FunctionName)
		}
		fmt.Fprintf(output, " (line %d)<br>", v.LineNumber)
		fmt.Fprintf(output, "📊 %v (threshold: %v)</small>", v.ActualValue, v.ThresholdValue)
		fmt.Fprintf(output, `
            </div>`)
	}

	fmt.Fprintf(output, `
        </div>
        
        <div class="recommendations">
            <h2>💡 Recommendations</h2>
            <ul>
                <li>Focus on reducing functions with cyclomatic complexity > 10</li>
                <li>Break down large functions into smaller, focused units</li>
                <li>Improve maintainability through better documentation</li>
                <li>Set up automated complexity monitoring in CI/CD</li>
            </ul>
        </div>
    </div>
</body>
</html>`)

	return nil
}
