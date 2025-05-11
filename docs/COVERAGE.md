# Go-Sentinel Code Coverage

## Overview

Go-Sentinel includes a comprehensive code coverage analysis feature that helps you understand how thoroughly your tests exercise your codebase. Similar to Jest's coverage reporting for JavaScript, Go-Sentinel's coverage feature provides detailed metrics and visualizations to identify untested code.

## Coverage Metrics

Go-Sentinel tracks four key coverage metrics:

1. **Statement Coverage**: The percentage of executable statements that were executed during tests.
2. **Branch Coverage**: The percentage of code branches (if/else/switch statements) that were tested.
3. **Function Coverage**: The percentage of functions or methods that were called during tests.
4. **Line Coverage**: The percentage of code lines that were executed during tests.

## Using Coverage in Go-Sentinel

### Keyboard Shortcuts

| Key | Function |
|-----|----------|
| `c` | Toggle coverage view on/off |
| `C` | Run tests with coverage for current package or all packages |
| `L` | Toggle filter to show only low-coverage files |
| Arrow keys | Navigate through files in coverage view |
| Enter | View detailed coverage for selected file |

### Running Tests with Coverage

1. Press `C` from anywhere in Go-Sentinel to run tests with coverage
2. Wait for the tests to complete and the coverage data to be processed
3. The coverage view will automatically appear showing overall metrics

### Understanding the Coverage Display

#### Color Coding

- **Green**: High coverage (≥80%)
- **Yellow**: Medium coverage (50%-79%)
- **Red**: Low coverage (<50%)

#### Execution Counts

Next to each line of code, you may see annotations like `1x`, `2x`, etc. These indicate how many times that line was executed during testing.

#### Uncovered Lines

Uncovered lines are highlighted in red, making it easy to identify code that needs additional test coverage.

## Examples

### Example 1: Running Coverage for All Tests

1. Press `C` from the main view
2. Coverage data will be generated and displayed
3. Press `c` to toggle between normal and coverage views

### Example 2: Focusing on Low Coverage Files

1. Run tests with coverage by pressing `C`
2. In the coverage view, press `L` to show only files with low coverage
3. Navigate to a file and press Enter to see detailed line-by-line coverage

## Technical Implementation

Go-Sentinel's coverage feature leverages Go's built-in coverage tools and enhances them with:

1. **Analysis**: Parses Go coverage profiles using the `golang.org/x/tools/cover` package
2. **Visualization**: Renders coverage data with color-coding in the TUI
3. **Integration**: Seamlessly connects with Go-Sentinel's test running capabilities

## Best Practices

1. **Regular Coverage Checks**: Run coverage analysis regularly to track testing progress
2. **Focus on Critical Areas**: Prioritize coverage for error handling and edge cases
3. **Set Coverage Goals**: Aim for specific coverage percentages for different parts of your codebase
4. **Don't Chase 100%**: Instead of targeting perfect coverage, focus on meaningful tests

## Troubleshooting

- If coverage data doesn't appear, ensure tests are passing or the coverage file was generated
- For large codebases, coverage generation might take some time
- If colors aren't displaying correctly, check your terminal's color support
