package runner

import (
	"fmt"
	"strings"
)

// FormatMillis converts seconds (float64) to milliseconds string, e.g. 0.0123 -> "12ms"
func FormatMillis(seconds float64) string {
	return fmt.Sprintf("%dms", int(seconds*1000))
}

// FormatCoverage takes a float64 (0.0-1.0 or 0-100) and returns a string with 2 decimal places and a percent sign.
func FormatCoverage(coverage float64) string {
	if coverage > 1.0 {
		return fmt.Sprintf("%.2f%%", coverage)
	}
	return fmt.Sprintf("%.2f%%", coverage*100)
}

// FormatTestOutput formats test events into a string that matches the standard Go test output format.
// It handles test status, package status, coverage information, and proper indentation.
func FormatTestOutput(events []TestEvent) string {
	if len(events) == 0 {
		return ""
	}

	pkgStatus := make(map[string]bool) // true = pass, false = fail
	pkgDuration := make(map[string]float64)
	pkgCoverage := make(map[string]float64)
	pkgOrder := make([]string, 0)
	pkgLines := make(map[string][]string)
	indent := "    "

	// First pass: collect package status and output
	for _, ev := range events {
		// Track package order and initialize package status
		if _, exists := pkgStatus[ev.Package]; !exists {
			pkgOrder = append(pkgOrder, ev.Package)
			pkgStatus[ev.Package] = true // Initialize as passing
		}

		// Add output to the package's output lines
		if ev.Output != "" {
			if strings.Contains(ev.Output, "coverage:") {
				var coverage float64
				if _, err := fmt.Sscanf(ev.Output, "coverage: %f%% of statements", &coverage); err == nil {
					pkgCoverage[ev.Package] = coverage
				}
			}
		}

		switch ev.Action {
		case "run":
			pkgLines[ev.Package] = append(pkgLines[ev.Package], fmt.Sprintf("=== RUN   %s", ev.Test))
		case "pass":
			if ev.Test != "" {
				prefix := "---"
				if strings.Contains(ev.Test, "/") {
					prefix = indent + prefix
				}
				pkgLines[ev.Package] = append(pkgLines[ev.Package], fmt.Sprintf("%s PASS: %s (%.3fs)", prefix, ev.Test, ev.Elapsed))
			}
			pkgDuration[ev.Package] = ev.Elapsed
		case "fail":
			pkgStatus[ev.Package] = false
			if ev.Test != "" {
				prefix := "---"
				if strings.Contains(ev.Test, "/") {
					prefix = indent + prefix
				}
				pkgLines[ev.Package] = append(pkgLines[ev.Package], fmt.Sprintf("%s FAIL: %s (%.3fs)", prefix, ev.Test, ev.Elapsed))
				if ev.Output != "" {
					prefix := indent
					if strings.Contains(ev.Test, "/") {
						prefix = indent + indent
					}
					pkgLines[ev.Package] = append(pkgLines[ev.Package], prefix+ev.Output)
				}
			}
			pkgDuration[ev.Package] = ev.Elapsed
		case "skip":
			if ev.Test != "" {
				pkgLines[ev.Package] = append(pkgLines[ev.Package], fmt.Sprintf("--- SKIP: %s (%.3fs)", ev.Test, ev.Elapsed))
				if ev.Output != "" {
					pkgLines[ev.Package] = append(pkgLines[ev.Package], indent+ev.Output)
				}
			}
		case "output":
			if ev.Output != "" && !strings.Contains(ev.Output, "coverage:") {
				prefix := ""
				if strings.Contains(ev.Test, "/") {
					prefix = indent
				}
				pkgLines[ev.Package] = append(pkgLines[ev.Package], prefix+ev.Output)
			}
		}
	}

	// Second pass: combine output for all packages
	var finalLines []string
	for _, pkg := range pkgOrder {
		// Add test output lines for this package
		finalLines = append(finalLines, pkgLines[pkg]...)

		// Add coverage information if available
		if coverage, ok := pkgCoverage[pkg]; ok {
			finalLines = append(finalLines, fmt.Sprintf("coverage: %.1f%% of statements", coverage))
		}

		// Add package status line
		if pkgStatus[pkg] {
			// Go uses three spaces after 'ok' and a tab after the package name
			line := fmt.Sprintf("ok   %s", pkg)
			if dur, ok := pkgDuration[pkg]; ok {
				line += fmt.Sprintf("\t%.3fs", dur)
			}
			finalLines = append(finalLines, line)
		} else {
			// Go uses tab after FAIL and after package name
			line := fmt.Sprintf("FAIL\t%s", pkg)
			if dur, ok := pkgDuration[pkg]; ok {
				line += fmt.Sprintf("\t%.3fs", dur)
			}
			finalLines = append(finalLines, "FAIL")
			finalLines = append(finalLines, line)
		}
	}

	// Add final FAIL marker if any package failed
	for _, pkg := range pkgOrder {
		if !pkgStatus[pkg] {
			finalLines = append(finalLines, "FAIL")
			break
		}
	}

	return strings.Join(finalLines, "\n")
}
